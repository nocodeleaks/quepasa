package whatsmeow

import (
	"fmt"
	"time"

	sipproxy "github.com/nocodeleaks/quepasa/sipproxy"
	log "github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow/binary"
	"go.mau.fi/whatsmeow/types"
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

	// Create SIP proxy integration using the existing singleton
	// Note: SIP proxy should already be initialized in main.go
	sipIntegration := NewSIPProxyIntegration(logger)

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
	cm.logger.Infof("📞 Attempting to accept WhatsApp call from %s (CallID: %s)", from, callID)

	// First try to accept the call directly in WhatsApp using SendNode
	if cm.connection != nil && cm.connection.Client != nil {
		client := cm.connection.Client
		ownID := client.Store.ID
		if ownID != nil {
			cm.logger.Infof("🔗 Sending WhatsApp call accept node...")

			// Create accept node
			acceptNode := binary.Node{
				Tag: "call",
				Attrs: binary.Attrs{
					"from": ownID.ToNonAD(),
					"to":   from,
					"id":   client.GenerateMessageID(),
				},
				Content: []binary.Node{{
					Tag: "accept",
					Attrs: binary.Attrs{
						"call-id":      callID,
						"call-creator": from,
					},
				}},
			}

			cm.logger.Infof("📨 Sending accept node: %+v", acceptNode)
			err := client.DangerousInternals().SendNode(acceptNode)
			if err != nil {
				cm.logger.Errorf("❌ Failed to send accept node: %v", err)
			} else {
				cm.logger.Infof("✅ Accept node sent successfully!")

				// Give some time for WhatsApp to process the accept
				time.Sleep(500 * time.Millisecond)

				// Now delegate to SIP proxy call answer manager
				cm.logger.Infof("📞 Delegating call acceptance to sipproxy for %s (CallID: %s)", from, callID)
				return cm.delegateToSIPProxy(from, callID)
			}
		}
	}

	// Fallback to old method if direct accept fails
	cm.logger.Warnf("⚠️ Direct WhatsApp accept failed, trying fallback method")
	return cm.delegateToSIPProxy(from, callID)
}

// delegateToSIPProxy handles the SIP proxy delegation logic
func (cm *WhatsmeowCallManager) delegateToSIPProxy(from types.JID, callID string) error {

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
	sipProxy := sipproxy.SIPProxy
	if sipProxy == nil {
		cm.logger.Errorf("❌ sipproxy singleton is nil!")
		return fmt.Errorf("sipproxy singleton not available")
	}

	cm.logger.Infof("✅ Got sipproxy singleton, creating answer manager")
	answerManager := sipproxy.NewSIPProxyCallAnswerManager(sipProxy)

	// 🔧 Get the WhatsApp receiver number (own number) from client store
	var toPhone string
	if cm.connection != nil && cm.connection.Client != nil && cm.connection.Client.Store != nil && cm.connection.Client.Store.ID != nil {
		toPhone = cm.connection.Client.Store.ID.User
		cm.logger.Infof("� WhatsApp receiver number: %s", toPhone)
	} else {
		cm.logger.Errorf("❌ Cannot get WhatsApp receiver number from client store")
		return fmt.Errorf("cannot get WhatsApp receiver number - client store not available")
	}

	cm.logger.Infof("�� Calling sipproxy AnswerCallWithReceiver: from=%s, to=%s, callID=%s", from.User, toPhone, callID)
	err := answerManager.AnswerCallWithReceiver(from.User, toPhone, callID)
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

// GetSIPProxyManager returns the SIP proxy integration (compatibility method)
func (cm *WhatsmeowCallManager) GetSIPProxyManager() *SIPProxyIntegration {
	return cm.sipIntegration
}

// SetSIPConfig delegates to sipproxy integration
func (cm *WhatsmeowCallManager) SetSIPConfig(host string, port int, listenerPort int) error {
	return cm.sipIntegration.InitializeSIPProxy(host, port, listenerPort)
}
