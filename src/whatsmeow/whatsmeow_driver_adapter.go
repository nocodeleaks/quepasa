package whatsmeow

import (
	"github.com/nocodeleaks/quepasa/ports"
	"github.com/nocodeleaks/quepasa/whatsapp"
)

// WhatsmeowDriverAdapter implements ports.WhatsappDriverFactory.
// This adapter breaks the models -> whatsmeow import cycle per PLAN P1.1.
type WhatsmeowDriverAdapter struct{}

// CreateEmptyConnection delegates to WhatsmeowService.
func (a *WhatsmeowDriverAdapter) CreateEmptyConnection() (whatsapp.IWhatsappConnection, error) {
	return WhatsmeowService.CreateEmptyConnection()
}

// CreateConnection delegates to WhatsmeowService.
func (a *WhatsmeowDriverAdapter) CreateConnection(options *whatsapp.WhatsappConnectionOptions) (whatsapp.IWhatsappConnection, error) {
	return WhatsmeowService.CreateConnection(options)
}

// Ensure WhatsmeowDriverAdapter implements ports.WhatsappDriverFactory at compile time.
var _ ports.WhatsappDriverFactory = (*WhatsmeowDriverAdapter)(nil)
