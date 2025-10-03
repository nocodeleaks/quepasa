package models

import whatsapp "github.com/nocodeleaks/quepasa/whatsapp"

type QpDispatchingHandlerInterface interface {

	// method for init process of dispatching messages
	HandleDispatching(*whatsapp.WhatsappMessage)
}
