package sipproxy

// SIPProxyConfig contains configuration for external SIP server
type SIPProxyConfig struct {
	ServerHost   string `json:"server_host"`
	ServerPort   int    `json:"server_port"`
	ListenerPort int    `json:"listener_port"` // Port to listen for SIP responses
	Username     string `json:"username,omitempty"`
	Password     string `json:"password,omitempty"`
	Protocol     string `json:"protocol"` // "UDP", "TCP", "TLS"
	Enabled      bool   `json:"enabled"`
}

// NewSIPProxyConfig creates a new SIP proxy configuration with default values
func NewSIPProxyConfig() *SIPProxyConfig {
	return &SIPProxyConfig{
		ServerHost:   "voip.sufficit.com.br",
		ServerPort:   26499,
		ListenerPort: 0, // Use 0 to force automatic port selection (avoid fixed port conflicts)
		Protocol:     "UDP",
		Enabled:      true,
	}
}

// SetCredentials sets the username and password for SIP authentication
func (spc *SIPProxyConfig) SetCredentials(username, password string) {
	spc.Username = username
	spc.Password = password
}

// SetServer sets the SIP server host and port
func (spc *SIPProxyConfig) SetServer(host string, port int) {
	spc.ServerHost = host
	spc.ServerPort = port
}

// SetProtocol sets the SIP protocol (UDP, TCP, TLS)
func (spc *SIPProxyConfig) SetProtocol(protocol string) {
	spc.Protocol = protocol
}

// IsValid checks if the configuration is valid
func (spc *SIPProxyConfig) IsValid() bool {
	return spc.ServerHost != "" && 
		   spc.ServerPort > 0 && 
		   spc.ListenerPort > 0 && 
		   spc.Protocol != "" &&
		   spc.Enabled
}
