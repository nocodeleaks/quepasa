package sipproxy

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/emiago/sipgo"
	"github.com/emiago/sipgo/sip"
	log "github.com/sirupsen/logrus"
)

// CallState represents the current state of a SIP call
type CallState int

const (
	CallStateInitiated CallState = iota
	CallStateInviting
	CallStateProceeding
	CallStateAccepted
	CallStateRejected
	CallStateTimeout
	CallStateCancelled
)

// CallInfo holds information about an active call
type CallInfo struct {
	CallID        string
	FromPhone     string
	ToPhone       string
	SIPTag        string // Generated SIP tag for this call
	LocalRTPPort  int    // RTP port announced in our SDP offer (SIP -> us)
	LocalRTPConn  *net.UDPConn
	State         CallState
	StartTime     time.Time
	LastUpdate    time.Time
	Context       context.Context
	CancelFunc    context.CancelFunc
	DialogSession *sipgo.DialogClientSession // SIP dialog session for BYE/CANCEL
}

func (scm *SIPCallManagerSipgo) setRTPMirrorPort(callID string, port int) {
	if scm == nil || callID == "" || port <= 0 {
		return
	}
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		scm.logger.Warnf("🎵 [RTP-MIRROR] Failed to resolve mirror addr for port=%d (CallID=%s): %v", port, callID, err)
		return
	}
	scm.rtpMirrorMutex.Lock()
	if scm.rtpMirrorAddr == nil {
		scm.rtpMirrorAddr = make(map[string]*net.UDPAddr)
	}
	scm.rtpMirrorAddr[callID] = addr
	scm.rtpMirrorMutex.Unlock()
	scm.logger.Infof("🎵 [RTP-MIRROR] Enabled: SIP RTP will be mirrored to %s (CallID=%s)", addr.String(), callID)
}

func (scm *SIPCallManagerSipgo) getRTPMirrorAddr(callID string) *net.UDPAddr {
	scm.rtpMirrorMutex.RLock()
	defer scm.rtpMirrorMutex.RUnlock()
	if scm.rtpMirrorAddr == nil {
		return nil
	}
	return scm.rtpMirrorAddr[callID]
}

func (scm *SIPCallManagerSipgo) allocateLocalRTPListener(callID string) (int, *net.UDPConn, error) {
	// Prefer even ports in the RTP range.
	start := RTP_MEDIA_PORT_MIN + int(time.Now().UnixNano()%1000)
	if start%2 != 0 {
		start++
	}
	for port := start; port <= RTP_MEDIA_PORT_MAX; port += 2 {
		addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("0.0.0.0:%d", port))
		if err != nil {
			continue
		}
		conn, err := net.ListenUDP("udp", addr)
		if err == nil {
			scm.logger.Infof("🎵 [RTP-MONITOR] Reserved local RTP port :%d (CallID=%s)", port, callID)
			return port, conn, nil
		}
	}
	// Wrap-around
	for port := RTP_MEDIA_PORT_MIN; port < start; port += 2 {
		addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("0.0.0.0:%d", port))
		if err != nil {
			continue
		}
		conn, err := net.ListenUDP("udp", addr)
		if err == nil {
			scm.logger.Infof("🎵 [RTP-MONITOR] Reserved local RTP port :%d (CallID=%s)", port, callID)
			return port, conn, nil
		}
	}
	return 0, nil, fmt.Errorf("no available RTP port in range %d-%d", RTP_MEDIA_PORT_MIN, RTP_MEDIA_PORT_MAX)
}

func (scm *SIPCallManagerSipgo) startLocalRTPMonitor(callInfo *CallInfo) {
	if callInfo == nil {
		return
	}
	port := callInfo.LocalRTPPort
	if port <= 0 {
		return
	}
	conn := callInfo.LocalRTPConn
	if conn == nil {
		scm.logger.Warnf("🎵 [RTP-MONITOR] No UDP conn reserved for :%d (CallID=%s)", port, callInfo.CallID)
		return
	}

	scm.logger.Infof("🎵 [RTP-MONITOR] Listening for SIP-side RTP on :%d (CallID=%s)", port, callInfo.CallID)

	go func() {
		buf := make([]byte, 2000)
		var packets int64
		var bytes int64
		var firstFrom string
		mirrorLogged := false
		for {
			select {
			case <-callInfo.Context.Done():
				return
			default:
			}
			_ = conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			n, from, err := conn.ReadFromUDP(buf)
			if err != nil {
				if ne, ok := err.(net.Error); ok && ne.Timeout() {
					continue
				}
				return
			}

			// Optional debug mirror: copy SIP-side RTP packets to a local UDP port (e.g. WhatsApp RTP bridge listener).
			if mirrorAddr := scm.getRTPMirrorAddr(callInfo.CallID); mirrorAddr != nil {
				if !mirrorLogged {
					mirrorLogged = true
					scm.logger.Infof("🎵 [RTP-MIRROR] Mirroring RTP stream to %s (CallID=%s)", mirrorAddr.String(), callInfo.CallID)
				}
				_, _ = conn.WriteToUDP(buf[:n], mirrorAddr)
			}

			packets++
			bytes += int64(n)
			if firstFrom == "" && from != nil {
				firstFrom = from.String()
				scm.logger.Infof("🎵 [RTP-MONITOR] First RTP packet: from=%s bytes=%d (CallID=%s)", firstFrom, n, callInfo.CallID)
			}
			if packets%50 == 0 {
				scm.logger.Infof("🎵 [RTP-MONITOR] RTP stats: packets=%d bytes=%d from=%s (CallID=%s)", packets, bytes, firstFrom, callInfo.CallID)
			}
		}
	}()
}

// SIPCallManagerSipgo manages SIP call lifecycle using sipgo package
type SIPCallManagerSipgo struct {
	logger          *log.Entry
	config          SIPProxySettings
	networkManager  *SIPProxyNetworkManager
	sipClient       *sipgo.Client
	userAgent       *sipgo.UserAgent
	dialogUA        *sipgo.DialogUA
	activeCalls     map[string]*CallInfo
	rtpMirrorMutex  sync.RWMutex
	rtpMirrorAddr   map[string]*net.UDPAddr
	cancelCallCount map[string]int // Track how many times CancelCall is called per CallID
	cancelMutex     sync.RWMutex   // Protect the counter
	defaultTimeout  time.Duration
	onCallRejected  SIPCallRejectedCallback // Callback for call rejection
	onCallAccepted  SIPCallAcceptedCallback // Callback for call acceptance
}

// SetRTPMirrorPort enables debug mirroring of SIP-side RTP packets to a local UDP port.
// Intended to feed other local listeners (e.g. WhatsApp RTP bridge) for correlation/debug.
func (scm *SIPCallManagerSipgo) SetRTPMirrorPort(callID string, port int) {
	scm.setRTPMirrorPort(callID, port)
}

// NewSIPCallManagerSipgo creates a new SIP call manager using sipgo
func NewSIPCallManagerSipgo(logger *log.Entry, config SIPProxySettings, networkManager *SIPProxyNetworkManager) *SIPCallManagerSipgo {
	// Get User-Agent from config
	userAgentName := config.UserAgent

	if !networkManager.IsConfigured() {
		err := networkManager.ConfigureNetwork()
		if err != nil {
			logger.Errorf("❌ Failed to configure network: %v", err)
			return nil
		}
	}

	// Get network configuration for UserAgent
	localIP := networkManager.GetLocalIP()
	localPort := networkManager.GetLocalPort()
	publicIP := networkManager.GetPublicIP()

	// LOG CONFIGURATION VALUES FOR DEBUGGING
	logger.Infof("🔍 SIPGO CONFIGURATION DEBUG:")
	logger.Infof("   🌐 ServerHost: %s", config.ServerHost)
	logger.Infof("   🌐 ServerPort: %d (should be 26499)", config.ServerPort)
	logger.Infof("   🏠 LocalIP: %s", localIP)
	logger.Infof("   🏠 LocalPort: %d", localPort)
	logger.Infof("   🏠 PublicIP: %s", publicIP)
	logger.Infof("   🏷️ UserAgent: %s", userAgentName)

	// Initialize sipgo UserAgent with complete configuration
	ua, err := sipgo.NewUA(
		sipgo.WithUserAgent(userAgentName),
		sipgo.WithUserAgentHostname(localIP),
	)
	if err != nil {
		logger.Errorf("❌ Failed to create sipgo UserAgent: %v", err)
		return nil
	}

	logger.Infof("✅ UserAgent configured: %s@%s:%d", userAgentName, localIP, localPort)

	// Create sipgo Client with explicit listen address
	// Try to bind to the specific IP:port to ensure Via header is correct
	listenAddr := fmt.Sprintf("%s:%d", localIP, localPort)
	client, err := sipgo.NewClient(ua, sipgo.WithClientAddr(listenAddr))
	if err != nil {
		logger.Errorf("❌ Failed to create sipgo Client with addr %s: %v", listenAddr, err)
		// Fallback to default client if binding fails
		client, err = sipgo.NewClient(ua)
		if err != nil {
			logger.Errorf("❌ Failed to create sipgo Client even with fallback: %v", err)
			return nil
		}
		logger.Warnf("⚠️ Using default client binding, Via header may show 0.0.0.0")
	} else {
		logger.Infof("🌐 SIP Client bound to specific address: %s", listenAddr)
	}

	// Create contact header for this client using UserAgent name
	contactHDR := sip.ContactHeader{
		Address: sip.Uri{
			Scheme: "sip",
			Host:   publicIP,
			Port:   localPort,
		}, Params: sip.NewParams().Add("transport", "udp"),
	}

	// Create DialogUA for managing calls
	dialogUA := &sipgo.DialogUA{
		Client:     client,
		ContactHDR: contactHDR,
	}

	logger.Info("✅ SIPCallManagerSipgo initialized with sipgo client")

	return &SIPCallManagerSipgo{
		logger:          logger,
		config:          config,
		networkManager:  networkManager,
		sipClient:       client,
		userAgent:       ua,
		dialogUA:        dialogUA,
		activeCalls:     make(map[string]*CallInfo),
		cancelCallCount: make(map[string]int),
		defaultTimeout:  60 * time.Second,
	}
}

// InitiateCallSipgo starts a new SIP call using sipgo
func (scm *SIPCallManagerSipgo) InitiateCallSipgo(callID, fromPhone, toPhone string) error {
	scm.logger.Infof("🚀 Initiating SIP call using sipgo: %s → %s (CallID: %s)", fromPhone, toPhone, callID)

	// =========================================================================
	// 🚫 CHECK IF CALL ALREADY EXISTS - PREVENT DUPLICATES
	// =========================================================================
	if existingCall, exists := scm.activeCalls[callID]; exists {
		scm.logger.Warnf("⚠️ DUPLICATE CALL PREVENTION: CallID %s already exists!", callID)
		scm.logger.Infof("📞 Existing call: From=%s, To=%s, State=%d",
			existingCall.FromPhone, existingCall.ToPhone, existingCall.State)
		scm.logger.Infof("📞 New call request: From=%s, To=%s", fromPhone, toPhone)

		// If it's the exact same call parameters, just return success
		if existingCall.FromPhone == fromPhone && existingCall.ToPhone == toPhone {
			scm.logger.Infof("✅ Call with same parameters already exists, skipping duplicate")
			return nil
		} else {
			scm.logger.Errorf("❌ CallID conflict: different call parameters for same CallID")
			return fmt.Errorf("CallID %s already exists with different parameters", callID)
		}
	}

	// Ensure network is configured
	if !scm.networkManager.IsConfigured() {
		scm.logger.Infof("🌐 Network not configured, setting up...")
		if err := scm.networkManager.ConfigureNetwork(); err != nil {
			return fmt.Errorf("failed to configure network: %v", err)
		}
	}

	// Create call context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), scm.defaultTimeout)

	// Create call info
	callInfo := &CallInfo{
		CallID:     callID,
		FromPhone:  fromPhone,
		ToPhone:    toPhone,
		State:      CallStateInitiated,
		StartTime:  time.Now(),
		LastUpdate: time.Now(),
		Context:    ctx,
		CancelFunc: cancel,
	}

	// Register the call (we already checked it doesn't exist)
	scm.activeCalls[callID] = callInfo

	// Create custom headers with SIP tag management
	headers := make([]sip.Header, 0)
	headers = SetCallIDHeader(headers, callID)
	headers = scm.SetViaHeader(headers)
	headers = scm.SetFromHeader(headers, fromPhone, callID)

	// Fix Contact header with correct IP from network manager
	localIP := scm.networkManager.GetLocalIP()
	localPort := scm.networkManager.GetLocalPort()
	headers = append(headers, &sip.ContactHeader{Address: sip.Uri{User: fromPhone, Host: localIP, Port: localPort}})
	// SDP body requires Content-Type.
	headers = append(headers, sip.NewHeader("Content-Type", "application/sdp"))

	// Reserve a local RTP port BEFORE generating SDP so SDP always matches a bound socket.
	if port, conn, err := scm.allocateLocalRTPListener(callID); err == nil {
		callInfo.LocalRTPPort = port
		callInfo.LocalRTPConn = conn
		scm.startLocalRTPMonitor(callInfo)
	} else {
		scm.logger.Errorf("🎵 [RTP-MONITOR] Failed to reserve local RTP port (CallID=%s): %v", callID, err)
	}

	recipient := scm.GetRecipient(toPhone)
	sdpBody := scm.CreateSDPOffer(fromPhone, callInfo.LocalRTPPort)
	if strings.TrimSpace(os.Getenv("SIPPROXY_LOG_SDP_OFFER")) == "1" {
		sdpOneLine := strings.ReplaceAll(strings.ReplaceAll(sdpBody, "\r", ""), "\n", "\\n")
		scm.logger.Infof("🎵 [SIP-SDP-OFFER] INVITE SDP offer (CallID=%s): %s", callID, sdpOneLine)
	}
	if callInfo.LocalRTPPort > 0 {
		scm.logger.Infof("🎵 [SDP] Local RTP port advertised: %d (CallID=%s)", callInfo.LocalRTPPort, callID)
	}

	// Send INVITE using sipgo Dialog API with SDP body and custom From header
	dialogSession, err := scm.dialogUA.Invite(ctx, recipient, []byte(sdpBody), headers...)
	if err != nil {
		scm.cleanupCall(callID)
		return fmt.Errorf("failed to send INVITE with sipgo: %v", err)
	}

	// Access the INVITE request headers after creation
	inviteReq := dialogSession.InviteRequest

	// Log unified Call-ID tracking
	scm.logger.Infof("🔄 Unified Call-ID Tracking:")
	scm.logger.Infof("   🆔 Call-ID: %s (same for WhatsApp and SIP)", callID)
	scm.logger.Infof("   ✅ Simplified tracking - no mapping needed!")
	scm.logger.Infof("   � Actual SIP Call-ID: %s", inviteReq.CallID().String())

	// Log the complete SIP INVITE request
	scm.logger.Infof("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	scm.logger.Infof("📨 SIPGO GENERATED SIP INVITE:")
	scm.logger.Infof("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	scm.logger.Infof("🎯 Request-URI: %s", inviteReq.Recipient.String())
	scm.logger.Infof("📞 Call-ID: %s", inviteReq.CallID().String())
	scm.logger.Infof("📤 From: %s", inviteReq.From().String())
	scm.logger.Infof("📥 To: %s", inviteReq.To().String())
	scm.logger.Infof("📞 Contact: %s", inviteReq.Contact().String())

	// Try to access Via header for NAT information
	if viaHeader := inviteReq.Via(); viaHeader != nil {
		scm.logger.Infof("🌐 Via: %s", viaHeader.String())
	}

	// Try to access User-Agent header
	if userAgent := inviteReq.GetHeader("User-Agent"); userAgent != nil {
		scm.logger.Infof("🏷️  User-Agent: %s", userAgent.String())
	}
	scm.logger.Infof("📄 Complete message:")
	for i, line := range strings.Split(inviteReq.String(), "\n") {
		scm.logger.Infof("📄 Line %d: %s", i+1, line)
	}
	scm.logger.Infof("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Update call state
	scm.updateCallState(callID, CallStateInviting)

	// Store the dialog session for future BYE/CANCEL operations
	callInfo.DialogSession = dialogSession

	// Start monitoring the dialog session responses
	go scm.monitorSipgoDialog(callInfo, dialogSession)

	scm.logger.Infof("✅ SIP INVITE sent using sipgo DialogUA, CallID: %s", callID)

	return nil
}

// updateCallState updates the state of a call
func (scm *SIPCallManagerSipgo) updateCallState(callID string, state CallState) {
	if callInfo, exists := scm.activeCalls[callID]; exists {
		callInfo.State = state
		callInfo.LastUpdate = time.Now()
		scm.logger.Infof("🔄 Call %s state updated to: %v", callID, state)
	}
}

// cleanupCall removes a call from active calls
func (scm *SIPCallManagerSipgo) cleanupCall(callID string) {
	if callInfo, exists := scm.activeCalls[callID]; exists {
		if callInfo.LocalRTPConn != nil {
			_ = callInfo.LocalRTPConn.Close()
			callInfo.LocalRTPConn = nil
		}
		if callInfo.CancelFunc != nil {
			callInfo.CancelFunc()
		}
		delete(scm.activeCalls, callID)
		scm.logger.Infof("🧹 Call %s cleaned up", callID)
	}

	// Also clean up the cancel call counter
	scm.resetCancelCallCount(callID)
}

// monitorSipgoDialog monitors a sipgo dialog session for responses
func (scm *SIPCallManagerSipgo) monitorSipgoDialog(callInfo *CallInfo, dialogSession *sipgo.DialogClientSession) {
	scm.logger.Infof("👂 Starting sipgo dialog monitoring for CallID: %s", callInfo.CallID)

	// Wait for responses using sipgo Dialog API with detailed error logging
	scm.logger.Infof("⏳ Waiting for SIP response from server for CallID: %s", callInfo.CallID)
	scm.logger.Infof("🔍 Dialog session state before wait...")
	scm.logger.Infof("📞 MONITORING: Calling WaitAnswer() to wait for 200 OK response...")

	err := dialogSession.WaitAnswer(callInfo.Context, sipgo.AnswerOptions{})

	scm.logger.Infof("📞 MONITORING: WaitAnswer() completed for CallID: %s", callInfo.CallID)
	if err == nil {
		scm.logger.Infof("📞 MONITORING: ✅ NO ERROR - This means we got 200 OK!")
	} else {
		scm.logger.Infof("📞 MONITORING: ❌ ERROR OCCURRED - %v", err)
	}

	if err != nil {
		scm.logger.Errorf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		scm.logger.Errorf("❌ SIP RESPONSE ERROR for CallID %s:", callInfo.CallID)
		scm.logger.Errorf("   🚨 Error details: %v", err)
		scm.logger.Errorf("   🚨 Error type: %T", err)
		scm.logger.Errorf("   🚨 Error string: %s", err.Error())

		// Check different types of errors that might contain response info
		switch e := err.(type) {
		case *sipgo.ErrDialogResponse:
			scm.logger.Errorf("   📨 Dialog Response Error detected")
			scm.logger.Errorf("   📨 Error message: %s", e.Error())
		default:
			scm.logger.Errorf("   📨 Generic error type")
		}

		// Try to parse error message for status codes
		errorMsg := err.Error()
		scm.logger.Errorf("   🔍 Parsing error message for SIP status codes...")
		if strings.Contains(errorMsg, "SIP/2.0") {
			scm.logger.Errorf("   📨 SIP Response found in error: %s", errorMsg)

			// Extract status code from error message and trigger WhatsApp rejection
			if strings.Contains(errorMsg, "603") {
				scm.logger.Errorf("   📨 ⚠️  STATUS CODE 603 DECLINED detected!")
				scm.logger.Errorf("   📨 ⚠️  This means the SIP server REJECTED the call")
				scm.logger.Errorf("   📨 🔄 Triggering WhatsApp call rejection...")

				// Call the rejection handler to reject the WhatsApp call
				if scm.onCallRejected != nil {
					scm.logger.Infof("📞❌ Calling WhatsApp rejection handler for CallID: %s", callInfo.CallID)
					scm.onCallRejected(callInfo.CallID, callInfo.FromPhone, callInfo.ToPhone, nil)
				} else {
					scm.logger.Errorf("❌ No rejection handler configured! Cannot reject WhatsApp call!")
				}
			} else if strings.Contains(errorMsg, "486") {
				scm.logger.Errorf("   📨 ⚠️  STATUS CODE 486 BUSY HERE detected!")
				scm.logger.Errorf("   📨 ⚠️  This means the SIP endpoint is busy")

				// Also trigger rejection for busy
				if scm.onCallRejected != nil {
					scm.logger.Infof("📞❌ Calling WhatsApp rejection handler for BUSY CallID: %s", callInfo.CallID)
					scm.onCallRejected(callInfo.CallID, callInfo.FromPhone, callInfo.ToPhone, nil)
				}
			} else if strings.Contains(errorMsg, "404") {
				scm.logger.Errorf("   📨 ⚠️  STATUS CODE 404 NOT FOUND detected!")
				scm.logger.Errorf("   📨 ⚠️  This means the SIP endpoint was not found")

				// Also trigger rejection for not found
				if scm.onCallRejected != nil {
					scm.logger.Infof("📞❌ Calling WhatsApp rejection handler for NOT FOUND CallID: %s", callInfo.CallID)
					scm.onCallRejected(callInfo.CallID, callInfo.FromPhone, callInfo.ToPhone, nil)
				}
			} else if strings.Contains(errorMsg, "408") {
				scm.logger.Errorf("   📨 ⚠️  STATUS CODE 408 TIMEOUT detected!")
				scm.logger.Errorf("   📨 ⚠️  This means the SIP request timed out")
			}
		}
		scm.logger.Errorf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		scm.updateCallState(callInfo.CallID, CallStateRejected)
		scm.cleanupCall(callInfo.CallID)
		return
	}

	// If we reach here, the call was answered successfully
	scm.logger.Infof("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	scm.logger.Infof("✅ SIP RESPONSE SUCCESS for CallID: %s", callInfo.CallID)
	scm.logger.Infof("   ✅ Call answered successfully!")
	scm.logger.Infof("   ✅ Status: 200 OK received from server")
	scm.logger.Infof("   ✅ This means the SIP server ACCEPTED the call")
	scm.logger.Infof("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Parse SDP answer (if any) to extract negotiated remote RTP endpoint.
	remoteRTPIP := ""
	remoteRTPPort := 0
	if dialogSession.InviteResponse != nil {
		body := dialogSession.InviteResponse.Body()
		remoteRTPIP, remoteRTPPort = parseSDPAudioEndpoint(body)
		if remoteRTPIP != "" || remoteRTPPort > 0 {
			scm.logger.Infof("🎵 [SIP-SDP] 200 OK remote media endpoint: %s:%d (CallID=%s)", remoteRTPIP, remoteRTPPort, callInfo.CallID)
		}
		if len(body) > 0 && strings.TrimSpace(os.Getenv("SIPPROXY_LOG_SDP_200OK")) == "1" {
			sdpOneLine := strings.ReplaceAll(strings.ReplaceAll(string(body), "\r", ""), "\n", "\\n")
			scm.logger.Infof("🎵 [SIP-SDP-200OK] 200 OK SDP (CallID=%s): %s", callInfo.CallID, sdpOneLine)
		}
		scm.logger.Infof("🎵 [SIP-RESP] source=%s dest=%s transport=%s (CallID=%s)", dialogSession.InviteResponse.Source(), dialogSession.InviteResponse.Destination(), dialogSession.InviteResponse.Transport(), callInfo.CallID)
	}

	scm.updateCallState(callInfo.CallID, CallStateAccepted)

	// 🎉 TRIGGER WHATSAPP CALL ACCEPTANCE!
	if scm.onCallAccepted != nil {
		scm.logger.Infof("📞✅ Calling WhatsApp acceptance handler for CallID: %s", callInfo.CallID)
		scm.onCallAccepted(callInfo.CallID, callInfo.FromPhone, callInfo.ToPhone, nil)

		// ✅ Keeping callback active for subsequent calls (removed loop prevention)
		scm.logger.Infof("✅ Callback executed successfully, keeping handler active for future calls")
	} else {
		scm.logger.Errorf("❌ No acceptance handler configured! Cannot accept WhatsApp call!")
	}

	// Send ACK
	err = dialogSession.Ack(callInfo.Context)
	if err != nil {
		scm.logger.Errorf("❌ Failed to send ACK for CallID %s: %v", callInfo.CallID, err)
	} else {
		scm.logger.Infof("📨 ACK sent for CallID: %s", callInfo.CallID)
	}

	// Optional RTP probe to help Asterisk Strict RTP / NAT learning.
	if os.Getenv("SIPPROXY_RTP_PROBE") == "1" {
		if callInfo.LocalRTPConn != nil && remoteRTPIP != "" && remoteRTPPort > 0 {
			// Primary target from SDP
			sendRTPProbe(scm.logger, callInfo.LocalRTPConn, remoteRTPIP, remoteRTPPort, callInfo.CallID)
			// If SDP uses loopback, also try local/public IP to avoid any routing oddities.
			if remoteRTPIP == "127.0.0.1" {
				alt1 := strings.TrimSpace(scm.networkManager.GetLocalIP())
				alt2 := strings.TrimSpace(scm.networkManager.GetPublicIP())
				if alt1 != "" && alt1 != remoteRTPIP {
					sendRTPProbe(scm.logger, callInfo.LocalRTPConn, alt1, remoteRTPPort, callInfo.CallID)
				}
				if alt2 != "" && alt2 != remoteRTPIP && alt2 != alt1 {
					sendRTPProbe(scm.logger, callInfo.LocalRTPConn, alt2, remoteRTPPort, callInfo.CallID)
				}
			}
		} else {
			scm.logger.Infof("🎵🧪 [RTP-PROBE] Skipped (missing socket or remote endpoint) localConn=%v remote=%s:%d (CallID=%s)", callInfo.LocalRTPConn != nil, remoteRTPIP, remoteRTPPort, callInfo.CallID)
		}
	}

	// Keep the call active - do not automatically hang up
	// The call will be managed by the SIP server and WhatsApp events
	scm.logger.Infof("📞 Call is now active and ready for media flow - CallID: %s", callInfo.CallID)

	// 🚨 CRITICAL: Stop monitoring after successful accept to prevent callback loops
	scm.logger.Infof("🛑 Stopping dialog monitoring after successful acceptance - CallID: %s", callInfo.CallID)
	return
}

// GetActiveCalls returns the list of active call IDs
func (scm *SIPCallManagerSipgo) GetActiveCalls() []string {
	callIDs := make([]string, 0, len(scm.activeCalls))
	for callID := range scm.activeCalls {
		callIDs = append(callIDs, callID)
	}
	return callIDs
}

// CancelCall cancels an active call by sending BYE/CANCEL to SIP server
func (scm *SIPCallManagerSipgo) CancelCall(callID string) error {
	callCount := scm.getCancelCallCount(callID)
	scm.logger.Infof("🔥🔥🔥 [CANCEL-CALL-ENTRY] CallID: %s - Entry #%d", callID, callCount)

	// Get call info
	callInfo, exists := scm.activeCalls[callID]
	if !exists {
		scm.logger.Warnf("📞⚠️ Call %s not found for cancellation - may have been removed already", callID)
		return nil // Not an error, call was already cleaned up
	}

	// =========================================================================
	// 🚫 SEND ACTUAL SIP BYE/CANCEL TO SERVER
	// =========================================================================
	if callInfo.DialogSession != nil {
		scm.logger.Infof("📞🚫 [BYE-SEND] Sending SIP BYE to server for call: %s (attempt #%d)", callID, callCount)

		// Create context for BYE request
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// =========================================================================
		// 🔍 WRAPPED BYE CALL WITH DETAILED LOGGING
		// =========================================================================
		scm.logger.Infof("🔥 [BYE-CALL] About to call DialogSession.Bye() for CallID: %s", callID)

		// Send BYE request to terminate the SIP session
		err := callInfo.DialogSession.Bye(ctx)

		scm.logger.Infof("🔥 [BYE-RETURN] DialogSession.Bye() returned for CallID: %s, err: %v", callID, err)

		if err != nil {
			scm.logger.Errorf("❌ Failed to send SIP BYE for call %s: %v", callID, err)
			// Continue with cleanup even if BYE fails
		} else {
			scm.logger.Infof("✅ [BYE-SUCCESS] SIP BYE sent successfully for call: %s (attempt #%d)", callID, callCount)
		}
	} else {
		scm.logger.Warnf("⚠️ No DialogSession available for call %s - cannot send SIP BYE", callID)
	}

	// Clean up local call data
	scm.cleanupCall(callID)
	scm.logger.Infof("🧹 Call %s cleaned up", callID)

	return nil
}

// SetCallRejectedHandler configures the callback for when SIP calls are rejected
func (scm *SIPCallManagerSipgo) SetCallRejectedHandler(handler SIPCallRejectedCallback) {
	scm.onCallRejected = handler
	scm.logger.Infof("📞❌ Call rejection handler configured for sipgo manager")
}

// SetCallAcceptedHandler configures the callback for when SIP calls are accepted
func (scm *SIPCallManagerSipgo) SetCallAcceptedHandler(handler SIPCallAcceptedCallback) {
	scm.onCallAccepted = handler
	scm.logger.Infof("📞✅ Call acceptance handler configured for sipgo manager")
}

// =========================================================================
// 🔢 CANCEL CALL COUNTER METHODS (for debugging multiple BYEs)
// =========================================================================

// getCancelCallCount gets and increments the cancel call counter for a CallID
func (scm *SIPCallManagerSipgo) getCancelCallCount(callID string) int {
	scm.cancelMutex.Lock()
	defer scm.cancelMutex.Unlock()

	scm.cancelCallCount[callID]++
	return scm.cancelCallCount[callID]
}

// resetCancelCallCount resets the counter for a CallID
func (scm *SIPCallManagerSipgo) resetCancelCallCount(callID string) {
	scm.cancelMutex.Lock()
	defer scm.cancelMutex.Unlock()

	delete(scm.cancelCallCount, callID)
}
