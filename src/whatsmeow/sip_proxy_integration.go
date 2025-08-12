package whatsmeow

import (
	"fmt"

	"github.com/emiago/sipgo/sip"
	sipproxy "github.com/nocodeleaks/quepasa/sipproxy"
	log "github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow/types"
)

// SIPProxyIntegration provides integration between WhatsApp calls and the singleton SIP proxy
type SIPProxyIntegration struct {
	logger     *log.Entry
	sipProxy   *sipproxy.SIPProxyManager
	connection *WhatsmeowConnection // WhatsApp connection for rejecting calls
}

// NewSIPProxyIntegration creates a new SIP proxy integration instance
func NewSIPProxyIntegration(logger *log.Entry) *SIPProxyIntegration {
	integration := &SIPProxyIntegration{
		logger:   logger.WithField("component", "sip_integration"),
		sipProxy: sipproxy.GetSIPProxyManager(),
	}

	integration.logger.Infof("🔗 SIP Proxy Integration created")
	return integration
}

// InitializeSIPProxy sets up the singleton SIP proxy with server configuration
func (si *SIPProxyIntegration) InitializeSIPProxy(serverHost string, serverPort int, listenerPort int) error {
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
	if !si.sipProxy.IsRunning() {
		return fmt.Errorf("SIP proxy is not running")
	}

	si.logger.Infof("📞 Forwarding WhatsApp call to SIP server")
	si.logger.Infof("   📞 CallID: %s", callID)
	si.logger.Infof("   📞 From: %s", fromUser)
	si.logger.Infof("   📞 To: %s", toUser)

	// Send SIP INVITE
	if err := si.sipProxy.SendSIPInvite(fromUser, toUser, callID); err != nil {
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

	si.logger.Infof("📞❌ Handling WhatsApp call termination for CallID: %s", callID)
	si.sipProxy.RemoveCall(callID)
	return nil
}

// GetActiveCalls returns all active calls from the SIP proxy
func (si *SIPProxyIntegration) GetActiveCalls() map[string]*sipproxy.SIPProxyCallData {
	return si.sipProxy.GetActiveCalls()
}

// IsReady returns true if the SIP proxy is ready to handle calls
func (si *SIPProxyIntegration) IsReady() bool {
	return si.sipProxy.IsRunning()
}

// GetStatus returns status information about the SIP proxy
func (si *SIPProxyIntegration) GetStatus() map[string]interface{} {
	status := make(map[string]interface{})

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

	// Store WhatsApp connection for rejecting calls
	if conn, ok := whatsappHandler.(*WhatsmeowConnection); ok {
		si.connection = conn
		si.logger.Infof("✅ WhatsApp connection stored for call rejection")
	} else {
		si.logger.Warnf("⚠️ WhatsApp handler is not a WhatsmeowConnection, call rejection may not work")
	}

	// Callback para quando uma chamada é aceita (200 OK)
	si.sipProxy.SetCallAcceptedHandler(func(callID, fromPhone, toPhone string, response *sip.Response) {
		si.logger.Infof("✅ SIP CALL ACCEPTED! CallID: %s, From: %s, To: %s", callID, fromPhone, toPhone)
		si.logger.Infof("📡 SIP Response: %d %s", response.StatusCode, response.Reason)

		// Aqui podemos notificar o WhatsApp que a chamada foi aceita
		si.onCallAccepted(callID, fromPhone, toPhone, response)
	})

	// Callback para quando uma chamada é rejeitada (>=400)
	si.sipProxy.SetCallRejectedHandler(func(callID, fromPhone, toPhone string, response *sip.Response) {
		si.logger.Infof("❌ SIP CALL REJECTED! CallID: %s, From: %s, To: %s", callID, fromPhone, toPhone)
		si.logger.Infof("📡 SIP Response: %d %s", response.StatusCode, response.Reason)

		// Aqui podemos notificar o WhatsApp que a chamada foi rejeitada
		si.onCallRejected(callID, fromPhone, toPhone, response)
	})
}

// onCallAccepted é chamado quando uma chamada SIP é aceita
func (si *SIPProxyIntegration) onCallAccepted(callID, fromPhone, toPhone string, response *sip.Response) {
	si.logger.Infof("🎉 CALL ACCEPTED EVENT - Ready to start WhatsApp ↔ SIP communication!")
	si.logger.Infof("📞 CallID: %s", callID)
	si.logger.Infof("📞 From: %s", fromPhone)
	si.logger.Infof("📞 To: %s", toPhone)
	si.logger.Infof("📡 SIP Status: %d %s", response.StatusCode, response.Reason)

	// TODO: Aqui implementaremos a comunicação bidirecional WhatsApp ↔ SIP
	// - Aceitar a chamada no WhatsApp
	// - Estabelecer ponte de áudio
	// - Gerenciar estado da chamada
}

// onCallRejected é chamado quando uma chamada SIP é rejeitada
func (si *SIPProxyIntegration) onCallRejected(callID, fromPhone, toPhone string, response *sip.Response) {
	si.logger.Infof("💔 CALL REJECTED EVENT")
	si.logger.Infof("📞 CallID: %s", callID)
	si.logger.Infof("📞 From: %s", fromPhone)
	si.logger.Infof("📞 To: %s", toPhone)
	si.logger.Infof("📡 SIP Status: %d %s", response.StatusCode, response.Reason)

	// 🚫 AUTOMATICALLY REJECT THE CALL IN WHATSAPP TOO
	si.logger.Infof("🚫 SIP rejected call, automatically rejecting in WhatsApp...")

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
			err := callManager.RejectCall(fromJID, callID)
			if err != nil {
				si.logger.Errorf("❌ Failed to reject WhatsApp call: %v", err)
			} else {
				si.logger.Infof("✅ Successfully rejected WhatsApp call from %s (CallID: %s)", fromPhone, callID)
				si.logger.Infof("🎯 Reason: SIP server responded with %d %s", response.StatusCode, response.Reason)
			}
		} else {
			si.logger.Errorf("❌ CallManager not available for rejecting WhatsApp call")
		}
	} else {
		si.logger.Errorf("❌ WhatsApp connection not available for rejecting call")
	}
}
