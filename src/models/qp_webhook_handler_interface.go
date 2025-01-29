package models

import whatsapp "github.com/nocodeleaks/quepasa/whatsapp"

type QpWebhookHandlerInterface interface {

	// method for init process of webhook messages
	HandleWebHook(*whatsapp.WhatsappMessage)
}
