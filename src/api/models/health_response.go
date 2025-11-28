package api

import (
	"encoding/json"
	"time"

	library "github.com/nocodeleaks/quepasa/library"
	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

type HealthResponse struct {
	models.QpResponse

	// -- general fields
	Timestamp time.Time         `json:"timestamp"`
	Uptime    *library.Duration `json:"uptime,omitempty" swaggertype:"object"`

	// -- single item fields
	Wid   string                            `json:"wid,omitempty"`
	State *whatsapp.WhatsappConnectionState `json:"state,omitempty"`

	// -- multiple items fields
	Items []models.QpHealthResponseItem `json:"items,omitempty"`
	Stats *HealthStats                  `json:"stats,omitempty"`
}

// MarshalJSON customizes JSON serialization to include computed state_code
func (h HealthResponse) MarshalJSON() ([]byte, error) {
	type Alias HealthResponse

	// Create auxiliary struct with state_code
	aux := &struct {
		*Alias
		StateCode *int `json:"state_code,omitempty"`
	}{
		Alias: (*Alias)(&h),
	}

	// Calculate state_code from state if present
	if h.State != nil {
		stateCode := h.State.EnumIndex()
		aux.StateCode = &stateCode
	}

	return json.Marshal(aux)
}
