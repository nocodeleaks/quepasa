package api

import (
	"time"

	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

type HealthResponse struct {
	models.QpResponse
	State     whatsapp.WhatsappConnectionState `json:"state,omitempty"`
	StateCode int                              `json:"state_code,omitempty"`
	Items     []models.QpHealthResponseItem    `json:"items,omitempty"`
	Stats     *HealthStats                     `json:"stats,omitempty"`
	Timestamp time.Time                        `json:"timestamp"`
	Version   string                           `json:"version"`
}
