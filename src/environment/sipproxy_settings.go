package environment

import (
	"strings"
)

// Environment variable constants for SIP proxy configuration
const (
	ENV_SIPPROXY_HOST           = "SIPPROXY_HOST"           // SIP server host (required for activation)
	ENV_SIPPROXY_PORT           = "SIPPROXY_PORT"           // SIP server port
	ENV_SIPPROXY_PROTOCOL       = "SIPPROXY_PROTOCOL"       // SIP server protocol
	ENV_SIPPROXY_LOCALPORT      = "SIPPROXY_LOCALPORT"      // local SIP port
	ENV_SIPPROXY_PUBLICIP       = "SIPPROXY_PUBLICIP"       // public IP for SIP
	ENV_SIPPROXY_STUNSERVER     = "SIPPROXY_STUNSERVER"     // STUN server for NAT discovery
	ENV_SIPPROXY_USEUPNP        = "SIPPROXY_USEUPNP"        // enable UPnP port forwarding
	ENV_SIPPROXY_MEDIAPORTS     = "SIPPROXY_MEDIAPORTS"     // RTP media port range
	ENV_SIPPROXY_CODECS         = "SIPPROXY_CODECS"         // supported audio codecs
	ENV_SIPPROXY_USERAGENT      = "SIPPROXY_USERAGENT"      // SIP User-Agent string
	ENV_SIPPROXY_LOGLEVEL       = "SIPPROXY_LOGLEVEL"       // SIP proxy log level
	ENV_SIPPROXY_TIMEOUT        = "SIPPROXY_TIMEOUT"        // SIP transaction timeout
	ENV_SIPPROXY_RETRIES        = "SIPPROXY_RETRIES"        // SIP INVITE retry attempts
	ENV_SIPPROXY_SDPSESSIONNAME = "SIPPROXY_SDPSESSIONNAME" // SDP session name
)

// SIPProxySettings holds all SIP proxy configuration loaded from environment
type SIPProxySettings struct {
	Enabled        bool   `json:"enabled"`
	Host           string `json:"host"`
	Protocol       string `json:"protocol"`
	Port           uint32 `json:"port"`
	LocalPort      uint32 `json:"local_port"`
	PublicIP       string `json:"public_ip"`
	STUNServer     string `json:"stun_server"`
	UseUPnP        bool   `json:"use_upnp"`
	MediaPorts     string `json:"media_ports"`
	Codecs         string `json:"codecs"`
	UserAgent      string `json:"user_agent"`
	LogLevel       string `json:"log_level"`
	Timeout        uint32 `json:"timeout"`
	Retries        uint32 `json:"retries"`
	SDPSessionName string `json:"sdp_session_name"` // Optional SDP session name for media
}

// NewSIPProxySettings creates a new SIP proxy settings by loading all values from environment
func NewSIPProxySettings() SIPProxySettings {
	host := getEnvOrDefaultString(ENV_SIPPROXY_HOST, "")
	enabled := len(strings.TrimSpace(host)) > 0

	return SIPProxySettings{
		Enabled:        enabled,
		Host:           host,
		Protocol:       getEnvOrDefaultString(ENV_SIPPROXY_PROTOCOL, "UDP"),
		Port:           getEnvOrDefaultUint32(ENV_SIPPROXY_PORT, 5060),
		LocalPort:      getEnvOrDefaultUint32(ENV_SIPPROXY_LOCALPORT, 5060),
		PublicIP:       getEnvOrDefaultString(ENV_SIPPROXY_PUBLICIP, ""),
		STUNServer:     getEnvOrDefaultString(ENV_SIPPROXY_STUNSERVER, "stun.l.google.com:19302"),
		UseUPnP:        getEnvOrDefaultBool(ENV_SIPPROXY_USEUPNP, true),
		MediaPorts:     getEnvOrDefaultString(ENV_SIPPROXY_MEDIAPORTS, "10000-20000"),
		Codecs:         getEnvOrDefaultString(ENV_SIPPROXY_CODECS, "PCMU,PCMA,G729"),
		UserAgent:      getEnvOrDefaultString(ENV_SIPPROXY_USERAGENT, "QuePasa-SIPProxy/1.0"),
		LogLevel:       getEnvOrDefaultString(ENV_SIPPROXY_LOGLEVEL, "info"),
		Timeout:        getEnvOrDefaultUint32(ENV_SIPPROXY_TIMEOUT, 30),
		Retries:        getEnvOrDefaultUint32(ENV_SIPPROXY_RETRIES, 3),
		SDPSessionName: getEnvOrDefaultString(ENV_SIPPROXY_SDPSESSIONNAME, "QuePasa SDP"),
	}
}
