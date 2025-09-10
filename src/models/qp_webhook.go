package models

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"reflect"
	"strings"
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
	
	// Retry on server errors (5xx) but not client errors (4xx)
	if statusCode >= 500 && statusCode < 600 {
		return true
	}
	
	// Don't retry on client errors (4xx) - these are permanent failures
	if statusCode >= 400 && statusCode < 500 {
		return false
	}
	
	// Retry on other non-200 responses (3xx, etc.)
	return statusCode != 200
}

func (source *QpWebhook) Post(message *whatsapp.WhatsappMessage) (err error) {
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
			logentry.Infof("error is not retryable, stopping attempts")
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
