package api

import (
	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// MessageResponse is the API transport shape for single-message reads.
type MessageResponse struct {
	models.QpResponse
	Message *whatsapp.WhatsappMessage `json:"message,omitempty"`
}
