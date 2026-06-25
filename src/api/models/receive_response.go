package api

import (
	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// ReceiveResponse is the API transport shape for message receive/history endpoints.
type ReceiveResponse struct {
	models.QpResponse
	Total      uint64                     `json:"total"`
	Page       int                        `json:"page,omitempty"`
	Limit      int                        `json:"limit,omitempty"`
	TotalPages int                        `json:"totalPages,omitempty"`
	Messages   []whatsapp.WhatsappMessage `json:"messages,omitempty"`
	Server     *models.QpServer           `json:"server,omitempty"`
}
