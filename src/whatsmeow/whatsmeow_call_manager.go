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

// NewWhatsmeowCallManager creates a new call manager instance
func NewWhatsmeowCallManager(connection *WhatsmeowConnection) *WhatsmeowCallManager {
	return &WhatsmeowCallManager{
		connection:     connection,
		logger:         log.WithField("service", "whatsmeow-call-manager"),
		sipIntegration: nil, // Will be set later via SetSIPIntegration
	}
}

// IsReady checks if the call manager is ready for operations
func (cm *WhatsmeowCallManager) IsReady() bool {
	return cm.connection != nil && cm.connection.Client != nil && cm.sipIntegration != nil && cm.sipIntegration.IsReady()
}

// SetSIPIntegration sets the SIP integration after initialization
func (cm *WhatsmeowCallManager) SetSIPIntegration(sipIntegration *SIPProxyIntegration) {
	cm.sipIntegration = sipIntegration
}

// GetSIPProxy returns the SIP proxy integration instance
func (cm *WhatsmeowCallManager) GetSIPProxy() *SIPProxyIntegration {
	return cm.sipIntegration
}

// PreAcceptCall sends a preaccept signal before accepting the call
// This follows the official WhatsApp call flow: offer -> preaccept -> accept -> transport
func (cm *WhatsmeowCallManager) PreAcceptCall(from types.JID, callID string) error {
	if cm.connection == nil || cm.connection.Client == nil {
		return fmt.Errorf("connection not available")
	}

	ownID := cm.connection.Client.Store.ID
	if ownID == nil {
		return fmt.Errorf("own ID not available")
	}

	cm.logger.Infof("📞⏳ [PRE-ACCEPT] Sending preaccept for call from %s (CallID: %s)", from, callID)

	// Create preaccept node following WhatsApp protocol
	preAcceptNode := binary.Node{
		Tag: "call",
		Attrs: binary.Attrs{
			"from": ownID.ToNonAD(),
			"to":   from,
			"id":   cm.connection.Client.GenerateMessageID(),
		},
		Content: []binary.Node{{
			Tag: "preaccept",
			Attrs: binary.Attrs{
				"call-id":      callID,
				"call-creator": from,
			},
		}},
	}

	cm.logger.Infof("📨⏳ [PRE-ACCEPT] Sending preaccept node: %+v", preAcceptNode)
	err := cm.connection.Client.DangerousInternals().SendNode(preAcceptNode)
	if err != nil {
		cm.logger.Errorf("❌⏳ [PRE-ACCEPT-ERROR] Failed to send preaccept node: %v", err)
		return err
	}

	cm.logger.Infof("✅⏳ [PRE-ACCEPT-SUCCESS] Preaccept sent successfully for %s (CallID: %s)", from, callID)
	return nil
}

// sendAcceptNode sends the actual accept node (separated for clarity)
func (cm *WhatsmeowCallManager) sendAcceptNode(from types.JID, callID string) error {
	ownID := cm.connection.Client.Store.ID
	if ownID == nil {
		return fmt.Errorf("own ID not available")
	}

	cm.logger.Infof("📞✅ [ACCEPT-NODE] Sending accept node for call from %s (CallID: %s)", from, callID)

	// Create accept node
	acceptNode := binary.Node{
		Tag: "call",
		Attrs: binary.Attrs{
			"from": ownID.ToNonAD(),
			"to":   from,
			"id":   cm.connection.Client.GenerateMessageID(),
		},
		Content: []binary.Node{{
			Tag: "accept",
			Attrs: binary.Attrs{
				"call-id":      callID,
				"call-creator": from,
			},
		}},
	}

	cm.logger.Infof("📨✅ [ACCEPT-NODE] Sending accept node: %+v", acceptNode)
	err := cm.connection.Client.DangerousInternals().SendNode(acceptNode)
	if err != nil {
		cm.logger.Errorf("❌✅ [ACCEPT-NODE-ERROR] Failed to send accept node: %v", err)
		return err
	}

	cm.logger.Infof("✅✅ [ACCEPT-NODE-SUCCESS] Accept node sent successfully!")
	return nil
}

// AcceptCall accepts an incoming call using the proper WhatsApp sequence: preaccept -> accept
// This implementation follows the WhatsApp Business API pattern discovered in Node.js code:
// 1. Immediate response (handled by SIP)
// 2. Pre-accept call (crucial step that stops ringing)
// 3. Accept call (final confirmation)
// 4. Bridge establishment (SIP delegation)
func (cm *WhatsmeowCallManager) AcceptCall(from types.JID, callID string) error {
	if cm.connection == nil || cm.connection.Client == nil {
		return fmt.Errorf("connection not available")
	}

	cm.logger.Infof("🔄📞 [CALL-ACCEPT-SEQUENCE] Starting proper WhatsApp call acceptance sequence...")

	// STEP 1: Send PreAccept first (this is crucial!)
	cm.logger.Infof("📞⏳ [STEP-1] Sending PreAccept...")
	err := cm.PreAcceptCall(from, callID)
	if err != nil {
		cm.logger.Errorf("❌⏳ [STEP-1-ERROR] PreAccept failed: %v", err)
		return fmt.Errorf("preaccept failed: %w", err)
	}

	// STEP 2: Wait a moment for WhatsApp to process preaccept
	cm.logger.Infof("⏰⏳ [STEP-1.5] Waiting for WhatsApp to process preaccept...")
	time.Sleep(1000 * time.Millisecond) // 1 second delay

	// STEP 3: Send actual Accept
	cm.logger.Infof("📞✅ [STEP-2] Sending Accept...")
	err = cm.sendAcceptNode(from, callID)
	if err != nil {
		cm.logger.Errorf("❌✅ [STEP-2-ERROR] Accept failed: %v", err)
		return fmt.Errorf("accept failed: %w", err)
	}

	// STEP 4: Wait a moment then delegate to SIP
	cm.logger.Infof("⏰📞 [STEP-2.5] Waiting before SIP delegation...")
	time.Sleep(500 * time.Millisecond) // 0.5 second delay

	// STEP 5: Delegate to SIP proxy
	cm.logger.Infof("📞🔗 [STEP-3] Delegating to SIP proxy...")
	err = cm.delegateToSIPProxy(from, callID)
	if err != nil {
		cm.logger.Errorf("❌🔗 [STEP-3-ERROR] SIP delegation failed: %v", err)
		return fmt.Errorf("SIP delegation failed: %w", err)
	}

	cm.logger.Infof("✅🎉 [CALL-ACCEPT-COMPLETE] Full call acceptance sequence completed successfully!")
	return nil
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

	// Get the WhatsApp receiver number (own number) from client store
	var toPhone string
	if cm.connection != nil && cm.connection.Client != nil && cm.connection.Client.Store != nil && cm.connection.Client.Store.ID != nil {
		toPhone = cm.connection.Client.Store.ID.User
		cm.logger.Infof("📱 WhatsApp receiver number: %s", toPhone)
	} else {
		cm.logger.Errorf("❌ Cannot get WhatsApp receiver number from client store")
		return fmt.Errorf("cannot get WhatsApp receiver number - client store not available")
	}

	cm.logger.Infof("📞📞 Calling sipproxy AnswerCallWithReceiver: from=%s, to=%s, callID=%s", from.User, toPhone, callID)
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
