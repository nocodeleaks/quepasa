package models

import (
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

type QpHealthResponseItem struct {
	// Public token
	Token string `json:"token"`

	// Whatsapp session id
	Wid string `json:"wid"`

	// Calculated current State of the connection
	State whatsapp.WhatsappConnectionState `json:"state"`

	// State code as integer
	StateCode int `json:"state_code,omitempty"`
}

// Check if the state is ready or manually stopped
func (source QpHealthResponseItem) GetHealth() bool {
	return source.State.IsHealthy()
}
