package environment

import (
	"strings"
)

// SIPProxyEnvironment handles all SIP proxy-related environment variables
type SIPProxyEnvironment struct{}

// SIP Proxy environment variable names
const (
	ENV_SIPPROXY_HOST       = "SIPPROXY_HOST"       // SIP server host (required for activation)
	ENV_SIPPROXY_PORT       = "SIPPROXY_PORT"       // SIP server port
	ENV_SIPPROXY_LOCALPORT  = "SIPPROXY_LOCALPORT"  // local SIP port
	ENV_SIPPROXY_PUBLICIP   = "SIPPROXY_PUBLICIP"   // public IP for SIP
	ENV_SIPPROXY_STUNSERVER = "SIPPROXY_STUNSERVER" // STUN server for NAT discovery
	ENV_SIPPROXY_USEUPNP    = "SIPPROXY_USEUPNP"    // enable UPnP port forwarding
	ENV_SIPPROXY_MEDIAPORTS = "SIPPROXY_MEDIAPORTS" // RTP media port range
	ENV_SIPPROXY_CODECS     = "SIPPROXY_CODECS"     // supported audio codecs
	ENV_SIPPROXY_USERAGENT  = "SIPPROXY_USERAGENT"  // SIP User-Agent string
	ENV_SIPPROXY_LOGLEVEL   = "SIPPROXY_LOGLEVEL"   // SIP proxy log level
	ENV_SIPPROXY_TIMEOUT    = "SIPPROXY_TIMEOUT"    // SIP transaction timeout
	ENV_SIPPROXY_RETRIES    = "SIPPROXY_RETRIES"    // SIP INVITE retry attempts
)

// Enabled checks if SIP proxy is enabled by checking if HOST is configured.
// If HOST exists, SIP Proxy is active. If HOST is empty, SIP Proxy is inactive.
func (env *SIPProxyEnvironment) Enabled() bool {
	host := env.Host()
	return len(strings.TrimSpace(host)) > 0
}

// Host returns the SIP server host. No default value - must be explicitly configured.
// If empty, SIP Proxy will be considered inactive.
func (env *SIPProxyEnvironment) Host() string {
	return getEnvOrDefaultString(ENV_SIPPROXY_HOST, "")
}

// Port returns the SIP server port. Defaults to 26499.
func (env *SIPProxyEnvironment) Port() uint32 {
	return getEnvOrDefaultUint32(ENV_SIPPROXY_PORT, 26499)
}

// LocalPort returns the local SIP port. Defaults to 5060.
func (env *SIPProxyEnvironment) LocalPort() uint32 {
	return getEnvOrDefaultUint32(ENV_SIPPROXY_LOCALPORT, 5060)
}

// PublicIP returns the configured public IP. Defaults to empty string (auto-discovery).
func (env *SIPProxyEnvironment) PublicIP() string {
	return getEnvOrDefaultString(ENV_SIPPROXY_PUBLICIP, "")
}

// STUNServer returns the STUN server for NAT discovery. Defaults to "stun.l.google.com:19302".
func (env *SIPProxyEnvironment) STUNServer() string {
	return getEnvOrDefaultString(ENV_SIPPROXY_STUNSERVER, "stun.l.google.com:19302")
}

// UseUPnP checks if UPnP port forwarding should be used. Defaults to true.
func (env *SIPProxyEnvironment) UseUPnP() bool {
	return getEnvOrDefaultBool(ENV_SIPPROXY_USEUPNP, true)
}

// MediaPorts returns the RTP media port range. Defaults to "10000-20000".
func (env *SIPProxyEnvironment) MediaPorts() string {
	return getEnvOrDefaultString(ENV_SIPPROXY_MEDIAPORTS, "10000-20000")
}

// Codecs returns the supported audio codecs. Defaults to "PCMU,PCMA,G729".
func (env *SIPProxyEnvironment) Codecs() string {
	return getEnvOrDefaultString(ENV_SIPPROXY_CODECS, "PCMU,PCMA,G729")
}

// UserAgent returns the SIP User-Agent string. Defaults to "QuePasa-SIP-Proxy/1.0".
func (env *SIPProxyEnvironment) UserAgent() string {
	return getEnvOrDefaultString(ENV_SIPPROXY_USERAGENT, "QuePasa-SIP-Proxy/1.0")
}

// LogLevel returns the SIP proxy log level. Defaults to "info".
func (env *SIPProxyEnvironment) LogLevel() string {
	return getEnvOrDefaultString(ENV_SIPPROXY_LOGLEVEL, "info")
}

// Timeout returns the SIP transaction timeout in seconds. Defaults to 30.
func (env *SIPProxyEnvironment) Timeout() uint32 {
	return getEnvOrDefaultUint32(ENV_SIPPROXY_TIMEOUT, 30)
}

// Retries returns the number of SIP INVITE retry attempts. Defaults to 3.
func (env *SIPProxyEnvironment) Retries() uint32 {
	return getEnvOrDefaultUint32(ENV_SIPPROXY_RETRIES, 3)
}
