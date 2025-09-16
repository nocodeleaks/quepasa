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

	req, err := http.NewRequest("POST", source.Url, bytes.NewBuffer(payloadJson))
	if err != nil {
		logentry.Errorf("failed to create HTTP request: %s", err.Error())
		return
	}

	req.Header.Set("User-Agent", "Quepasa")
	req.Header.Set("X-QUEPASA-WID", source.Wid)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	timeout := time.Duration(environment.Settings.API.GetWebhookTimeout()) * time.Second
	client.Timeout = timeout

	logentry.Debugf("executing HTTP request to: %s, timeout: %v", source.Url, timeout)
	resp, err := client.Do(req)

	// Always increment webhooks sent counter
	metrics.WebhooksSent.Inc()
	logentry.Debugf("webhook sent counter incremented")

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
	} else {
		// Webhook successful
		metrics.WebhookSuccess.Inc()
		source.Failure = nil
		source.Success = &currentTime
		logentry.Infof("webhook posted successfully (status: %d, duration: %v)", statusCode, duration)
	}

	return
}

// ToDispatching converts webhook configuration to dispatching format
func (source *QpWebhook) ToDispatching() *QpDispatching {
	return &QpDispatching{
		LogStruct:        source.LogStruct,
		WhatsappOptions:  source.WhatsappOptions,
		ConnectionString: source.Url,
		Type:             DispatchingTypeWebhook,
		ForwardInternal:  source.ForwardInternal,
		TrackId:          source.TrackId,
		Extra:            source.Extra,
		Failure:          source.Failure,
		Success:          source.Success,
		Timestamp:        source.Timestamp,
		Wid:              source.Wid,
	}
}
