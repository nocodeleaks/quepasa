package sipproxy

import (
	"context"
	"fmt"
	"strings"
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
	CallID     string
	FromPhone  string
	ToPhone    string
	SIPTag     string // Generated SIP tag for this call
	State      CallState
	StartTime  time.Time
	LastUpdate time.Time
	Context    context.Context
	CancelFunc context.CancelFunc
}

// SIPCallManagerSipgo manages SIP call lifecycle using sipgo package
type SIPCallManagerSipgo struct {
	logger         *log.Entry
	config         SIPProxySettings
	networkManager *SIPProxyNetworkManager
	sipClient      *sipgo.Client
	userAgent      *sipgo.UserAgent
	dialogUA       *sipgo.DialogUA
	activeCalls    map[string]*CallInfo
	defaultTimeout time.Duration
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
	logger.Infof("   🏷️  UserAgent: %s", userAgentName)

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
		logger:         logger,
		config:         config,
		networkManager: networkManager,
		sipClient:      client,
		userAgent:      ua,
		dialogUA:       dialogUA,
		activeCalls:    make(map[string]*CallInfo),
		defaultTimeout: 60 * time.Second,
	}
}

// InitiateCallSipgo starts a new SIP call using sipgo
func (scm *SIPCallManagerSipgo) InitiateCallSipgo(callID, fromPhone, toPhone string) error {
	scm.logger.Infof("🚀 Initiating SIP call using sipgo: %s → %s (CallID: %s)", fromPhone, toPhone, callID)

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

	// Register the call
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

	recipient := scm.GetRecipient(toPhone)
	sdpBody := scm.CreateSDPOffer(fromPhone)

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
		if callInfo.CancelFunc != nil {
			callInfo.CancelFunc()
		}
		delete(scm.activeCalls, callID)
		scm.logger.Infof("🧹 Call %s cleaned up", callID)
	}
}

// monitorSipgoDialog monitors a sipgo dialog session for responses
func (scm *SIPCallManagerSipgo) monitorSipgoDialog(callInfo *CallInfo, dialogSession *sipgo.DialogClientSession) {
	scm.logger.Infof("👂 Starting sipgo dialog monitoring for CallID: %s", callInfo.CallID)

	// Wait for responses using sipgo Dialog API
	err := dialogSession.WaitAnswer(callInfo.Context, sipgo.AnswerOptions{})
	if err != nil {
		scm.logger.Errorf("❌ Dialog wait error for CallID %s: %v", callInfo.CallID, err)
		scm.updateCallState(callInfo.CallID, CallStateRejected)
		scm.cleanupCall(callInfo.CallID)
		return
	}

	// If we reach here, the call was answered successfully
	scm.logger.Infof("✅ Call answered successfully for CallID: %s", callInfo.CallID)
	scm.updateCallState(callInfo.CallID, CallStateAccepted)

	// Send ACK
	err = dialogSession.Ack(callInfo.Context)
	if err != nil {
		scm.logger.Errorf("❌ Failed to send ACK for CallID %s: %v", callInfo.CallID, err)
	} else {
		scm.logger.Infof("📨 ACK sent for CallID: %s", callInfo.CallID)
	}

	// Keep the call active - do not automatically hang up
	// The call will be managed by the SIP server and WhatsApp events
	scm.logger.Infof("📞 Call is now active and ready for media flow - CallID: %s", callInfo.CallID)
}

// GetActiveCalls returns the list of active call IDs
func (scm *SIPCallManagerSipgo) GetActiveCalls() []string {
	callIDs := make([]string, 0, len(scm.activeCalls))
	for callID := range scm.activeCalls {
		callIDs = append(callIDs, callID)
	}
	return callIDs
}

// CancelCall cancels an active call
func (scm *SIPCallManagerSipgo) CancelCall(callID string) error {
	scm.logger.Infof("❌ Canceling call: %s", callID)
	scm.cleanupCall(callID)
	return nil
}
