package models

import (
	"reflect"
	"strings"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
)

type QPWebhookHandler struct {
	server *QpWhatsappServer
}

func (w *QPWebhookHandler) Handle(payload *whatsapp.WhatsappMessage) {
	if !w.HasWebhook() {
		return
	}

	if payload.Type == whatsapp.DiscardMessageType|whatsapp.UnknownMessageType {
		log.Debugf("ignoring unknown message type on webhook request: %v", reflect.TypeOf(&payload))
		return
	}

	if payload.Type == whatsapp.TextMessageType && len(strings.TrimSpace(payload.Text)) <= 0 {
		log.Debugf("ignoring empty text message on webhook request: %s", payload.Id)
		return
	}

	if payload.Chat.Id == "status@broadcast" && !w.server.HandleBroadcast {
		log.Debugf("ignoring broadcast message on webhook request: %s", payload.Id)
		return
	}

	PostToWebHookFromServer(w.server, payload)
}

func (w *QPWebhookHandler) HasWebhook() bool {
	if w.server != nil {
		return len(w.server.Webhooks) > 0
	}
	return false
}
