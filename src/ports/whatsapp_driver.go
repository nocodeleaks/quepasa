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

// WhatsappDeviceInfo is a transport-agnostic view of a driver device session.
// It lets the domain reason about paired devices without importing whatsmeow types.
type WhatsappDeviceInfo struct {
	JID      string
	PushName string
	Platform string
}

// WhatsappDriverService exposes driver-level queries the domain needs without
// importing whatsmeow directly. Owned by the domain, implemented by whatsmeow,
// injected at startup. Completes the models -> whatsmeow decoupling (PLAN P1.1).
type WhatsappDriverService interface {
	// GetContactManagerForWid returns a contact manager for a wid, falling back
	// to store-only access when there is no active connection.
	GetContactManagerForWid(wid string, conn whatsapp.IWhatsappConnection) (whatsapp.WhatsappContactManagerInterface, error)

	// ResolveMigratedWid returns the current device wid for a migrated phone number.
	ResolveMigratedWid(phone string) (string, error)

	// ListDevices returns every device session known to the driver database.
	ListDevices() ([]WhatsappDeviceInfo, error)
}

// GlobalWhatsappDriverService is injected at startup by main.go.
// Transitional global until constructor wiring lands (P1.2).
var GlobalWhatsappDriverService WhatsappDriverService
