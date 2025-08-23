package whatsmeow

import (
	"fmt"

	"go.mau.fi/whatsmeow/binary"
	"go.mau.fi/whatsmeow/types"
)

// AcceptCall aceita uma chamada WhatsApp usando estrutura exata do WA-JS
func (cm *WhatsmeowCallManager) AcceptCall(from types.JID, callID string) error {
	cm.logger.Infof("📞 Aceitando chamada de %v (CallID: %s)", from, callID)

	ownID := cm.connection.Client.Store.ID
	if ownID == nil {
		return fmt.Errorf("own ID not available")
	}

	// Tradução exata da estrutura WA-JS para Go
	acceptNode := binary.Node{
		Tag: "call",
		Attrs: binary.Attrs{
			"id": cm.connection.Client.GenerateMessageID(),
			"to": from,
		},
		Content: []binary.Node{{
			Tag: "accept",
			Attrs: binary.Attrs{
				"call-creator": from,
				"call-id":      callID,
			},
			Content: []binary.Node{
				{Tag: "audio", Attrs: binary.Attrs{"enc": "opus", "rate": "16000"}},
				{Tag: "audio", Attrs: binary.Attrs{"enc": "opus", "rate": "8000"}},
				{Tag: "net", Attrs: binary.Attrs{"medium": "3"}},
				{Tag: "encopt", Attrs: binary.Attrs{"keygen": "2"}},
			},
		}},
	}

	return cm.connection.Client.DangerousInternals().SendNode(acceptNode)
}
