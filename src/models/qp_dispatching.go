package models

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/nocodeleaks/quepasa/environment"
	"github.com/nocodeleaks/quepasa/library"
	metrics "github.com/nocodeleaks/quepasa/metrics"
	rabbitmq "github.com/nocodeleaks/quepasa/rabbitmq"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
)

// Dispatching types
const (
	DispatchingTypeWebhook  = "webhook"
	DispatchingTypeRabbitMQ = "rabbitmq"
)

type QpDispatching struct {
	library.LogStruct // logging

	// Optional whatsapp options
	// ------------------------
	whatsapp.WhatsappOptions

	ConnectionString string      `db:"connection_string" json:"connection_string,omitempty"` // destination URL (webhook) or connection string (rabbitmq)
	Type             string      `db:"type" json:"type,omitempty"`                           // webhook or rabbitmq
	ForwardInternal  bool        `db:"forwardinternal" json:"forwardinternal,omitempty"`     // forward internal msg from api
	TrackId          string      `db:"trackid" json:"trackid,omitempty"`                     // identifier of remote system to avoid loop
	Extra            interface{} `db:"extra" json:"extra,omitempty"`                         // extra info to append on payload
	Failure          *time.Time  `json:"failure,omitempty"`                                  // first failure timestamp
	Success          *time.Time  `json:"success,omitempty"`                                  // last success timestamp
	Timestamp        *time.Time  `db:"timestamp" json:"timestamp,omitempty"`

	// just for logging and response headers
	Wid string `json:"-"`
}

// custom log entry with fields: wid & connection_string
func (source *QpDispatching) GetLogger() *log.Entry {
	if source != nil && source.LogEntry != nil {
		return source.LogEntry
	}

	logentry := log.WithContext(context.Background())

	if source != nil {
		logentry = logentry.WithField(LogFields.WId, source.Wid)
		logentry = logentry.WithField(LogFields.Url, source.ConnectionString)
		source.LogEntry = logentry
	}

	logentry.Level = log.ErrorLevel
	logentry.Infof("generating new log entry for %s, with level: %s", reflect.TypeOf(source), logentry.Level)

	return logentry
}

//#region VIEWS TRICKS

func (source QpDispatching) GetReadReceipts() bool {
	return source.ReadReceipts.Boolean()
}

func (source QpDispatching) IsSetReadReceipts() bool {
	return source.ReadReceipts != whatsapp.UnSetBooleanType
}

func (source QpDispatching) GetGroups() bool {
	return source.Groups.Boolean()
}

func (source QpDispatching) IsSetGroups() bool {
	return source.Groups != whatsapp.UnSetBooleanType
}

func (source QpDispatching) GetBroadcasts() bool {
	return source.Broadcasts.Boolean()
}

func (source QpDispatching) IsSetBroadcasts() bool {
	return source.Broadcasts != whatsapp.UnSetBooleanType
}

func (source QpDispatching) GetCalls() bool {
	return source.Calls.Boolean()
}

func (source QpDispatching) IsSetCalls() bool {
	return source.Calls != whatsapp.UnSetBooleanType
}

func (source QpDispatching) IsSetExtra() bool {
	return source.Extra != nil
}

// GetExtraText converts extra field to JSON string
func (source *QpDispatching) GetExtraText() string {
	if source.Extra == nil {
		return ""
	}

	extraJson, err := json.Marshal(&source.Extra)
	if err != nil {
		return fmt.Sprintf("%v", source.Extra)
	}
	return string(extraJson)
}

// ParseExtra tries to parse extra field from JSON string
func (source *QpDispatching) ParseExtra() {
	if source.Extra == nil {
		return
	}

	extraText := fmt.Sprintf("%v", source.Extra)
	if extraText == "" || extraText == "<nil>" {
		source.Extra = nil
		return
	}

	var extraJson interface{}
	err := json.Unmarshal([]byte(extraText), &extraJson)
	if err != nil {
		// If it's not valid JSON, keep as string
		source.Extra = extraText
	} else {
		// If it's valid JSON, keep as parsed object
		source.Extra = extraJson
	}
}

func (source QpDispatching) IsWebhook() bool {
	return source.Type == DispatchingTypeWebhook
}

func (source QpDispatching) IsRabbitMQ() bool {
	return source.Type == DispatchingTypeRabbitMQ
}

//#endregion

// Dispatch sends the message based on the type
func (source *QpDispatching) Dispatch(message *whatsapp.WhatsappMessage) (err error) {
	switch source.Type {
	case DispatchingTypeWebhook:
		return source.PostWebhook(message)
	case DispatchingTypeRabbitMQ:
		return source.PublishRabbitMQ(message)
	default:
		return errors.New("unsupported dispatching type: " + source.Type)
	}
}

// PostWebhook sends message via HTTP webhook
func (source *QpDispatching) PostWebhook(message *whatsapp.WhatsappMessage) (err error) {
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

	req, err := http.NewRequest("POST", source.ConnectionString, bytes.NewBuffer(payloadJson))
	if err != nil {
		return
	}

	req.Header.Set("User-Agent", "Quepasa")
	req.Header.Set("X-QUEPASA-WID", source.Wid)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	timeout := time.Duration(environment.Settings.API.GetWebhookTimeout()) * time.Second
	client.Timeout = timeout
	resp, err := client.Do(req)

	// Always increment webhooks sent counter
	metrics.WebhooksSent.Inc()

	// Record latency
	duration := time.Since(startTime)
	metrics.WebhookLatency.Observe(duration.Seconds())

	var statusCode int
	if resp != nil {
		statusCode = resp.StatusCode
		defer resp.Body.Close()
	}

	if err != nil {
		logentry.Warnf("error at post webhook: %s", err.Error())

		// Check if it's a timeout error
		if netErr, ok := err.(interface{ Timeout() bool }); ok && netErr.Timeout() {
			metrics.WebhookTimeouts.Inc()
			logentry.Warnf("webhook timeout after %v", timeout)
		}
	}

	if resp != nil && statusCode != 200 {
		err = ErrInvalidResponse
		// Record HTTP error with status code
		metrics.WebhookHTTPErrors.WithLabelValues(fmt.Sprintf("%d", statusCode)).Inc()
	}

	currentTime := time.Now().UTC()
	if err != nil {
		metrics.WebhookSendErrors.Inc()
		if source.Failure == nil {
			source.Failure = &currentTime
		}
		logentry.Errorf("webhook failed with status %d: %s", statusCode, err.Error())
		// Mark dispatch error on message
		if message != nil {
			message.MarkDispatchError()
		}
	} else {
		// Webhook successful
		metrics.WebhookSuccess.Inc()
		source.Failure = nil
		source.Success = &currentTime
		logentry.Infof("webhook posted successfully (status: %d, duration: %v)", statusCode, duration)
		// Clear dispatch error on message
		if message != nil {
			message.ClearDispatchError()
		}
	}

	return
}

// PublishRabbitMQ sends message via RabbitMQ using QuePasa fixed Exchange and routing key with intelligent routing
func (source *QpDispatching) PublishRabbitMQ(message *whatsapp.WhatsappMessage) (err error) {
	startTime := time.Now()

	// updating log
	logentry := source.LogWithField(LogFields.MessageId, message.Id)

	// Determine the routing key based on message type
	routingKey := source.DetermineRoutingKey(message)

	logentry.Infof("publishing to QuePasa Exchange: %s with routing key: %s using connection: %s", rabbitmq.QuePasaExchangeName, routingKey, source.ConnectionString)

	payload := &QpRabbitMQPayload{
		WhatsappMessage: message,
		Extra:           source.Extra,
	}

	// Calculate payload size for metrics
	payloadJson, marshalErr := json.Marshal(&payload)
	var payloadSizeBytes float64
	if marshalErr == nil {
		payloadSizeBytes = float64(len(payloadJson))
	}

	// Get or create RabbitMQ client for this specific connection string
	client := rabbitmq.GetRabbitMQClient(source.ConnectionString)
	if client == nil {
		err = errors.New("failed to get rabbitmq client for connection: " + source.ConnectionString)
		logentry.Errorf("rabbitmq client not available for connection %s: %s", source.ConnectionString, err.Error())

		// Record RabbitMQ publish error
		metrics.RecordRabbitMQPublishError(routingKey, rabbitmq.QuePasaExchangeName, "client_unavailable", message.Type.String())

		currentTime := time.Now().UTC()
		if source.Failure == nil {
			source.Failure = &currentTime
		}
		// Mark dispatch error on message
		if message != nil {
			message.MarkDispatchError()
		}
		return err
	}

	// Try to ensure QuePasa Exchange and Queues exist
	// If channel not ready, message will be cached automatically
	err = client.EnsureExchangeAndQueuesWithRetry()
	if err != nil {
		logentry.Warnf("QuePasa setup not ready yet, message will be cached: %s", err.Error())
		// Don't return error - let the publish method handle caching
		// Record as warning but not as error since message will be cached
		metrics.RecordRabbitMQPublishError(routingKey, rabbitmq.QuePasaExchangeName, "setup_not_ready", message.Type.String())
	}

	// Publish to QuePasa Exchange with routing key
	// This will cache the message if connection is not ready
	client.PublishQuePasaMessage(routingKey, payload)

	// Check if connection is still ready after publish attempt
	// If not ready, it means message was cached and we should mark as failure
	if !client.IsConnectionReady() {
		logentry.Warnf("rabbitmq connection lost during publish for %s, message was cached", source.ConnectionString)

		// Record RabbitMQ publish error - connection lost/cached
		metrics.RecordRabbitMQPublishError(routingKey, rabbitmq.QuePasaExchangeName, "connection_lost_cached", message.Type.String())

		// Mark as failure even though message was cached
		currentTime := time.Now().UTC()
		source.Failure = &currentTime
		source.Success = nil

		// Mark dispatch error on message
		if message != nil {
			message.MarkDispatchError()
		}

		// Still record metrics since message was processed
		metrics.RecordRabbitMQMessagePublished(routingKey, rabbitmq.QuePasaExchangeName, routingKey, message.Type.String())
		duration := time.Since(startTime)
		metrics.ObserveRabbitMQPublishDuration(routingKey, rabbitmq.QuePasaExchangeName, message.Type.String(), duration.Seconds())
		if payloadSizeBytes > 0 {
			messageType := message.Type.String()
			metrics.ObserveRabbitMQMessageSize(routingKey, messageType, payloadSizeBytes)
		}

		logentry.Infof("message cached for QuePasa exchange: %s with routing key: %s (duration: %v, size: %.0f bytes) - marked as failure due to connection issue", rabbitmq.QuePasaExchangeName, routingKey, duration, payloadSizeBytes)
		return errors.New("rabbitmq connection not available, message cached")
	}

	// Always increment RabbitMQ messages published counter
	metrics.RecordRabbitMQMessagePublished(routingKey, rabbitmq.QuePasaExchangeName, routingKey, message.Type.String())

	// Record publish duration
	duration := time.Since(startTime)
	metrics.ObserveRabbitMQPublishDuration(routingKey, rabbitmq.QuePasaExchangeName, message.Type.String(), duration.Seconds())

	// Record message size if we have it
	if payloadSizeBytes > 0 {
		messageType := message.Type.String()
		metrics.ObserveRabbitMQMessageSize(routingKey, messageType, payloadSizeBytes)
	}

	// Mark as success only if connection is ready and message was truly published
	currentTime := time.Now().UTC()
	source.Failure = nil
	source.Success = &currentTime

	// Clear dispatch error on message
	if message != nil {
		message.ClearDispatchError()
	}

	logentry.Infof("message published to QuePasa exchange: %s with routing key: %s (duration: %v, size: %.0f bytes)", rabbitmq.QuePasaExchangeName, routingKey, duration, payloadSizeBytes)

	return nil
}

// DetermineRoutingKey determines the appropriate routing key based on message type and content
// Returns one of the QuePasa standard routing keys using a rule-based approach for better performance
func (source *QpDispatching) DetermineRoutingKey(message *whatsapp.WhatsappMessage) string {
	// Priority 1: History messages always go to history queue
	if message.FromHistory {
		return rabbitmq.QuePasaRoutingKeyHistory
	}

	// Priority 2: Event messages - using lookup map for better performance
	if source.isEventMessage(message) {
		return rabbitmq.QuePasaRoutingKeyEvents
	}

	// Priority 3: Default to production queue for normal messages
	return rabbitmq.QuePasaRoutingKeyProd
}

// isEventMessage determines if a message should be routed to the events queue
// Uses efficient type checking and specific business rules
func (source *QpDispatching) isEventMessage(message *whatsapp.WhatsappMessage) bool {
	// Define event message types in a map for O(1) lookup
	eventTypes := map[whatsapp.WhatsappMessageType]bool{
		whatsapp.UnhandledMessageType: true,
	}

	// Check if message type is in event types
	if eventTypes[message.Type] {
		return true
	}

	// Special case: Contact message with specific conditions
	if message.Type == whatsapp.ContactMessageType {
		return message.Edited && message.Attachment != nil
	}

	return false
}
