package whatsmeow

import (
	"fmt"
	"sync"

	"github.com/emiago/sipgo/sip"
	sipproxy "github.com/nocodeleaks/quepasa/sipproxy"
	log "github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow/types"
)

// SIPProxyIntegration provides integration between WhatsApp calls and the singleton SIP proxy
type SIPProxyIntegration struct {
	logger             *log.Entry
	sipProxy           *sipproxy.SIPProxyManager
	connection         *WhatsmeowConnection // WhatsApp connection for rejecting calls
	processingCallsMap map[string]bool      // Track calls being processed to avoid loops
	processingMutex    sync.RWMutex         // Mutex to protect the processing map
}

// NewSIPProxyIntegration creates a new SIP proxy integration instance
func NewSIPProxyIntegration(logger *log.Entry) *SIPProxyIntegration {
	integration := &SIPProxyIntegration{
		logger:             logger.WithField("component", "sip_integration"),
		sipProxy:           sipproxy.SIPProxy,
		processingCallsMap: make(map[string]bool),
	}
	return integration
}

// InitializeSIPProxy sets up the singleton SIP proxy with server configuration
func (si *SIPProxyIntegration) InitializeSIPProxy(serverHost string, serverPort int, listenerPort int) error {
	if si.sipProxy == nil {
		return fmt.Errorf("SIP proxy singleton is nil - cannot initialize")
	}

	// The singleton already has default configuration, we just initialize it
	si.logger.Infof("🔧 Initializing SIP Proxy Singleton...")
	if err := si.sipProxy.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize SIP proxy: %v", err)
	}

	// SIP proxy starts automatically on Initialize()
	si.logger.Infof("✅ SIP Proxy Singleton initialized successfully")

	return nil
}

// ThrowSIPProxy forwards a WhatsApp call to the SIP server
func (si *SIPProxyIntegration) ThrowSIPProxy(callID, fromUser, toUser string) error {
	if si.sipProxy == nil {
		return fmt.Errorf("SIP proxy singleton is nil - cannot forward call")
	}

	if !si.sipProxy.IsRunning() {
		return fmt.Errorf("SIP proxy is not running")
	}

	si.logger.Infof("📞 Forwarding WhatsApp call to SIP server")
	si.logger.Infof("   📞 CallID: %s", callID)
	si.logger.Infof("   📞 From: %s", fromUser)
	si.logger.Infof("   📞 To: %s", toUser)

	si.logger.Infof("🚀 Sending SIP INVITE to server...")
	si.logger.Infof("🔍 DEBUG: Verificando parâmetros antes de enviar:")
	si.logger.Infof("   ✅ CallID correto: %s", callID)
	si.logger.Infof("   ✅ From correto: %s", fromUser)
	si.logger.Infof("   ✅ To correto: %s", toUser)

	// Send SIP INVITE - Corrected parameter order
	if err := si.sipProxy.SendSIPInvite(callID, fromUser, toUser); err != nil {
		si.logger.Errorf("❌ Failed to send SIP INVITE: %v", err)
		return err
	}

	si.logger.Infof("✅ SIP INVITE sent successfully")
	return nil
}

// HandleWhatsAppCallTermination notifies SIP proxy of call termination
func (si *SIPProxyIntegration) HandleWhatsAppCallTermination(callID string) error {
	if !si.sipProxy.IsRunning() {
		return fmt.Errorf("SIP proxy is not running")
	}

	// =========================================================================
	// 🚫 DUPLICATE TERMINATION PREVENTION
	// =========================================================================
	terminationKey := callID + "_terminating"
	si.logger.Infof("🔍🔍🔍 [TERMINATION-ENTRY] CallID: %s (key: %s)", callID, terminationKey)

	if si.isCallBeingProcessed(terminationKey) {
		si.logger.Warnf("⚠️⚠️⚠️ DUPLICATE TERMINATION PREVENTION: Call %s is already being terminated, skipping", callID)
		si.logger.Infof("🔍 Current processing calls: %+v", si.processingCallsMap)
		return nil
	}

	// Mark call as being terminated to prevent duplicates
	if !si.markCallAsProcessing(terminationKey) {
		si.logger.Warnf("⚠️ DUPLICATE TERMINATION PREVENTION: Failed to mark call %s as terminating", callID)
		return nil
	}

	si.logger.Infof("🔒🔒🔒 Call %s marked as terminating", callID)

	// Ensure we unmark the call when done
	defer func() {
		si.unmarkCallAsProcessing(terminationKey)
		si.logger.Infof("🔓🔓🔓 Call %s unmarked from terminating", callID)
	}()

	si.logger.Infof("📞❌ Handling WhatsApp call termination for CallID: %s", callID)

	// Send BYE/CANCEL to SIP server (this will also clean up locally)
	if err := si.sipProxy.CancelCall(callID); err != nil {
		si.logger.Errorf("❌ Failed to cancel SIP call %s: %v", callID, err)
		return err
	}

	si.logger.Infof("✅ SIP BYE/CANCEL sent successfully for CallID: %s", callID)
	return nil
}

// =========================================================================
// 🔄 LOOP PREVENTION METHODS
// =========================================================================

// markCallAsProcessing marks a call as being processed to avoid loops
func (si *SIPProxyIntegration) markCallAsProcessing(callID string) bool {
	si.processingMutex.Lock()
	defer si.processingMutex.Unlock()

	if si.processingCallsMap[callID] {
		return false // Already being processed
	}

	si.processingCallsMap[callID] = true
	si.logger.Infof("🔒 Call %s marked as processing", callID)
	return true
}

// unmarkCallAsProcessing removes a call from the processing list
func (si *SIPProxyIntegration) unmarkCallAsProcessing(callID string) {
	si.processingMutex.Lock()
	defer si.processingMutex.Unlock()

	delete(si.processingCallsMap, callID)
	si.logger.Infof("🔓 Call %s unmarked from processing", callID)
}

// isCallBeingProcessed checks if a call is currently being processed
func (si *SIPProxyIntegration) isCallBeingProcessed(callID string) bool {
	si.processingMutex.RLock()
	defer si.processingMutex.RUnlock()

	return si.processingCallsMap[callID]
}

// GetActiveCalls returns all active calls from the SIP proxy
func (si *SIPProxyIntegration) GetActiveCalls() map[string]*sipproxy.SIPProxyCallData {
	return si.sipProxy.GetActiveCalls()
}

// IsReady returns true if the SIP proxy is ready to handle calls
func (si *SIPProxyIntegration) IsReady() bool {
	if si.sipProxy == nil {
		return false
	}
	return si.sipProxy.IsRunning()
}

// GetStatus returns status information about the SIP proxy
func (si *SIPProxyIntegration) GetStatus() map[string]interface{} {
	status := make(map[string]interface{})

	if si.sipProxy == nil {
		status["running"] = false
		status["configured"] = false
		status["error"] = "SIP proxy singleton is nil"
		return status
	}

	status["running"] = si.sipProxy.IsRunning()
	status["configured"] = true // Singleton is always configured
	status["public_ip"] = si.sipProxy.GetPublicIP()

	activeCalls := si.sipProxy.GetActiveCalls()
	status["active_calls"] = len(activeCalls)

	return status
}

// SetupCallbacks configura callbacks para quando chamadas SIP são aceitas/rejeitadas
func (si *SIPProxyIntegration) SetupCallbacks(whatsappHandler interface{}) {
	si.logger.Infof("🎯 Setting up SIP call status callbacks")

	// Check if SIP proxy is available
	if si.sipProxy == nil {
		si.logger.Warnf("⚠️ SIP Proxy is nil - callbacks cannot be set up")
		return
	}

	// Store WhatsApp connection for rejecting calls
	if conn, ok := whatsappHandler.(*WhatsmeowConnection); ok {
		si.connection = conn
		si.logger.Infof("✅ WhatsApp connection stored for call events")
	} else {
		si.logger.Warnf("⚠️ WhatsApp handler is not a WhatsmeowConnection, call actions may not work")
	}

	// Callback para quando uma chamada é aceita (200 OK)
	si.sipProxy.SetCallAcceptedHandler(func(callID, fromPhone, toPhone string, response *sip.Response) {
		si.logger.Infof("✅ SIP CALL ACCEPTED! CallID: %s, From: %s, To: %s", callID, fromPhone, toPhone)

		// Safe response handling
		if response != nil {
			si.logger.Infof("📡 SIP Response: %d %s", response.StatusCode, response.Reason)
		} else {
			si.logger.Infof("📡 SIP Response: 200 OK (confirmed by sipgo dialog)")
		}

		// Aqui podemos notificar o WhatsApp que a chamada foi aceita
		si.onCallAccepted(callID, fromPhone, toPhone, response)
	})

	// Callback para quando uma chamada é rejeitada (>=400)
	si.sipProxy.SetCallRejectedHandler(func(callID, fromPhone, toPhone string, response *sip.Response) {
		si.logger.Infof("❌ SIP CALL REJECTED! CallID: %s, From: %s, To: %s", callID, fromPhone, toPhone)

		// Verificar se response não é nil antes de acessar
		if response != nil {
			si.logger.Infof("📡 SIP Response: %d %s", response.StatusCode, response.Reason)
		} else {
			si.logger.Infof("📡 SIP Response: (details not available)")
		}

		// Aqui podemos notificar o WhatsApp que a chamada foi rejeitada
		si.onCallRejected(callID, fromPhone, toPhone, response)
	})
}

// onCallAccepted é chamado quando uma chamada SIP é aceita
func (si *SIPProxyIntegration) onCallAccepted(callID, fromPhone, toPhone string, response *sip.Response) {
	// =========================================================================
	// 🚫 LOOP PREVENTION: Check if call is already being processed
	// =========================================================================
	if si.isCallBeingProcessed(callID) {
		si.logger.Warnf("⚠️ LOOP PREVENTION: Call %s is already being processed, skipping WhatsApp acceptance", callID)
		return
	}

	// Mark call as being processed to prevent loops
	if !si.markCallAsProcessing(callID) {
		si.logger.Warnf("⚠️ LOOP PREVENTION: Failed to mark call %s as processing (concurrent access?)", callID)
		return
	}

	// Ensure we unmark the call when done
	defer si.unmarkCallAsProcessing(callID)

	si.logger.Infof("🎉 CALL ACCEPTED EVENT - SIP server authorized the call!")
	si.logger.Infof("📞 CallID: %s", callID)
	si.logger.Infof("📞 From (caller): %s", fromPhone)
	si.logger.Infof("📞 To (receiver): %s", toPhone)

	// Safe response handling
	if response != nil {
		si.logger.Infof("📡 SIP Status: %d %s", response.StatusCode, response.Reason)
	} else {
		si.logger.Infof("📡 SIP Status: 200 OK (confirmed by sipgo dialog)")
	}

	// ✅ SIP authorized call - Call was already accepted in immediate acceptance
	si.logger.Infof("✅ SIP authorized call, call already handled by immediate acceptance")
	si.logger.Infof("� SKIPPING SECOND ACCEPT - Call already accepted immediately on CallOffer")
	si.logger.Infof("🎯 Reason: SIP server responded with 200 OK")
	si.logger.Infof("🔗 Bridge established between WhatsApp and SIP server")
}

// onCallRejected é chamado quando uma chamada SIP é rejeitada
func (si *SIPProxyIntegration) onCallRejected(callID, fromPhone, toPhone string, response *sip.Response) {
	// =========================================================================
	// 🚫 LOOP PREVENTION: Check if call is already being processed for rejection
	// =========================================================================
	rejectionKey := fmt.Sprintf("rejection_%s", callID)
	if si.isCallBeingProcessed(rejectionKey) {
		si.logger.Warnf("⚠️⚠️⚠️ LOOP PREVENTION: Call %s is already being processed for REJECTION, skipping WhatsApp rejection", callID)
		return
	}

	// Mark call as being processed to prevent loops
	if !si.markCallAsProcessing(rejectionKey) {
		si.logger.Warnf("⚠️⚠️⚠️ LOOP PREVENTION: Failed to mark call %s as processing rejection (concurrent access?)", callID)
		return
	}

	// Ensure we unmark the call when done
	defer si.unmarkCallAsProcessing(rejectionKey)

	si.logger.Infof("💔💔💔 [REJECTION-ENTRY] SIP server rejected the call!")
	si.logger.Infof("📞 CallID: %s", callID)
	si.logger.Infof("📞 From (caller): %s", fromPhone)
	si.logger.Infof("📞 To (receiver): %s", toPhone)

	// Verificar se response não é nil antes de acessar
	if response != nil {
		si.logger.Infof("📡 SIP Status: %d %s", response.StatusCode, response.Reason)
	} else {
		si.logger.Infof("📡 SIP Status: (details not available)")
	}

	// 🚫 AUTOMATICALLY REJECT THE CALL IN WHATSAPP TOO
	si.logger.Infof("🚫🚫🚫 [WHATSAPP-REJECTION] SIP rejected call, automatically rejecting in WhatsApp...")

	// Aqui precisamos acessar o WhatsApp connection para rejeitar a chamada
	// O fromPhone é quem está ligando, então vamos rejeitar a chamada vinda dele
	if si.connection != nil {
		// Converter número de telefone para JID WhatsApp
		fromJID, err := types.ParseJID(fromPhone + "@s.whatsapp.net")
		if err != nil {
			si.logger.Errorf("❌ Failed to parse fromPhone JID: %v", err)
			return
		}

		// Usar o CallManager para rejeitar a chamada no WhatsApp
		if callManager := si.connection.GetCallManager(); callManager != nil {
			si.logger.Infof("🔄🔄🔄 [WHATSAPP-CALL-REJECT] Attempting to reject WhatsApp call from %s...", fromPhone)
			err := callManager.RejectCall(fromJID, callID)
			if err != nil {
				si.logger.Errorf("❌❌❌ [REJECTION-ERROR] Failed to reject WhatsApp call: %v", err)
				// Tentar métodos alternativos de rejeição
				si.logger.Infof("🔄 Trying alternative rejection method...")
				// Aqui podemos implementar um método alternativo se necessário
			} else {
				si.logger.Infof("✅✅✅ [REJECTION-SUCCESS] Successfully rejected WhatsApp call from %s (CallID: %s)", fromPhone, callID)
				if response != nil {
					si.logger.Infof("🎯 Reason: SIP server responded with %d %s", response.StatusCode, response.Reason)
				} else {
					si.logger.Infof("🎯 Reason: SIP server rejected the call")
				}
				si.logger.Infof("📞📞📞 [WHATSAPP-TERMINATED] WhatsApp call should now be terminated")
			}
		} else {
			si.logger.Errorf("❌ CallManager not available for rejecting WhatsApp call")
		}
	} else {
		si.logger.Errorf("❌ WhatsApp connection not available for rejecting call")
	}
}
