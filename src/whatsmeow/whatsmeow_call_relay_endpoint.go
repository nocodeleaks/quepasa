package whatsmeow

import "time"

// RelayEndpoint represents a relay server endpoint learned from CallRelayLatency (<te> payload).
// Relay-only calls may rely on these endpoints instead of ICE candidates.
type RelayEndpoint struct {
	RelayName  string    `json:"relay_name,omitempty"`
	IP         string    `json:"ip,omitempty"`
	Port       int       `json:"port,omitempty"`
	Endpoint   string    `json:"endpoint,omitempty"`
	LatencyRaw string    `json:"latency_raw,omitempty"`
	ObservedAt time.Time `json:"observed_at"`
}
