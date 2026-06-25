package sipproxy

import (
	"fmt"
	"net"
	"sync"
	"time"

	qplog "github.com/nocodeleaks/quepasa/qplog"
)

// RTPProxy manages the RTP media bridge between WhatsApp and SIP server
type RTPProxy struct {
	logger        qplog.Logger
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
func NewRTPProxy(logger qplog.Logger, localIP, publicIP string) *RTPProxy {
	return &RTPProxy{
		logger:        logger,
		activeStreams: make(map[string]*RTPStream),
		localIP:       localIP,
		publicIP:      publicIP,
	}
}

// bindHost returns the concrete local IP to bind RTP sockets to. Binding to a
// specific IP (not 0.0.0.0) forces outbound RTP to source from that address,
// keeping it consistent with the SDP/Contact address when the host has
// multiple interfaces or default routes.
func (rtp *RTPProxy) bindHost() string {
	return pickBindIP(rtp.localIP, rtp.publicIP)
}

// pickBindIP chooses the local address to bind SIP/RTP sockets to. It prefers
// the configured public IP when that address is actually assigned to a local
// interface (the common single-host case where "public IP" is the server's own
// routable address) so the source IP matches the SDP/Contact and the SIP
// server's IP ACL. Otherwise it falls back to the auto-discovered local IP, and
// finally to 0.0.0.0. Auto-discovery via the default route is unreliable on
// multi-homed hosts, so the explicitly configured public IP wins when usable.
func pickBindIP(localIP, publicIP string) string {
	if isLocalInterfaceIP(publicIP) {
		return publicIP
	}
	if isLocalInterfaceIP(localIP) {
		return localIP
	}
	if localIP != "" {
		return localIP
	}
	if publicIP != "" {
		return publicIP
	}
	return "0.0.0.0"
}

// isLocalInterfaceIP reports whether ip is assigned to a local network
// interface (and is therefore bindable as a socket source address).
func isLocalInterfaceIP(ip string) bool {
	if ip == "" {
		return false
	}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return false
	}
	for _, a := range addrs {
		if ipn, ok := a.(*net.IPNet); ok && ipn.IP.String() == ip {
			return true
		}
	}
	return false
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

	// Bind WhatsApp and SIP RTP sockets on free ports (0.0.0.0 = accept from any IP).
	// bindAvailablePort returns the live socket, so there is no rebind race.
	whatsAppConn, whatsAppPort, err := rtp.bindAvailablePort("0.0.0.0", 0)
	if err != nil {
		return nil, fmt.Errorf("failed to bind WhatsApp RTP port: %v", err)
	}

	sipConn, sipPort, err := rtp.bindAvailablePort("0.0.0.0", whatsAppPort)
	if err != nil {
		whatsAppConn.Close()
		return nil, fmt.Errorf("failed to bind SIP RTP port: %v", err)
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

// CreateRTPStreamRaw creates UDP sockets for a call WITHOUT starting the
// automatic forwarding goroutines. The caller (VoIPBridge) reads and writes
// packets directly via RTPStream.WhatsAppConn and RTPStream.SIPConn.
//
// This is used by the native VoIP bridge where the calls module handles codec
// conversion and the SIP side needs raw socket access.
func (rtp *RTPProxy) CreateRTPStreamRaw(callID string, remoteHost string, remotePort int) (*RTPStream, error) {
	rtp.streamMutex.Lock()
	defer rtp.streamMutex.Unlock()

	// Check if stream already exists
	if existingStream, exists := rtp.activeStreams[callID]; exists {
		rtp.logger.Infof("🎵 RTP stream already exists for CallID: %s, WhatsApp port: %d", callID, existingStream.WhatsAppPort)
		return existingStream, nil
	}

	// Bind WhatsApp and SIP RTP sockets to a concrete IP so outbound RTP sources
	// consistently. bindAvailablePort returns the live socket (no rebind race).
	host := rtp.bindHost()
	whatsAppConn, whatsAppPort, err := rtp.bindAvailablePort(host, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to bind WhatsApp RTP port: %v", err)
	}

	sipConn, sipPort, err := rtp.bindAvailablePort(host, whatsAppPort)
	if err != nil {
		whatsAppConn.Close()
		return nil, fmt.Errorf("failed to bind SIP RTP port: %v", err)
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

	// NOTE: No forwarding goroutines started — the caller owns I/O.
	rtp.logger.Infof("🎵✅ RTP raw stream created for CallID: %s (no forwarders)", callID)
	rtp.logger.Infof("   📥 WhatsApp port %d ← raw socket", whatsAppPort)
	rtp.logger.Infof("   📤 SIP port %d → %s:%d (raw socket)", sipPort, remoteHost, remotePort)

	return stream, nil
}

// bindAvailablePort scans the RTP media port range for a free even port on host,
// binds it, and returns the LIVE socket. The caller takes ownership of the conn.
//
// Unlike a test-bind/close/rebind probe, the returned socket is the one that
// stays bound, so there is no time-of-check/time-of-use window where another
// caller (or process) can steal the port between the check and the real bind.
// excludePort (0 = none) is skipped so the WhatsApp and SIP legs of the same
// stream never land on the same port. Caller must hold streamMutex.
func (rtp *RTPProxy) bindAvailablePort(host string, excludePort int) (*net.UDPConn, int, error) {
	for port := RTP_MEDIA_PORT_MIN; port <= RTP_MEDIA_PORT_MAX; port += 2 { // RTP uses even ports
		if port == excludePort {
			continue
		}

		// Skip ports already owned by an active stream.
		inUse := false
		for _, stream := range rtp.activeStreams {
			if stream.WhatsAppPort == port || stream.SIPPort == port {
				inUse = true
				break
			}
		}
		if inUse {
			continue
		}

		addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, port))
		if err != nil {
			continue
		}
		conn, err := net.ListenUDP("udp", addr)
		if err != nil {
			continue // port not available
		}
		return conn, port, nil
	}

	if excludePort != 0 {
		return nil, 0, fmt.Errorf("no available ports in range %d-%d excluding %d", RTP_MEDIA_PORT_MIN, RTP_MEDIA_PORT_MAX, excludePort)
	}
	return nil, 0, fmt.Errorf("no available ports in range %d-%d", RTP_MEDIA_PORT_MIN, RTP_MEDIA_PORT_MAX)
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

	// Create UDP listener for WhatsApp RTP using the specified (fixed) port.
	whatsAppAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", rtp.localIP, whatsAppPort))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve WhatsApp UDP address: %v", err)
	}

	whatsAppConn, err := net.ListenUDP("udp", whatsAppAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create WhatsApp UDP listener on port %d: %v", whatsAppPort, err)
	}

	// Bind the SIP RTP socket on a free port (different from the WhatsApp port).
	sipConn, sipPort, err := rtp.bindAvailablePort(rtp.localIP, whatsAppPort)
	if err != nil {
		whatsAppConn.Close()
		return nil, fmt.Errorf("failed to bind SIP RTP port: %v", err)
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
