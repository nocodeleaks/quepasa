package ports

import (
	"github.com/nocodeleaks/quepasa/whatsapp"
)

// WhatsappDriverFactory creates WhatsApp connections.
// This interface is owned by the domain (models) and implemented by whatsmeow.
// Breaking the models <-> whatsmeow import cycle per ADR-0003 and PLAN P1.1.
type WhatsappDriverFactory interface {
	// CreateEmptyConnection creates a new unpaired connection with callback.
	CreateEmptyConnection() (whatsapp.IWhatsappConnection, error)

	// CreateConnection creates a connection from options.
	CreateConnection(options *whatsapp.WhatsappConnectionOptions) (whatsapp.IWhatsappConnection, error)
}

// GlobalWhatsappDriverFactory is injected at startup by main.go.
// This is a transitional global until dependency injection is fully wired (P1.2).
var GlobalWhatsappDriverFactory WhatsappDriverFactory
