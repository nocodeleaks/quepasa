package models

import (
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

type QpHealthResponseItem struct {
	// Public token
	Token string `json:"token"`

	// Whatsapp session id
	Wid string `json:"wid"`

	Status whatsapp.WhatsappConnectionState `json:"status"`
}

// Check if the state is ready or manually stopped
func (source QpHealthResponseItem) GetHealth() bool {
	return source.Status.IsHealthy()
}
