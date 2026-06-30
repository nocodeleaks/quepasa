package sipproxy

import (
	"strconv"
	"strings"

	environment "github.com/nocodeleaks/quepasa/environment"
)

// RTP Media Port Configuration.
//
// The range is configurable via the SIPPROXY_MEDIAPORTS environment variable
// ("start-end", e.g. "10000-20000"), defining the start of the RTP port range
// and its maximum. Defaults to 10000-20000 when unset or invalid. These are
// package vars (not consts) so they can be initialised from the environment.
var (
	// RTP_MEDIA_PORT_MIN is the start (minimum) port for RTP media streams.
	RTP_MEDIA_PORT_MIN = 10000

	// RTP_MEDIA_PORT_MAX is the maximum port for RTP media streams.
	RTP_MEDIA_PORT_MAX = 20000

	// RTP_MEDIA_PORT_RANGE is the total span of available ports.
	RTP_MEDIA_PORT_RANGE = RTP_MEDIA_PORT_MAX - RTP_MEDIA_PORT_MIN
)

func init() {
	loadRTPMediaPortRange(environment.Settings.SIPProxy.MediaPorts)
}

// loadRTPMediaPortRange parses a "start-end" spec and applies it to the RTP port
// range vars. Invalid/empty input keeps the defaults.
func loadRTPMediaPortRange(spec string) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return
	}
	parts := strings.SplitN(spec, "-", 2)
	if len(parts) != 2 {
		return
	}
	start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	end, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil || start <= 0 || end <= start {
		return
	}
	RTP_MEDIA_PORT_MIN = start
	RTP_MEDIA_PORT_MAX = end
	RTP_MEDIA_PORT_RANGE = end - start
}

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
func (spc *SIPProxySettings) SetServer(host string, port int) {
	spc.ServerHost = host
	spc.ServerPort = port
}

// SetProtocol sets the SIP protocol (UDP, TCP, TLS)
func (spc *SIPProxySettings) SetProtocol(protocol string) {
	spc.Protocol = protocol
}

// IsValid checks if the configuration is valid
func (spc SIPProxySettings) IsValid() bool {
	return spc.ServerHost != "" &&
		spc.ServerPort > 0 &&
		spc.Protocol != ""
}
