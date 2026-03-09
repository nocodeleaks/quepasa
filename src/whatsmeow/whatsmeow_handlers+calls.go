// =============================================================================
// Arquivo: whatsmeow_handlers+calls.go
// Propósito:
//
//	Centralizar APENAS o tratamento BÁSICO (mínimo) dos eventos de chamadas
//	recebidos do WhatsApp (CallOffer, CallAccept, CallReject, Transport, etc.).
//	Aqui deve existir somente:
//	  - Logging padronizado de cada evento
//	  - Extração simples de campos essenciais (quando necessário)
//	  - Acionamento de rotinas triviais (ex: marcar estado, aceitar direto, fila leve)
//	NÃO colocar aqui lógica complexa, fluxos extensos, handshake detalhado,
//	montagem de nós (nodes) avançados, análise profunda de media ou integração
//	externa pesada. Qualquer processamento mais elaborado deve ser extraído
//	para arquivos/funções específicas (ex: call_accept_flow.go, call_transport_parser.go).
//
// Diretriz:
//
//	Mantenha este arquivo pequeno, legível e focado em orquestrar o básico.
//	Ele é a “camada fina” entre o evento bruto e módulos especializados.
//
// =============================================================================
package whatsmeow

import (
	"context"
	stdbinary "encoding/binary"
	"fmt"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/nocodeleaks/quepasa/library"
	"go.mau.fi/whatsmeow/binary"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

func normalizePhoneFromJIDString(value string) string {
	v := strings.TrimSpace(value)
	if v == "" {
		return ""
	}
	// Typical values: "5571...@s.whatsapp.net" or "5571..."
	if i := strings.IndexByte(v, '@'); i > 0 {
		v = v[:i]
	}
	return strings.TrimSpace(v)
}

func (source *WhatsmeowHandlers) HandleCallOffer(evt *events.CallOffer) {
	if source == nil || evt == nil {
		return
	}

	logentry := source.GetLogger()
	logentry.Infof("[CALL] Offer: from=%s callID=%s ts=%v", evt.From, evt.CallID, evt.Timestamp)

	callOffer := NewWhatsmeowCallOffer(evt)
	if !callOffer.IsValid() {
		logentry.Debugf("[CALL] Offer not valid (expired or not joinable): callID=%s", evt.CallID)
		return
	}

	// =========================================================================
	// HISTORICAL CALL FILTER - Critical anti-ghost-call mechanism
	// =========================================================================
	//
	// PROBLEM CONTEXT:
	// When this service starts/restarts, WhatsApp server sends "recent call history"
	// as part of sync. These historical CallOffer events look IDENTICAL to fresh
	// incoming calls, but they represent ALREADY FINISHED calls that happened before
	// our service started. If we process them, we will:
	//   - Send ACCEPT to calls that nobody is making anymore
	//   - Waste resources trying to establish media for ghost calls
	//   - Generate confusing logs about "calls that never connect"
	//
	// DETECTION STRATEGY:
	// We use TWO time references to distinguish historical vs fresh calls:
	//   1. startupTime (global var) = Unix timestamp when this process started
	//   2. evt.Timestamp = Unix timestamp when the WhatsApp call was CREATED
	//
	// LOGIC:
	//   Historical call = (service just started) AND (call created BEFORE startup)
	//
	// IMPLEMENTATION:
	//   - startupGracePeriod: window after startup where we check for historical calls
	//     (15 seconds is enough to receive all sync/history without blocking real calls)
	//   - timeSinceStartup: how long this service has been running
	//   - If we're still in grace period AND call timestamp is BEFORE our startup time,
	//     then this is DEFINITELY a historical/ghost call → DISCARD
	//
	// EXAMPLE SCENARIO:
	//   10:50:00 - User makes WhatsApp call (evt.Timestamp = 10:50:00)
	//   10:50:10 - User hangs up (call ends)
	//   10:51:00 - Our service starts (startupTime = 10:51:00)
	//   10:51:05 - WhatsApp sends CallOffer for the 10:50:00 call as history
	//
	//   Detection: timeSinceStartup=5s (< 15s grace period)
	//              evt.Timestamp (10:50:00) < startupTime (10:51:00)
	//              → HISTORICAL → Discard without processing
	//
	// WHY NOT RELY ON IsValid() ALONE?
	//   IsValid() checks if call is < 90 seconds old, but a call from 72 seconds ago
	//   would still pass validation even if it's historical. We need the startup
	//   reference to be certain.
	//
	// EDGE CASES:
	//   - Service running > 15s: grace period expires, all CallOffers are treated as fresh
	//     (this is correct - if service is stable, new offers are truly new)
	//   - Call timestamp AFTER startup: impossible to be historical, always fresh
	//   - Timezone issues: both timestamps use UTC (Go time.Now().Unix()), safe
	//
	// POSSIBLE IMPROVEMENTS:
	//   1. Use WhatsApp's own "history sync" event to mark historical batches explicitly
	//   2. Track "last call ID seen before shutdown" and reject any ID <= that threshold
	//   3. Add whitelist mode: only accept calls from specific contacts during grace period
	//   4. Persist startup time to disk to handle rapid restart scenarios better
	//
	// MONITORING:
	//   Look for "📜 [CALL] Discarding HISTORICAL offer" logs after service (re)starts.
	//   If you see these during normal operation (uptime > 15s), investigate timing bugs.
	//
	// RELATED CODE:
	//   - startupTime variable: defined in whatsmeow_handlers.go (global init)
	//   - IsValid() method: whatsmeow_call_offer.go (structural + age validation)
	//
	const startupGracePeriod = 15 * time.Second // Adjust if sync takes longer
	timeSinceStartup := time.Duration(time.Now().Unix()-startupTime) * time.Second
	callAge := time.Since(evt.Timestamp)

	if timeSinceStartup < startupGracePeriod && evt.Timestamp.Before(time.Unix(startupTime, 0)) {
		logentry.Warnf("📜 [CALL] Discarding HISTORICAL offer (received %v after startup, call was %v before startup): callID=%s from=%s",
			timeSinceStartup.Round(time.Second),
			time.Unix(startupTime, 0).Sub(evt.Timestamp).Round(time.Second),
			evt.CallID, evt.From)
		logentry.Infof("    ℹ️  This is a ghost call from WhatsApp sync/history, not a real incoming call. Safe to ignore.")
		return
	}

	logentry.Infof("✅ [CALL] Offer is FRESH (service uptime=%v, call age=%v): callID=%s - proceeding with call handling",
		timeSinceStartup.Round(time.Second),
		callAge.Round(time.Second),
		evt.CallID)
	// End of historical call filter
	// =========================================================================

	if envTruthy("QP_CALL_OFFER_SUMMARY") {
		full := envTruthy("QP_CALL_OFFER_SUMMARY_FULL")
		// Offer net medium (if present)
		offerMedium := ""
		if d := callOffer.GetData(); d != nil {
			if n := d.FindFirst("net"); n != nil {
				offerMedium = strings.TrimSpace(n.Attrs["medium"])
			}
		}

		relayUUID := ""
		relaySelfPID := ""
		relayPeerPID := ""
		relayKey := ""
		hbhKey := ""
		te2Count := 0
		authTokenCount := 0
		tokenCount := 0
		protocols := []string{}
		if rb := callOffer.GetRelayBlock(); rb != nil {
			relayUUID = rb.UUID
			relaySelfPID = rb.SelfPID
			relayPeerPID = rb.PeerPID
			relayKey = rb.Key
			hbhKey = rb.HBHKey
			te2Count = len(rb.TE2)
			tokenCount = len(rb.Tokens)
			authTokenCount = len(rb.Auth)
			protocols = rb.Protocols
		}

		// Audio codecs
		codecs := []string{}
		if d := callOffer.GetData(); d != nil {
			for _, c := range d.Content {
				if c.Tag != "audio" {
					continue
				}
				enc := strings.TrimSpace(c.Attrs["enc"])
				rate := strings.TrimSpace(c.Attrs["rate"])
				if enc != "" && rate != "" {
					codecs = append(codecs, enc+"@"+rate)
				}
			}
		}
		if len(codecs) > 1 {
			// small de-dup
			uniq := map[string]struct{}{}
			out := make([]string, 0, len(codecs))
			for _, c := range codecs {
				if _, ok := uniq[c]; ok {
					continue
				}
				uniq[c] = struct{}{}
				out = append(out, c)
			}
			codecs = out
		}

		tokens := callOffer.GetRelayTokens()
		redTokens := []string{}
		for i := 0; i < len(tokens) && i < 3; i++ {
			// GetRelayTokens returns base64 strings (stable representation).
			redTokens = append(redTokens, redactValue(tokens[i], full))
		}

		encType := ""
		encV := ""
		encLen := 0
		encKind := ""
		if d := callOffer.GetData(); d != nil {
			if enc := d.ExtractEncBlock(); enc != nil {
				encType = enc.Type
				encV = enc.V
				encLen = enc.RawLen
				encKind = enc.ContentKind
			}
		}

		logentry.Infof(
			"🧊 [OFFER-SUMMARY] callID=%s offer.medium=%s disable_p2p=%v relay_candidates=%v relays=%v relay_tokens=%d sample_tokens_b64=%v enc.type=%s enc.v=%s enc.len=%d enc.kind=%s relay.uuid=%s relay.self_pid=%s relay.peer_pid=%s relay.te2=%d relay.protocols=%v relay.token_nodes=%d relay.auth_token_nodes=%d has_relay_key=%t has_hbh_key=%t codecs=%v",
			evt.CallID,
			offerMedium,
			callOffer.IsP2PDisabledCached(),
			callOffer.HasRelayCandidatesCached(),
			callOffer.RelayNamesCached(),
			len(tokens),
			redTokens,
			encType,
			encV,
			encLen,
			encKind,
			relayUUID,
			relaySelfPID,
			relayPeerPID,
			te2Count,
			protocols,
			tokenCount,
			authTokenCount,
			strings.TrimSpace(relayKey) != '',
			strings.TrimSpace(hbhKey) != '',
			codecs,
		)
	}

	// Quick diagnosis: relay-only calls won't have media unless relay/ICE/SRTP is implemented.
	p2pDisabled := callOffer.IsP2PDisabledCached()
	hasRelay := callOffer.HasRelayCandidatesCached()
	relayNames := callOffer.RelayNamesCached()
	if p2pDisabled || hasRelay {
		logentry.Warnf("[CALL] Offer diagnostics: callID=%s relay_candidates=%v disable_p2p=%v relays=%v", evt.CallID, hasRelay, p2pDisabled, relayNames)
	}

	// SIP forwarding (signaling bridge). Prefer caller phone from offer attrs (caller_pn) to avoid @lid.
	if source.WhatsmeowConnection != nil {
		if envTruthy("QP_CALL_DISABLE_SIP_FORWARDING") {
			logentry.Warnf("[CALL] SIP forwarding disabled by env (QP_CALL_DISABLE_SIP_FORWARDING=1): callID=%s", evt.CallID)
		} else {
			sipCallManager := source.WhatsmeowConnection.GetSIPCallManager()
			if sipCallManager != nil && sipCallManager.IsEnabled() {
				callerPN := ""
				if callOffer != nil {
					callerPN = callOffer.GetData().Attrs["caller_pn"]
				}
				fromPhone := normalizePhoneFromJIDString(callerPN)
				if fromPhone == "" {
					fromPhone = normalizePhoneFromJIDString(evt.From.User)
				}

				toPhone := ""
				if statusManager := source.GetStatusManager(); statusManager != nil {
					if wid, err := statusManager.GetWidInternal(); err == nil {
						toPhone = normalizePhoneFromJIDString(wid)
					}
				}

				if fromPhone != "" && toPhone != "" {
					if err := sipCallManager.ProcessIncomingCall(evt.CallID, fromPhone, toPhone); err != nil {
						logentry.Errorf("[CALL] SIP forwarding failed: callID=%s err=%v", evt.CallID, err)
					} else {
						logentry.Infof("[CALL] SIP forwarding started: callID=%s from=%s to=%s", evt.CallID, fromPhone, toPhone)
					}
				} else {
					logentry.Warnf("[CALL] SIP forwarding skipped (missing phones): callID=%s fromPhone='%s' toPhone='%s'", evt.CallID, fromPhone, toPhone)
				}
			}
		}
	}

	if os.Getenv("QP_CALL_DUMP_OFFER") == "1" {
		path, err := DumpCallOfferEvent(evt, callOffer)
		if err != nil {
			logentry.Errorf("[CALL] Offer dump failed: callID=%s err=%v", evt.CallID, err)
		} else {
			logentry.Infof("[CALL] Offer dumped: callID=%s path=%s", evt.CallID, path)
		}
	}

	if source.WhatsmeowConnection != nil {
		if envTruthy("QP_CALL_OBSERVE_ONLY") {
			logentry.Warnf("[CALL] Observe-only enabled (QP_CALL_OBSERVE_ONLY=1): skipping WhatsApp accept flow (CallID=%s)", evt.CallID)
			return
		}
		if cm := source.WhatsmeowConnection.GetCallManager(); cm != nil {
			if rb := callOffer.GetRelayBlock(); rb != nil {
				cm.setCallRelayBlock(evt.CallID, rb)
			}
			if d := callOffer.GetData(); d != nil {
				if enc := d.ExtractEncBlock(); enc != nil && enc.RawLen > 0 {
					cm.setCallOfferEnc(evt.CallID, enc)
				}
			}
			cm.StartIncomingCallFlow(evt.From, evt.CallID)
		} else {
			logentry.Debug("[CALL] CallManager indisponível no momento da oferta")
		}
	}
}

func (source *WhatsmeowHandlers) HandleCallOfferNotice(evt *events.CallOfferNotice) {
	if source == nil || evt == nil {
		return
	}
	logentry := source.GetLogger()
	logentry.Infof("[CALL] OfferNotice: from=%s callID=%s", evt.From, evt.CallID)
	// OfferNotice pode chegar antes/depois — apenas log; fluxo principal já tratado no Offer
}

func (source *WhatsmeowHandlers) HandleCallRelayLatency(evt *events.CallRelayLatency) {
	if source == nil || evt == nil {
		return
	}
	logentry := source.GetLogger()
	logentry.Infof("[CALL] RelayLatency: from=%s callID=%s data=%v", evt.From, evt.CallID, evt.Data)

	// Decode relay endpoints from te payload (usually 4 bytes IPv4 + 2 bytes port).
	// This is useful for relay-only calls where the peer never provides ICE candidates.
	decoded := make([]RelayEndpoint, 0, 3)
	{
		nodeType := reflect.TypeOf(binary.Node{})
		rv := reflect.ValueOf(evt.Data)
		var relayNode *binary.Node
		if rv.IsValid() {
			if rv.Type() == nodeType {
				n := rv.Interface().(binary.Node)
				relayNode = &n
			} else if rv.Kind() == reflect.Ptr && rv.Elem().IsValid() && rv.Elem().Type() == nodeType {
				n := rv.Elem().Interface().(binary.Node)
				relayNode = &n
			}
		}
		if relayNode != nil {
			if children, ok := relayNode.Content.([]binary.Node); ok {
				for _, ch := range children {
					if ch.Tag != "te" {
						continue
					}
					relayName := ""
					latency := ""
					if ch.Attrs != nil {
						if v, ok2 := ch.Attrs["relay_name"]; ok2 {
							relayName = fmt.Sprintf("%v", v)
						}
						if v, ok2 := ch.Attrs["latency"]; ok2 {
							latency = fmt.Sprintf("%v", v)
						}
					}

					var payload []byte
					switch v := ch.Content.(type) {
					case []byte:
						payload = v
					case []interface{}:
						payload = make([]byte, 0, len(v))
						for _, it := range v {
							switch n := it.(type) {
							case byte:
								payload = append(payload, n)
							case int:
								payload = append(payload, byte(n))
							case int64:
								payload = append(payload, byte(n))
							case float64:
								payload = append(payload, byte(n))
							}
						}
					}

					endpoint := ""
					if len(payload) >= 6 {
						ip := net.IPv4(payload[0], payload[1], payload[2], payload[3]).String()
						port := int(stdbinary.BigEndian.Uint16(payload[4:6]))
						endpoint = net.JoinHostPort(ip, strconv.Itoa(port))
						ep := RelayEndpoint{
							RelayName:  relayName,
							IP:         ip,
							Port:       port,
							Endpoint:   endpoint,
							LatencyRaw: latency,
							ObservedAt: time.Now().UTC(),
						}
						decoded = append(decoded, ep)
					}
					if relayName != "" || endpoint != "" {
						logentry.Infof("[CALL] RelayLatency decoded: callID=%s relay=%s endpoint=%s latency=%s bytes=%d", evt.CallID, relayName, endpoint, latency, len(payload))
					}
				}
			}
		}
	}

	observeOnly := envTruthy("QP_CALL_OBSERVE_ONLY")
	if !observeOnly && len(decoded) > 0 && source.WhatsmeowConnection != nil {
		if cm := source.WhatsmeowConnection.GetCallManager(); cm != nil {
			for _, ep := range decoded {
				cm.addRelayEndpoint(evt.CallID, ep)
				go cm.ProbeRelaySTUNEndpoint(evt.CallID, ep)
			}
		}
	}
	if envTruthy("QP_CALL_DUMP_RELAY_LATENCY") {
		path, err := DumpCallRelayLatencyEvent(evt, decoded)
		if err != nil {
			logentry.Errorf("[CALL] RelayLatency dump failed: callID=%s err=%v", evt.CallID, err)
		} else {
			logentry.Infof("[CALL] RelayLatency dumped: callID=%s path=%s", evt.CallID, path)
		}
	}
	if observeOnly {
		logentry.Warnf("[CALL] Observe-only enabled (QP_CALL_OBSERVE_ONLY=1): skipping RelayLatency STUN probe and echo (CallID=%s)", evt.CallID)
		return
	}

	if !envTruthy("QP_CALL_ECHO_RELAY_LATENCY") {
		return
	}
	if source.WhatsmeowConnection == nil || source.WhatsmeowConnection.Client == nil {
		logentry.Warnf("[CALL] RelayLatency echo skipped: no connection/client (callID=%s)", evt.CallID)
		return
	}
	ownID := source.WhatsmeowConnection.Client.Store.ID
	if ownID == nil {
		logentry.Warnf("[CALL] RelayLatency echo skipped: own ID not available (callID=%s)", evt.CallID)
		return
	}

	// evt.Data type depends on whatsmeow versions; accept either binary.Node or *binary.Node.
	nodeType := reflect.TypeOf(binary.Node{})
	rv := reflect.ValueOf(evt.Data)
	if !rv.IsValid() {
		logentry.Warnf("[CALL] RelayLatency echo skipped: empty data (callID=%s)", evt.CallID)
		return
	}
	var relayNode binary.Node
	if rv.Type() == nodeType {
		relayNode = rv.Interface().(binary.Node)
	} else if rv.Kind() == reflect.Ptr && rv.Elem().IsValid() && rv.Elem().Type() == nodeType {
		relayNode = rv.Elem().Interface().(binary.Node)
	} else {
		logentry.Warnf("[CALL] RelayLatency echo skipped: unsupported data type=%s (callID=%s)", rv.Type().String(), evt.CallID)
		return
	}

	echo := binary.Node{Tag: "call", Attrs: binary.Attrs{
		"to":   evt.From.ToNonAD(),
		"from": ownID.ToNonAD(),
		"id":   source.WhatsmeowConnection.Client.GenerateMessageID(),
	}, Content: []binary.Node{relayNode}}
	if err := source.WhatsmeowConnection.Client.DangerousInternals().SendNode(context.Background(), echo); err != nil {
		logentry.Errorf("[CALL] RelayLatency echo send failed: callID=%s err=%v", evt.CallID, err)
		return
	}
	logentry.Infof("[CALL] RelayLatency echo sent: callID=%s", evt.CallID)
}

func (source *WhatsmeowHandlers) HandleCallTerminate(evt *events.CallTerminate) {
	if source == nil || evt == nil {
		return
	}
	logentry := source.GetLogger()
	logentry.Infof("[CALL] Terminate: from=%s callID=%s reason=%v", evt.From, evt.CallID, evt.Reason)
	if envTruthy("QP_CALL_DUMP_TERMINATE") {
		if path, err := DumpCallTerminateEvent(evt); err != nil {
			logentry.Errorf("[CALL] Terminate dump failed: callID=%s err=%v", evt.CallID, err)
		} else {
			logentry.Infof("[CALL] Terminate dumped: callID=%s path=%s", evt.CallID, path)
		}
	}
	if source.WhatsmeowConnection != nil {
		if cm := source.WhatsmeowConnection.GetCallManager(); cm != nil {
			cm.markCallTerminated(evt.CallID)
		}
	}
}

func (source *WhatsmeowHandlers) HandleCallAccept(evt *events.CallAccept) {
	if source == nil || evt == nil {
		return
	}
	logentry := source.GetLogger()
	logentry.Infof("[CALL] Accept: from=%s callID=%s", evt.From, evt.CallID)

	// Capture any <te> payloads from the accept node for TURN username experiments.
	if source.WhatsmeowConnection != nil {
		if cm := source.WhatsmeowConnection.GetCallManager(); cm != nil {
			teValues := extractCallAcceptTEValues(evt)
			if len(teValues) > 0 {
				cm.addCallAcceptTE(evt.CallID, teValues)
				logentry.Infof("[CALL] Accept TE captured: callID=%s count=%d", evt.CallID, len(teValues))
			}
		}
	}

	if envTruthy("QP_CALL_DUMP_ACCEPT_RECEIVED") {
		var ownID *types.JID
		if source.WhatsmeowConnection != nil && source.WhatsmeowConnection.Client != nil {
			ownID = source.WhatsmeowConnection.Client.Store.ID
		}
		if path, err := DumpCallAcceptReceivedEvent(evt, ownID); err != nil {
			logentry.Errorf("[CALL] Accept dump failed: callID=%s err=%v", evt.CallID, err)
		} else {
			logentry.Infof("[CALL] Accept dumped: callID=%s path=%s", evt.CallID, path)
		}
	}
}

func extractCallAcceptTEValues(evt *events.CallAccept) []string {
	if evt == nil || evt.Data == nil {
		return nil
	}
	// evt.Data type depends on whatsmeow versions; accept either binary.Node or *binary.Node.
	nodeType := reflect.TypeOf(binary.Node{})
	rv := reflect.ValueOf(evt.Data)
	if !rv.IsValid() {
		return nil
	}
	var acceptNode *binary.Node
	if rv.Type() == nodeType {
		n := rv.Interface().(binary.Node)
		acceptNode = &n
	} else if rv.Kind() == reflect.Ptr && rv.Elem().IsValid() && rv.Elem().Type() == nodeType {
		n := rv.Elem().Interface().(binary.Node)
		acceptNode = &n
	}
	if acceptNode == nil {
		return nil
	}
	children, ok := acceptNode.Content.([]binary.Node)
	if !ok || len(children) == 0 {
		return nil
	}
	vals := make([]string, 0, 4)
	for _, ch := range children {
		if ch.Tag != "te" {
			continue
		}
		s := ""
		switch v := ch.Content.(type) {
		case string:
			s = v
		case []byte:
			s = string(v)
		default:
			s = fmt.Sprint(v)
		}
		s = strings.TrimSpace(s)
		if s != "" {
			vals = append(vals, s)
		}
	}
	return vals
}

func (source *WhatsmeowHandlers) HandleCallReject(evt *events.CallReject) {
	if source == nil || evt == nil {
		return
	}
	logentry := source.GetLogger()
	logentry.Infof("[CALL] Reject: from=%s callID=%s", evt.From, evt.CallID)
}

func (source *WhatsmeowHandlers) HandleCallTransport(evt *events.CallTransport) {
	if source == nil || evt == nil {
		return
	}

	logentry := source.GetLogger()
	logentry.Infof("[CALL] Transport: from=%s callID=%s ts=%v", evt.From, evt.CallID, evt.Timestamp)

	callTransport := NewWhatsmeowCallTransport(evt)
	if !callTransport.IsValid() {
		logentry.Debugf("[CALL] Transport not valid (expired or malformed): callID=%s", evt.CallID)
		return
	}

	if os.Getenv("QP_CALL_DUMP_TRANSPORT") == "1" {
		path, err := DumpCallTransportEvent(evt, callTransport)
		if err != nil {
			logentry.Errorf("[CALL] Transport dump failed: callID=%s err=%v", evt.CallID, err)
		} else {
			logentry.Infof("[CALL] Transport dumped: callID=%s path=%s", evt.CallID, path)
		}
	}

	if os.Getenv("QP_CALL_LOG_TRANSPORT_JSON") == "1" {
		json := library.ToJson(evt)
		if json != "" {
			logentry.Infof("[CALL] Transport JSON: %s", json)
		}
	}

	if envTruthy("QP_CALL_OBSERVE_ONLY") {
		logentry.Warnf("[CALL] Observe-only enabled (QP_CALL_OBSERVE_ONLY=1): skipping transport processing/send (CallID=%s)", evt.CallID)
		return
	}

	if source.WhatsmeowConnection != nil {
		if cm := source.WhatsmeowConnection.GetCallManager(); cm != nil {
			if err := cm.HandleCallTransport(evt.From, evt.CallID, evt.Data); err != nil {
				logentry.Errorf("[CALL] Transport processing failed: callID=%s err=%v", evt.CallID, err)
			}
		}
	}

	// Log básico somente. Parsing profundo mover para outro arquivo (ex: call_transport_parse.go)
	size := 0
	if evt.Data != nil {
		size = len(fmt.Sprintf("%v", evt.Data))
	}
	logentry.Infof("[CALL] Transport: from=%s callID=%s dataSize=%d", evt.From, evt.CallID, size)
}

func (source *WhatsmeowHandlers) HandleCallUnknown(evt *events.UnknownCallEvent) {
	if source == nil || evt == nil {
		return
	}
	logentry := source.GetLogger()
	logentry.Infof("[CALL] UnknownCallEvent: raw=%+v", evt)
}
