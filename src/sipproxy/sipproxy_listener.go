package sipproxy

import (
	"context"
	"fmt"
	"net"

	"github.com/emiago/sipgo"
	log "github.com/sirupsen/logrus"
)

// SIPListener handles SIP server and raw UDP listener functionality
type SIPListener struct {
	logger             *log.Entry
	server             *sipgo.Server
	userAgent          *sipgo.UserAgent
	actualListenerPort int
	isRunning          bool
	stopChannel        chan bool
}

// NewSIPListener creates a new SIP listener instance
func NewSIPListener(logger *log.Entry) *SIPListener {
	return &SIPListener{
		logger:      logger,
		stopChannel: make(chan bool, 1),
	}
}

// FindAvailableUDPPort finds an available UDP port for the SIP listener
// Tries preferred SIP port range first (5060-5080), then fallback range
func (sl *SIPListener) FindAvailableUDPPort() (int, error) {
	sl.logger.Infof("🔍 Finding available UDP port for SIP listener...")

	// First try preferred SIP port range (5060-5080)
	minPort, maxPort := GetSIPPortRange()
	sl.logger.Infof("🎯 Trying preferred SIP port range: %d-%d", minPort, maxPort)

	for port := minPort; port <= maxPort; port++ {
		if sl.isPortAvailable(port) {
			sl.logger.Infof("✅ Found available UDP port %d in preferred range", port)
			return port, nil
		}
	}

	sl.logger.Warnf("⚠️ Preferred SIP port range (%d-%d) unavailable, trying fallback range", minPort, maxPort)

	// Try fallback range if preferred range is not available
	fallbackMin, fallbackMax := GetSIPPortFallbackRange()
	sl.logger.Infof("🔄 Trying fallback port range: %d-%d", fallbackMin, fallbackMax)

	for port := fallbackMin; port <= fallbackMax; port++ {
		if sl.isPortAvailable(port) {
			sl.logger.Infof("✅ Found available UDP port %d in fallback range", port)
			return port, nil
		}
	}

	// Last resort: let OS choose any available port
	sl.logger.Warnf("⚠️ Fallback range also unavailable, letting OS choose port")
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, fmt.Errorf("failed to listen on TCP: %v", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	sl.logger.Infof("✅ OS assigned port %d", port)
	return port, nil
}

// isPortAvailable checks if a specific UDP port is available
func (sl *SIPListener) isPortAvailable(port int) bool {
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}

	udpListener, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return false
	}
	defer udpListener.Close()

	return true
}

// StartListener starts the SIP server and raw UDP listener
func (sl *SIPListener) StartListener(config SIPProxySettings) error {
	// Try to find available port with preference for standard SIP range
	availablePort, err := sl.FindAvailableUDPPort()
	if err != nil {
		return fmt.Errorf("failed to find available UDP port: %v", err)
	}
	sl.actualListenerPort = availablePort

	// Show port selection info
	minPort, maxPort := GetSIPPortRange()
	if sl.actualListenerPort >= minPort && sl.actualListenerPort <= maxPort {
		sl.logger.Infof("✅ Using preferred SIP port: %d (standard SIP range)", sl.actualListenerPort)
	} else {
		fallbackMin, fallbackMax := GetSIPPortFallbackRange()
		if sl.actualListenerPort >= fallbackMin && sl.actualListenerPort <= fallbackMax {
			sl.logger.Infof("🔄 Using fallback port: %d (preferred ports unavailable)", sl.actualListenerPort)
		} else {
			sl.logger.Infof("🎲 Using OS-assigned port: %d (all preferred ranges unavailable)", sl.actualListenerPort)
		}
	}

	// Create SIPgo UserAgent
	ua, err := sipgo.NewUA()
	if err != nil {
		return fmt.Errorf("failed to create UserAgent: %v", err)
	}
	sl.userAgent = ua

	// Create server using the UserAgent
	server, err := sipgo.NewServer(ua)
	if err != nil {
		return fmt.Errorf("failed to create SIP server: %v", err)
	}
	sl.server = server

	// Listen on available port
	listenAddr := fmt.Sprintf(":%d", sl.actualListenerPort)
	sl.logger.Infof("🚀 Starting SIP server listener on %s", listenAddr)
	sl.logger.Infof("🔍 UserAgent and Server will share port %d for bidirectional communication", sl.actualListenerPort)

	// Start SIP server listener (this will handle all UDP traffic on the port)
	go func() {
		if err := sl.server.ListenAndServe(context.Background(), config.Protocol, listenAddr); err != nil {
			sl.logger.Errorf("❌ SIP server listen error: %v", err)
		}
	}()

	sl.isRunning = true
	sl.logger.Infof("✅ SIP Listener started successfully on port %d", sl.actualListenerPort)

	return nil
}

// Stop stops the SIP listener
func (sl *SIPListener) Stop() error {
	if !sl.isRunning {
		return nil
	}

	sl.logger.Infof("🛑 Stopping SIP Listener...")

	// Stop raw UDP listener
	select {
	case sl.stopChannel <- true:
	default:
	}

	if sl.server != nil {
		sl.server.Close()
	}

	sl.isRunning = false
	sl.logger.Infof("✅ SIP Listener stopped successfully")

	return nil
}

// GetUserAgent returns the SIP UserAgent
func (sl *SIPListener) GetUserAgent() *sipgo.UserAgent {
	return sl.userAgent
}

// GetServer returns the SIP Server
func (sl *SIPListener) GetServer() *sipgo.Server {
	return sl.server
}

// GetActualListenerPort returns the actual port being used
func (sl *SIPListener) GetActualListenerPort() int {
	return sl.actualListenerPort
}

// GetPort returns the actual port being used (alias for GetActualListenerPort)
func (sl *SIPListener) GetPort() int {
	return sl.actualListenerPort
}

// IsRunning returns true if the listener is running
func (sl *SIPListener) IsRunning() bool {
	return sl.isRunning
}
