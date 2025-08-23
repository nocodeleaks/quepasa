package sipproxy

import (
	"time"
)

// RTP Media Port Configuration
const (
	// RTP_MEDIA_PORT_MIN defines the minimum port for RTP media streams
	RTP_MEDIA_PORT_MIN = 10000

	// RTP_MEDIA_PORT_MAX defines the maximum port for RTP media streams
	RTP_MEDIA_PORT_MAX = 20000

	// RTP_MEDIA_PORT_RANGE defines the total range of available ports
	RTP_MEDIA_PORT_RANGE = RTP_MEDIA_PORT_MAX - RTP_MEDIA_PORT_MIN
)

// SIP Call Configuration
const (
	// SIP_INVITE_MAX_ATTEMPTS defines the maximum number of SIP INVITE attempts per call
	SIP_INVITE_MAX_ATTEMPTS = 1

	// SIP_INVITE_RETRY_INTERVAL defines the interval between SIP INVITE attempts (in seconds)
	SIP_INVITE_RETRY_INTERVAL = 5
)

// SIP Port Configuration
const (
	// SIP_PORT_MIN defines the minimum preferred port for SIP listener
	SIP_PORT_MIN = 5060

	// SIP_PORT_MAX defines the maximum preferred port for SIP listener
	SIP_PORT_MAX = 5080

	// SIP_PORT_FALLBACK_MIN defines fallback minimum port if preferred range fails
	SIP_PORT_FALLBACK_MIN = 10000

	// SIP_PORT_FALLBACK_MAX defines fallback maximum port if preferred range fails
	SIP_PORT_FALLBACK_MAX = 11000
)

// SIPProxySettings contains configuration for external SIP server
type SIPProxySettings struct {
	SIPProxyNetworkManagerSettings

	SDPSessionName string `json:"sdp_session_name"` // Optional SDP session name for media
	UserAgent      string `json:"user_agent"`
	ServerHost     string `json:"server_host"`
	ServerPort     int    `json:"server_port"`
	ListenerPort   int    `json:"listener_port"` // Port to listen for SIP responses
	Protocol       string `json:"protocol"`      // "UDP", "TCP", "TLS"
}

// GetRandomRTPMediaPort returns a random port within the RTP media port range
func GetRandomRTPMediaPort() int {
	// Generate random offset within the range
	offset := generatePortOffset()
	return RTP_MEDIA_PORT_MIN + (offset % RTP_MEDIA_PORT_RANGE)
}

// generatePortOffset generates a pseudo-random offset for port selection
func generatePortOffset() int {
	// Simple pseudo-random based on current time nanoseconds
	// This ensures different ports for concurrent calls
	return int(time.Now().UnixNano() % int64(RTP_MEDIA_PORT_RANGE))
}

// GetSIPInviteMaxAttempts returns the maximum number of SIP INVITE attempts
func GetSIPInviteMaxAttempts() int {
	return SIP_INVITE_MAX_ATTEMPTS
}

// GetSIPInviteRetryInterval returns the retry interval for SIP INVITE attempts
func GetSIPInviteRetryInterval() int {
	return SIP_INVITE_RETRY_INTERVAL
}

// GetSIPPortRange returns the preferred SIP port range
func GetSIPPortRange() (int, int) {
	return SIP_PORT_MIN, SIP_PORT_MAX
}

// GetSIPPortFallbackRange returns the fallback SIP port range
func GetSIPPortFallbackRange() (int, int) {
	return SIP_PORT_FALLBACK_MIN, SIP_PORT_FALLBACK_MAX
}

// SetServer sets the SIP server host and port
func (spc SIPProxySettings) SetServer(host string, port int) {
	spc.ServerHost = host
	spc.ServerPort = port
}

// SetProtocol sets the SIP protocol (UDP, TCP, TLS)
func (spc SIPProxySettings) SetProtocol(protocol string) {
	spc.Protocol = protocol
}

// IsValid checks if the configuration is valid
func (spc SIPProxySettings) IsValid() bool {
	return spc.ServerHost != "" &&
		spc.ServerPort > 0 &&
		spc.Protocol != ""
}
