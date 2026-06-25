package sipproxy

import (
	"time"
)

// SIPProxyLatencyData contains call performance metrics
type SIPProxyLatencyData struct {
	Latency    uint32    `json:"latency"`
	Timestamp  time.Time `json:"timestamp"`
	RawContent []byte    `json:"raw_content,omitempty"`
}

// NewSIPProxyLatencyData creates a new latency data instance
func NewSIPProxyLatencyData(latency uint32) *SIPProxyLatencyData {
	return &SIPProxyLatencyData{
		Latency:   latency,
		Timestamp: time.Now(),
	}
}

// SetRawContent sets the raw content for the latency data
func (ld *SIPProxyLatencyData) SetRawContent(content []byte) {
	ld.RawContent = content
}

// IsExpired checks if the latency data is older than the specified duration
func (ld *SIPProxyLatencyData) IsExpired(duration time.Duration) bool {
	return time.Since(ld.Timestamp) > duration
}
