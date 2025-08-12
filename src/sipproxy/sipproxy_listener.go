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
func (sl *SIPListener) FindAvailableUDPPort() (int, error) {
	sl.logger.Infof("🔍 Finding available UDP port for SIP listener...")

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, fmt.Errorf("failed to listen on TCP: %v", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port

	// Verificar se a porta UDP também está disponível
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		return 0, fmt.Errorf("failed to resolve UDP address: %v", err)
	}

	udpListener, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return 0, fmt.Errorf("failed to listen on UDP: %v", err)
	}
	defer udpListener.Close()

	port = udpListener.LocalAddr().(*net.UDPAddr).Port
	sl.logger.Infof("✅ Found available UDP port %d for SIP listener", port)
	return port, nil
}

// StartListener starts the SIP server and raw UDP listener
func (sl *SIPListener) StartListener(config *SIPProxyConfig) error {
	// Sempre usar porta aleatória disponível para evitar conflitos com outros serviços SIP
	availablePort, err := sl.FindAvailableUDPPort()
	if err != nil {
		return fmt.Errorf("failed to find available UDP port: %v", err)
	}
	sl.actualListenerPort = availablePort
	sl.logger.Infof("🎲 Using random available port: %d (avoiding fixed port conflicts like 5060)", sl.actualListenerPort)

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

	// Start raw UDP listener for debugging incoming packets
	go sl.startRawUDPListener(listenAddr)

	// Start SIP server listener
	go func() {
		if err := sl.server.ListenAndServe(context.Background(), config.Protocol, listenAddr); err != nil {
			sl.logger.Errorf("❌ SIP server listen error: %v", err)
		}
	}()

	sl.isRunning = true
	sl.logger.Infof("✅ SIP Listener started successfully on port %d", sl.actualListenerPort)

	return nil
}

// startRawUDPListener starts a raw UDP listener for debugging
func (sl *SIPListener) startRawUDPListener(listenAddr string) {
	rawAddr, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		sl.logger.Errorf("❌ Failed to resolve raw UDP address: %v", err)
		return
	}

	rawConn, err := net.ListenUDP("udp", rawAddr)
	if err != nil {
		sl.logger.Errorf("❌ Failed to create raw UDP listener: %v", err)
		return
	}
	defer rawConn.Close()

	sl.logger.Infof("🔍🔍🔍 RAW UDP Debug listener started on %s", listenAddr)

	buf := make([]byte, 4096)
	for {
		select {
		case <-sl.stopChannel:
			sl.logger.Infof("🔍🔍🔍 RAW UDP Debug listener stopped")
			return
		default:
			n, addr, err := rawConn.ReadFromUDP(buf)
			if err != nil {
				sl.logger.Errorf("❌ Raw UDP read error: %v", err)
				continue
			}

			message := string(buf[:n])
			sl.logger.Infof("🔍🔍🔍 RAW UDP received %d bytes from %s:", n, addr)
			sl.logger.Infof("🔍🔍🔍 RAW UDP message: %s", message)
		}
	}
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
