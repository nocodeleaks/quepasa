package sipproxy

import (
	"time"
)

// SIPProxyCallTestResult represents the result of a call test
type SIPProxyCallTestResult struct {
	TestID      string        `json:"test_id"`
	CallID      string        `json:"call_id"`
	StartTime   time.Time     `json:"start_time"`
	EndTime     time.Time     `json:"end_time"`
	Duration    time.Duration `json:"duration"`
	Success     bool          `json:"success"`
	ErrorMsg    string        `json:"error_msg,omitempty"`
	SIPResponse string        `json:"sip_response,omitempty"`
}
