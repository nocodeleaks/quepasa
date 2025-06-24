package models

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"time"

	"github.com/nocodeleaks/quepasa/library"
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
	req.Header.Set("User-Agent", "Quepasa")
	req.Header.Set("X-QUEPASA-WID", source.Wid)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	client.Timeout = time.Second * 10
	resp, err := client.Do(req)
	if err != nil {
		logentry.Warnf("error at post webhook: %s", err.Error())
	}

	if resp != nil {
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			err = ErrInvalidResponse
		}
	}

	time := time.Now().UTC()
	if err != nil {
		if source.Failure == nil {
			source.Failure = &time
		}
	} else {
		source.Failure = nil
		source.Success = &time
	}

	return
}
