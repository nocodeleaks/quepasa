package api

import (
	"time"

	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

type HealthResponse struct {
	models.QpResponse

	// -- single item fields
	Wid       string                           `json:"wid,omitempty"`
	State     whatsapp.WhatsappConnectionState `json:"state,omitempty"`
	StateCode int                              `json:"state_code,omitempty"`

	// -- multiple items fields
	Items []models.QpHealthResponseItem `json:"items,omitempty"`
	Stats *HealthStats                  `json:"stats,omitempty"`

	// -- general fields
	Timestamp   time.Time            `json:"timestamp"`
	Version     string               `json:"version"`
	Uptime      time.Duration        `json:"uptime"`
	Environment *EnvironmentSettings `json:"environment,omitempty"`
}
