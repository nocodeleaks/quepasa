package sipproxy

// SIPProxyNetworkManagerSettings holds configuration settings for the SIP proxy network manager
type SIPProxyNetworkManagerSettings struct {
	StunServer       string `json:"stun_server"`
	SIPServer        string `json:"sip_server"`
	SIPPort          int    `json:"sip_port"`
	LocalPort        int    `json:"local_port"`
	PublicIP         string `json:"public_ip"`
	LocalIP          string `json:"local_ip"`
	IsSTUNConfigured bool   `json:"is_stun_configured"`
}
