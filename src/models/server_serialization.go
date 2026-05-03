package models

import (
	"encoding/json"
	"time"

	library "github.com/nocodeleaks/quepasa/library"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// MarshalJSON customizes JSON serialization to include only dispatching field instead of webhooks.
func (source QpWhatsappServer) MarshalJSON() ([]byte, error) {
	// Create a custom struct to control serialization
	type customServer struct {
		whatsapp.WhatsappOptions
		Token       string           `json:"token"`
		Wid         string           `json:"wid,omitempty"`
		Verified    bool             `json:"verified"`
		Devel       bool             `json:"devel"`
		Metadata    QpMetadata       `json:"metadata,omitempty"`
		User        string           `json:"user,omitempty"`
		Timestamp   time.Time        `json:"timestamp,omitempty"`
		Reconnect   bool             `json:"reconnect"`
		StartTime   time.Time        `json:"starttime,omitempty"`
		Timestamps  QpTimestamps     `json:"timestamps"`
		Dispatching []*QpDispatching `json:"dispatching,omitempty"`
		Uptime      library.Duration `json:"uptime"`
	}

	// Get dispatching data from memory (includes real-time failure/success updates)
	var dispatchingData []*QpDispatching
	if source.QpDataDispatching.Dispatching != nil {
		// Use in-memory dispatching data with real-time status
		dispatchingData = source.QpDataDispatching.Dispatching
	}

	// Prepare timestamps for serialization
	timestamps := source.Timestamps
	timestamps.Update = source.Timestamp

	// Calculate uptime
	uptime := time.Duration(0)
	if !timestamps.Start.IsZero() {
		uptime = time.Since(timestamps.Start)
	}

	payload := customServer{
		WhatsappOptions: source.WhatsappOptions,
		Token:           source.Token,
		Wid:             source.GetWId(),
		Verified:        source.Verified,
		Devel:           source.Devel,
		Metadata:        source.Metadata,
		User:            source.GetUser(),
		Timestamp:       source.Timestamp,
		Reconnect:       source.Reconnect,
		StartTime:       timestamps.Start,
		Timestamps:      timestamps,
		Dispatching:     dispatchingData,
		Uptime:          library.Duration(uptime),
	}

	return json.Marshal(payload)
}
