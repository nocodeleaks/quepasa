package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	dispatchservice "github.com/nocodeleaks/quepasa/dispatch/service"
	environment "github.com/nocodeleaks/quepasa/environment"
	events "github.com/nocodeleaks/quepasa/events"
	library "github.com/nocodeleaks/quepasa/library"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
)

// Dispatching types
const (
	DispatchingTypeWebhook              = "webhook"
	DispatchingTypeRabbitMQ             = "rabbitmq"
	webhookSequentialFailureBlockWindow = 48 * time.Hour
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
	Failure          *time.Time  `db:"failure" json:"failure,omitempty"`                     // first failure timestamp in the current failure streak
	Success          *time.Time  `db:"success" json:"success,omitempty"`                     // last success timestamp
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

func (source QpDispatching) GetDirect() bool {
	return source.Direct.Boolean()
}

func (source QpDispatching) IsSetDirect() bool {
	return source.Direct != whatsapp.UnSetBooleanType
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

func (source *QpDispatching) GetDispatchType() string {
	if source == nil {
		return ""
	}
	return source.Type
}

func (source *QpDispatching) GetWid() string {
	if source == nil {
		return ""
	}
	return source.Wid
}

func (source *QpDispatching) SetWid(wid string) {
	if source == nil {
		return
	}
	source.Wid = wid
}

func (source *QpDispatching) IsFromInternalForwardEnabled() bool {
	if source == nil {
		return false
	}
	return source.ForwardInternal
}

func (source *QpDispatching) GetTrackId() string {
	if source == nil {
		return ""
	}
	return source.TrackId
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

// IsFailureMoreRecent reports whether the current failure streak is newer than
// the last success. This is the predicate used to detect sequential failures.
func (source *QpDispatching) IsFailureMoreRecent() bool {
	if source == nil || source.Failure == nil {
		return false
	}

	if source.Success == nil {
		return true
	}

	return source.Failure.After(*source.Success)
}

// IsWebhookBlockedAt blocks webhook delivery when the same endpoint has been
// failing continuously for a long time without any newer success.
func (source *QpDispatching) IsWebhookBlockedAt(now time.Time) bool {
	if source == nil || !source.IsWebhook() || !source.IsFailureMoreRecent() {
		return false
	}

	if now.Before(*source.Failure) {
		return false
	}

	return now.Sub(*source.Failure) >= webhookSequentialFailureBlockWindow
}

// PostWebhook sends message via HTTP webhook
func (source *QpDispatching) PostWebhook(message *whatsapp.WhatsappMessage) (err error) {
	// updating log
	logentry := source.LogWithField(LogFields.MessageId, message.Id)
	currentTime := time.Now().UTC()
	if source.IsWebhookBlockedAt(currentTime) {
		source.publishDispatchingEvent("dispatch.webhook.blocked", "blocked", 0, map[string]string{
			"dispatch_type": source.Type,
			"reason":        "sequential_failures",
		})
		logentry.Warnf("webhook dispatch blocked after sequential failures since %s", source.Failure.Format(time.RFC3339))
		if message != nil {
			message.MarkExceptionsWithMessage("Webhook dispatch blocked after sequential failures")
		}
		return nil
	}

	logentry.Infof("posting webhook")

	timeout := time.Duration(environment.Settings.API.WebhookTimeout) * time.Millisecond
	result, err := dispatchservice.SendWebhook(message, &dispatchservice.WebhookRequest{
		ConnectionString: source.ConnectionString,
		Wid:              source.Wid,
		Extra:            source.Extra,
		Timeout:          timeout,
	}, logentry)

	// Always increment webhooks sent counter
	WebhooksSent.Inc()

	// Record latency
	duration := time.Duration(0)
	statusCode := 0
	if result != nil {
		duration = result.Duration
		statusCode = result.StatusCode
	}
	WebhookLatency.WithLabelValues().Observe(duration.Seconds())
	eventAttributes := map[string]string{
		"dispatch_type": source.Type,
	}
	if statusCode > 0 {
		eventAttributes["http_status"] = fmt.Sprintf("%d", statusCode)
	}
	if result != nil && result.TimedOut {
		eventAttributes["timed_out"] = "true"
	}

	if err != nil {
		logentry.Warnf("error at post webhook: %s", err.Error())

		// Check if it's a timeout error
		if result != nil && result.TimedOut {
			WebhookTimeouts.Inc()
			logentry.Warnf("webhook timeout after %v", timeout)
		}
	}

	if statusCode > 0 && statusCode != 200 {
		err = ErrInvalidResponse
		// Record HTTP error with status code
		WebhookHTTPErrors.WithLabelValues(fmt.Sprintf("%d", statusCode)).Inc()
	}

	if err != nil {
		WebhookSendErrors.Inc()
		if source.Failure == nil {
			source.Failure = &currentTime
		}
		source.publishDispatchingEvent("dispatch.webhook.delivery", "error", duration, eventAttributes)
		logentry.Errorf("webhook failed with status %d: %s", statusCode, err.Error())
		// Mark exceptions on message
		if message != nil {
			message.MarkExceptionsWithMessage(fmt.Sprintf("Webhook failed with status %d: %s", statusCode, err.Error()))
		}
	} else {
		// Webhook successful
		WebhookSuccess.Inc()
		source.Failure = nil
		source.Success = &currentTime
		source.publishDispatchingEvent("dispatch.webhook.delivery", "success", duration, eventAttributes)
		logentry.Infof("webhook posted successfully (status: %d, duration: %v)", statusCode, duration)
		// Clear exceptions on message
		if message != nil {
			message.ClearExceptions()
		}
	}

	return
}

// PublishRabbitMQ sends message via RabbitMQ using QuePasa fixed Exchange and routing key with intelligent routing
func (source *QpDispatching) PublishRabbitMQ(message *whatsapp.WhatsappMessage) (err error) {
	// updating log
	var messageIdForLog string
	if message != nil {
		messageIdForLog = message.Id
	}
	logentry := source.LogWithField(LogFields.MessageId, messageIdForLog)
	result, err := dispatchservice.PublishRabbitMQ(message, &dispatchservice.RabbitMQRequest{
		ConnectionString: source.ConnectionString,
		Extra:            source.Extra,
	}, logentry)

	// Mark as success only if connection is ready and message was truly published
	currentTime := time.Now().UTC()
	if err != nil {
		if source.Failure == nil {
			source.Failure = &currentTime
		}
		source.Success = nil

		eventAttributes := map[string]string{
			"dispatch_type": source.Type,
		}
		resultDuration := time.Duration(0)
		if result != nil {
			eventAttributes["routing_key"] = result.RoutingKey
			resultDuration = result.Duration
			if result.Cached {
				eventAttributes["cached"] = "true"
			}
		}
		source.publishDispatchingEvent("dispatch.rabbitmq.publish", "error", resultDuration, eventAttributes)

		if message != nil {
			if result != nil && result.Cached {
				message.MarkExceptionsWithMessage("RabbitMQ connection lost - message cached")
			} else {
				message.MarkExceptionsWithMessage(fmt.Sprintf("RabbitMQ publish failed for connection %s", source.ConnectionString))
			}
		}

		return err
	}

	source.Failure = nil
	source.Success = &currentTime

	eventAttributes := map[string]string{
		"dispatch_type": source.Type,
	}
	resultDuration := time.Duration(0)
	if result != nil {
		eventAttributes["routing_key"] = result.RoutingKey
		resultDuration = result.Duration
		if result.Cached {
			eventAttributes["cached"] = "true"
		}
	}
	source.publishDispatchingEvent("dispatch.rabbitmq.publish", "success", resultDuration, eventAttributes)

	if message != nil {
		message.ClearExceptions()
	}

	return nil
}

func (source *QpDispatching) publishDispatchingEvent(name string, status string, duration time.Duration, attributes map[string]string) {
	if source == nil {
		return
	}

	events.Publish(events.Event{
		Name:       name,
		Source:     "models.qp_dispatching",
		Status:     status,
		Duration:   duration,
		Attributes: attributes,
	})
}

// DetermineRoutingKey determines the appropriate routing key based on message type and content
// Returns one of the QuePasa standard routing keys using a rule-based approach for better performance
func (source *QpDispatching) DetermineRoutingKey(message *whatsapp.WhatsappMessage) string {
	return dispatchservice.DetermineRoutingKey(message)
}
