package models

import (
	"github.com/nocodeleaks/quepasa/library"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

type QPDispatchingHandler struct {
	library.LogStruct // logging
	server            *QpWhatsappServer
}

func (source *QPDispatchingHandler) HandleDispatching(payload *whatsapp.WhatsappMessage) {
	if !source.HasDispatching() {
		return
	}

	// updating log
	logentry := source.GetLogger()
	loglevel := logentry.Level
	logentry = logentry.WithField(LogFields.MessageId, payload.Id)
	logentry.Level = loglevel

	err := PostToDispatchingFromServer(source.server, payload)
	if err != nil {
		logentry.Errorf("error on handle dispatching distributions: %s", err.Error())
	}
}

func (source *QPDispatchingHandler) HasDispatching() bool {
	if source.server != nil {
		return len(source.server.QpDataDispatching.Dispatching) > 0
	}
	return false
}

func (source *QPDispatchingHandler) HasWebhook() bool {
	if source.server != nil {
		webhooks := source.server.QpDataDispatching.GetWebhooks()
		return len(webhooks) > 0
	}
	return false
}
