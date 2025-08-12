package sipproxy

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/emiago/sipgo"
	log "github.com/sirupsen/logrus"
)

// SIPProxyManager is the main SIP proxy manager with refactored components
type SIPProxyManager struct {
	activeCalls map[string]*SIPProxyCallData
	mutex       sync.RWMutex
	logger      *log.Entry
	config      *SIPProxyConfig
	client      *sipgo.Client
	isRunning   bool
	publicIP    string

	// Refactored components
	stunDiscovery      *STUNDiscovery
	upnpManager        *UPnPManager
	sipListener        *SIPListener
	transactionMonitor *SIPTransactionMonitor
}

var (
	managerInstance *SIPProxyManager
	managerOnce     sync.Once
)

// GetSIPProxyManager returns the singleton instance of SIP proxy manager
func GetSIPProxyManager() *SIPProxyManager {
	managerOnce.Do(func() {
		logger := log.WithField("component", "sipproxy")

		managerInstance = &SIPProxyManager{
			activeCalls: make(map[string]*SIPProxyCallData),
			logger:      logger,
			config:      NewSIPProxyConfig(),
		}

		// Initialize refactored components
		managerInstance.stunDiscovery = NewSTUNDiscovery(logger)
		managerInstance.upnpManager = NewUPnPManager(logger)
		managerInstance.sipListener = NewSIPListener(logger)
		managerInstance.transactionMonitor = NewSIPTransactionMonitor(logger)
	})
	return managerInstance
}

// SetCallAcceptedHandler sets the callback for when calls are accepted
func (m *SIPProxyManager) SetCallAcceptedHandler(handler SIPCallAcceptedCallback) {
	m.transactionMonitor.SetCallbacks(handler, m.transactionMonitor.callRejectedHandler)
}

// SetCallRejectedHandler sets the callback for when calls are rejected
func (m *SIPProxyManager) SetCallRejectedHandler(handler SIPCallRejectedCallback) {
	m.transactionMonitor.SetCallbacks(m.transactionMonitor.callAcceptedHandler, handler)
}

// Initialize initializes the SIP proxy manager with all components
func (m *SIPProxyManager) Initialize() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.isRunning {
		return nil // Already initialized
	}

	m.logger.Infof("🎯 Initializing SIP Proxy Manager (IPv4 only)")
	m.logger.Infof("📡 Server: %s:%d", m.config.ServerHost, m.config.ServerPort)

	// 1. Discover public IP via STUN
	publicIP, err := m.stunDiscovery.DiscoverPublicIPv4()
	if err != nil {
		return fmt.Errorf("failed to discover public IP: %v", err)
	}
	m.publicIP = publicIP
	m.logger.Infof("🌐 Public IP: %s", m.publicIP)

	// 2. Setup UPnP for automatic port forwarding
	if err := m.upnpManager.Setup(); err != nil {
		m.logger.Warnf("⚠️ UPnP setup failed (continuing without UPnP): %v", err)
	}

	// 3. Start SIP listener with random port
	if err := m.sipListener.StartListener(m.config); err != nil {
		return fmt.Errorf("failed to start SIP listener: %v", err)
	}

	// 4. Open UPnP port for the listener
	if m.upnpManager.client != nil {
		if err := m.upnpManager.OpenPort(m.sipListener.GetActualListenerPort(), "UDP"); err != nil {
			m.logger.Warnf("⚠️ UPnP port opening failed: %v", err)
		}
	}

	// 5. Create SIP client using the UserAgent from listener
	client, err := sipgo.NewClient(m.sipListener.GetUserAgent())
	if err != nil {
		return fmt.Errorf("failed to create SIP client: %v", err)
	}
	m.client = client

	// 6. Setup transaction monitoring
	m.transactionMonitor.SetClient(m.client)

	m.isRunning = true
	m.logger.Infof("✅ SIP Proxy Manager initialized successfully")

	return nil
}

// SendSIPInvite sends a SIP INVITE to initiate a call
func (m *SIPProxyManager) SendSIPInvite(fromPhone, toPhone, callID string) error {
	if !m.isRunning {
		return fmt.Errorf("SIP proxy manager not initialized")
	}

	if m.client == nil {
		return fmt.Errorf("SIP client not available")
	}

	m.logger.Infof("📞 Sending SIP INVITE from %s to %s (CallID: %s)", fromPhone, toPhone, callID)

	// Get configuration details
	sipServer := "voip.sufficit.com.br:26499"
	localPort := m.sipListener.GetActualListenerPort()
	
	m.logger.Infof("🌐 SIP Server Target: %s", sipServer)
	m.logger.Infof("🌐 Public IP: %s", m.publicIP)
	m.logger.Infof("🌐 Local Port: %d", localPort)
	
	// Create SIP URI for destination
	destURI := fmt.Sprintf("sip:%s@%s", toPhone, sipServer)
	fromURI := fmt.Sprintf("sip:%s@%s:%d", fromPhone, m.publicIP, localPort)
	contactURI := fmt.Sprintf("sip:%s@%s:%d", fromPhone, m.publicIP, localPort)
	
	m.logger.Infof("📋 SIP INVITE Details:")
	m.logger.Infof("   🎯 Destination URI: %s", destURI)
	m.logger.Infof("   📤 From URI: %s", fromURI)
	m.logger.Infof("   📞 Contact URI: %s", contactURI)
	m.logger.Infof("   🆔 Call-ID: %s", callID)
	
	// Generate tags and sequence
	fromTag := generateTag()
	cseq := "1 INVITE"
	
	m.logger.Infof("   🏷️ From Tag: %s", fromTag)
	m.logger.Infof("   🔢 CSeq: %s", cseq)
	
	// Create SDP body
	sdpBody := fmt.Sprintf(`v=0
o=QuePasa %s 1 IN IP4 %s
s=QuePasa SIP Call
c=IN IP4 %s
t=0 0
m=audio %d RTP/AVP 0 8 18
a=rtpmap:0 PCMU/8000
a=rtpmap:8 PCMA/8000
a=rtpmap:18 G729/8000
a=sendrecv`, callID, m.publicIP, m.publicIP, localPort+1000)

	m.logger.Infof("📋 SDP Body Created:")
	m.logger.Infof("   📊 Media Port: %d", localPort+1000)
	m.logger.Infof("   🎵 Codecs: PCMU/8000, PCMA/8000, G729/8000")
	
	// Log the complete SIP INVITE message that would be sent
	m.logger.Infof("📨 Complete SIP INVITE Message:")
	m.logger.Infof("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	m.logger.Infof("INVITE %s SIP/2.0", destURI)
	m.logger.Infof("Via: SIP/2.0/UDP %s:%d;branch=z9hG4bK%s", m.publicIP, localPort, fromTag)
	m.logger.Infof("From: <%s>;tag=%s", fromURI, fromTag)
	m.logger.Infof("To: <%s>", destURI)
	m.logger.Infof("Call-ID: %s", callID)
	m.logger.Infof("CSeq: %s", cseq)
	m.logger.Infof("Contact: <%s>", contactURI)
	m.logger.Infof("Max-Forwards: 70")
	m.logger.Infof("User-Agent: QuePasa-SIP-Proxy/1.0")
	m.logger.Infof("Content-Type: application/sdp")
	m.logger.Infof("Content-Length: %d", len(sdpBody))
	m.logger.Infof("")
	
	// Log SDP body line by line
	sdpLines := strings.Split(sdpBody, "\n")
	for _, line := range sdpLines {
		m.logger.Infof("%s", line)
	}
	m.logger.Infof("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	// Simulate sending (since we don't have the actual SIP implementation yet)
	m.logger.Infof("🚀 Attempting to send SIP INVITE to %s", sipServer)
	m.logger.Infof("📡 Network target: voip.sufficit.com.br port 26499")
	
	// Monitor the transaction
	m.transactionMonitor.MonitorTransaction(fromPhone, toPhone, callID, "INVITE")

	m.logger.Infof("✅ SIP INVITE prepared and monitored successfully (CallID: %s)", callID)
	m.logger.Infof("⚠️ Note: Actual network transmission requires full SIP client implementation")
	
	return nil
}

// generateTag generates a random tag for SIP headers
func generateTag() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// GetPublicIP returns the public IPv4 address
func (m *SIPProxyManager) GetPublicIP() string {
	return m.publicIP
}

// GetPort returns the SIP listener port
func (m *SIPProxyManager) GetPort() int {
	if m.sipListener == nil {
		return 0
	}
	return m.sipListener.GetPort()
}

// IsRunning returns whether the manager is running
func (m *SIPProxyManager) IsRunning() bool {
	return m.isRunning
}

// Stop stops the SIP proxy manager and cleans up resources
func (m *SIPProxyManager) Stop() error {
	if !m.isRunning {
		return nil
	}

	m.logger.Infof("🛑 Stopping SIP Proxy Manager...")

	// Stop components in reverse order
	if m.transactionMonitor != nil {
		m.transactionMonitor.Stop()
	}

	if m.sipListener != nil {
		m.sipListener.Stop()
	}

	if m.upnpManager != nil {
		m.upnpManager.ClosePort()
	}

	if m.client != nil {
		m.client.Close()
	}

	m.isRunning = false
	m.logger.Infof("✅ SIP Proxy Manager stopped successfully")

	return nil
}

// RemoveCall removes a call from active calls
func (m *SIPProxyManager) RemoveCall(callID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.activeCalls[callID]; exists {
		delete(m.activeCalls, callID)
		m.logger.Infof("📞❌ Removed call: %s", callID)
	}
}

// GetActiveCalls returns a copy of active calls
func (m *SIPProxyManager) GetActiveCalls() map[string]*SIPProxyCallData {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Return a copy to avoid race conditions
	activeCalls := make(map[string]*SIPProxyCallData)
	for key, value := range m.activeCalls {
		activeCalls[key] = value
	}

	return activeCalls
}
