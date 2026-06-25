package whatsapp

// WhatsappMessageDebug contains debug information for unhandled events
type WhatsappMessageDebug struct {
	Event  string `json:"event"`
	Reason string `json:"reason"`         // Reason for the debug event
	Info   any    `json:"info,omitempty"` // Additional information about the event
}
