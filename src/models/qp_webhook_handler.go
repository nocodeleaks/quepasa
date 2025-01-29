package models

import (
	"reflect"
	"strings"

	"github.com/nocodeleaks/quepasa/library"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

type QPWebhookHandler struct {
	library.LogStruct // logging
	server            *QpWhatsappServer
}

func (source *QPWebhookHandler) HandleWebHook(payload *whatsapp.WhatsappMessage) {
	if !source.HasWebhook() {
		return
	}

	// updating log
	logentry := source.GetLogger()
	loglevel := logentry.Level
	logentry = logentry.WithField(LogFields.MessageId, payload.Id)
	logentry.Level = loglevel

	if payload.Type == whatsapp.DiscardMessageType || payload.Type == whatsapp.UnknownMessageType {
		logentry.Debugf("ignoring discard|unknown message type on webhook request: %v", reflect.TypeOf(&payload))
		return
	}

	if payload.Type == whatsapp.TextMessageType && len(strings.TrimSpace(payload.Text)) <= 0 {
		logentry.Debugf("ignoring empty text message on webhook request: %s", payload.Id)
		return
	}

	err := PostToWebHookFromServer(source.server, payload)
	if err != nil {
		logentry.Errorf("error on handle webhook distributions: %s", err.Error())
	}
}

func (source *QPWebhookHandler) HasWebhook() bool {
	if source.server != nil {
		return len(source.server.Webhooks) > 0
	}
	return false
}
