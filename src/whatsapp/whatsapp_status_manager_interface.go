package whatsapp

// WhatsappStatusManagerInterface defines the interface for connection status and information management operations
// This interface should be implemented by the status manager in the whatsmeow package
type WhatsappStatusManagerInterface interface {
	// Get WhatsApp connection version
	GetVersion() string

	// Get WhatsApp ID (WID)
	GetWid() string

	// Get WhatsApp ID with error handling
	GetWidInternal() (string, error)

	// Check if connection is valid (connected and logged in)
	IsValid() bool

	// Check if connection is established
	IsConnected() bool

	// Get current connection status
	GetStatus() WhatsappConnectionState

	// Get auto-reconnect setting
	GetReconnect() bool
}

// IWhatsappConnectionWithStatus extends IWhatsappConnection with status management
// Use this interface when you need both connection and status operations
type IWhatsappConnectionWithStatus interface {
	IWhatsappConnection

	// GetStatusManager returns the status manager for status operations
	GetStatusManager() WhatsappStatusManagerInterface
}
