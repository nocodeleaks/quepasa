package api

import (
	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// PictureResponse is the API transport shape for profile-picture endpoints.
type PictureResponse struct {
	models.QpResponse
	Info *whatsapp.WhatsappProfilePicture `json:"info,omitempty"`
}
