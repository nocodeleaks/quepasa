package sipproxy

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// RTPProxy manages the RTP media bridge between WhatsApp and SIP server
type RTPProxy struct {
	logger        *logrus.Logger
	activeStreams map[string]*RTPStream
	streamMutex   sync.RWMutex
	localIP       string
	publicIP      string
}

// RTPStream represents an active RTP media stream for a call
type RTPStream struct {
	CallID           string
	WhatsAppPort     int          // Porta para receber do WhatsApp
	SIPPort          int          // Porta para enviar ao servidor SIP
	RemoteHost       string       // Host do servidor SIP
	RemotePort       int          // Porta do servidor SIP
	WhatsAppConn     *net.UDPConn // Conexão para WhatsApp
	SIPConn          *net.UDPConn // Conexão para servidor SIP
	isActive         bool
	lastPacketTime   time.Time
	packetsForwarded int64
	bytesForwarded   int64
}

// NewRTPProxy creates a new RTP proxy instance
func NewRTPProxy(logger *logrus.Logger, localIP, publicIP string) *RTPProxy {
	return &RTPProxy{
		logger:        logger,
		activeStreams: make(map[string]*RTPStream),
		localIP:       localIP,
		publicIP:      publicIP,
	}
}

// CreateRTPStream creates a new RTP stream for a call
func (rtp *RTPProxy) CreateRTPStream(callID string, remoteHost string, remotePort int) (*RTPStream, error) {
	rtp.streamMutex.Lock()
	defer rtp.streamMutex.Unlock()

	// Check if stream already exists
	if existingStream, exists := rtp.activeStreams[callID]; exists {
		rtp.logger.Infof("🎵 RTP stream already exists for CallID: %s, WhatsApp port: %d", callID, existingStream.WhatsAppPort)
		return existingStream, nil
	}

	// Find available port for WhatsApp connection
	whatsAppPort, err := rtp.findAvailablePort()
	if err != nil {
		return nil, fmt.Errorf("failed to find available WhatsApp RTP port: %v", err)
	}

	// Find available port for SIP connection (different from WhatsApp port)
	sipPort, err := rtp.findAvailablePortExcluding(whatsAppPort)
	if err != nil {
		return nil, fmt.Errorf("failed to find available SIP RTP port: %v", err)
	}

	// Create UDP listener for WhatsApp RTP - usar 0.0.0.0 para aceitar conexões de qualquer IP
	whatsAppAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("0.0.0.0:%d", whatsAppPort))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve WhatsApp UDP address: %v", err)
	}

	whatsAppConn, err := net.ListenUDP("udp", whatsAppAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create WhatsApp UDP listener: %v", err)
	}

	// Create UDP listener for SIP RTP
	sipAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("0.0.0.0:%d", sipPort))
	if err != nil {
		whatsAppConn.Close()
		return nil, fmt.Errorf("failed to resolve SIP UDP address: %v", err)
	}

	sipConn, err := net.ListenUDP("udp", sipAddr)
	if err != nil {
		whatsAppConn.Close()
		return nil, fmt.Errorf("failed to create SIP UDP listener: %v", err)
	}

	stream := &RTPStream{
		CallID:         callID,
		WhatsAppPort:   whatsAppPort,
		SIPPort:        sipPort,
		RemoteHost:     remoteHost,
		RemotePort:     remotePort,
		WhatsAppConn:   whatsAppConn,
		SIPConn:        sipConn,
		isActive:       true,
		lastPacketTime: time.Now(),
	}

	rtp.activeStreams[callID] = stream

	// Start RTP forwarding in background
	go rtp.startRTPForwarding(stream)

	rtp.logger.Infof("🎵✅ RTP stream created for CallID: %s", callID)
	rtp.logger.Infof("   � WhatsApp port %d ← WhatsApp RTP packets", whatsAppPort)
	rtp.logger.Infof("   � SIP port %d → %s:%d (SIP server)", sipPort, remoteHost, remotePort)

	return stream, nil
}

// findAvailablePort finds an available port in the RTP range
func (rtp *RTPProxy) findAvailablePort() (int, error) {
	for port := RTP_MEDIA_PORT_MIN; port <= RTP_MEDIA_PORT_MAX; port += 2 { // RTP uses even ports
		// Check if port is already in use
		inUse := false
		for _, stream := range rtp.activeStreams {
			if stream.WhatsAppPort == port || stream.SIPPort == port {
				inUse = true
				break
			}
		}

		if !inUse {
			// Test if port is actually available
			testAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
			if err != nil {
				continue
			}

			testConn, err := net.ListenUDP("udp", testAddr)
			if err != nil {
				continue // Port not available
			}
			testConn.Close()

			return port, nil
		}
	}

	return 0, fmt.Errorf("no available ports in range %d-%d", RTP_MEDIA_PORT_MIN, RTP_MEDIA_PORT_MAX)
}

// findAvailablePortExcluding finds an available port excluding a specific port
func (rtp *RTPProxy) findAvailablePortExcluding(excludePort int) (int, error) {
	for port := RTP_MEDIA_PORT_MIN; port <= RTP_MEDIA_PORT_MAX; port += 2 { // RTP uses even ports
		if port == excludePort {
			continue // Skip the excluded port
		}

		// Check if port is already in use
		inUse := false
		for _, stream := range rtp.activeStreams {
			if stream.WhatsAppPort == port || stream.SIPPort == port {
				inUse = true
				break
			}
		}

		if !inUse {
			// Test if port is actually available
			testAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
			if err != nil {
				continue
			}

			testConn, err := net.ListenUDP("udp", testAddr)
			if err != nil {
				continue // Port not available
			}
			testConn.Close()

			return port, nil
		}
	}

	return 0, fmt.Errorf("no available ports in range %d-%d excluding %d", RTP_MEDIA_PORT_MIN, RTP_MEDIA_PORT_MAX, excludePort)
}

// startRTPForwarding starts bidirectional RTP packet forwarding between WhatsApp and SIP server
func (rtp *RTPProxy) startRTPForwarding(stream *RTPStream) {
	defer stream.WhatsAppConn.Close()
	defer stream.SIPConn.Close()

	rtp.logger.Infof("🎵🚀 RTP FORWARDING STARTED for CallID: %s", stream.CallID)
	rtp.logger.Infof("   🎵 WhatsApp port %d ← WhatsApp RTP packets", stream.WhatsAppPort)
	rtp.logger.Infof("   🎵 SIP port %d → %s:%d (SIP server)", stream.SIPPort, stream.RemoteHost, stream.RemotePort)

	// Create remote connection for forwarding to SIP server
	remoteAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", stream.RemoteHost, stream.RemotePort))
	if err != nil {
		rtp.logger.Errorf("❌ Failed to resolve remote SIP address: %v", err)
		return
	}

	// Create connection to SIP server for outbound packets
	remoteSIPConn, err := net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		rtp.logger.Errorf("❌ Failed to create remote SIP connection: %v", err)
		return
	}
	defer remoteSIPConn.Close()

	rtp.logger.Infof("✅ RTP connections ready - monitoring for packets...")

	// Start two goroutines for bidirectional forwarding
	// 1. WhatsApp → SIP Server
	go rtp.forwardWhatsAppToSIP(stream, remoteSIPConn)

	// 2. SIP Server → WhatsApp
	go rtp.forwardSIPToWhatsApp(stream, remoteSIPConn)

	// Keep the main goroutine alive
	for stream.isActive {
		time.Sleep(1 * time.Second)
	}
}

// forwardWhatsAppToSIP forwards RTP packets from WhatsApp to SIP server
func (rtp *RTPProxy) forwardWhatsAppToSIP(stream *RTPStream, remoteSIPConn *net.UDPConn) {
	buffer := make([]byte, 1500) // Standard MTU size for RTP packets
	stream.WhatsAppConn.SetReadDeadline(time.Now().Add(30 * time.Second))

	packetCount := 0
	lastLogTime := time.Now()

	for stream.isActive {
		// Read RTP packet from WhatsApp
		n, clientAddr, err := stream.WhatsAppConn.ReadFromUDP(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// Extend deadline if no packets received
				stream.WhatsAppConn.SetReadDeadline(time.Now().Add(30 * time.Second))
				rtp.logger.Infof("🔍 RTP DEBUG: No WhatsApp packets received on port %d for 30s, waiting...", stream.WhatsAppPort)
				continue
			}
			rtp.logger.Debugf("🎵 WhatsApp RTP read error for CallID %s: %v", stream.CallID, err)
			break
		}

		packetCount++
		stream.lastPacketTime = time.Now()
		stream.packetsForwarded++
		stream.bytesForwarded += int64(n)

		// Log first few packets and periodic updates
		if packetCount <= 5 || time.Since(lastLogTime) > 10*time.Second {
			rtp.logger.Infof("🎵📦 WhatsApp→SIP: Packet #%d (CallID=%s, WhatsAppPort=%d, Size=%d bytes, From=%s)",
				packetCount, stream.CallID, stream.WhatsAppPort, n, clientAddr.String())
			if packetCount > 5 {
				lastLogTime = time.Now()
			}
		}

		// Forward packet to SIP server
		_, err = remoteSIPConn.Write(buffer[:n])
		if err != nil {
			rtp.logger.Errorf("❌ Failed to forward WhatsApp packet to SIP server: %v", err)
			continue
		}

		// Reset read deadline
		stream.WhatsAppConn.SetReadDeadline(time.Now().Add(30 * time.Second))
	}

	rtp.logger.Infof("🎵🛑 WhatsApp→SIP forwarding stopped for CallID: %s", stream.CallID)
}

// forwardSIPToWhatsApp forwards RTP packets from SIP server to WhatsApp
func (rtp *RTPProxy) forwardSIPToWhatsApp(stream *RTPStream, remoteSIPConn *net.UDPConn) {
	buffer := make([]byte, 1500) // Standard MTU size for RTP packets

	var whatsAppAddr *net.UDPAddr // Store WhatsApp address from first received packet

	for stream.isActive {
		// Read RTP packet from SIP server
		remoteSIPConn.SetReadDeadline(time.Now().Add(1 * time.Second))
		n, err := remoteSIPConn.Read(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue // Timeout is expected, just continue
			}
			rtp.logger.Debugf("🎵 SIP RTP read error for CallID %s: %v", stream.CallID, err)
			continue
		}

		// If we don't have WhatsApp address yet, we can't forward
		if whatsAppAddr == nil {
			rtp.logger.Debugf("🎵 SIP→WhatsApp: Received packet but no WhatsApp address yet, dropping")
			continue
		}

		// Forward packet back to WhatsApp
		_, err = stream.WhatsAppConn.WriteToUDP(buffer[:n], whatsAppAddr)
		if err != nil {
			rtp.logger.Errorf("❌ Failed to forward SIP packet to WhatsApp: %v", err)
			continue
		}

		rtp.logger.Debugf("🎵📦 SIP→WhatsApp: Packet forwarded (Size=%d bytes, To=%s)", n, whatsAppAddr.String())
	}

	rtp.logger.Infof("🎵🛑 SIP→WhatsApp forwarding stopped for CallID: %s", stream.CallID)
}

// StopRTPStream stops and removes an RTP stream
func (rtp *RTPProxy) StopRTPStream(callID string) {
	rtp.streamMutex.Lock()
	defer rtp.streamMutex.Unlock()

	stream, exists := rtp.activeStreams[callID]
	if !exists {
		rtp.logger.Warnf("⚠️ RTP stream not found for CallID: %s", callID)
		return
	}

	stream.isActive = false
	stream.WhatsAppConn.Close()
	stream.SIPConn.Close()
	delete(rtp.activeStreams, callID)

	rtp.logger.Infof("🎵✅ RTP stream stopped for CallID: %s", callID)
	rtp.logger.Infof("   📊 Final stats: %d packets, %d bytes", stream.packetsForwarded, stream.bytesForwarded)
}

// GetActiveStreams returns the number of active RTP streams
func (rtp *RTPProxy) GetActiveStreams() int {
	rtp.streamMutex.RLock()
	defer rtp.streamMutex.RUnlock()
	return len(rtp.activeStreams)
}

// GetStreamInfo returns information about a specific stream
func (rtp *RTPProxy) GetStreamInfo(callID string) *RTPStream {
	rtp.streamMutex.RLock()
	defer rtp.streamMutex.RUnlock()
	return rtp.activeStreams[callID]
}

// UpdateWhatsAppEndpoint updates the WhatsApp endpoint for an existing stream
func (rtp *RTPProxy) UpdateWhatsAppEndpoint(callID, whatsappIP string, whatsappPort int) error {
	rtp.streamMutex.Lock()
	defer rtp.streamMutex.Unlock()

	stream, exists := rtp.activeStreams[callID]
	if !exists {
		return fmt.Errorf("RTP stream not found for callID: %s", callID)
	}

	rtp.logger.Infof("🎵🔧 [RTP-UPDATE] Updating WhatsApp endpoint for %s: %s:%d", callID, whatsappIP, whatsappPort)
	rtp.logger.Infof("🎵📡 [RTP-ENDPOINT] Stream %s (port %d) configurado para encaminhar para %s:%d",
		callID, stream.WhatsAppPort, whatsappIP, whatsappPort)

	// A configuração do endpoint específico pode ser feita aqui
	// Por enquanto apenas logamos a informação

	return nil
}

// UpdateServerEndpoint updates the SIP server endpoint for an existing stream
func (rtp *RTPProxy) UpdateServerEndpoint(callID, serverHost string, serverPort int) error {
	rtp.streamMutex.Lock()
	defer rtp.streamMutex.Unlock()

	stream, exists := rtp.activeStreams[callID]
	if !exists {
		return fmt.Errorf("RTP stream not found for callID: %s", callID)
	}

	rtp.logger.Infof("🎵🔧 [RTP-SERVER-UPDATE] Updating SIP server endpoint for %s: %s:%d → %s:%d",
		callID, stream.RemoteHost, stream.RemotePort, serverHost, serverPort)

	// Update the stream's remote endpoint
	stream.RemoteHost = serverHost
	stream.RemotePort = serverPort

	rtp.logger.Infof("🎵✅ [RTP-SERVER-UPDATED] SIP server endpoint updated successfully for CallID: %s", callID)
	rtp.logger.Infof("🎵📡 [RTP-NEW-FLOW] WhatsApp port %d ← → SIP port %d → %s:%d",
		stream.WhatsAppPort, stream.SIPPort, serverHost, serverPort)

	return nil
}

// CreateRTPStreamWithLocalPort creates an RTP stream using a specific local port for WhatsApp
// This ensures the WhatsApp port from INVITE SDP matches the RTP bridge port
func (rtp *RTPProxy) CreateRTPStreamWithLocalPort(callID string, whatsAppPort int, remoteHost string, remotePort int) (*RTPStream, error) {
	rtp.streamMutex.Lock()
	defer rtp.streamMutex.Unlock()

	// Check if stream already exists
	if existingStream, exists := rtp.activeStreams[callID]; exists {
		rtp.logger.Infof("🎵 RTP stream already exists for CallID: %s, WhatsApp port: %d", callID, existingStream.WhatsAppPort)
		return existingStream, nil
	}

	rtp.logger.Infof("🎵🔢 Creating RTP stream with specified WhatsApp port: %d (ensuring port consistency)", whatsAppPort)

	// Find available port for SIP connection (different from WhatsApp port)
	sipPort, err := rtp.findAvailablePortExcluding(whatsAppPort)
	if err != nil {
		return nil, fmt.Errorf("failed to find available SIP RTP port: %v", err)
	}

	// Create UDP listener for WhatsApp RTP using the specified port
	whatsAppAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", rtp.localIP, whatsAppPort))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve WhatsApp UDP address: %v", err)
	}

	whatsAppConn, err := net.ListenUDP("udp", whatsAppAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create WhatsApp UDP listener on port %d: %v", whatsAppPort, err)
	}

	// Create UDP listener for SIP RTP
	sipAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", rtp.localIP, sipPort))
	if err != nil {
		whatsAppConn.Close()
		return nil, fmt.Errorf("failed to resolve SIP UDP address: %v", err)
	}

	sipConn, err := net.ListenUDP("udp", sipAddr)
	if err != nil {
		whatsAppConn.Close()
		return nil, fmt.Errorf("failed to create SIP UDP listener: %v", err)
	}

	stream := &RTPStream{
		CallID:         callID,
		WhatsAppPort:   whatsAppPort,
		SIPPort:        sipPort,
		RemoteHost:     remoteHost,
		RemotePort:     remotePort,
		WhatsAppConn:   whatsAppConn,
		SIPConn:        sipConn,
		isActive:       true,
		lastPacketTime: time.Now(),
	}

	// Store stream
	rtp.activeStreams[callID] = stream

	// Start bidirectional RTP forwarding
	go rtp.startRTPForwarding(stream)

	rtp.logger.Infof("🎵✅ RTP stream created with port consistency - CallID: %s, WhatsApp: %d, SIP: %d, Remote: %s:%d",
		callID, whatsAppPort, sipPort, remoteHost, remotePort)

	return stream, nil
}
