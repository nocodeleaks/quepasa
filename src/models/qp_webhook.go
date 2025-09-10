package models

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/nocodeleaks/quepasa/environment"
	"github.com/nocodeleaks/quepasa/library"
	metrics "github.com/nocodeleaks/quepasa/metrics"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
)

type QpWebhook struct {
	library.LogStruct // logging

	// Optional whatsapp options
	// ------------------------
	whatsapp.WhatsappOptions

	Url             string      `db:"url" json:"url,omitempty"`                         // destination
	ForwardInternal bool        `db:"forwardinternal" json:"forwardinternal,omitempty"` // forward internal msg from api
	TrackId         string      `db:"trackid" json:"trackid,omitempty"`                 // identifier of remote system to avoid loop
	Extra           interface{} `db:"extra" json:"extra,omitempty"`                     // extra info to append on payload
	Failure         *time.Time  `json:"failure,omitempty"`                              // first failure timestamp
	Success         *time.Time  `json:"success,omitempty"`                              // last success timestamp
	Timestamp       *time.Time  `db:"timestamp" json:"timestamp,omitempty"`

	// just for logging and response headers
	Wid string `json:"-"`
}

// custom log entry with fields: wid & url
func (source *QpWebhook) GetLogger() *log.Entry {
	if source != nil && source.LogEntry != nil {
		return source.LogEntry
	}

	logentry := log.WithContext(context.Background())

	if source != nil {
		logentry = logentry.WithField(LogFields.WId, source.Wid)
		logentry = logentry.WithField(LogFields.Url, source.Url)
		source.LogEntry = logentry
	}

	logentry.Level = log.ErrorLevel
	logentry.Infof("generating new log entry for %s, with level: %s", reflect.TypeOf(source), logentry.Level)

	return logentry
}

//#region VIEWS TRICKS

func (source QpWebhook) GetReadReceipts() bool {
	return source.ReadReceipts.Boolean()
}

func (source QpWebhook) IsSetReadReceipts() bool {
	return source.ReadReceipts != whatsapp.UnSetBooleanType
}

func (source QpWebhook) GetGroups() bool {
	return source.Groups.Boolean()
}

func (source QpWebhook) IsSetGroups() bool {
	return source.Groups != whatsapp.UnSetBooleanType
}

func (source QpWebhook) GetBroadcasts() bool {
	return source.Broadcasts.Boolean()
}

func (source QpWebhook) IsSetBroadcasts() bool {
	return source.Broadcasts != whatsapp.UnSetBooleanType
}

func (source QpWebhook) GetCalls() bool {
	return source.Calls.Boolean()
}

func (source QpWebhook) IsSetCalls() bool {
	return source.Calls != whatsapp.UnSetBooleanType
}

func (source QpWebhook) IsSetExtra() bool {
	return source.Extra != nil
}

//#endregion

var ErrInvalidResponse error = errors.New("the requested url do not return 200 status code")

// WebhookQueueMessage represents a webhook message in the queue
type WebhookQueueMessage struct {
	ID          string                    `json:"id"`
	Webhook     *QpWebhook                `json:"webhook"`
	Message     *whatsapp.WhatsappMessage `json:"message"`
	Timestamp   time.Time                 `json:"timestamp"`
	RetryCount  int                       `json:"retry_count"`
	LastAttempt time.Time                 `json:"last_attempt"`
	Status      string                    `json:"status"` // "queued", "processing", "completed", "failed", "discarded"
}

// WebhookQueueClient manages asynchronous webhook processing with caching
type WebhookQueueClient struct {
	messageCache    chan WebhookQueueMessage // In-memory queue for webhook messages
	cacheProcessing sync.Once                // Ensures cache processor starts only once
	maxCacheSize    int                      // Maximum number of messages to hold in cache
	closed          chan struct{}            // Signals that the client should stop
	wg              sync.WaitGroup           // To wait for goroutines to finish
	processingDelay time.Duration            // Delay between processing messages
}

// Global webhook queue client instance
var WebhookQueueClientInstance *WebhookQueueClient
var webhookQueueOnce sync.Once

// InitializeWebhookQueue initializes the global webhook queue client
func InitializeWebhookQueue() {
	webhookQueueOnce.Do(func() {
		// Only initialize if queue is enabled
		if !environment.Settings.API.WebhookQueueEnabled {
			log.Info("Webhook queue system is disabled")
			return
		}

		size := environment.Settings.API.GetWebhookQueueSize()
		delay := time.Duration(environment.Settings.API.GetWebhookQueueDelay()) * time.Second
		workers := environment.Settings.API.GetWebhookWorkers()

		// Validate configuration
		if size <= 0 {
			log.Warnf("Invalid webhook queue size %d, using default 100", size)
			size = 100
		}
		if workers <= 0 {
			log.Warnf("Invalid webhook workers count %d, using default 1", workers)
			workers = 1
		}

		WebhookQueueClientInstance = &WebhookQueueClient{
			messageCache:    make(chan WebhookQueueMessage, size),
			maxCacheSize:    size,
			closed:          make(chan struct{}),
			processingDelay: delay,
		}

		// Start multiple cache processors (workers)
		WebhookQueueClientInstance.cacheProcessing.Do(func() {
			for i := 0; i < workers; i++ {
				WebhookQueueClientInstance.wg.Add(1)
				go WebhookQueueClientInstance.processCache(i)
			}
		})

		log.Infof("Webhook queue client initialized with size %d, delay %v, and %d workers", size, delay, workers)
	})
}

// AddToCache adds a webhook message to the cache
// Returns true if added to cache, false if cache is full
func (w *WebhookQueueClient) AddToCache(webhook *QpWebhook, message *whatsapp.WhatsappMessage) bool {
	msg := WebhookQueueMessage{
		ID:          fmt.Sprintf("webhook-%d", time.Now().UnixNano()),
		Webhook:     webhook,
		Message:     message,
		Timestamp:   time.Now(),
		RetryCount:  0,
		LastAttempt: time.Time{},
		Status:      "queued",
	}

	select {
	case w.messageCache <- msg:
		// Log queue status
		queueSize := len(w.messageCache)
		log.WithFields(log.Fields{
			"webhook_id": msg.ID,
			"url":        webhook.Url,
			"queue_size": queueSize,
			"max_size":   w.maxCacheSize,
			"status":     "enqueued",
		}).Infof("Webhook enqueued for processing (Queue: %d/%d)", queueSize, w.maxCacheSize)

		metrics.WebhookQueueSize.Set(float64(queueSize))
		return true
	default:
		// Cache is full, discard message
		log.WithFields(log.Fields{
			"webhook_id": msg.ID,
			"url":        webhook.Url,
			"queue_size": len(w.messageCache),
			"max_size":   w.maxCacheSize,
			"status":     "discarded",
		}).Warnf("Webhook queue full, discarding message (Queue: %d/%d)", len(w.messageCache), w.maxCacheSize)

		metrics.WebhookQueueDiscarded.Inc()
		return false
	}
}

// processCache processes messages from the cache
func (w *WebhookQueueClient) processCache(workerId int) {
	defer w.wg.Done()
	log.Infof("Webhook queue processor (worker %d) started", workerId)

	for {
		select {
		case <-w.closed:
			log.Infof("Webhook queue processor (worker %d) shutting down", workerId)
			return

		case msg := <-w.messageCache:
			w.processMessage(msg)

			// Apply processing delay if configured
			if w.processingDelay > 0 {
				time.Sleep(w.processingDelay)
			}
		}
	}
}

// processMessage processes a single webhook message
func (w *WebhookQueueClient) processMessage(msg WebhookQueueMessage) {
	// Update message status to processing
	msg.Status = "processing"
	msg.LastAttempt = time.Now()

	log.WithFields(log.Fields{
		"webhook_id":  msg.ID,
		"url":         msg.Webhook.Url,
		"status":      msg.Status,
		"retry_count": msg.RetryCount,
	}).Info("Processing webhook from queue")

	// Process the webhook
	err := msg.Webhook.postWebhook(msg.Message)

	// Update metrics and status based on result
	if err != nil {
		// postWebhook already exhausted all retry attempts
		// No need to retry at queue level - mark as final failure
		msg.Status = "failed_final"

		log.WithFields(log.Fields{
			"webhook_id":  msg.ID,
			"url":         msg.Webhook.Url,
			"status":      msg.Status,
			"retry_count": msg.RetryCount,
		}).Error("Webhook failed after all retry attempts - marking as final failure")

		metrics.WebhookQueueFailed.Inc()
	} else {
		msg.Status = "completed"
		log.WithFields(log.Fields{
			"webhook_id": msg.ID,
			"url":        msg.Webhook.Url,
			"status":     msg.Status,
		}).Info("Webhook processed successfully")
		metrics.WebhookQueueCompleted.Inc()
	}

	// Update processed counter
	metrics.WebhookQueueProcessed.Inc()

	// Update queue size metric
	metrics.WebhookQueueSize.Set(float64(len(w.messageCache)))
}

// shouldRetryMessage determines if a failed message should be retried
func (w *WebhookQueueClient) shouldRetryMessage(msg WebhookQueueMessage) bool {
	maxRetries := environment.Settings.API.GetWebhookRetryCount()
	return msg.RetryCount < maxRetries
}

// GetQueueStatus returns current queue status information
func (w *WebhookQueueClient) GetQueueStatus() map[string]interface{} {
	workers := environment.Settings.API.GetWebhookWorkers()
	return map[string]interface{}{
		"current_size":     len(w.messageCache),
		"max_size":         w.maxCacheSize,
		"utilization":      float64(len(w.messageCache)) / float64(w.maxCacheSize) * 100,
		"is_enabled":       environment.Settings.API.WebhookQueueEnabled,
		"processing_delay": w.processingDelay.String(),
		"workers":          workers,
	}
}

// Close shuts down the webhook queue client gracefully
func (w *WebhookQueueClient) Close() {
	log.Info("Closing webhook queue client...")
	close(w.closed)
	
	// Wait for workers to finish with a timeout
	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		log.Info("All webhook queue workers finished gracefully")
	case <-time.After(30 * time.Second):
		log.Warn("Timeout waiting for webhook queue workers to finish")
	}
	
	log.Info("Webhook queue client closed")
}

// shouldRetry determines if an error should trigger a retry attempt
func shouldRetry(err error, statusCode int) bool {
	if err == nil && statusCode == 200 {
		return false // Success, no retry needed
	}

	if err != nil {
		errStr := strings.ToLower(err.Error())

		// Retry on timeout errors
		if strings.Contains(errStr, "timeout") ||
			strings.Contains(errStr, "deadline exceeded") ||
			strings.Contains(errStr, "context deadline exceeded") {
			return true
		}

		// Retry on network errors
		if strings.Contains(errStr, "connection refused") ||
			strings.Contains(errStr, "connection reset") ||
			strings.Contains(errStr, "no such host") ||
			strings.Contains(errStr, "network is unreachable") {
			return true
		}

		// Check for net.Error timeout
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return true
		}

		// Don't retry on other client errors (malformed URL, etc.)
		return false
	}

	// Don't retry on client errors (4xx) - these are permanent failures
	if statusCode >= 400 && statusCode < 500 {
		return false
	}

	// Retry on server errors (5xx)
	if statusCode >= 500 && statusCode < 600 {
		return true
	}

	// Retry on other non-200 responses (3xx, etc.)
	return statusCode != 200
}

func (source *QpWebhook) Post(message *whatsapp.WhatsappMessage) (err error) {
	// Check if queue system is enabled
	if WebhookQueueClientInstance != nil && environment.Settings.API.WebhookQueueEnabled {
		// Enqueue the webhook for asynchronous processing
		if WebhookQueueClientInstance.AddToCache(source, message) {
			return nil // Success, enqueued
		} else {
			// Queue full, fallback to direct processing
			logentry := source.LogWithField(LogFields.MessageId, message.Id)
			logentry.Warn("Webhook queue full, processing directly")
		}
	}

	// Direct processing (original behavior or fallback)
	return source.postWebhook(message)
}

func (source *QpWebhook) postWebhook(message *whatsapp.WhatsappMessage) (err error) {
	startTime := time.Now()

	// updating log
	logentry := source.LogWithField(LogFields.MessageId, message.Id)
	logentry.Infof("posting webhook")

	payload := &QpWebhookPayload{
		WhatsappMessage: message,
		Extra:           source.Extra,
	}

	payloadJson, err := json.Marshal(&payload)
	if err != nil {
		return
	}

	// logging webhook payload
	logentry.Debugf("posting webhook payload: %s", payloadJson)

	// Check if retry system is enabled
	if !environment.Settings.API.IsWebhookRetryEnabled() {
		// Use original single-attempt logic
		return source.postSingleAttempt(payloadJson, logentry, startTime)
	}

	// Use retry logic
	retryCount := environment.Settings.API.GetWebhookRetryCount()
	retryDelay := time.Duration(environment.Settings.API.GetWebhookRetryDelay()) * time.Second
	timeout := time.Duration(environment.Settings.API.GetWebhookTimeout()) * time.Second

	// Retry logic
	attemptsMade := 0
	for attempt := 0; attempt <= retryCount; attempt++ {
		attemptsMade++

		if attempt > 0 {
			metrics.WebhookRetryAttempts.Inc()
			logentry.Infof("webhook retry attempt %d/%d after %v delay", attempt, retryCount, retryDelay)
			time.Sleep(retryDelay)
		}

		req, reqErr := http.NewRequest("POST", source.Url, bytes.NewBuffer(payloadJson))
		if reqErr != nil {
			err = reqErr
			// Don't retry on request creation errors (bad URL, etc.)
			logentry.Errorf("failed to create webhook request: %s", reqErr.Error())
			break
		}

		req.Header.Set("User-Agent", "Quepasa")
		req.Header.Set("X-QUEPASA-WID", source.Wid)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		client.Timeout = timeout
		resp, clientErr := client.Do(req)

		metrics.WebhooksSent.Inc()

		var statusCode int
		if resp != nil {
			statusCode = resp.StatusCode
			defer resp.Body.Close()
		}

		// Determine if we should retry
		if clientErr == nil && statusCode == 200 {
			// Success! Break out of retry loop
			err = nil
			if attempt > 0 {
				metrics.WebhookRetriesSuccessful.Inc()
			}
			logentry.Debugf("webhook success on attempt %d", attempt+1)
			break
		}

		// Determine the error to log and check
		var currentErr error
		if clientErr != nil {
			currentErr = clientErr
			err = clientErr
		} else if statusCode != 200 {
			currentErr = ErrInvalidResponse
			err = ErrInvalidResponse
		} else {
			currentErr = errors.New("no response received")
			err = currentErr
		}

		// Log the specific error
		if clientErr != nil {
			logentry.Warnf("webhook request error (attempt %d/%d): %s", attempt+1, retryCount+1, clientErr.Error())
		} else {
			logentry.Warnf("webhook returned status %d (attempt %d/%d)", statusCode, attempt+1, retryCount+1)
		}

		// Check if we should retry this error
		if !shouldRetry(currentErr, statusCode) {
			if statusCode >= 400 && statusCode < 500 {
				logentry.Warnf("client error (4xx) detected - not retryable (status: %d)", statusCode)
			} else {
				logentry.Infof("error is not retryable, stopping attempts")
			}
			break
		}

		// If this is the last attempt, don't continue
		if attempt == retryCount {
			logentry.Warnf("max retry attempts reached")
			break
		}
	}

	// Record metrics
	duration := time.Since(startTime)
	metrics.WebhookLatency.Observe(duration.Seconds())

	currentTime := time.Now().UTC()
	if err != nil {
		metrics.WebhookSendErrors.Inc()
		if attemptsMade > 1 {
			metrics.WebhookRetryFailures.Inc()
		}
		if source.Failure == nil {
			source.Failure = &currentTime
		}
		logentry.Errorf("webhook failed after %d attempts: %s", retryCount+1, err.Error())
	} else {
		source.Failure = nil
		source.Success = &currentTime
		logentry.Infof("webhook posted successfully")
	}

	return
}

// postSingleAttempt handles the original single-attempt webhook logic (no retry)
func (source *QpWebhook) postSingleAttempt(payloadJson []byte, logentry *log.Entry, startTime time.Time) (err error) {

	timeout := time.Duration(environment.Settings.API.GetWebhookTimeout()) * time.Second

	req, err := http.NewRequest("POST", source.Url, bytes.NewBuffer(payloadJson))
	if err != nil {
		return
	}

	req.Header.Set("User-Agent", "Quepasa")
	req.Header.Set("X-QUEPASA-WID", source.Wid)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	client.Timeout = timeout
	resp, err := client.Do(req)

	metrics.WebhooksSent.Inc()

	if err != nil {
		logentry.Warnf("error at post webhook: %s", err.Error())
	}

	if resp != nil {
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			err = ErrInvalidResponse
		}
	}

	// Record metrics
	duration := time.Since(startTime)
	metrics.WebhookLatency.Observe(duration.Seconds())

	currentTime := time.Now().UTC()
	if err != nil {
		metrics.WebhookSendErrors.Inc()
		if source.Failure == nil {
			source.Failure = &currentTime
		}
	} else {
		source.Failure = nil
		source.Success = &currentTime
		logentry.Infof("webhook posted successfully")
	}

	return
}

// Initialize webhook queue if enabled
func init() {
	// Only initialize if queue is enabled
	if environment.Settings.API.WebhookQueueEnabled {
		InitializeWebhookQueue()
	}
}

var WebhookQueueDiscarded = metrics.WebhookQueueDiscarded

// GetWebhookQueueStatus returns current webhook queue status
func GetWebhookQueueStatus() map[string]interface{} {
	if WebhookQueueClientInstance == nil {
		return map[string]interface{}{
			"enabled": false,
			"status":  "not_initialized",
			"message": "Webhook queue system is not enabled or initialized",
		}
	}

	status := WebhookQueueClientInstance.GetQueueStatus()
	status["enabled"] = environment.Settings.API.WebhookQueueEnabled
	status["status"] = "active"
	status["message"] = "Webhook queue system is active and processing messages"

	// Add health check information
	queueSize := status["current_size"].(int)
	maxSize := status["max_size"].(int)
	utilization := status["utilization"].(float64)

	if utilization > 90.0 {
		status["health"] = "critical"
		status["message"] = "Queue utilization is critically high"
	} else if utilization > 75.0 {
		status["health"] = "warning"
		status["message"] = "Queue utilization is high"
	} else {
		status["health"] = "healthy"
	}

	// Add queue statistics
	status["statistics"] = map[string]interface{}{
		"queue_size":          queueSize,
		"max_size":            maxSize,
		"available_slots":     maxSize - queueSize,
		"utilization_percent": utilization,
		"processing_delay":    status["processing_delay"],
	}

	return status
}

// CleanupWebhookQueue shuts down the global webhook queue client
func CleanupWebhookQueue() {
	if WebhookQueueClientInstance != nil {
		WebhookQueueClientInstance.Close()
		WebhookQueueClientInstance = nil
		log.Info("Webhook queue client cleaned up")
	}
}

// RestartWebhookQueue shuts down and reinitializes the webhook queue
// Useful for configuration changes without full application restart
func RestartWebhookQueue() {
	log.Info("Restarting webhook queue system...")
	CleanupWebhookQueue()

	// Reset the sync.Once to allow reinitialization
	webhookQueueOnce = sync.Once{}

	// Only restart if queue is enabled
	if environment.Settings.API.WebhookQueueEnabled {
		InitializeWebhookQueue()
		log.Info("Webhook queue system restarted successfully")
	} else {
		log.Info("Webhook queue system remains disabled after restart")
	}
}
