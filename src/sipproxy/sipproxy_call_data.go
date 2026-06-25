package sipproxy

import (
	"time"
)

// SIPProxyCallData represents SIP/RTP information captured from WhatsApp calls
type SIPProxyCallData struct {
	CallID       string                 `json:"call_id"`
	From         string                 `json:"from"`
	To           string                 `json:"to"`
	StartTime    time.Time              `json:"start_time"`
	EndTime      *time.Time             `json:"end_time,omitempty"`
	Status       string                 `json:"status"` // "offered", "accepted", "terminated", "relay"
	RTPData      map[string]interface{} `json:"rtp_data,omitempty"`
	SIPHeaders   map[string]string      `json:"sip_headers,omitempty"`
	LatencyInfo  *SIPProxyLatencyData   `json:"latency_info,omitempty"`
	RawEventData interface{}            `json:"raw_event_data"`
	ServerHost   string                 `json:"server_host"`
	ServerPort   int                    `json:"server_port"`
}

// NewSIPProxyCallData creates a new SIP call data instance
func NewSIPProxyCallData(callID, from, to string) *SIPProxyCallData {
	return &SIPProxyCallData{
		CallID:     callID,
		From:       from,
		To:         to,
		StartTime:  time.Now(),
		Status:     "offered",
		RTPData:    make(map[string]interface{}),
		SIPHeaders: make(map[string]string),
	}
}

// SetStatus updates the call status
func (scd *SIPProxyCallData) SetStatus(status string) {
	scd.Status = status
	if status == "terminated" || status == "rejected" {
		now := time.Now()
		scd.EndTime = &now
	}
}

// AddRTPData adds RTP data to the call
func (scd *SIPProxyCallData) AddRTPData(key string, value interface{}) {
	scd.RTPData[key] = value
}

// AddSIPHeader adds a SIP header to the call
func (scd *SIPProxyCallData) AddSIPHeader(key, value string) {
	scd.SIPHeaders[key] = value
}

// GetDuration returns the call duration
func (scd *SIPProxyCallData) GetDuration() time.Duration {
	if scd.EndTime != nil {
		return scd.EndTime.Sub(scd.StartTime)
	}
	return time.Since(scd.StartTime)
}

// GetSessionID returns a session ID based on the call start time
func (scd *SIPProxyCallData) GetSessionID() int64 {
	return scd.StartTime.Unix()
}
