package api

import (
	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// HealthResponseItem is the API projection of one WhatsApp server health state.
type HealthResponseItem struct {
	// Public token.
	Token string `json:"token"`

	// WhatsApp session id.
	Wid string `json:"wid"`

	// Current connection state.
	State whatsapp.WhatsappConnectionState `json:"state"`

	// Numeric representation of the connection state.
	StateCode int `json:"state_code,omitempty"`
}

// GetHealth reports whether the represented server is currently healthy.
func (source HealthResponseItem) GetHealth() bool {
	return source.State.IsHealthy()
}

// NewHealthResponseItem projects a live server into the API transport shape.
func NewHealthResponseItem(server *models.QpWhatsappServer) HealthResponseItem {
	state := server.GetState()
	return HealthResponseItem{
		Token:     server.Token,
		Wid:       server.Wid,
		State:     state,
		StateCode: int(state),
	}
}
