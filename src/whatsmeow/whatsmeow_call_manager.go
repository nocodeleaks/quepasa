package whatsmeow

import (
	"fmt"

	sipproxy "github.com/nocodeleaks/quepasa/sipproxy"
	log "github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

// WhatsmeowCallManager is a lightweight adapter for the sipproxy call management
// All actual VoIP functionality is delegated to the sipproxy singleton
type WhatsmeowCallManager struct {
	connection     *WhatsmeowConnection
	logger         *log.Entry
	sipIntegration *SIPProxyIntegration
}

// NewWhatsmeowCallManager creates a lightweight call manager adapter
func NewWhatsmeowCallManager(conn *WhatsmeowConnection) *WhatsmeowCallManager {
	logger := conn.GetLogger().WithField("component", "call_adapter")

	// Create SIP proxy integration using the singleton
	sipIntegration := NewSIPProxyIntegration(logger)

	// Initialize the singleton SIP proxy with server configuration
	err := sipIntegration.InitializeSIPProxy("voip.sufficit.com.br", 26499, 5060)
	if err != nil {
		logger.Errorf("Failed to initialize SIP proxy: %v", err)
	} else {
		logger.Infof("🚀 SIP proxy singleton ready via adapter")
	}

	// Setup callbacks para capturar quando chamadas SIP são aceitas/rejeitadas
	sipIntegration.SetupCallbacks(conn)

	return &WhatsmeowCallManager{
		connection:     conn,
		logger:         logger,
		sipIntegration: sipIntegration,
	}
}

// GetSIPProxy returns the SIP proxy integration
func (cm *WhatsmeowCallManager) GetSIPProxy() *SIPProxyIntegration {
	return cm.sipIntegration
}

// AcceptCall delegates to sipproxy call answer manager
func (cm *WhatsmeowCallManager) AcceptCall(from types.JID, callID string) error {
	cm.logger.Infof("📞 Delegating call acceptance to sipproxy for %s (CallID: %s)", from, callID)

	// Check if SIP integration is ready
	if cm.sipIntegration == nil {
		cm.logger.Errorf("❌ SIP integration is nil!")
		return fmt.Errorf("SIP integration not initialized")
	}

	if !cm.sipIntegration.IsReady() {
		cm.logger.Errorf("❌ SIP integration is not ready!")
		return fmt.Errorf("SIP integration not ready")
	}

	cm.logger.Infof("✅ SIP integration is ready, proceeding with call acceptance")

	// Use sipproxy's call answer manager
	sipProxy := sipproxy.GetSIPProxyManager()
	if sipProxy == nil {
		cm.logger.Errorf("❌ sipproxy singleton is nil!")
		return fmt.Errorf("sipproxy singleton not available")
	}

	cm.logger.Infof("✅ Got sipproxy singleton, creating answer manager")
	answerManager := sipproxy.NewSIPProxyCallAnswerManager(sipProxy)

	cm.logger.Infof("🚀 Calling sipproxy AnswerCall for user: %s, callID: %s", from.User, callID)
	err := answerManager.AnswerCall(from.User, callID)
	if err != nil {
		cm.logger.Errorf("❌ SIP AnswerCall failed: %v", err)
		return err
	}

	cm.logger.Infof("✅ SIP AnswerCall succeeded!")
	return nil
}

// RejectCall rejects an incoming call using WhatsApp client
func (cm *WhatsmeowCallManager) RejectCall(from types.JID, callID string) error {
	if cm.connection == nil || cm.connection.Client == nil {
		return fmt.Errorf("connection not available")
	}

	cm.logger.Infof("❌ Rejecting call from %s (CallID: %s)", from, callID)
	return cm.connection.Client.RejectCall(from, callID)
}

// LogCallEvent logs call events and delegates to sipproxy
func (cm *WhatsmeowCallManager) LogCallEvent(eventType string, evt interface{}) {
	cm.logger.Infof("🔍 CALL EVENT: %s", eventType)

	// Delegate to sipproxy integration for all SIP handling
	switch e := evt.(type) {
	case *events.CallOffer:
		cm.logger.Infof("📞 CallOffer from %s (ID: %s)", e.From, e.CallID)
		// NOTE: ThrowSIPProxy is now called directly in the main handler, not here

	case *events.CallOfferNotice:
		cm.logger.Infof("📞 CallOfferNotice from %s (ID: %s)", e.From, e.CallID)

	case *events.CallTerminate:
		cm.logger.Infof("📞❌ CallTerminate from %s (ID: %s) - Reason: %v", e.From, e.CallID, e.Reason)
		if err := cm.sipIntegration.HandleWhatsAppCallTermination(e.CallID); err != nil {
			cm.logger.Errorf("Failed to handle call termination: %v", err)
		}

	case *events.CallAccept:
		cm.logger.Infof("📞✅ CallAccept from %s (ID: %s)", e.From, e.CallID)

	case *events.CallRelayLatency:
		cm.logger.Infof("📊 CallRelayLatency for %s (ID: %s)", e.From, e.CallID)

	default:
		cm.logger.Infof("📞 Unknown Call Event: %T", evt)
	}
}

// GetSIPProxyManager returns the SIP proxy integration (compatibility method)
func (cm *WhatsmeowCallManager) GetSIPProxyManager() *SIPProxyIntegration {
	return cm.sipIntegration
}

// SetSIPConfig delegates to sipproxy integration
func (cm *WhatsmeowCallManager) SetSIPConfig(host string, port int, listenerPort int) error {
	return cm.sipIntegration.InitializeSIPProxy(host, port, listenerPort)
}
