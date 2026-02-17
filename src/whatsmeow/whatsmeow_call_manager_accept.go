package whatsmeow

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"go.mau.fi/whatsmeow/binary"
	"go.mau.fi/whatsmeow/types"
)

// AcceptCall aceita uma chamada WhatsApp usando estrutura exata do WA-JS.
// IMPORTANT: A simples emissão de um nó <accept> minimalista tende a manter a UI em "connecting"
// porque não prepara corretamente a negociação de transporte/candidatos.
func (cm *WhatsmeowCallManager) AcceptCall(from types.JID, callID string) error {
	cm.logger.Infof("📞 Aceitando chamada de %v (CallID: %s)", from, callID)
	if callID == "" {
		return fmt.Errorf("callID is empty")
	}
	return cm.executeWAJSAcceptStructure(from, callID, 0)
}

// AcceptCallMinimal sends a minimal <accept> node (no candidates).
// This is useful as a compatibility fallback to trigger the peer to send a CallTransport.
func (cm *WhatsmeowCallManager) AcceptCallMinimal(from types.JID, callID string) error {
	cm.logger.Infof("📞⚡ [MINIMAL-ACCEPT] Sending minimal ACCEPT (CallID: %s)", callID)
	if callID == "" {
		return fmt.Errorf("callID is empty")
	}

	ownID := cm.connection.Client.Store.ID
	if ownID == nil {
		return fmt.Errorf("own ID not available")
	}

	acceptNode := binary.Node{
		Tag: "call",
		Attrs: binary.Attrs{
			"id":   cm.connection.Client.GenerateMessageID(),
			"from": ownID.ToNonAD(),
			"to":   from.ToNonAD(),
		},
		Content: []binary.Node{{
			Tag: "accept",
			Attrs: binary.Attrs{
				"call-creator": from.ToNonAD(),
				"call-id":      callID,
			},
			Content: []binary.Node{
				{Tag: "audio", Attrs: binary.Attrs{"enc": "opus", "rate": "16000"}},
				{Tag: "audio", Attrs: binary.Attrs{"enc": "opus", "rate": "8000"}},
				{Tag: "net", Attrs: binary.Attrs{"medium": cm.getNetMediumForCall(callID), "protocol": "0"}},
				{Tag: "encopt", Attrs: binary.Attrs{"keygen": "2"}},
			},
		}},
	}

	return cm.connection.Client.DangerousInternals().SendNode(context.Background(), acceptNode)
}

// AcceptCallLegacySimple sends WhatsApp call signaling nodes exactly like the older legacy implementation:
// <preaccept> (minimal) -> wait -> <accept> (minimal).
//
// This is useful to validate UI state transitions (e.g. "calling" -> "connecting")
// without involving candidate/transport structures.
func (cm *WhatsmeowCallManager) AcceptCallLegacySimple(from types.JID, callID string) error {
	if cm == nil || cm.connection == nil || cm.connection.Client == nil {
		return fmt.Errorf("connection not available")
	}
	cm.logger.Infof("📞🕰️ [LEGACY-ACCEPT] Sending legacy preaccept->accept (CallID=%s)", callID)
	if callID == "" {
		return fmt.Errorf("callID is empty")
	}

	ownID := cm.connection.Client.Store.ID
	if ownID == nil {
		return fmt.Errorf("own ID not available")
	}

	// Use caller's JID directly (LID format if that's what we received)
	targetJID := from
	cm.logger.Infof("📞 [ACCEPT-TARGET] Using caller format: %s", targetJID.ToNonAD())

	// IMPORTANT: keep this as close as possible to the older legacy implementation.
	// In particular, use the original JID in "to" and "call-creator" (no ToNonAD conversion).
	netMedium := cm.getNetMediumForCall(callID)
	cm.logger.Infof("🧪📡 [LEGACY-NET] Using net medium=%s protocol=0 (CallID=%s)", netMedium, callID)

	// Some older WA-JS implementations send <call> stanzas with only {to,id} at the top-level
	// (no explicit "from" attribute). This is gated to avoid breaking other flows.
	legacyWAJSAttrs := envTruthy("QP_CALL_LEGACY_WAJS_ATTRS")
	if legacyWAJSAttrs {
		cm.logger.Warnf("🧪 [LEGACY-WAJS-ATTRS] Using WA-JS-like <call> attrs (no from) (CallID=%s)", callID)
	}
	preacceptCallAttrs := binary.Attrs{
		"to": targetJID.ToNonAD(),
		"id": cm.connection.Client.GenerateMessageID(),
	}
	if !legacyWAJSAttrs {
		preacceptCallAttrs["from"] = ownID.ToNonAD()
	}
	preacceptNode := binary.Node{
		Tag:   "call",
		Attrs: preacceptCallAttrs,
		Content: []binary.Node{{
			Tag: "preaccept",
			Attrs: binary.Attrs{
				"call-id":      callID,
				"call-creator": targetJID.ToNonAD(),
			},
		}},
	}
	if strings.TrimSpace(os.Getenv("QP_CALL_DUMP_ACCEPT")) == "1" {
		cm.logger.Infof("🧪 [LEGACY-PREACCEPT-DUMP] Node (WILL SEND):\n%s", cm.debugFormatNode(preacceptNode))
		if path, err := DumpPreacceptSent(callID, targetJID, *ownID, preacceptNode); err == nil {
			cm.logger.Infof("💾 [PREACCEPT-DUMP] Saved to %s", path)
		} else {
			cm.logger.Warnf("⚠️ [PREACCEPT-DUMP] Failed: %v", err)
		}
	}

	if err := cm.connection.Client.DangerousInternals().SendNode(context.Background(), preacceptNode); err != nil {
		return fmt.Errorf("failed to send legacy preaccept: %w", err)
	}

	delayMS := 1000
	if raw := strings.TrimSpace(os.Getenv("QP_CALL_LEGACY_DELAY_MS")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			if v < 0 {
				v = 0
			}
			if v > 5000 {
				v = 5000
			}
			delayMS = v
		}
	}
	if delayMS > 0 {
		time.Sleep(time.Duration(delayMS) * time.Millisecond)
	}

	acceptCallAttrs := binary.Attrs{
		"to": targetJID.ToNonAD(),
		"id": cm.connection.Client.GenerateMessageID(),
	}
	if !legacyWAJSAttrs {
		acceptCallAttrs["from"] = ownID.ToNonAD()
	}
	acceptNode := binary.Node{
		Tag:   "call",
		Attrs: acceptCallAttrs,
		Content: []binary.Node{{
			Tag: "accept",
			Attrs: binary.Attrs{
				"call-id":      callID,
				"call-creator": targetJID.ToNonAD(),
			},
			Content: []binary.Node{
				{Tag: "audio", Attrs: binary.Attrs{"enc": "opus", "rate": "16000"}},
				{Tag: "audio", Attrs: binary.Attrs{"enc": "opus", "rate": "8000"}},
				{Tag: "net", Attrs: binary.Attrs{"medium": netMedium, "protocol": "0"}},
				{Tag: "encopt", Attrs: binary.Attrs{"keygen": "2"}},
			},
		}},
	}
	if strings.TrimSpace(os.Getenv("QP_CALL_DUMP_ACCEPT")) == "1" {
		cm.logger.Infof("🧪 [LEGACY-ACCEPT-DUMP] Node (WILL SEND):\n%s", cm.debugFormatNode(acceptNode))
		if path, err := DumpAcceptSent(callID, targetJID, *ownID, acceptNode); err == nil {
			cm.logger.Infof("💾 [ACCEPT-DUMP] Saved to %s", path)
		} else {
			cm.logger.Warnf("⚠️ [ACCEPT-DUMP] Failed: %v", err)
		}
	}

	if err := cm.connection.Client.DangerousInternals().SendNode(context.Background(), acceptNode); err != nil {
		return fmt.Errorf("failed to send legacy accept: %w", err)
	}

	cm.logger.Infof("✅📞🕰️ [LEGACY-ACCEPT] Legacy accept dispatched using %s (delayMS=%d) (CallID=%s)", targetJID.ToNonAD(), delayMS, callID)

	// Minimal relay/media-plane step: validate outbound UDP reachability to relay endpoints.
	// This does NOT implement SRTP; it only probes and listens for relay responses.
	cm.MaybeStartRelaySessionProbe(callID)

	// Optional follow-up: send an initial <transport> after the legacy accept.
	// Some relay-only offers may require additional signaling to progress UI state.
	if envTruthy("QP_CALL_LEGACY_TRANSPORT_AFTER") {
		delay2 := 200
		if raw := strings.TrimSpace(os.Getenv("QP_CALL_LEGACY_TRANSPORT_DELAY_MS")); raw != "" {
			if v, err := strconv.Atoi(raw); err == nil {
				if v < 0 {
					v = 0
				}
				if v > 5000 {
					v = 5000
				}
				delay2 = v
			}
		}
		go func() {
			if delay2 > 0 {
				time.Sleep(time.Duration(delay2) * time.Millisecond)
			}
			cm.logger.Infof("🎵🚛 [LEGACY-TRANSPORT] Sending initial TRANSPORT after legacy accept (delayMS=%d) (CallID=%s)", delay2, callID)
			if err := cm.sendTransportInfo(targetJID, callID, 0); err != nil {
				cm.logger.Warnf("⚠️🎵🚛 [LEGACY-TRANSPORT] Failed to send transport after legacy accept: %v (CallID=%s)", err, callID)
			}
		}()
	}
	return nil
}
