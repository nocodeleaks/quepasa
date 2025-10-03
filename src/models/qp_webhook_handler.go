package models

import (
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

	err := PostToWebhooksModern(source.server, payload)
	if err != nil {
		logentry.Errorf("error on handle webhook distributions: %s", err.Error())
	}
}

func (source *QPWebhookHandler) HasWebhook() bool {
	if source.server != nil {
		webhooks := source.server.GetWebhooks()
		return len(webhooks) > 0
	}
	return false
}
