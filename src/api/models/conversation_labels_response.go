package api

import models "github.com/nocodeleaks/quepasa/models"

// ConversationLabelsResponse is the API transport shape for conversation label endpoints.
type ConversationLabelsResponse struct {
	models.QpResponse
	Affected uint                          `json:"affected,omitempty"`
	ChatID   string                        `json:"chatid,omitempty"`
	Label    *models.QpConversationLabel   `json:"label,omitempty"`
	Labels   []*models.QpConversationLabel `json:"labels,omitempty"`
}
