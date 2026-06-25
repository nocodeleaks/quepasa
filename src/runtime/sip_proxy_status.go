package runtime

import (
	environment "github.com/nocodeleaks/quepasa/environment"
	sipproxy "github.com/nocodeleaks/quepasa/sipproxy"
)

// SIPProxyStatus is the business view of the SIP proxy: whether it is configured
// (via environment) and whether the live proxy is currently running. It never
// carries secrets — host/port are operational info already exposed to operators.
type SIPProxyStatus struct {
	Configured bool   `json:"configured"`
	Running    bool   `json:"running"`
	Host       string `json:"host,omitempty"`
	Port       int    `json:"port,omitempty"`
	Protocol   string `json:"protocol,omitempty"`
}

// GetSIPProxyStatus reports the SIP proxy state for UI/health consumers.
//
// "Configured" is derived from the environment (SIPPROXY_HOST present), so it is
// true even before any instance with VoIP enabled has spun up the proxy.
// "Running" reflects the live sipproxy singleton.
func GetSIPProxyStatus() SIPProxyStatus {
	env := environment.NewSIPProxySettings()
	live := sipproxy.CurrentStatus()

	return SIPProxyStatus{
		Configured: env.Enabled,
		Running:    live.Running,
		Host:       env.Host,
		Port:       int(env.Port),
		Protocol:   env.Protocol,
	}
}
