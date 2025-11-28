package api

import (
	models "github.com/nocodeleaks/quepasa/models"
)

// InformationResponse represents bot/server information with optional environment settings
type InformationResponse struct {
	models.QpResponse
	Version string                   `json:"version"`
	Server  *models.QpWhatsappServer `json:"server,omitempty"`
}

// ParseSuccess populates the response with server info
func (source *InformationResponse) ParseSuccess(server *models.QpWhatsappServer) {
	source.QpResponse.ParseSuccess("follow server information")
	source.Server = server
}

// PatchSuccess populates the response with server info and custom message
func (source *InformationResponse) PatchSuccess(server *models.QpWhatsappServer, message string) {
	source.QpResponse.ParseSuccess(message)
	source.Server = server
}
