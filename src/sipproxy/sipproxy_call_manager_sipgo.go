package sipproxy

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/emiago/sipgo"
	"github.com/emiago/sipgo/sip"
	qplog "github.com/nocodeleaks/quepasa/qplog"
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
	logger          qplog.Logger
	config          SIPProxySettings
	networkManager  *SIPProxyNetworkManager
	sipClient       *sipgo.Client
	userAgent       *sipgo.UserAgent
	dialogUA        *sipgo.DialogUA
	activeCalls     map[string]*CallInfo
	callsMutex      sync.RWMutex   // Protect activeCalls and CallInfo.State across concurrent calls
	cancelCallCount map[string]int // Track how many times CancelCall is called per CallID
	cancelMutex     sync.RWMutex   // Protect the counter
	defaultTimeout  time.Duration
	// Per-call handlers map - supports multiple handlers per call ID
	callAcceptedHandlers map[string][]SIPCallAcceptedCallback
	callRejectedHandlers map[string][]SIPCallRejectedCallback
	// Global fallback handlers (for backward compatibility)
	onCallRejected SIPCallRejectedCallback // Callback for call rejection
	onCallAccepted SIPCallAcceptedCallback // Callback for call acceptance
	handlerMutex   sync.RWMutex            // Protect handler maps
	// localRTPPorts maps callID → the local UDP port the audio bridge listens
	// on for this call's SIP RTP. CreateSDPOffer advertises this exact port so
	// the SIP server sends its RTP to the socket the bridge actually reads.
	localRTPPorts sync.Map
	// remoteRTPAddrs maps callID → the SIP server's RTP address ("ip:port")
	// parsed from the 200 OK SDP answer. The audio bridge sends its WhatsApp→SIP
	// RTP here from the start, so the SIP server (which may itself wait for our
	// RTP before sending) doesn't deadlock waiting on us.
	remoteRTPAddrs sync.Map
}

// SetLocalRTPPort registers the local RTP port to advertise in the SDP offer
// for callID. Must be called before InitiateCallSipgo for that call.
func (scm *SIPCallManagerSipgo) SetLocalRTPPort(callID string, port int) {
	scm.localRTPPorts.Store(callID, port)
}

// GetRemoteRTPAddr returns the SIP server's RTP address ("ip:port") learned
// from the 200 OK SDP answer for callID, once available.
func (scm *SIPCallManagerSipgo) GetRemoteRTPAddr(callID string) (string, bool) {
	if v, ok := scm.remoteRTPAddrs.Load(callID); ok {
		return v.(string), true
	}
	return "", false
}

// parseSDPRTPAddr extracts the audio RTP address ("ip:port") from an SDP body
// using its c= connection line and m=audio media line.
func parseSDPRTPAddr(sdp string) (string, bool) {
	var ip string
	var port string
	for _, line := range strings.Split(sdp, "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.HasPrefix(line, "c=IN IP4 ") {
			ip = strings.TrimSpace(strings.TrimPrefix(line, "c=IN IP4 "))
		} else if strings.HasPrefix(line, "m=audio ") {
			f := strings.Fields(line)
			if len(f) >= 2 {
				port = f[1]
			}
		}
	}
	if ip != "" && port != "" {
		return ip + ":" + port, true
	}
	return "", false
}

// NewSIPCallManagerSipgo creates a new SIP call manager using sipgo
func NewSIPCallManagerSipgo(logger qplog.Logger, config SIPProxySettings, networkManager *SIPProxyNetworkManager) *SIPCallManagerSipgo {
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

	// The SIP socket MUST bind to a concrete local IP, not 0.0.0.0. When the
	// host has multiple interfaces/default routes, a 0.0.0.0 bind lets the
	// kernel pick the egress source IP by route — which can differ from the
	// address the SIP server's IP ACL expects (causing 403/603). pickBindIP
	// prefers the configured public IP when it's a local interface address so
	// the source IP stays consistent with the Via/Contact/SDP.
	bindIP := pickBindIP(localIP, publicIP)

	// LOG CONFIGURATION VALUES FOR DEBUGGING
	logger.Infof("🔍 SIPGO CONFIGURATION DEBUG:")
	logger.Infof("   🌐 ServerHost: %s", config.ServerHost)
	logger.Infof("   🌐 ServerPort: %d", config.ServerPort)
	logger.Infof("   🏠 LocalIP: %s", localIP)
	logger.Infof("   🏠 LocalPort: %d", localPort)
	logger.Infof("   🏠 PublicIP: %s", publicIP)
	logger.Infof("   🏷️ UserAgent: %s", userAgentName)

	// Initialize sipgo UserAgent with complete configuration
	ua, err := sipgo.NewUA(
		sipgo.WithUserAgent(userAgentName),
		sipgo.WithUserAgentHostname(bindIP),
	)
	if err != nil {
		logger.Errorf("❌ Failed to create sipgo UserAgent: %v", err)
		return nil
	}

	logger.Infof("✅ UserAgent configured: %s@%s:%d", userAgentName, bindIP, localPort)

	// Create sipgo Client with explicit listen address
	// Try to bind to the specific IP:port to ensure Via header is correct
	listenAddr := fmt.Sprintf("%s:%d", bindIP, localPort)
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

	logger.Infof("✅ SIPCallManagerSipgo initialized with sipgo client")

	return &SIPCallManagerSipgo{
		logger:               logger,
		config:               config,
		networkManager:       networkManager,
		sipClient:            client,
		userAgent:            ua,
		dialogUA:             dialogUA,
		activeCalls:          make(map[string]*CallInfo),
		cancelCallCount:      make(map[string]int),
		defaultTimeout:       60 * time.Second,
		callAcceptedHandlers: make(map[string][]SIPCallAcceptedCallback),
		callRejectedHandlers: make(map[string][]SIPCallRejectedCallback),
	}
}

// InitiateCallSipgo starts a new SIP call using sipgo
func (scm *SIPCallManagerSipgo) InitiateCallSipgo(callID, fromPhone, toPhone string) error {
	return scm.InitiateCallSipgoWithHeaders(callID, fromPhone, toPhone, nil)
}

// InitiateCallSipgoWithHeaders starts a new SIP call using sipgo and attaches
// additional SIP headers to the outbound INVITE.
func (scm *SIPCallManagerSipgo) InitiateCallSipgoWithHeaders(callID, fromPhone, toPhone string, extraHeaders map[string]string) error {
	scm.logger.Infof("🚀 Initiating SIP call using sipgo: %s → %s (CallID: %s)", fromPhone, toPhone, callID)

	// =========================================================================
	// 🚫 CHECK IF CALL ALREADY EXISTS - PREVENT DUPLICATES
	// =========================================================================
	scm.callsMutex.RLock()
	existingCall, exists := scm.activeCalls[callID]
	var exFrom, exTo string
	if exists {
		exFrom, exTo = existingCall.FromPhone, existingCall.ToPhone
	}
	scm.callsMutex.RUnlock()
	if exists {
		scm.logger.Warnf("⚠️ DUPLICATE CALL PREVENTION: CallID %s already exists (From=%s To=%s)", callID, exFrom, exTo)

		// If it's the exact same call parameters, just return success
		if exFrom == fromPhone && exTo == toPhone {
			scm.logger.Infof("✅ Call with same parameters already exists, skipping duplicate")
			return nil
		}
		scm.logger.Errorf("❌ CallID conflict: different call parameters for same CallID")
		return fmt.Errorf("CallID %s already exists with different parameters", callID)
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

	// Register the call atomically; re-check under the write lock in case another
	// goroutine raced in with the same CallID between the check above and here.
	scm.callsMutex.Lock()
	if _, dup := scm.activeCalls[callID]; dup {
		scm.callsMutex.Unlock()
		cancel()
		scm.logger.Infof("✅ Call %s registered concurrently, skipping duplicate", callID)
		return nil
	}
	scm.activeCalls[callID] = callInfo
	scm.callsMutex.Unlock()

	// Create custom headers with SIP tag management
	headers := make([]sip.Header, 0)
	headers = SetCallIDHeader(headers, callID)
	headers = scm.SetViaHeader(headers)
	headers = scm.SetFromHeader(headers, fromPhone, callID)

	// Fix Contact header with correct IP from network manager
	localIP := scm.networkManager.GetLocalIP()
	localPort := scm.networkManager.GetLocalPort()
	headers = append(headers, &sip.ContactHeader{Address: sip.Uri{Scheme: "sip", User: fromPhone, Host: localIP, Port: localPort}})

	for name, value := range extraHeaders {
		name = strings.TrimSpace(name)
		value = strings.TrimSpace(value)
		if name == "" || value == "" {
			continue
		}
		header := sip.NewHeader(name, value)
		if header != nil {
			headers = append(headers, header)
		}
	}

	// Content-Type MUST be application/sdp so the SIP server treats the body as
	// the SDP offer. Without it, asterisk ignores the offer, puts its own offer
	// in the 200 OK, and then aborts the dialog with "incomplete SDP
	// negotiation" when our ACK carries no answer — an immediate BYE after ACK.
	ctype := sip.ContentTypeHeader("application/sdp")
	headers = append(headers, &ctype)

	recipient := scm.GetRecipient(toPhone)
	sdpBody, err := scm.CreateSDPOffer(callID, fromPhone)
	if err != nil {
		scm.cleanupCall(callID)
		return fmt.Errorf("failed to build SDP offer: %v", err)
	}

	// Send INVITE using sipgo Dialog API with SDP body and custom From header
	dialogSession, err := scm.dialogUA.Invite(ctx, recipient, []byte(sdpBody), headers...)
	if err != nil {
		scm.cleanupCall(callID)
		return fmt.Errorf("failed to send INVITE with sipgo: %v", err)
	}

	// Access the INVITE request headers after creation
	inviteReq := dialogSession.InviteRequest

	// Concise per-call line at info level. The WhatsApp Call-ID is reused as the
	// SIP Call-ID, so a single identifier tracks both legs.
	scm.logger.Infof("📞 SIP INVITE → %s (Call-ID: %s)", inviteReq.Recipient.String(), callID)

	// The full INVITE (all headers + SDP) is verbose; emit it once at debug
	// level only, as a single message, instead of one info line per header/line.
	scm.logger.Debugf("SIPGO generated SIP INVITE:\n%s", inviteReq.String())

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
	scm.callsMutex.Lock()
	callInfo, exists := scm.activeCalls[callID]
	if exists {
		callInfo.State = state
		callInfo.LastUpdate = time.Now()
	}
	scm.callsMutex.Unlock()
	if exists {
		scm.logger.Infof("🔄 Call %s state updated to: %v", callID, state)
	}
}

// cleanupCall removes a call from active calls
func (scm *SIPCallManagerSipgo) cleanupCall(callID string) {
	scm.callsMutex.Lock()
	callInfo, exists := scm.activeCalls[callID]
	if exists {
		if callInfo.CancelFunc != nil {
			callInfo.CancelFunc()
		}
		delete(scm.activeCalls, callID)
	}
	scm.callsMutex.Unlock()
	if exists {
		scm.logger.Infof("🧹 Call %s cleaned up", callID)
	}

	// Also clean up the cancel call counter
	scm.resetCancelCallCount(callID)

	// Release the registered local RTP port and remote RTP address for this call.
	scm.localRTPPorts.Delete(callID)
	scm.remoteRTPAddrs.Delete(callID)
}

// monitorSipgoDialog monitors a sipgo dialog session for responses
func (scm *SIPCallManagerSipgo) monitorSipgoDialog(callInfo *CallInfo, dialogSession *sipgo.DialogClientSession) {
	scm.logger.Debugf("dialog monitor: waiting for SIP answer (Call-ID: %s)", callInfo.CallID)

	err := dialogSession.WaitAnswer(callInfo.Context, sipgo.AnswerOptions{})

	if err != nil {
		errorMsg := err.Error()
		scm.logger.Debugf("dialog monitor: WaitAnswer error for Call-ID %s: %v (type %T)", callInfo.CallID, err, err)

		// Map the SIP status carried in the error to a concise log line, and
		// reject the WhatsApp leg for the failure codes that warrant it.
		reject := false
		switch {
		case strings.Contains(errorMsg, "603"):
			scm.logger.Warnf("SIP rejected: 603 Declined (Call-ID: %s)", callInfo.CallID)
			reject = true
		case strings.Contains(errorMsg, "486"):
			scm.logger.Warnf("SIP rejected: 486 Busy Here (Call-ID: %s)", callInfo.CallID)
			reject = true
		case strings.Contains(errorMsg, "404"):
			scm.logger.Warnf("SIP rejected: 404 Not Found (Call-ID: %s)", callInfo.CallID)
			reject = true
		case strings.Contains(errorMsg, "408"):
			scm.logger.Warnf("SIP failed: 408 Timeout (Call-ID: %s)", callInfo.CallID)
		default:
			scm.logger.Errorf("SIP response error (Call-ID: %s): %v", callInfo.CallID, err)
		}

		if reject {
			// Per-call rejection handlers first; fall back to the global handler.
			if !scm.invokeCallRejectedHandlers(callInfo.CallID, callInfo.FromPhone, callInfo.ToPhone, nil) {
				if scm.onCallRejected != nil {
					scm.onCallRejected(callInfo.CallID, callInfo.FromPhone, callInfo.ToPhone, nil)
				} else {
					scm.logger.Errorf("no rejection handler configured; cannot reject WhatsApp call (Call-ID: %s)", callInfo.CallID)
				}
			}
		}

		scm.updateCallState(callInfo.CallID, CallStateRejected)
		scm.cleanupCall(callInfo.CallID)
		return
	}

	// 200 OK — the SIP server accepted the call.
	scm.logger.Infof("✅ SIP 200 OK accepted (Call-ID: %s)", callInfo.CallID)
	scm.updateCallState(callInfo.CallID, CallStateAccepted)

	// Publish the SIP server's RTP address from the 200 OK SDP answer so the
	// audio bridge can start sending WhatsApp→SIP RTP immediately (breaking the
	// mutual "wait for the other side's RTP" deadlock).
	if dialogSession.InviteResponse != nil {
		if addr, ok := parseSDPRTPAddr(string(dialogSession.InviteResponse.Body())); ok {
			scm.remoteRTPAddrs.Store(callInfo.CallID, addr)
			scm.logger.Debugf("remote SIP RTP address from 200 OK SDP: %s (Call-ID: %s)", addr, callInfo.CallID)
		} else {
			scm.logger.Warnf("could not parse RTP address from 200 OK SDP (Call-ID: %s)", callInfo.CallID)
		}
	}

	// Trigger WhatsApp acceptance: per-call handlers first, global fallback otherwise.
	scm.handlerMutex.RLock()
	perCallHandlers := scm.callAcceptedHandlers[callInfo.CallID]
	scm.handlerMutex.RUnlock()

	handlerCalled := false
	for _, handler := range perCallHandlers {
		handler(callInfo.CallID, callInfo.FromPhone, callInfo.ToPhone, nil)
		handlerCalled = true
	}
	if !handlerCalled && scm.onCallAccepted != nil {
		scm.onCallAccepted(callInfo.CallID, callInfo.FromPhone, callInfo.ToPhone, nil)
	} else if !handlerCalled {
		scm.logger.Errorf("no acceptance handler configured (Call-ID: %s); cannot accept WhatsApp call", callInfo.CallID)
	}

	// Send ACK.
	if err = dialogSession.Ack(callInfo.Context); err != nil {
		scm.logger.Errorf("failed to send ACK (Call-ID: %s): %v", callInfo.CallID, err)
	}

	// Call stays active; managed by SIP server and WhatsApp events. Stop
	// monitoring here to avoid callback loops.
	scm.logger.Infof("📞 SIP call active (Call-ID: %s)", callInfo.CallID)
}

// GetActiveCalls returns the list of active call IDs
func (scm *SIPCallManagerSipgo) GetActiveCalls() []string {
	scm.callsMutex.RLock()
	defer scm.callsMutex.RUnlock()
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
	scm.callsMutex.RLock()
	callInfo, exists := scm.activeCalls[callID]
	scm.callsMutex.RUnlock()
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

// SetCallRejectedHandler configures the GLOBAL callback for when SIP calls are rejected
// This is a fallback handler for backward compatibility
func (scm *SIPCallManagerSipgo) SetCallRejectedHandler(handler SIPCallRejectedCallback) {
	scm.handlerMutex.Lock()
	defer scm.handlerMutex.Unlock()
	scm.onCallRejected = handler
	scm.logger.Infof("📞❌ GLOBAL call rejection handler configured for sipgo manager")
}

// SetCallAcceptedHandler configures the GLOBAL callback for when SIP calls are accepted
// This is a fallback handler for backward compatibility
func (scm *SIPCallManagerSipgo) SetCallAcceptedHandler(handler SIPCallAcceptedCallback) {
	scm.handlerMutex.Lock()
	defer scm.handlerMutex.Unlock()
	scm.onCallAccepted = handler
	scm.logger.Infof("📞✅ GLOBAL call acceptance handler configured for sipgo manager")
}

// RegisterCallAcceptedHandler registers a callback for a SPECIFIC call ID
// Multiple handlers can be registered for the same call ID
func (scm *SIPCallManagerSipgo) RegisterCallAcceptedHandler(callID string, handler SIPCallAcceptedCallback) {
	scm.handlerMutex.Lock()
	defer scm.handlerMutex.Unlock()
	if scm.callAcceptedHandlers[callID] == nil {
		scm.callAcceptedHandlers[callID] = make([]SIPCallAcceptedCallback, 0)
	}
	scm.callAcceptedHandlers[callID] = append(scm.callAcceptedHandlers[callID], handler)
	scm.logger.Infof("📞✅ Per-call acceptance handler registered for CallID: %s (total: %d)", callID, len(scm.callAcceptedHandlers[callID]))
}

// RegisterCallRejectedHandler registers a callback for a SPECIFIC call ID
// Multiple handlers can be registered for the same call ID
func (scm *SIPCallManagerSipgo) RegisterCallRejectedHandler(callID string, handler SIPCallRejectedCallback) {
	scm.handlerMutex.Lock()
	defer scm.handlerMutex.Unlock()
	if scm.callRejectedHandlers[callID] == nil {
		scm.callRejectedHandlers[callID] = make([]SIPCallRejectedCallback, 0)
	}
	scm.callRejectedHandlers[callID] = append(scm.callRejectedHandlers[callID], handler)
	scm.logger.Infof("📞❌ Per-call rejection handler registered for CallID: %s (total: %d)", callID, len(scm.callRejectedHandlers[callID]))
}

// ClearCallHandlers removes all per-call handlers for a specific call ID
// This is exported to allow the VoIP manager to clean up handlers after call completion
func (scm *SIPCallManagerSipgo) ClearCallHandlers(callID string) {
	scm.handlerMutex.Lock()
	defer scm.handlerMutex.Unlock()
	delete(scm.callAcceptedHandlers, callID)
	delete(scm.callRejectedHandlers, callID)
	scm.logger.Infof("📞🧹 Cleared handlers for CallID: %s", callID)
}

// invokeCallAcceptedHandlers invokes all per-call acceptance handlers for a call ID
// Returns true if at least one handler was called
func (scm *SIPCallManagerSipgo) invokeCallAcceptedHandlers(callID, fromPhone, toPhone string, resp *sip.Response) bool {
	scm.handlerMutex.RLock()
	perCallHandlers := scm.callAcceptedHandlers[callID]
	scm.handlerMutex.RUnlock()

	handlerCalled := false
	for _, handler := range perCallHandlers {
		scm.logger.Infof("📞✅ Calling per-call acceptance handler for CallID: %s", callID)
		handler(callID, fromPhone, toPhone, resp)
		handlerCalled = true
	}
	return handlerCalled
}

// invokeCallRejectedHandlers invokes all per-call rejection handlers for a call ID
// Returns true if at least one handler was called
func (scm *SIPCallManagerSipgo) invokeCallRejectedHandlers(callID, fromPhone, toPhone string, resp *sip.Response) bool {
	scm.handlerMutex.RLock()
	perCallHandlers := scm.callRejectedHandlers[callID]
	scm.handlerMutex.RUnlock()

	handlerCalled := false
	for _, handler := range perCallHandlers {
		scm.logger.Infof("📞❌ Calling per-call rejection handler for CallID: %s", callID)
		handler(callID, fromPhone, toPhone, resp)
		handlerCalled = true
	}
	return handlerCalled
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
