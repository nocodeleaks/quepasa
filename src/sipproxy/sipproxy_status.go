package sipproxy

// ProxyStatus is a snapshot of the SIP proxy singleton's live state. It carries
// only what callers need to report health/configuration — never credentials.
type ProxyStatus struct {
	Initialized bool   // the proxy singleton has been created
	Running     bool   // the proxy transport/listener is up
	Configured  bool   // the manager holds a valid SIP server configuration
	ServerHost  string // configured SIP server host
	ServerPort  int    // configured SIP server port
	Protocol    string // configured SIP transport (UDP/TCP/TLS)
}

// Status returns this manager's live status snapshot.
func (m *SIPProxyManager) Status() ProxyStatus {
	if m == nil {
		return ProxyStatus{}
	}
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return ProxyStatus{
		Initialized: true,
		Running:     m.isRunning,
		Configured:  m.config.IsValid(),
		ServerHost:  m.config.ServerHost,
		ServerPort:  m.config.ServerPort,
		Protocol:    m.config.Protocol,
	}
}

// CurrentStatus reports the singleton proxy's live state WITHOUT forcing its
// creation. When the proxy has never been initialized (e.g. no instance has
// VoIP enabled) the zero value is returned (Initialized/Running false).
func CurrentStatus() ProxyStatus {
	// managerInstance is only assigned inside managerOnce.Do during startup;
	// a plain read here is sufficient for a status probe.
	m := managerInstance
	if m == nil {
		return ProxyStatus{}
	}
	return m.Status()
}
