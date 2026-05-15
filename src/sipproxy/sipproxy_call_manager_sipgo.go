package sipproxy

import (
	"context"
	"fmt"
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
	State         CallState
	StartTime     time.Time
	LastUpdate    time.Time
	Context       context.Context
	CancelFunc    context.CancelFunc
	DialogSession *sipgo.DialogClientSession // SIP dialog session for BYE/CANCEL
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
	activeCallsMu   sync.RWMutex
	cancelCallCount map[string]int // Track how many times CancelCall is called per CallID
	cancelMutex     sync.RWMutex   // Protect the counter
	defaultTimeout  time.Duration
	onCallRejected  SIPCallRejectedCallback // Callback for call rejection
	onCallAccepted  SIPCallAcceptedCallback // Callback for call acceptance
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
	if existingCall, exists := scm.getActiveCall(callID); exists {
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
	scm.setActiveCall(callID, callInfo)

	// Create custom headers with SIP tag management
	headers := make([]sip.Header, 0)
	headers = SetCallIDHeader(headers, callID)
	headers = scm.SetViaHeader(headers)
	headers = scm.SetFromHeader(headers, fromPhone, callID)
	headers = scm.SetIdentityAndTraceHeaders(headers, fromPhone, toPhone, callID)
	contactUser := fromPhone
	if scm.config.FromUser != "" {
		contactUser = scm.config.FromUser
	}

	// Fix Contact header with correct IP from network manager
	localIP := scm.networkManager.GetLocalIP()
	localPort := scm.networkManager.GetLocalPort()
	headers = append(headers, &sip.ContactHeader{Address: sip.Uri{User: contactUser, Host: localIP, Port: localPort}})

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
	scm.logger.Debugf("📤 From: %s", inviteReq.From().String())
	scm.logger.Debugf("📥 To: %s", inviteReq.To().String())
	scm.logger.Debugf("📞 Contact: %s", inviteReq.Contact().String())

	// Try to access Via header for NAT information
	if viaHeader := inviteReq.Via(); viaHeader != nil {
		scm.logger.Infof("🌐 Via: %s", viaHeader.String())
	}

	// Try to access User-Agent header
	if userAgent := inviteReq.GetHeader("User-Agent"); userAgent != nil {
		scm.logger.Infof("🏷️  User-Agent: %s", userAgent.String())
	}
	if assertedID := inviteReq.GetHeader("P-Asserted-Identity"); assertedID != nil {
		scm.logger.Debugf("🪪 P-Asserted-Identity: %s", assertedID.String())
	}
	if remotePartyID := inviteReq.GetHeader("Remote-Party-ID"); remotePartyID != nil {
		scm.logger.Debugf("🪪 Remote-Party-ID: %s", remotePartyID.String())
	}
	if waCaller := inviteReq.GetHeader("X-QuePasa-WA-Caller"); waCaller != nil {
		scm.logger.Debugf("🔎 X-QuePasa-WA-Caller: %s", waCaller.String())
	}
	if waCalled := inviteReq.GetHeader("X-QuePasa-WA-Called"); waCalled != nil {
		scm.logger.Debugf("🔎 X-QuePasa-WA-Called: %s", waCalled.String())
	}
	if waSession := inviteReq.GetHeader("X-QuePasa-Session"); waSession != nil {
		scm.logger.Debugf("🔎 X-QuePasa-Session: %s", waSession.String())
	}
	if waCallID := inviteReq.GetHeader("X-QuePasa-CallID"); waCallID != nil {
		scm.logger.Debugf("🔎 X-QuePasa-CallID: %s", waCallID.String())
	}
	scm.logger.Debugf("📄 Complete message:")
	for i, line := range strings.Split(inviteReq.String(), "\n") {
		scm.logger.Debugf("📄 Line %d: %s", i+1, line)
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
	if callInfo, exists := scm.getActiveCall(callID); exists {
		callInfo.State = state
		callInfo.LastUpdate = time.Now()
		scm.logger.Infof("🔄 Call %s state updated to: %v", callID, state)
	}
}

// cleanupCall removes a call from active calls
func (scm *SIPCallManagerSipgo) cleanupCall(callID string) {
	if callInfo, exists := scm.getActiveCall(callID); exists {
		if callInfo.CancelFunc != nil {
			callInfo.CancelFunc()
		}
		scm.deleteActiveCall(callID)
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

	answerOptions := sipgo.AnswerOptions{}
	if scm.config.AuthPassword != "" {
		answerOptions.Username = scm.config.AuthUsername
		answerOptions.Password = scm.config.AuthPassword
		scm.logger.Infof("🔐 SIP digest auth enabled for INVITE answer flow (username=%s)", scm.config.AuthUsername)
	}

	err := dialogSession.WaitAnswer(callInfo.Context, answerOptions)

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

	// Keep the call active - do not automatically hang up
	// The call will be managed by the SIP server and WhatsApp events
	scm.logger.Infof("📞 Call is now active and ready for media flow - CallID: %s", callInfo.CallID)

	// 🚨 CRITICAL: Stop monitoring after successful accept to prevent callback loops
	scm.logger.Infof("🛑 Stopping dialog monitoring after successful acceptance - CallID: %s", callInfo.CallID)
	return
}

// GetActiveCalls returns the list of active call IDs
func (scm *SIPCallManagerSipgo) GetActiveCalls() []string {
	return scm.listActiveCallIDs()
}

// CancelCall cancels an active call by sending BYE/CANCEL to SIP server
func (scm *SIPCallManagerSipgo) CancelCall(callID string) error {
	callCount := scm.getCancelCallCount(callID)
	scm.logger.Infof("🔥🔥🔥 [CANCEL-CALL-ENTRY] CallID: %s - Entry #%d", callID, callCount)

	// Get call info
	callInfo, exists := scm.getActiveCall(callID)
	if !exists {
		scm.logger.Warnf("📞⚠️ Call %s not found for cancellation - may have been removed already", callID)
		return nil // Not an error, call was already cleaned up
	}

	if callInfo.State != CallStateAccepted {
		scm.logger.Infof("📞↩️ Call %s is not in accepted state (state=%v), skipping SIP BYE and cleaning up", callID, callInfo.State)
		scm.cleanupCall(callID)
		return nil
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

func (scm *SIPCallManagerSipgo) getActiveCall(callID string) (*CallInfo, bool) {
	scm.activeCallsMu.RLock()
	defer scm.activeCallsMu.RUnlock()

	callInfo, exists := scm.activeCalls[callID]
	return callInfo, exists
}

func (scm *SIPCallManagerSipgo) setActiveCall(callID string, callInfo *CallInfo) {
	scm.activeCallsMu.Lock()
	defer scm.activeCallsMu.Unlock()

	scm.activeCalls[callID] = callInfo
}

func (scm *SIPCallManagerSipgo) deleteActiveCall(callID string) {
	scm.activeCallsMu.Lock()
	defer scm.activeCallsMu.Unlock()

	delete(scm.activeCalls, callID)
}

func (scm *SIPCallManagerSipgo) listActiveCallIDs() []string {
	scm.activeCallsMu.RLock()
	defer scm.activeCallsMu.RUnlock()

	callIDs := make([]string, 0, len(scm.activeCalls))
	for callID := range scm.activeCalls {
		callIDs = append(callIDs, callID)
	}

	return callIDs
}
