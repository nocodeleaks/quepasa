package models

import whatsapp "github.com/nocodeleaks/quepasa/whatsapp"

type QpPictureResponse struct {
	QpResponse
	Info *whatsapp.WhatsappProfilePicture `json:"info,omitempty"`
}
