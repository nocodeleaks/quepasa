package models

import (
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

type QpMessageResponse struct {
	QpResponse
	Message *whatsapp.WhatsappMessage `json:"message,omitempty"`
}
