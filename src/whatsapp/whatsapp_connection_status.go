package whatsapp

import (
	"time"
)

// WhatsappConnectionStatus represents detailed connection status information
// that can be used to analyze connection health and reconnection attempts
type WhatsappConnectionStatus struct {
	// Basic connection state
	State WhatsappConnectionState `json:"state"`

	// Connection health indicators
	IsConnected     bool `json:"is_connected"`
	IsAuthenticated bool `json:"is_authenticated"`
	IsConnecting    bool `json:"is_connecting"`
	IsReconnecting  bool `json:"is_reconnecting"`
	IsValid         bool `json:"is_valid"`

	// WhatsApp information (from StatusManager)
	Platform string `json:"platform,omitempty"`

	// 55254525452:{session_id}@s.whatsapp.net
	SessionId string `json:"session_id,omitempty"`

	ReconnectAttempts    uint32 `json:"reconnect_attempts"`
	AutoReconnectEnabled bool   `json:"auto_reconnect_enabled"`
	ReconnectErrors      uint32 `json:"reconnect_errors"`

	// Timing information
	LastSuccessfulConnect *time.Time     `json:"last_successful_connect,omitempty"`
	ConnectionUptime      *time.Duration `json:"connection_uptime,omitempty"`

	// Additional status indicators
	FailedToken bool `json:"failed_token"`
}
