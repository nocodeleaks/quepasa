package whatsmeow

import (
	"bytes"
	"context"
	stdbin "encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pion/stun"
	log "github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow/binary"
	"go.mau.fi/whatsmeow/types"
)

// WhatsmeowCallManager gerencia chamadas WhatsApp com estrutura WA-JS comprovada
// Todas as funcionalidades de VoIP sÃ£o delegadas para o sipproxy
type WhatsmeowCallManager struct {
	connection     *WhatsmeowConnection
	logger         *log.Entry
	sipIntegration *SIPProxyIntegration
	activeRTPPorts map[int]bool // Rastrear portas RTP ativas

	// Handshake state tracking
	handshakeStates map[string]*CallHandshakeState
	hsMutex         sync.Mutex
	offersSeen      map[string]time.Time // rastrear CallOffers jÃ¡ processadas
}

func (cm *WhatsmeowCallManager) getNetMedium() string {
	return cm.getNetMediumForCall("")
}

func (cm *WhatsmeowCallManager) getNetMediumForCall(callID string) string {
	// Observed values:
	//  - 1: direct/host candidate
	//  - 2: relay
	//  - 3: default (legacy)
	//
	// New: "auto" chooses relay (2) when we saw relay material in CallOffer.
	v := strings.TrimSpace(os.Getenv("QP_CALL_NET_MEDIUM"))
	switch v {
	case "1", "2", "3":
		return v
	}
	if strings.EqualFold(v, "auto") {
		if callID != "" {
			cm.hsMutex.Lock()
			st := cm.handshakeStates[callID]
			cm.hsMutex.Unlock()
			if st != nil && st.Relay != nil && len(st.Relay.TE2) > 0 {
				return "2"
			}
		}
		return "3"
	}
	return "3"
}

// CallHandshakeState tracks per-call handshake progress
type CallHandshakeState struct {
	RemoteTransportReceived bool
	Attempts                int
	CreatedAt               time.Time
	LastAttempt             time.Time
	AcceptSent              bool // se jÃ¡ enviamos ACCEPT (ex: modo direct)
	Terminated              bool
	TerminatedAt            time.Time
	TerminateCleanupPlanned bool
	OfferFrom               types.JID
	OfferFromSet            bool

	// Relay-only material captured from CallOffer (<relay> node). Do not log raw values.
	Relay *RelayBlock
	// OfferEnc contains the opaque <enc> payload observed in CallOffer (often type=pkmsg).
	// This is a prime suspect for relay/TURN short-term integrity material.
	OfferEnc                 *EncBlock
	OfferEncDecrypted        []byte
	OfferEncDecryptedSource  string
	RelayEndpoints           []RelayEndpoint
	RelaySTUNProbedEndpoints map[string]bool

	// Relay session probe is a minimal UDP connectivity test toward relay endpoints
	// to validate outbound UDP reachability before attempting SRTP/relay media-plane.
	RelaySessionProbeStarted   bool
	RelaySessionProbeStartedAt time.Time
	RelaySessionProbeEndpoint  string
	RelaySessionProbePending   bool

	// AcceptTE contains <te> payloads observed in incoming CallAccept events.
	// These values may be used as candidate TURN USERNAMEs in relay session probing.
	AcceptTE []string

	// Media port/mapping must be consistent across PREACCEPT/TRANSPORT/ACCEPT.
	LocalMediaPort  int
	PublicMediaIP   string
	PublicMediaPort int
}

func (cm *WhatsmeowCallManager) setCallOfferEnc(callID string, enc *EncBlock) {
	if cm == nil || callID == "" || enc == nil {
		return
	}
	cm.hsMutex.Lock()
	defer cm.hsMutex.Unlock()
	st, ok := cm.handshakeStates[callID]
	if !ok || st == nil {
		st = &CallHandshakeState{CreatedAt: time.Now(), LastAttempt: time.Now()}
		cm.handshakeStates[callID] = st
	}
	st.OfferEnc = enc
}

func (cm *WhatsmeowCallManager) setCallOfferEncDecrypted(callID string, plaintext []byte, source string) {
	if cm == nil || callID == "" || len(plaintext) == 0 {
		return
	}
	cm.hsMutex.Lock()
	defer cm.hsMutex.Unlock()
	st, ok := cm.handshakeStates[callID]
	if !ok || st == nil {
		st = &CallHandshakeState{CreatedAt: time.Now(), LastAttempt: time.Now()}
		cm.handshakeStates[callID] = st
	}
	st.OfferEncDecrypted = append([]byte(nil), plaintext...)
	st.OfferEncDecryptedSource = strings.TrimSpace(source)
}

func (cm *WhatsmeowCallManager) addCallAcceptTE(callID string, teValues []string) {
	if cm == nil || callID == "" || len(teValues) == 0 {
		return
	}
	cm.hsMutex.Lock()
	defer cm.hsMutex.Unlock()
	st, ok := cm.handshakeStates[callID]
	if !ok || st == nil {
		st = &CallHandshakeState{CreatedAt: time.Now(), LastAttempt: time.Now()}
		cm.handshakeStates[callID] = st
	}
	seen := map[string]bool{}
	for _, v := range st.AcceptTE {
		seen[v] = true
	}
	for _, v := range teValues {
		v = strings.TrimSpace(v)
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true
		st.AcceptTE = append(st.AcceptTE, v)
	}
}

func (cm *WhatsmeowCallManager) getCallAcceptTE(callID string) []string {
	if cm == nil || callID == "" {
		return nil
	}
	cm.hsMutex.Lock()
	st := cm.handshakeStates[callID]
	cm.hsMutex.Unlock()
	if st == nil || len(st.AcceptTE) == 0 {
		return nil
	}
	out := make([]string, 0, len(st.AcceptTE))
	out = append(out, st.AcceptTE...)
	return out
}

func (cm *WhatsmeowCallManager) getRelayEndpoints(callID string) []RelayEndpoint {
	if cm == nil || callID == "" {
		return nil
	}
	cm.hsMutex.Lock()
	st := cm.handshakeStates[callID]
	cm.hsMutex.Unlock()
	if st == nil || len(st.RelayEndpoints) == 0 {
		return nil
	}
	out := make([]RelayEndpoint, 0, len(st.RelayEndpoints))
	out = append(out, st.RelayEndpoints...)
	return out
}

func (cm *WhatsmeowCallManager) getRelayBlock(callID string) *RelayBlock {
	if cm == nil || callID == "" {
		return nil
	}
	cm.hsMutex.Lock()
	st := cm.handshakeStates[callID]
	cm.hsMutex.Unlock()
	if st == nil || st.Relay == nil {
		return nil
	}
	return st.Relay
}

func (cm *WhatsmeowCallManager) markRelaySessionProbeStarted(callID string, endpoint string) bool {
	if cm == nil || callID == "" {
		return false
	}
	cm.hsMutex.Lock()
	defer cm.hsMutex.Unlock()
	st := cm.handshakeStates[callID]
	if st == nil {
		return false
	}
	if st.RelaySessionProbeStarted {
		return false
	}
	st.RelaySessionProbeStarted = true
	st.RelaySessionProbePending = false
	st.RelaySessionProbeStartedAt = time.Now().UTC()
	st.RelaySessionProbeEndpoint = endpoint
	return true
}

func (cm *WhatsmeowCallManager) markRelaySessionProbePending(callID string) bool {
	if cm == nil || callID == "" {
		return false
	}
	cm.hsMutex.Lock()
	defer cm.hsMutex.Unlock()
	st := cm.handshakeStates[callID]
	if st == nil {
		return false
	}
	if st.RelaySessionProbeStarted || st.RelaySessionProbePending {
		return false
	}
	st.RelaySessionProbePending = true
	return true
}

func (cm *WhatsmeowCallManager) hasCallState(callID string) bool {
	if cm == nil || callID == "" {
		return false
	}
	cm.hsMutex.Lock()
	_, ok := cm.handshakeStates[callID]
	cm.hsMutex.Unlock()
	return ok
}

func (cm *WhatsmeowCallManager) shouldProbeRelayEndpoint(callID string, endpoint string) bool {
	if cm == nil || callID == "" {
		return false
	}
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return false
	}
	cm.hsMutex.Lock()
	defer cm.hsMutex.Unlock()
	st, ok := cm.handshakeStates[callID]
	if !ok || st == nil {
		return false
	}
	if st.RelaySTUNProbedEndpoints == nil {
		st.RelaySTUNProbedEndpoints = map[string]bool{}
	}
	if st.RelaySTUNProbedEndpoints[endpoint] {
		return false
	}
	st.RelaySTUNProbedEndpoints[endpoint] = true
	return true
}

func (cm *WhatsmeowCallManager) ProbeRelaySTUNEndpoint(callID string, ep RelayEndpoint) {
	if cm == nil || callID == "" {
		return
	}
	if !envTruthy("QP_CALL_RELAY_STUN_PROBE") {
		return
	}
	addr := strings.TrimSpace(ep.Endpoint)
	if !cm.shouldProbeRelayEndpoint(callID, addr) {
		return
	}

	cm.logger.Infof("ðŸ§ª [RELAY-STUN-PROBE] Probing relay=%s endpoint=%s (CallID=%s)", ep.RelayName, addr, callID)
	ip, port, localPort, err := cm.performRealSTUNRequest(addr)
	if err != nil {
		cm.logger.Warnf("âš ï¸ [RELAY-STUN-PROBE] relay=%s endpoint=%s failed: %v (CallID=%s)", ep.RelayName, addr, err, callID)
		if envTruthy("QP_CALL_RELAY_STUN_PROBE_CONTROL") {
			control := "stun.cloudflare.com:3478"
			cm.logger.Infof("ðŸ§ª [RELAY-STUN-CONTROL] Probing control STUN endpoint=%s (CallID=%s)", control, callID)
			ip2, port2, localPort2, err2 := cm.performRealSTUNRequest(control)
			if err2 != nil {
				cm.logger.Warnf("âš ï¸ [RELAY-STUN-CONTROL] endpoint=%s failed: %v (CallID=%s)", control, err2, callID)
			} else {
				cm.logger.Infof("âœ… [RELAY-STUN-CONTROL] endpoint=%s mapped=%s:%d localPort=%d (CallID=%s)", control, ip2, port2, localPort2, callID)
			}
		}
		return
	}
	cm.logger.Infof("âœ… [RELAY-STUN-PROBE] relay=%s endpoint=%s mapped=%s:%d localPort=%d (CallID=%s)", ep.RelayName, addr, ip, port, localPort, callID)
}

func (cm *WhatsmeowCallManager) setOfferFrom(callID string, from types.JID) {
	if cm == nil || callID == "" {
		return
	}
	cm.hsMutex.Lock()
	defer cm.hsMutex.Unlock()
	st, ok := cm.handshakeStates[callID]
	if !ok {
		st = &CallHandshakeState{CreatedAt: time.Now(), LastAttempt: time.Now()}
		cm.handshakeStates[callID] = st
	}
	st.OfferFrom = from
	st.OfferFromSet = true
}

func (cm *WhatsmeowCallManager) getOfferFrom(callID string) (types.JID, bool) {
	if cm == nil || callID == "" {
		return types.JID{}, false
	}
	cm.hsMutex.Lock()
	defer cm.hsMutex.Unlock()
	st, ok := cm.handshakeStates[callID]
	if !ok || st == nil || !st.OfferFromSet {
		return types.JID{}, false
	}
	return st.OfferFrom, true
}

func (cm *WhatsmeowCallManager) markCallTerminated(callID string) {
	if cm == nil || callID == "" {
		return
	}
	keepSeconds := 0
	if raw := strings.TrimSpace(os.Getenv("QP_CALL_KEEP_STATE_AFTER_TERMINATE_SECONDS")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			if v < 0 {
				v = 0
			}
			if v > 300 {
				v = 300
			}
			keepSeconds = v
		}
	}

	cm.hsMutex.Lock()
	st, ok := cm.handshakeStates[callID]
	if ok && st != nil {
		st.Terminated = true
		st.TerminatedAt = time.Now().UTC()
		if keepSeconds <= 0 {
			delete(cm.handshakeStates, callID)
			cm.hsMutex.Unlock()
			return
		}
		if !st.TerminateCleanupPlanned {
			st.TerminateCleanupPlanned = true
			delay := time.Duration(keepSeconds) * time.Second
			cm.hsMutex.Unlock()
			go func() {
				time.Sleep(delay)
				cm.hsMutex.Lock()
				if st2, ok2 := cm.handshakeStates[callID]; ok2 && st2 != nil && st2.Terminated {
					delete(cm.handshakeStates, callID)
				}
				cm.hsMutex.Unlock()
			}()
			return
		}
	}
	cm.hsMutex.Unlock()
}

func (cm *WhatsmeowCallManager) addRelayEndpoint(callID string, ep RelayEndpoint) {
	if cm == nil || callID == "" {
		return
	}
	if ep.ObservedAt.IsZero() {
		ep.ObservedAt = time.Now()
	}
	cm.hsMutex.Lock()
	defer cm.hsMutex.Unlock()
	st, ok := cm.handshakeStates[callID]
	if !ok {
		st = &CallHandshakeState{CreatedAt: time.Now(), LastAttempt: time.Now()}
		cm.handshakeStates[callID] = st
	}
	// Deduplicate by relay_name + endpoint.
	for _, existing := range st.RelayEndpoints {
		if existing.RelayName == ep.RelayName && existing.Endpoint == ep.Endpoint && existing.IP == ep.IP && existing.Port == ep.Port {
			return
		}
	}
	st.RelayEndpoints = append(st.RelayEndpoints, ep)
}

func envTruthy(name string) bool {
	v := strings.TrimSpace(os.Getenv(name))
	return v == "1" || strings.EqualFold(v, "true") || strings.EqualFold(v, "yes")
}

func envTruthyDefault(name string, defaultValue bool) bool {
	v := strings.TrimSpace(os.Getenv(name))
	if v == "" {
		return defaultValue
	}
	return v == "1" || strings.EqualFold(v, "true") || strings.EqualFold(v, "yes")
}

func callIsLIDJID(jid types.JID) bool {
	return strings.EqualFold(strings.TrimSpace(jid.Server), "lid") || strings.Contains(jid.String(), "@lid")
}

// callReplyJID chooses which JID form should be used when replying to a call peer.
// When QP_CALL_REPLY_USE_LID=1 and the peer is a LID JID, we reply using the raw LID JID.
// Otherwise, we use ToNonAD() (phone-number JID when mapping exists).
func (cm *WhatsmeowCallManager) callReplyJID(peer types.JID) types.JID {
	if envTruthy("QP_CALL_REPLY_USE_LID") && callIsLIDJID(peer) {
		return peer
	}
	return peer.ToNonAD()
}

func redactValue(v string, full bool) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	if full {
		return v
	}
	// Keep a short prefix/suffix to correlate without leaking full secrets.
	if len(v) <= 10 {
		return "***"
	}
	return v[:3] + "..." + v[len(v)-3:]
}

type callTransportSummary struct {
	NetMedium     string
	NetProtocol   string
	ICEUfrag      string
	ICEPwd        string
	Fingerprints  []string
	Candidates    []string
	SecretsFound  []string
	AttrSnapshots []string
}

func anyToString(v any) (string, bool) {
	switch t := v.(type) {
	case string:
		return t, true
	case []byte:
		return string(t), true
	default:
		return "", false
	}
}

func attrString(attrs map[string]any, key string) string {
	if attrs == nil {
		return ""
	}
	v, ok := attrs[key]
	if !ok {
		return ""
	}
	if s, ok := anyToString(v); ok {
		return strings.TrimSpace(s)
	}
	return strings.TrimSpace(fmt.Sprint(v))
}

func summarizeCallTransportNode(node *binary.Node, full bool) callTransportSummary {
	var s callTransportSummary
	if node == nil {
		return s
	}

	// Walk the full node tree and extract likely ICE/candidate/crypto fields.
	stack := []*binary.Node{node}
	for len(stack) > 0 {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		// Common places where children are stored.
		if children, ok := n.Content.([]binary.Node); ok {
			for i := len(children) - 1; i >= 0; i-- {
				child := children[i]
				stack = append(stack, &child)
			}
		}

		// Tag-based extraction.
		if strings.EqualFold(n.Tag, "net") {
			if s.NetMedium == "" {
				s.NetMedium = attrString(n.Attrs, "medium")
			}
			if s.NetProtocol == "" {
				s.NetProtocol = attrString(n.Attrs, "protocol")
			}
		}

		// Attr-based extraction.
		for k, vAny := range n.Attrs {
			key := strings.ToLower(strings.TrimSpace(k))
			val := ""
			if s1, ok := anyToString(vAny); ok {
				val = strings.TrimSpace(s1)
			} else {
				val = strings.TrimSpace(fmt.Sprint(vAny))
			}
			switch key {
			case "ufrag", "ice_ufrag", "ice-ufrag":
				if s.ICEUfrag == "" {
					s.ICEUfrag = redactValue(val, full)
				}
			case "pwd", "ice_pwd", "ice-pwd":
				if s.ICEPwd == "" {
					s.ICEPwd = redactValue(val, full)
				}
			case "fingerprint":
				if val != "" {
					s.Fingerprints = append(s.Fingerprints, redactValue(val, full))
				}
			}

			// Detect likely secret-bearing fields.
			if !full {
				if strings.Contains(key, "token") || strings.Contains(key, "secret") || strings.Contains(key, "key") || strings.Contains(key, "pwd") {
					if val != "" {
						s.SecretsFound = append(s.SecretsFound, k+"="+redactValue(val, false))
					}
				}
			}
		}

		// Candidate-like node snapshot.
		if strings.Contains(strings.ToLower(n.Tag), "candidate") {
			ip := attrString(n.Attrs, "ip")
			port := attrString(n.Attrs, "port")
			candType := attrString(n.Attrs, "type")
			proto := attrString(n.Attrs, "proto")
			if ip != "" && port != "" {
				c := ip + ":" + port
				if proto != "" {
					c = proto + " " + c
				}
				if candType != "" {
					c = c + " (" + candType + ")"
				}
				s.Candidates = append(s.Candidates, c)
			} else {
				s.AttrSnapshots = append(s.AttrSnapshots, fmt.Sprintf("%s attrs=%v", n.Tag, n.Attrs))
			}
		}
	}

	// De-dup small slices.
	uniq := func(in []string) []string {
		seen := map[string]bool{}
		out := make([]string, 0, len(in))
		for _, v := range in {
			v = strings.TrimSpace(v)
			if v == "" || seen[v] {
				continue
			}
			seen[v] = true
			out = append(out, v)
		}
		return out
	}
	s.Fingerprints = uniq(s.Fingerprints)
	s.Candidates = uniq(s.Candidates)
	s.SecretsFound = uniq(s.SecretsFound)
	s.AttrSnapshots = uniq(s.AttrSnapshots)

	return s
}

// disableSTUN verifica se STUN deve ser pulado (modo simulaÃ§Ã£o)
func (cm *WhatsmeowCallManager) disableSTUN() bool {
	return os.Getenv("QP_CALL_DISABLE_STUN") == "1"
}

// initHandshakeState initializes per-call state
func (cm *WhatsmeowCallManager) initHandshakeState(callID string) {
	cm.hsMutex.Lock()
	defer cm.hsMutex.Unlock()
	if _, exists := cm.handshakeStates[callID]; !exists {
		cm.handshakeStates[callID] = &CallHandshakeState{CreatedAt: time.Now(), LastAttempt: time.Now()}
		cm.logger.Infof("ðŸ§ª [HS-INIT] Handshake state inicializado para CallID=%s", callID)
	}
}

func (cm *WhatsmeowCallManager) setCallMediaPort(callID string, port int) {
	if callID == "" || port <= 0 {
		return
	}
	cm.hsMutex.Lock()
	defer cm.hsMutex.Unlock()
	st, ok := cm.handshakeStates[callID]
	if !ok {
		st = &CallHandshakeState{CreatedAt: time.Now(), LastAttempt: time.Now()}
		cm.handshakeStates[callID] = st
	}
	if st.LocalMediaPort == 0 {
		st.LocalMediaPort = port
		cm.logger.Infof("ðŸŽµðŸ”’ [MEDIA-PORT] Locked media port=%d for CallID=%s", port, callID)
		if cm.sipIntegration != nil {
			cm.sipIntegration.SetSIPRTPMirrorPort(callID, port)
		}
	}
}

func (cm *WhatsmeowCallManager) getCallMediaPort(callID string) int {
	cm.hsMutex.Lock()
	defer cm.hsMutex.Unlock()
	if st, ok := cm.handshakeStates[callID]; ok {
		return st.LocalMediaPort
	}
	return 0
}

func (cm *WhatsmeowCallManager) setCallPublicMapping(callID string, publicIP string, publicPort int) {
	if callID == "" || publicIP == "" || publicPort <= 0 {
		return
	}
	cm.hsMutex.Lock()
	defer cm.hsMutex.Unlock()
	st, ok := cm.handshakeStates[callID]
	if !ok {
		st = &CallHandshakeState{CreatedAt: time.Now(), LastAttempt: time.Now()}
		cm.handshakeStates[callID] = st
	}
	if st.PublicMediaIP == "" && st.PublicMediaPort == 0 {
		st.PublicMediaIP = publicIP
		st.PublicMediaPort = publicPort
		cm.logger.Infof("ðŸŽµðŸŒ [MEDIA-MAP] Locked srflx=%s:%d for CallID=%s", publicIP, publicPort, callID)
	}
}

func (cm *WhatsmeowCallManager) setCallRelayBlock(callID string, rb *RelayBlock) {
	if callID == "" || rb == nil {
		return
	}
	cm.hsMutex.Lock()
	defer cm.hsMutex.Unlock()
	st, ok := cm.handshakeStates[callID]
	if !ok {
		st = &CallHandshakeState{CreatedAt: time.Now(), LastAttempt: time.Now()}
		cm.handshakeStates[callID] = st
	}
	if st.Relay == nil {
		st.Relay = rb
	}
}

func (cm *WhatsmeowCallManager) getCallPublicMapping(callID string) (string, int) {
	cm.hsMutex.Lock()
	defer cm.hsMutex.Unlock()
	if st, ok := cm.handshakeStates[callID]; ok {
		return st.PublicMediaIP, st.PublicMediaPort
	}
	return "", 0
}

// markRemoteTransportReceived marks that remote transport arrived
func (cm *WhatsmeowCallManager) markRemoteTransportReceived(callID string) {
	cm.hsMutex.Lock()
	defer cm.hsMutex.Unlock()
	if st, ok := cm.handshakeStates[callID]; ok {
		if !st.RemoteTransportReceived {
			st.RemoteTransportReceived = true
			cm.logger.Infof("ðŸ§ª [HS-REMOTE-TRANSPORT] Transport remoto recebido para CallID=%s", callID)
		}
	}
}

// monitorTransportHandshake retries sending transport if remote not responding
func (cm *WhatsmeowCallManager) monitorTransportHandshake(callID string, from types.JID, rtpPort int) {
	// Permitir desabilitar monitor via env (ex: para modos accept-immediate / preaccept-only)
	mode := strings.TrimSpace(strings.ToLower(os.Getenv("QP_CALL_HANDSHAKE_MODE")))
	if mode == "preaccept-only" || mode == "accept-immediate" || os.Getenv("QP_CALL_DISABLE_MONITOR") == "1" {
		cm.logger.Infof("ðŸ§ª [HS-MONITOR-SKIP] Monitor desabilitado para modo=%s (CallID=%s)", mode, callID)
		return
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	maxAttempts := 5

	for range ticker.C {
		cm.hsMutex.Lock()
		st, ok := cm.handshakeStates[callID]
		if !ok {
			cm.hsMutex.Unlock()
			return
		}
		if st.RemoteTransportReceived {
			cm.logger.Infof("ðŸ§ª [HS-MONITOR-END] Remote transport recebido, encerrando monitor para CallID=%s", callID)
			delete(cm.handshakeStates, callID)
			cm.hsMutex.Unlock()
			return
		}
		if st.Terminated {
			cm.logger.Infof("ðŸ§ª [HS-MONITOR-END] Call already terminated, stopping monitor (CallID=%s)", callID)
			delete(cm.handshakeStates, callID)
			cm.hsMutex.Unlock()
			return
		}
		if st.Attempts >= maxAttempts {
			cm.logger.Warnf("âš ï¸ [HS-MONITOR-ABORT] Limite de tentativas atingido sem transport remoto (CallID=%s)", callID)
			delete(cm.handshakeStates, callID)
			cm.hsMutex.Unlock()
			return
		}
		st.Attempts++
		st.LastAttempt = time.Now()
		attempt := st.Attempts
		cm.hsMutex.Unlock()

		cm.logger.Warnf("âš ï¸ [HS-RETRY-%d] Reenviando transport inicial (CallID=%s) aguardando resposta remota", attempt, callID)
		if err := cm.sendTransportInfo(from, callID, rtpPort); err != nil {
			cm.logger.Errorf("âŒ [HS-RETRY-ERROR] Falha ao reenviar transport: %v", err)
		}
	}
}

// NewWhatsmeowCallManager cria um novo gerenciador de chamadas WhatsApp
func NewWhatsmeowCallManager(connection *WhatsmeowConnection, logger *log.Entry, sipIntegration *SIPProxyIntegration) *WhatsmeowCallManager {
	if logger == nil {
		logger = log.WithField("service", "whatsmeow-call-manager")
	}

	if os.Getenv("QP_CALL_META_ONLY") == "1" {
		logger.Infof("ðŸ§ª [ISOLATION] META-only ativo: evitar referÃªncias ao pacote 'models' (fluxo reduzido)")
	}

	return &WhatsmeowCallManager{
		connection:      connection,
		logger:          logger,
		sipIntegration:  sipIntegration,
		activeRTPPorts:  make(map[int]bool),
		handshakeStates: make(map[string]*CallHandshakeState),
		offersSeen:      make(map[string]time.Time),
	}
}

// hasSeenOffer verifica se jÃ¡ processamos esta oferta
func (cm *WhatsmeowCallManager) hasSeenOffer(callID string) bool {
	cm.hsMutex.Lock()
	defer cm.hsMutex.Unlock()
	_, exists := cm.offersSeen[callID]
	return exists
}

// markOffer registra oferta como vista
func (cm *WhatsmeowCallManager) markOffer(callID string) {
	cm.hsMutex.Lock()
	cm.offersSeen[callID] = time.Now()
	cm.hsMutex.Unlock()
}

// StartIncomingCallFlow decide qual fluxo iniciar para uma CallOffer nova
func (cm *WhatsmeowCallManager) StartIncomingCallFlow(from types.JID, callID string) {
	if cm == nil {
		return
	}
	if callID == "" {
		cm.logger.Warn("[CALL-FLOW] callID vazio ignorado")
		return
	}
	if envTruthy("QP_CALL_OBSERVE_ONLY") {
		cm.logger.Warnf("[CALL-FLOW] Observe-only enabled (QP_CALL_OBSERVE_ONLY=1): not sending PREACCEPT/ACCEPT/TRANSPORT (CallID=%s)", callID)
		return
	}
	if cm.hasSeenOffer(callID) {
		cm.logger.Infof("[CALL-FLOW] Oferta duplicada ignorada (CallID=%s)", callID)
		return
	}
	cm.markOffer(callID)
	cm.initHandshakeState(callID)
	cm.setOfferFrom(callID, from)
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("QP_CALL_ACCEPT_MODE"))) // minimal | direct | handshake (default)
	if mode == "legacy" {
		cm.logger.Infof("[CALL-FLOW] Modo=legacy â†’ enviando PREACCEPT/ACCEPT minimalistas (CallID=%s)", callID)
		if err := cm.AcceptCallLegacySimple(from, callID); err != nil {
			cm.logger.Errorf("[CALL-FLOW-ERROR] Legacy accept falhou: %v", err)
		}
		return
	}
	if mode == "sip" {
		cm.logger.Infof("[CALL-FLOW] Modo=sip â†’ aguardando SIP 200 OK para enviar ACCEPT no WhatsApp (CallID=%s)", callID)
		if envTruthyDefault("QP_CALL_SIP_PREACCEPT", true) {
			cm.logger.Infof("[CALL-FLOW] Modo=sip â†’ enviando PREACCEPT-only para tentar mover UI para 'connecting' (CallID=%s)", callID)
			if err := cm.SendPreacceptOnly(from, callID); err != nil {
				cm.logger.Errorf("[CALL-FLOW-ERROR] Modo=sip â†’ PREACCEPT-only falhou: %v", err)
			}
			if envTruthyDefault("QP_CALL_SIP_TRANSPORT_AFTER_PREACCEPT", true) {
				go func() {
					time.Sleep(250 * time.Millisecond)
					cm.logger.Infof("[CALL-FLOW] Modo=sip â†’ enviando TRANSPORT inicial apÃ³s PREACCEPT-only (CallID=%s)", callID)
					if err := cm.sendTransportInfo(from, callID, 0); err != nil {
						cm.logger.Errorf("[CALL-FLOW-ERROR] Modo=sip â†’ TRANSPORT pÃ³s-PREACCEPT-only falhou: %v", err)
					}
				}()
			}
		}
		return
	}
	if mode == "minimal" {
		cm.logger.Infof("[CALL-FLOW] Modo=minimal â†’ enviando ACCEPT minimalista (CallID=%s)", callID)
		if err := cm.AcceptCallMinimal(from, callID); err != nil {
			cm.logger.Errorf("[CALL-FLOW-ERROR] Minimal accept falhou: %v", err)
			return
		}
		if envTruthyDefault("QP_CALL_MINIMAL_PREACCEPT_AFTER", true) {
			// Hybrid approach: after minimal accept (to trigger remote transport), also send PREACCEPT-only.
			go func() {
				time.Sleep(150 * time.Millisecond)
				cm.logger.Infof("[CALL-FLOW] Modo=minimal â†’ enviando PREACCEPT (preaccept-only) apÃ³s ACCEPT minimal (CallID=%s)", callID)
				if err := cm.executeWAJSAcceptStructure(from, callID, 0); err != nil {
					cm.logger.Errorf("[CALL-FLOW-ERROR] PREACCEPT-only pÃ³s-minimal falhou: %v", err)
				}
			}()
		} else {
			cm.logger.Warnf("[CALL-FLOW] Modo=minimal â†’ PREACCEPT pÃ³s-minimal desabilitado (QP_CALL_MINIMAL_PREACCEPT_AFTER=0) (CallID=%s)", callID)
		}
		return
	}
	if mode == "direct" {
		cm.logger.Infof("[CALL-FLOW] Modo=direct â†’ enviando ACCEPT direto (CallID=%s)", callID)
		if err := cm.AcceptDirectCall(from, callID); err != nil {
			cm.logger.Errorf("[CALL-FLOW-ERROR] Direct accept falhou: %v", err)
		}
		return
	}
	// fallback / handshake completo
	cm.logger.Infof("[CALL-FLOW] Modo=handshake (ou vazio) â†’ iniciando AcceptCall padrÃ£o (CallID=%s)", callID)
	if err := cm.AcceptCall(from, callID); err != nil {
		cm.logger.Errorf("[CALL-FLOW-ERROR] AcceptCall falhou: %v", err)
	}
}

// SendPreacceptOnly sends only the <preaccept> node (no transport, no accept).
// Intended to move the UI away from ringing while still waiting for SIP-side answer.
func (cm *WhatsmeowCallManager) SendPreacceptOnly(from types.JID, callID string) error {
	if cm == nil {
		return fmt.Errorf("call manager is nil")
	}
	if callID == "" {
		return fmt.Errorf("callID is empty")
	}
	ownID := cm.connection.Client.Store.ID
	if ownID == nil {
		return fmt.Errorf("own ID not available")
	}

	// Determine local/public mapping similarly to WA-JS flow.
	localIP := cm.getLocalNetworkIP()
	if localIP == "" {
		return fmt.Errorf("failed to determine local IPv4 address")
	}

	var publicIP string
	var publicPort int
	var localPort int
	allocatedPort := 0
	if cm.disableSTUN() {
		allocatedPort = 64006
		cm.logger.Infof("ðŸ§ª [STUN-DISABLED] PREACCEPT-only usando porta estÃ¡tica %d", allocatedPort)
	} else {
		var err error
		publicIP, publicPort, localPort, err = cm.performSTUNDiscovery()
		if err != nil || publicIP == "" {
			cm.logger.Errorf("âŒðŸŽµ [PREACCEPT-ONLY-STUN-ERROR] Falha no STUN: %v", err)
			allocatedPort = 64006
		} else if localPort > 0 {
			allocatedPort = localPort
		}
	}
	if allocatedPort <= 0 {
		allocatedPort = 64006
	}
	cm.setCallMediaPort(callID, allocatedPort)
	cm.setCallPublicMapping(callID, publicIP, publicPort)

	includeSrflx := envTruthy("QP_CALL_INCLUDE_SRFLX")
	candidates := cm.buildCandidates(localIP, allocatedPort, publicIP, publicPort, includeSrflx)
	if envTruthy("QP_CALL_PREACCEPT_RELAY_EMPTY_NET") && cm.getNetMediumForCall(callID) == "2" {
		cm.logger.Warnf("ðŸ§Š [PREACCEPT-ONLY-RELAY-EMPTY-NET] Sending PREACCEPT with empty net (no candidates) for relay call (CallID=%s)", callID)
		candidates = nil
	}
	cm.logger.Infof("ðŸ§ª [PREACCEPT-ONLY-NET] medium=%s candidates=%d (CallID=%s)", cm.getNetMediumForCall(callID), len(candidates), callID)

	replyTo := cm.callReplyJID(from)
	preacceptNode := binary.Node{Tag: "call", Attrs: binary.Attrs{"to": replyTo, "from": ownID.ToNonAD(), "id": cm.connection.Client.GenerateMessageID()}, Content: []binary.Node{{
		Tag:   "preaccept",
		Attrs: binary.Attrs{"call-id": callID, "call-creator": replyTo},
		Content: []binary.Node{
			{Tag: "audio", Attrs: binary.Attrs{"enc": "opus", "rate": "16000"}},
			{Tag: "audio", Attrs: binary.Attrs{"enc": "opus", "rate": "8000"}},
			cm.wrapCandidatesInNetNode(callID, candidates),
			{Tag: "encopt", Attrs: binary.Attrs{"keygen": "2"}},
		},
	}}}

	if err := cm.connection.Client.DangerousInternals().SendNode(context.Background(), preacceptNode); err != nil {
		return err
	}
	cm.logger.Infof("âœ…ðŸŽµ [PREACCEPT-ONLY-SENT] PREACCEPT-only enviado (CallID=%s)", callID)
	return nil
}

// executeWAJSAcceptStructure implementa a estrutura exata do WA-JS
func (cm *WhatsmeowCallManager) executeWAJSAcceptStructure(from types.JID, callID string, rtpPort int) error {
	cm.logger.Infof("ðŸŽ¯ðŸ”¥ [WA-JS-ACCEPT] === EXECUTANDO ESTRUTURA EXATA DO WA-JS ===")

	ownID := cm.connection.Client.Store.ID
	if ownID == nil {
		return fmt.Errorf("own ID not available")
	}

	cm.logger.Infof("ðŸŽ¯ðŸ”¥ [PREACCEPT-FLOW] === USANDO PREACCEPT PRIMEIRO ===")
	cm.logger.Infof("ðŸŽ¯ðŸ’¡ [STRATEGY] PREACCEPT para preparar a chamada e receber transport")
	cm.logger.Infof("ðŸŽ¯ðŸ“ [FLOW] PREACCEPT â†’ aguardar TRANSPORT â†’ responder com ACCEPT")

	rawMode := os.Getenv("QP_CALL_HANDSHAKE_MODE")
	mode := strings.TrimSpace(strings.ToLower(rawMode)) // preaccept+transport (default), preaccept-only, accept-early, accept-immediate
	includeSrflx := envTruthy("QP_CALL_INCLUDE_SRFLX")
	cm.logger.Infof("ðŸ§ª [HS-MODE] raw='%s' normalized='%s' INCLUDE_SRFLX=%v", rawMode, mode, includeSrflx)

	// ðŸŽµðŸš€ STEP 1: Enviar PREACCEPT primeiro com informaÃ§Ãµes de rede (candidatos ICE)
	cm.logger.Infof("ðŸŽµðŸš€ [PREACCEPT] Enviando PREACCEPT com dados de rede/IP/porta (modo=%s)...", mode)

	// Obter IP local e porta para incluir no preaccept
	localIP := cm.getLocalNetworkIP()
	if localIP == "" {
		return fmt.Errorf("failed to determine local IPv4 address")
	}

	var publicIP string
	var publicPort int
	var localPort int
	var allocatedPort int
	var err error
	if cm.disableSTUN() {
		allocatedPort = 64006
		cm.logger.Infof("ðŸ§ª [STUN-DISABLED] QP_CALL_DISABLE_STUN=1 -> usando porta estÃ¡tica %d sem discovery", allocatedPort)
	} else {
		publicIP, publicPort, localPort, err = cm.performSTUNDiscovery()
		if err != nil || publicIP == "" {
			cm.logger.Errorf("âŒðŸŽµ [PREACCEPT-STUN-ERROR] Falha no STUN: %v", err)
			allocatedPort = 64006
		} else if localPort > 0 {
			allocatedPort = localPort
		}
	}
	cm.logger.Infof("ðŸŽµðŸ”§ [PREACCEPT-NETWORK] IP local: %s, porta local: %d, public=%s:%d (stunDisabled=%v)", localIP, allocatedPort, publicIP, publicPort, cm.disableSTUN())
	cm.setCallMediaPort(callID, allocatedPort)
	cm.setCallPublicMapping(callID, publicIP, publicPort)

	candidates := cm.buildCandidates(localIP, allocatedPort, publicIP, publicPort, includeSrflx)
	if envTruthy("QP_CALL_PREACCEPT_RELAY_EMPTY_NET") && cm.getNetMediumForCall(callID) == "2" {
		cm.logger.Warnf("ðŸ§Š [PREACCEPT-RELAY-EMPTY-NET] Sending PREACCEPT with empty net (no candidates) for relay call (CallID=%s)", callID)
		candidates = nil
	}
	cm.logger.Infof("ðŸ§ª [PREACCEPT-NET] medium=%s candidates=%d (CallID=%s)", cm.getNetMediumForCall(callID), len(candidates), callID)

	replyTo := cm.callReplyJID(from)
	preacceptNode := binary.Node{Tag: "call", Attrs: binary.Attrs{"to": replyTo, "from": ownID.ToNonAD(), "id": cm.connection.Client.GenerateMessageID()}, Content: []binary.Node{{
		Tag:   "preaccept",
		Attrs: binary.Attrs{"call-id": callID, "call-creator": replyTo},
		Content: []binary.Node{
			{Tag: "audio", Attrs: binary.Attrs{"enc": "opus", "rate": "16000"}},
			{Tag: "audio", Attrs: binary.Attrs{"enc": "opus", "rate": "8000"}},
			cm.wrapCandidatesInNetNode(callID, candidates),
			{Tag: "encopt", Attrs: binary.Attrs{"keygen": "2"}},
		},
	}}}

	// Dump tambÃ©m o ACCEPT hipotÃ©tico se variÃ¡vel de debug estiver ativa
	if os.Getenv("QP_CALL_DUMP_ACCEPT") == "1" {
		acceptPreview := cm.buildAcceptNode(replyTo, ownID.ToNonAD().String(), callID, replyTo.String(), candidates)
		cm.logger.Infof("ðŸ§ª [ACCEPT-DUMP-PREVIEW] Modo dump ativo (QP_CALL_DUMP_ACCEPT=1). Node hipotÃ©tico abaixo (NÃƒO ENVIADO):\n%s", cm.debugFormatNode(acceptPreview))
	}
	for i, c := range candidates {
		cm.logger.Infof("ðŸŽµðŸ” [PREACCEPT-CANDIDATE-%d] %v", i, c.Attrs)
	}

	err = cm.connection.Client.DangerousInternals().SendNode(context.Background(), preacceptNode)
	if err != nil {
		cm.logger.Errorf("âŒðŸŽµ [PREACCEPT-ERROR] Falha ao enviar preaccept: %v", err)
		return err
	}

	cm.logger.Infof("âœ…ðŸŽµ [PREACCEPT-SENT] PREACCEPT enviado com dados de rede: %s:%d! (aguardando TRANSPORT antes do ACCEPT)", localIP, allocatedPort)

	// Relay-only calls frequently arrive with disable_p2p=true and peer net medium=2.
	// In those cases, waiting for remote transport after PREACCEPT may time out; optionally send ACCEPT immediately.
	if cm.getNetMediumForCall(callID) == "2" && envTruthy("QP_CALL_RELAY_ACCEPT_IMMEDIATE") {
		cm.logger.Warnf("ðŸ§ª [RELAY-ACCEPT-IMMEDIATE] medium=2 detected; sending ACCEPT immediately after PREACCEPT (CallID=%s)", callID)
		acceptNode := cm.buildAcceptNodeWithMedium(replyTo, ownID.ToNonAD().String(), callID, replyTo.String(), nil, "2")
		if err2 := cm.connection.Client.DangerousInternals().SendNode(context.Background(), acceptNode); err2 != nil {
			cm.logger.Errorf("âŒ [RELAY-ACCEPT-IMMEDIATE-ERROR] %v", err2)
		} else {
			cm.logger.Infof("âœ… [RELAY-ACCEPT-IMMEDIATE-SENT] ACCEPT sent for relay-only call")
		}
	}

	skipTransport := mode == "preaccept-only"
	if mode == "accept-early" {
		cm.logger.Warnf("ðŸ§ª [HS-MODE=accept-early] Enviando ACCEPT antecipado (experimental)")
		go func() {
			acceptNode := cm.buildAcceptNode(replyTo, ownID.ToNonAD().String(), callID, replyTo.String(), candidates)
			if err2 := cm.connection.Client.DangerousInternals().SendNode(context.Background(), acceptNode); err2 != nil {
				cm.logger.Errorf("âŒ [ACCEPT-EARLY-ERROR] %v", err2)
			} else {
				cm.logger.Infof("âœ… [ACCEPT-EARLY-SENT] ACCEPT enviado antecipadamente")
			}
		}()
	}
	if mode == "accept-immediate" {
		cm.logger.Warnf("ðŸ§ª [HS-MODE=accept-immediate] Enviando ACCEPT imediatamente apÃ³s PREACCEPT (sem esperar transport remoto)")
		acceptNode := cm.buildAcceptNode(replyTo, ownID.ToNonAD().String(), callID, replyTo.String(), candidates)
		if err2 := cm.connection.Client.DangerousInternals().SendNode(context.Background(), acceptNode); err2 != nil {
			cm.logger.Errorf("âŒ [ACCEPT-IMMEDIATE-ERROR] %v", err2)
		} else {
			cm.logger.Infof("âœ… [ACCEPT-IMMEDIATE-SENT] ACCEPT enviado logo apÃ³s PREACCEPT")
		}
	}
	if !skipTransport && (mode == "" || strings.Contains(mode, "transport")) {
		time.Sleep(250 * time.Millisecond)
		cm.logger.Infof("ðŸŽµðŸ›  [TRANSPORT] (modo=%s) Enviando transport apÃ³s preaccept...", mode)
		if err = cm.sendTransportInfo(from, callID, rtpPort); err != nil {
			cm.logger.Errorf("âŒðŸŽµ [TRANSPORT-ERROR] Falha ao enviar transport: %v", err)
			return err
		}
		cm.logger.Infof("âœ…ðŸŽ‰ [SEQUENCE-COMPLETE] SequÃªncia preaccept â†’ transport enviada! (modo=%s)", mode)
	} else {
		cm.logger.Warnf("ðŸ§ª [HS-MODE=preaccept-only] NÃ£o enviando transport inicial; aguardando remoto")
	}

	// Initialize handshake state tracking and start monitor
	cm.initHandshakeState(callID)
	go cm.monitorTransportHandshake(callID, from, rtpPort)

	// ï¿½ RTP bridge e otimizaÃ§Ãµes desativadas em modo de isolamento (META-only) para focar no handshake WA
	if os.Getenv("QP_CALL_META_ONLY") != "1" {
		cm.logger.Infof("ðŸŽµðŸ”¥ [RTP-BRIDGE] === INICIANDO PONTE RTP VOIP â†” WHATSAPP ===")
		cm.logger.Infof("ðŸŽµðŸ”§ [FASE-2] === INICIANDO OTIMIZAÃ‡ÃƒO RTP/CODEC ===")
		cm.sendCodecPreferences(from, callID)
		cm.sendAdvancedRTPConfig(from, callID)
		cm.sendQualityOfService(from, callID)
		cm.logger.Infof("ðŸŽ‰ðŸŽµ [RTP-OPTIMIZE-COMPLETE] === OTIMIZAÃ‡ÃƒO RTP/CODEC CONCLUÃDA ===")
		go cm.startVoIPRTPBridge(from, callID, allocatedPort)
	} else {
		cm.logger.Infof("ðŸ§ª [ISOLATION] Skipping RTP bridge & codec optimization (QP_CALL_META_ONLY=1)")
	}

	return nil
}

// buildCandidates cria lista de candidates (host + opcional srflx)
// IMPORTANT: host candidate must use the LOCAL UDP port, while srflx uses the STUN-mapped port.
func (cm *WhatsmeowCallManager) buildCandidates(localIP string, localPort int, publicIP string, publicPort int, includeSrflx bool) []binary.Node {
	candidates := []binary.Node{{
		Tag: "candidate",
		Attrs: binary.Attrs{
			"generation": "0",
			"id":         "1",
			"ip":         localIP,
			"network":    "1",
			"port":       fmt.Sprintf("%d", localPort),
			"priority":   "2130706431",
			"protocol":   "udp",
			"type":       "host",
		},
	}}
	includeSrflxAlways := envTruthy("QP_CALL_INCLUDE_SRFLX_ALWAYS")
	if includeSrflx && publicIP != "" && publicPort > 0 && (includeSrflxAlways || publicIP != localIP || publicPort != localPort) {
		candidates = append(candidates, binary.Node{Tag: "candidate", Attrs: binary.Attrs{
			"generation": "0",
			"id":         "2",
			"ip":         publicIP,
			"network":    "1",
			"port":       fmt.Sprintf("%d", publicPort),
			"priority":   "2130706430",
			"protocol":   "udp",
			"type":       "srflx",
			"rel-addr":   localIP,
			"rel-port":   fmt.Sprintf("%d", localPort),
		}})
	}
	return candidates
}

// wrapCandidatesInNetNode envolve candidates em nÃ³ net com medium configurÃ¡vel
func (cm *WhatsmeowCallManager) wrapCandidatesInNetNode(callID string, candidates []binary.Node) binary.Node {
	return binary.Node{Tag: "net", Attrs: binary.Attrs{"medium": cm.getNetMediumForCall(callID), "protocol": "0"}, Content: candidates}
}

// sendTransportInfo envia informaÃ§Ãµes de transporte usando IP LOCAL do dispositivo WhatsApp
func (cm *WhatsmeowCallManager) sendTransportInfo(from types.JID, callID string, rtpPort int) error {
	cm.logger.Infof("ðŸŽµðŸš€ [TRANSPORT-LOCAL-IP] === USANDO IP LOCAL DO DISPOSITIVO WHATSAPP ===")
	cm.logger.Infof("ðŸŽµðŸ“‹ [TRANSPORT-PARAMS] From: %v, CallID: %s, RTPPort: %d", from, callID, rtpPort)

	// CRITICAL: Usar IP LOCAL da rede (onde o WhatsApp device estÃ¡ executando)
	// O WhatsApp device precisa informar onde ELE vai escutar RTP, nÃ£o o IP pÃºblico
	localIP := cm.getLocalNetworkIP()
	if localIP == "" {
		return fmt.Errorf("failed to determine local IPv4 address")
	}

	// Prefer a locked media port/mapping for this CallID to avoid mismatches.
	lockedPort := cm.getCallMediaPort(callID)
	publicIP, publicPort := cm.getCallPublicMapping(callID)
	allocatedPort := lockedPort
	if allocatedPort <= 0 {
		if cm.disableSTUN() {
			allocatedPort = 64006
			cm.logger.Infof("ðŸ§ª [STUN-DISABLED] Usando porta estÃ¡tica %d para transport (sem STUN)", allocatedPort)
		} else {
			ipTmp, pubPortTmp, locPortTmp, err2 := cm.performSTUNDiscovery()
			if err2 != nil {
				cm.logger.Errorf("âŒðŸŽµ [STUN-ERROR] Falha no STUN discovery: %v", err2)
				allocatedPort = 64006 // Fallback
			} else {
				publicIP = ipTmp
				publicPort = pubPortTmp
				if locPortTmp > 0 {
					allocatedPort = locPortTmp
				} else {
					allocatedPort = 64006
				}
				cm.setCallMediaPort(callID, allocatedPort)
				cm.setCallPublicMapping(callID, publicIP, publicPort)
			}
		}
	}
	cm.logger.Infof("âœ…ðŸ” [LOCAL-IP-SUCCESS] IP local para WhatsApp: %s:%d public=%s:%d (stunDisabled=%v)", localIP, allocatedPort, publicIP, publicPort, cm.disableSTUN())

	// Use port from parameter or STUN discovery
	var finalPort int
	if rtpPort > 0 {
		finalPort = rtpPort
	} else {
		finalPort = allocatedPort
	}

	cm.logger.Infof("ðŸŽµðŸ”§ [RTP-PORT-LOCAL] Usando IP local: %s, porta final: %d", localIP, finalPort)
	cm.logger.Infof("ðŸŽµðŸš› [TRANSPORT] Enviando transport INICIAL com IP LOCAL do dispositivo WhatsApp...")

	includeSrflx := envTruthy("QP_CALL_INCLUDE_SRFLX")
	transportCandidates := cm.buildCandidates(localIP, finalPort, publicIP, publicPort, includeSrflx)
	if cm.getNetMediumForCall(callID) == "2" && envTruthy("QP_CALL_TRANSPORT_RELAY_EMPTY_NET") {
		cm.logger.Warnf("ðŸ§Š [TRANSPORT-RELAY-EMPTY-NET] Sending initial transport with empty net (no candidates) for relay call (CallID=%s)", callID)
		transportCandidates = nil
	}
	if shouldIncludeCompactTransportNodes() && transportCandidates != nil {
		transportCandidates = appendCompactTransportNodes(transportCandidates, localIP, finalPort, publicIP, publicPort, includeSrflx)
	}

	ownID := cm.connection.Client.Store.ID
	if ownID == nil {
		return fmt.Errorf("own ID not available")
	}

	// CRITICAL: Transport INICIAL sem transport-message-type (ou tipo 0)
	// Este Ã© o transport que enviamos PRIMEIRO, antes de receber o deles
	legacyWAJSAttrs := envTruthy("QP_CALL_LEGACY_WAJS_ATTRS")
	replyTo := cm.callReplyJID(from)
	transportCallAttrs := binary.Attrs{
		"id": cm.connection.Client.GenerateMessageID(),
		"to": replyTo,
	}
	if !legacyWAJSAttrs {
		transportCallAttrs["from"] = ownID.ToNonAD()
	}

	transportNode := binary.Node{
		Tag:   "call",
		Attrs: transportCallAttrs,
		Content: []binary.Node{{
			Tag: "transport",
			Attrs: binary.Attrs{
				"call-id":      callID,
				"call-creator": replyTo,
				// NÃ£o incluir transport-message-type no initial transport
			},
			Content: []binary.Node{{
				Tag: "net",
				Attrs: binary.Attrs{
					"medium":   cm.getNetMediumForCall(callID),
					"protocol": "0",
				},
				Content: transportCandidates,
			}},
		}},
	}

	if errSend := cm.connection.Client.DangerousInternals().SendNode(context.Background(), transportNode); errSend != nil {
		return fmt.Errorf("failed to send transport: %w", errSend)
	}

	// Dump transport sent when configured
	if os.Getenv("QP_CALL_DUMP_TRANSPORT") == "1" {
		if p, e := DumpTransportSent(callID, from, *ownID, transportNode); e == nil {
			cm.logger.Infof("ðŸ’¾ [TRANSPORT-SENT-DUMP] Saved to %s", p)
		} else {
			cm.logger.Warnf("âš ï¸ [TRANSPORT-SENT-DUMP] Failed: %v", e)
		}
	}

	cm.logger.Infof("âœ… [TRANSPORT-SENT] Transport INICIAL enviado usando padrÃ£o whatsmeow!")
	cm.logger.Infof("ðŸŽµðŸ” [TRANSPORT-INFO] RTP: %s:%d, CallID: %s", localIP, finalPort, callID)
	cm.logger.Infof("ðŸŽ¯ðŸ’¡ [TRANSPORT-STRATEGY] Aguardando transport deles para responder com handshake")

	return nil
}

// sendTransportInfoResponse sends a TRANSPORT message in response to a peer-provided transport.
// For relay-only calls (medium=2), the peer transport often carries no ICE candidates; in this case,
// we default to sending an empty <net> (no candidates) to avoid mixing P2P candidates inside relay medium.
func (cm *WhatsmeowCallManager) sendTransportInfoResponse(from types.JID, callID string, receivedTransport *binary.Node) error {
	if cm == nil || cm.connection == nil || cm.connection.Client == nil {
		return fmt.Errorf("connection not available")
	}
	if callID == "" {
		return fmt.Errorf("callID is empty")
	}
	if receivedTransport == nil {
		return fmt.Errorf("receivedTransport is nil")
	}

	ownID := cm.connection.Client.Store.ID
	if ownID == nil {
		return fmt.Errorf("own ID not available")
	}

	// Extract round/type from the received transport attrs
	p2pRound := ""
	if receivedTransport.Attrs != nil {
		if v, ok := receivedTransport.Attrs["p2p-cand-round"]; ok {
			p2pRound = fmt.Sprintf("%v", v)
		}
	}

	// Extract medium from <net> child if present, fallback to computed
	transportMedium := ""
	if children, ok := receivedTransport.Content.([]binary.Node); ok {
		for _, ch := range children {
			if ch.Tag != "net" {
				continue
			}
			if ch.Attrs != nil {
				if m, ok2 := ch.Attrs["medium"]; ok2 {
					transportMedium = fmt.Sprintf("%v", m)
				}
			}
			break
		}
	}
	if strings.TrimSpace(transportMedium) == "" {
		transportMedium = cm.getNetMediumForCall(callID)
	}
	if strings.TrimSpace(transportMedium) == "" {
		transportMedium = "3"
	}

	// Default: relay transport response sends empty net (no candidates)
	includeCandidates := envTruthy("QP_CALL_TRANSPORT_RESPONSE_INCLUDE_CANDIDATES")
	transportCandidates := []binary.Node(nil)
	if includeCandidates {
		localIP := cm.getLocalNetworkIP()
		if localIP == "" {
			return fmt.Errorf("failed to determine local IPv4 address")
		}
		lockedPort := cm.getCallMediaPort(callID)
		publicIP, publicPort := cm.getCallPublicMapping(callID)
		port := lockedPort
		if port <= 0 {
			if cm.disableSTUN() {
				port = 64006
			} else {
				var errStun error
				var localPort int
				publicIP, publicPort, localPort, errStun = cm.performSTUNDiscovery()
				if errStun != nil || publicIP == "" {
					port = 64006
				} else if localPort > 0 {
					port = localPort
				}
				cm.setCallMediaPort(callID, port)
				cm.setCallPublicMapping(callID, publicIP, publicPort)
			}
		}
		includeSrflx := envTruthy("QP_CALL_INCLUDE_SRFLX")
		transportCandidates = cm.buildCandidates(localIP, port, publicIP, publicPort, includeSrflx)
		if shouldIncludeCompactTransportNodes() {
			transportCandidates = appendCompactTransportNodes(transportCandidates, localIP, port, publicIP, publicPort, includeSrflx)
		}
	}

	responseType := strings.TrimSpace(os.Getenv("QP_CALL_TRANSPORT_RESPONSE_TYPE"))
	if responseType == "" {
		responseType = "2"
	}
	replyTo := cm.callReplyJID(from)
	transportAttrs := binary.Attrs{
		"call-id":                callID,
		"call-creator":           replyTo,
		"transport-message-type": responseType,
	}
	if p2pRound != "" {
		transportAttrs["p2p-cand-round"] = p2pRound
	}

	legacyWAJSAttrs := envTruthy("QP_CALL_LEGACY_WAJS_ATTRS")
	callAttrs := binary.Attrs{
		"id": cm.connection.Client.GenerateMessageID(),
		"to": replyTo,
	}
	if !legacyWAJSAttrs {
		callAttrs["from"] = ownID.ToNonAD()
	}

	transportNode := binary.Node{
		Tag:   "call",
		Attrs: callAttrs,
		Content: []binary.Node{{
			Tag:   "transport",
			Attrs: transportAttrs,
			Content: []binary.Node{{
				Tag: "net",
				Attrs: binary.Attrs{
					"medium":   strings.TrimSpace(transportMedium),
					"protocol": "0",
				},
				Content: transportCandidates,
			}},
		}},
	}

	if errSend := cm.connection.Client.DangerousInternals().SendNode(context.Background(), transportNode); errSend != nil {
		return fmt.Errorf("failed to send transport response: %w", errSend)
	}
	if os.Getenv("QP_CALL_DUMP_TRANSPORT") == "1" {
		if p, e := DumpTransportSent(callID, from, *ownID, transportNode); e == nil {
			cm.logger.Infof("ðŸ’¾ [TRANSPORT-RESPONSE-DUMP] Saved to %s", p)
		} else {
			cm.logger.Warnf("âš ï¸ [TRANSPORT-RESPONSE-DUMP] Failed: %v", e)
		}
	}

	cm.logger.Infof("âœ…ðŸŽµðŸš› [TRANSPORT-RESPONSE-SENT] transport-message-type=2 medium=%s candidates=%d (CallID=%s)", strings.TrimSpace(transportMedium), len(transportCandidates), callID)

	// Next step in connecting state: validate UDP reachability to relay endpoints and capture first inbound packet.
	cm.MaybeStartRelaySessionProbe(callID)
	return nil
}

// performSTUNDiscovery faz STUN discovery e retorna:
//   - publicIP/publicPort: XOR-MAPPED (srflx)
//   - localPort: porta local real do socket UDP usado no STUN (host)
func (cm *WhatsmeowCallManager) performSTUNDiscovery() (string, int, int, error) {
	// Novo: fallback mÃºltiplo opcional
	fallbackEnabled := os.Getenv("QP_CALL_STUN_FALLBACK") == "1"
	primary := "stun1.l.google.com:19302"
	servers := []string{primary}
	if fallbackEnabled {
		servers = append(servers, "stun1.l.google.com:19302", "stun2.l.google.com:19302", "stun.l.google.com:19302", "stun.cloudflare.com:3478")
	}

	for i, srv := range servers {
		tag := "PRIMARY"
		if i > 0 {
			tag = fmt.Sprintf("FALLBACK-%d", i)
		}
		cm.logger.Infof("ðŸ”ðŸŒ [STUN-%s] Tentando servidor STUN: %s", tag, srv)
		ip, publicPort, localPort, err := cm.performRealSTUNRequest(srv)
		if err != nil {
			cm.logger.Errorf("âŒðŸ” [STUN-%s-FAIL] %v", tag, err)
			continue
		}
		cm.logger.Infof("âœ…ðŸ” [STUN-%s-SUCCESS] public=%s:%d localPort=%d", tag, ip, publicPort, localPort)
		return ip, publicPort, localPort, nil
	}
	return "", 0, 0, fmt.Errorf("todos servidores STUN falharam (fallback=%v)", fallbackEnabled)
}

// performRealSTUNRequest faz uma consulta STUN REAL e retorna tambÃ©m a porta local do socket.
func (cm *WhatsmeowCallManager) performRealSTUNRequest(stunServer string) (string, int, int, error) {
	cm.logger.Infof("ðŸ”ðŸš€ [STUN-REAL-REQ] Iniciando consulta STUN REAL para: %s", stunServer)

	// Resolver endereÃ§o do servidor STUN
	serverAddr, err := net.ResolveUDPAddr("udp", stunServer)
	if err != nil {
		return "", 0, 0, fmt.Errorf("erro ao resolver servidor STUN %s: %w", stunServer, err)
	}

	// Criar conexÃ£o UDP
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		return "", 0, 0, fmt.Errorf("erro ao conectar ao servidor STUN %s: %w", stunServer, err)
	}
	defer conn.Close()
	localPort := 0
	if la, ok := conn.LocalAddr().(*net.UDPAddr); ok {
		localPort = la.Port
	}

	// Criar mensagem STUN de binding request
	message := stun.MustBuild(stun.TransactionID, stun.BindingRequest)

	cm.logger.Infof("ðŸ”ðŸ“¤ [STUN-REAL-SEND] Enviando STUN Binding Request para %s", stunServer)

	// Enviar mensagem STUN
	_, err = conn.Write(message.Raw)
	if err != nil {
		return "", 0, localPort, fmt.Errorf("erro ao enviar STUN request: %w", err)
	}

	// Configurar timeout para resposta (mais rÃ¡pido para Meta, normal para outros)
	var timeout time.Duration
	if strings.Contains(stunServer, "157.240.226.62") {
		timeout = 2 * time.Second // Timeout mais rÃ¡pido para Meta
		cm.logger.Infof("ðŸ”â±ï¸ [STUN-TIMEOUT] Usando timeout rÃ¡pido (2s) para Meta")
	} else {
		timeout = 5 * time.Second // Timeout normal para outros
		cm.logger.Infof("ðŸ”â±ï¸ [STUN-TIMEOUT] Usando timeout normal (5s)")
	}

	conn.SetReadDeadline(time.Now().Add(timeout))

	// Buffer para resposta
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		return "", 0, localPort, fmt.Errorf("erro ao ler resposta STUN: %w", err)
	}

	cm.logger.Infof("ðŸ”ðŸ“¥ [STUN-REAL-RECV] Recebida resposta STUN de %d bytes", n)

	// Decodificar resposta STUN
	var stunResponse stun.Message
	stunResponse.Raw = buffer[:n]

	if err := stunResponse.Decode(); err != nil {
		return "", 0, localPort, fmt.Errorf("erro ao decodificar resposta STUN: %w", err)
	}

	// Extrair endereÃ§o mapeado (XOR-MAPPED-ADDRESS)
	var xorAddr stun.XORMappedAddress
	if err := xorAddr.GetFrom(&stunResponse); err != nil {
		// Tentar MAPPED-ADDRESS como fallback
		var mappedAddr stun.MappedAddress
		if err := mappedAddr.GetFrom(&stunResponse); err != nil {
			return "", 0, localPort, fmt.Errorf("erro ao extrair endereÃ§o mapeado: %w", err)
		}

		cm.logger.Infof("âœ…ðŸ” [STUN-REAL-MAPPED] EndereÃ§o descoberto (MAPPED): %s", mappedAddr.IP.String())
		return mappedAddr.IP.String(), mappedAddr.Port, localPort, nil
	}

	cm.logger.Infof("âœ…ðŸ” [STUN-REAL-XOR] EndereÃ§o descoberto (XOR-MAPPED): %s:%d (localPort=%d)", xorAddr.IP.String(), xorAddr.Port, localPort)
	return xorAddr.IP.String(), xorAddr.Port, localPort, nil
}

// getLocalNetworkIP obtÃ©m o IP local da rede onde o WhatsApp device estÃ¡ executando
func (cm *WhatsmeowCallManager) getLocalNetworkIP() string {
	cm.logger.Infof("ðŸ”ðŸ  [LOCAL-IP] Descobrindo IP local da rede...")

	// MÃ©todo 1: Conectar ao Google DNS para descobrir interface de rede local
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		cm.logger.Errorf("âŒðŸ  [LOCAL-IP-DIAL] Falha ao conectar: %v", err)
		return cm.getLocalNetworkIPFromInterfaces()
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	localIP := localAddr.IP.String()

	cm.logger.Infof("âœ…ðŸ  [LOCAL-IP-SUCCESS] IP local descoberto: %s", localIP)
	return localIP
}

// getLocalNetworkIPFromInterfaces fallback para obter IP local via interfaces de rede
func (cm *WhatsmeowCallManager) getLocalNetworkIPFromInterfaces() string {
	cm.logger.Infof("ðŸ”ðŸ  [LOCAL-IP-INTERFACES] Buscando via interfaces de rede...")

	interfaces, err := net.Interfaces()
	if err != nil {
		cm.logger.Errorf("âŒðŸ  [LOCAL-IP-INTERFACES] Falha ao listar interfaces: %v", err)
		return ""
	}

	var fallbackIPv4 string
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					ipStr := ipnet.IP.String()
					// Prefer RFC1918 addresses first.
					if strings.HasPrefix(ipStr, "192.168.") {
						cm.logger.Infof("âœ…ðŸ  [LOCAL-IP-FOUND] RFC1918 IPv4: %s", ipStr)
						return ipStr
					}
					if strings.HasPrefix(ipStr, "10.") {
						cm.logger.Infof("âœ…ðŸ  [LOCAL-IP-FOUND] RFC1918 IPv4: %s", ipStr)
						return ipStr
					}
					if strings.HasPrefix(ipStr, "172.") {
						// 172.16.0.0/12
						parts := strings.Split(ipStr, ".")
						if len(parts) >= 2 {
							second := parts[1]
							switch second {
							case "16", "17", "18", "19", "20", "21", "22", "23", "24", "25", "26", "27", "28", "29", "30", "31":
								cm.logger.Infof("âœ…ðŸ  [LOCAL-IP-FOUND] RFC1918 IPv4: %s", ipStr)
								return ipStr
							}
						}
					}

					// Keep a non-loopback IPv4 as fallback (may be public on servers).
					if fallbackIPv4 == "" {
						fallbackIPv4 = ipStr
					}
				}
			}
		}
	}
	if fallbackIPv4 != "" {
		cm.logger.Warnf("âš ï¸ðŸ  [LOCAL-IP-FALLBACK] Nenhum RFC1918 IPv4 encontrado; usando IPv4=%s", fallbackIPv4)
		return fallbackIPv4
	}

	cm.logger.Errorf("âŒðŸ  [LOCAL-IP-ERROR] Nenhum IPv4 nÃ£o-loopback encontrado")
	return ""
}

// startVoIPRTPBridge cria ponte RTP entre servidor VoIP e dispositivo WhatsApp
func (cm *WhatsmeowCallManager) startVoIPRTPBridge(from types.JID, callID string, rtpPort int) {
	cm.logger.Infof("ðŸŽµðŸš€ [RTP-BRIDGE] === INICIANDO PONTE RTP VOIP â†” WHATSAPP === (CallID=%s)", callID)
	cm.logger.Infof("ðŸŽµðŸ“‹ [RTP-BRIDGE-PARAMS] From: %v, CallID: %s, Port: %d", from, callID, rtpPort)

	// Use provided port first. Otherwise prefer the locked media port for this CallID.
	bridgePort := rtpPort
	if bridgePort <= 0 {
		bridgePort = cm.getCallMediaPort(callID)
	}
	if bridgePort <= 0 {
		if cm.disableSTUN() {
			bridgePort = 64006
			cm.logger.Infof("ðŸ§ª [STUN-DISABLED] Ponte RTP usando porta estÃ¡tica %d (sem STUN)", bridgePort)
		} else {
			if _, _, portTmp, err2 := cm.performSTUNDiscovery(); err2 != nil {
				cm.logger.Errorf("âŒðŸŽµ [RTP-BRIDGE-ERROR] Falha no STUN discovery: %v", err2)
				bridgePort = 64006
			} else {
				bridgePort = portTmp
			}
			cm.setCallMediaPort(callID, bridgePort)
		}
	}

	cm.logger.Infof("ðŸŽµðŸŒ‰ [RTP-BRIDGE] Criando ponte RTP na porta: %d (CallID=%s)", bridgePort, callID)
	cm.logger.Infof("ðŸŽµðŸ“¡ [RTP-BRIDGE-FLOW] VoIP Server â†’ :%d â†’ WhatsApp Device (CallID=%s)", bridgePort, callID)

	// ðŸŽµ STEP 2: Iniciar listener UDP real para RTP
	cm.logger.Infof("ðŸŽµðŸ‘‚ [RTP-LISTEN] Criando listener UDP real na porta %d... (CallID=%s)", bridgePort, callID)

	// Criar listener UDP
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", bridgePort))
	if err != nil {
		cm.logger.Errorf("âŒðŸŽµ [UDP-ERROR] Erro ao resolver endereÃ§o UDP: %v", err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		cm.logger.Errorf("âŒðŸŽµ [UDP-ERROR] Erro ao criar listener UDP: %v (CallID=%s)", err, callID)
		return
	}
	defer conn.Close()

	cm.logger.Infof("âœ…ðŸŽµ [UDP-SUCCESS] Listener UDP ativo em %s (CallID=%s)", conn.LocalAddr().String(), callID)
	cm.logger.Infof("ðŸŽµðŸ“¡ [RTP-READY] Aguardando RTP do servidor VoIP... (CallID=%s)", callID)

	// Buffer para receber pacotes RTP
	buffer := make([]byte, 1500) // MTU padrÃ£o

	// Timeout para evitar hang infinito
	timeout := time.After(2 * time.Minute)

	var wavDumper *rtpWavDumper
	wavDumpEnabled := strings.TrimSpace(os.Getenv("QP_CALL_RTP_DUMP_WAV"))
	if wavDumpEnabled == "1" || strings.EqualFold(wavDumpEnabled, "true") {
		dumpDir := strings.TrimSpace(os.Getenv("QP_CALL_DUMP_DIR"))
		if dumpDir == "" {
			dumpDir = "../.dist/call_dumps"
		}
		maxSeconds := 10
		if v := strings.TrimSpace(os.Getenv("QP_CALL_RTP_DUMP_WAV_SECONDS")); v != "" {
			if i, err := strconv.Atoi(v); err == nil {
				if i < 1 {
					i = 1
				}
				if i > 60 {
					i = 60
				}
				maxSeconds = i
			}
		}
		wavDumper = newRTPWavDumper(cm.logger, callID, dumpDir, time.Duration(maxSeconds)*time.Second)
		cm.logger.Infof("ðŸŽµðŸŽ§ [RTP-WAV-DUMP] Enabled: dumping up to %ds of RTP audio to %s (CallID=%s)", maxSeconds, dumpDir, callID)
	}

	firstPacketDumped := false
	for {
		select {
		case <-timeout:
			cm.logger.Infof("ðŸŽµâ° [RTP-TIMEOUT] Timeout de 2 minutos atingido (CallID=%s)", callID)
			return
		default:
			// Set timeout de 1 segundo para cada leitura
			conn.SetReadDeadline(time.Now().Add(1 * time.Second))

			n, clientAddr, err := conn.ReadFromUDP(buffer)
			if err != nil {
				// Timeout Ã© esperado, continuar
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				cm.logger.Errorf("âŒðŸŽµ [UDP-READ-ERROR] Erro ao ler UDP: %v (CallID=%s)", err, callID)
				continue
			}

			if wavDumper != nil {
				payload, pt, ok := extractRTPPayload(buffer[:n])
				if ok {
					wavDumper.HandleRTP(payload, pt)
				}
			}

			dumpEnabled := strings.TrimSpace(os.Getenv("QP_CALL_RTP_DUMP_FIRST"))
			if !firstPacketDumped && (dumpEnabled == "1" || strings.EqualFold(dumpEnabled, "true")) {
				firstPacketDumped = true
				dumpLen := n
				if dumpLen > 32 {
					dumpLen = 32
				}

				rtpVersion := -1
				rtpMarker := -1
				rtpPayloadType := -1
				rtpSeq := -1
				rtpTimestamp := int64(-1)
				rtpSSRC := int64(-1)
				rtpPayloadLen := -1
				if n >= 12 {
					rtpVersion = int((buffer[0] >> 6) & 0x3)
					rtpMarker = int((buffer[1] >> 7) & 0x1)
					rtpPayloadType = int(buffer[1] & 0x7F)
					rtpSeq = int((uint16(buffer[2]) << 8) | uint16(buffer[3]))
					rtpTimestamp = int64((uint32(buffer[4]) << 24) | (uint32(buffer[5]) << 16) | (uint32(buffer[6]) << 8) | uint32(buffer[7]))
					rtpSSRC = int64((uint32(buffer[8]) << 24) | (uint32(buffer[9]) << 16) | (uint32(buffer[10]) << 8) | uint32(buffer[11]))
					rtpPayloadLen = n - 12
				}

				cm.logger.Infof(
					"ðŸŽµðŸ§ª [RTP-DUMP-FIRST] from=%v bytes=%d v=%d m=%d pt=%d seq=%d ts=%d ssrc=%d payload=%d first=%x (CallID=%s)",
					clientAddr,
					n,
					rtpVersion,
					rtpMarker,
					rtpPayloadType,
					rtpSeq,
					rtpTimestamp,
					rtpSSRC,
					rtpPayloadLen,
					buffer[:dumpLen],
					callID,
				)
			}

			cm.logger.Infof("ðŸŽµðŸ“¥ [RTP-RECEIVED] %d bytes de %v - reenviando para WhatsApp... (CallID=%s)", n, clientAddr, callID)
			cm.forwardRealRTPToWhatsApp(from, callID, buffer[:n], clientAddr.String())
		}
	}
}

type rtpWavDumper struct {
	logger     *log.Entry
	callID     string
	dumpDir    string
	maxDur     time.Duration
	startTime  time.Time
	writer     *wavWriter
	filePath   string
	closed     bool
	bytesWrote int
}

func newRTPWavDumper(logger *log.Entry, callID string, dumpDir string, maxDur time.Duration) *rtpWavDumper {
	return &rtpWavDumper{logger: logger, callID: callID, dumpDir: dumpDir, maxDur: maxDur}
}

func (d *rtpWavDumper) HandleRTP(payload []byte, payloadType int) {
	if d == nil || d.closed {
		return
	}
	if payloadType != 0 {
		return
	}
	if d.startTime.IsZero() {
		d.startTime = time.Now()
		_ = os.MkdirAll(filepath.Clean(d.dumpDir), 0o755)
		d.filePath = filepath.Join(filepath.Clean(d.dumpDir), fmt.Sprintf("rtp_%s_%s.wav", d.callID, time.Now().Format("20060102_150405")))
		w, err := newWavWriter(d.filePath, 8000, 1)
		if err != nil {
			d.logger.Errorf("âŒðŸŽµðŸŽ§ [RTP-WAV-DUMP-ERROR] Failed to create wav: %v (CallID=%s)", err, d.callID)
			d.closed = true
			return
		}
		d.writer = w
		d.logger.Infof("ðŸŽµðŸŽ§ [RTP-WAV-DUMP] Started: %s (CallID=%s)", d.filePath, d.callID)
	}

	if time.Since(d.startTime) > d.maxDur {
		d.close()
		return
	}

	if d.writer == nil {
		return
	}

	pcm := decodePCMUToPCM16(payload)
	if len(pcm) == 0 {
		return
	}
	if err := d.writer.WriteSamples(pcm); err != nil {
		d.logger.Errorf("âŒðŸŽµðŸŽ§ [RTP-WAV-DUMP-ERROR] Failed to write wav samples: %v (CallID=%s)", err, d.callID)
		d.close()
		return
	}
	d.bytesWrote += len(payload)
}

func (d *rtpWavDumper) close() {
	if d == nil || d.closed {
		return
	}
	d.closed = true
	if d.writer != nil {
		_ = d.writer.Close()
	}
	if d.filePath != "" {
		d.logger.Infof("ðŸŽµðŸŽ§ [RTP-WAV-DUMP] Done: wrote=%d payload-bytes file=%s (CallID=%s)", d.bytesWrote, d.filePath, d.callID)
	}
}

type wavWriter struct {
	f         *os.File
	dataBytes uint32
}

func newWavWriter(path string, sampleRate uint32, channels uint16) (*wavWriter, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	ww := &wavWriter{f: f, dataBytes: 0}
	if err := ww.writeHeader(sampleRate, channels, 16); err != nil {
		_ = f.Close()
		return nil, err
	}
	return ww, nil
}

func (w *wavWriter) writeHeader(sampleRate uint32, channels uint16, bitsPerSample uint16) error {
	byteRate := sampleRate * uint32(channels) * uint32(bitsPerSample) / 8
	blockAlign := channels * bitsPerSample / 8

	header := make([]byte, 44)
	copy(header[0:4], []byte("RIFF"))
	// chunk size at [4:8] filled on Close
	copy(header[8:12], []byte("WAVE"))
	copy(header[12:16], []byte("fmt "))
	stdbin.LittleEndian.PutUint32(header[16:20], 16)
	stdbin.LittleEndian.PutUint16(header[20:22], 1)
	stdbin.LittleEndian.PutUint16(header[22:24], channels)
	stdbin.LittleEndian.PutUint32(header[24:28], sampleRate)
	stdbin.LittleEndian.PutUint32(header[28:32], byteRate)
	stdbin.LittleEndian.PutUint16(header[32:34], blockAlign)
	stdbin.LittleEndian.PutUint16(header[34:36], bitsPerSample)
	copy(header[36:40], []byte("data"))
	// data size at [40:44] filled on Close

	_, err := w.f.Write(header)
	return err
}

func (w *wavWriter) WriteSamples(samples []int16) error {
	if w == nil || w.f == nil {
		return io.ErrClosedPipe
	}
	if len(samples) == 0 {
		return nil
	}

	buf := make([]byte, len(samples)*2)
	for i, s := range samples {
		stdbin.LittleEndian.PutUint16(buf[i*2:i*2+2], uint16(s))
	}

	n, err := w.f.Write(buf)
	if err != nil {
		return err
	}
	w.dataBytes += uint32(n)
	return nil
}

func (w *wavWriter) Close() error {
	if w == nil || w.f == nil {
		return nil
	}

	chunkSize := uint32(36) + w.dataBytes
	if _, err := w.f.Seek(4, 0); err != nil {
		_ = w.f.Close()
		w.f = nil
		return err
	}
	if err := stdbin.Write(w.f, stdbin.LittleEndian, chunkSize); err != nil {
		_ = w.f.Close()
		w.f = nil
		return err
	}
	if _, err := w.f.Seek(40, 0); err != nil {
		_ = w.f.Close()
		w.f = nil
		return err
	}
	if err := stdbin.Write(w.f, stdbin.LittleEndian, w.dataBytes); err != nil {
		_ = w.f.Close()
		w.f = nil
		return err
	}

	err := w.f.Close()
	w.f = nil
	return err
}

func extractRTPPayload(pkt []byte) ([]byte, int, bool) {
	if len(pkt) < 12 {
		return nil, -1, false
	}
	version := (pkt[0] >> 6) & 0x3
	if version != 2 {
		return nil, -1, false
	}
	padding := (pkt[0]>>5)&0x1 == 1
	extension := (pkt[0]>>4)&0x1 == 1
	csrcCount := int(pkt[0] & 0x0F)
	pt := int(pkt[1] & 0x7F)

	idx := 12 + 4*csrcCount
	if len(pkt) < idx {
		return nil, pt, false
	}
	if extension {
		if len(pkt) < idx+4 {
			return nil, pt, false
		}
		extLenWords := int(stdbin.BigEndian.Uint16(pkt[idx+2 : idx+4]))
		idx += 4 + 4*extLenWords
		if len(pkt) < idx {
			return nil, pt, false
		}
	}

	end := len(pkt)
	if padding {
		padLen := int(pkt[len(pkt)-1])
		if padLen <= 0 || padLen > end-idx {
			return nil, pt, false
		}
		end -= padLen
	}
	if end <= idx {
		return nil, pt, false
	}
	return pkt[idx:end], pt, true
}

func decodePCMUToPCM16(pcmu []byte) []int16 {
	if len(pcmu) == 0 {
		return nil
	}
	out := make([]int16, len(pcmu))
	for i, b := range pcmu {
		out[i] = muLawDecode(b)
	}
	return out
}

func muLawDecode(uLaw byte) int16 {
	u := ^uLaw
	sign := u & 0x80
	exponent := (u >> 4) & 0x07
	mantissa := u & 0x0F

	sample := ((int(mantissa) << 3) + 0x84) << exponent
	sample -= 0x84
	if sign != 0 {
		sample = -sample
	}
	return int16(sample)
}

// forwardRealRTPToWhatsApp reenvÃ­a RTP real recebido do VoIP para o dispositivo WhatsApp
func (cm *WhatsmeowCallManager) forwardRealRTPToWhatsApp(from types.JID, callID string, rtpData []byte, sourceAddr string) {
	cm.logger.Infof("ðŸŽµðŸ“¤ [RTP-FORWARD-REAL] Reenviando %d bytes de %s para WhatsApp", len(rtpData), sourceAddr)

	// TODO: Implementar forwarding RTP real para WhatsApp
	// O RTP deve ser enviado diretamente para o dispositivo WhatsApp
	// usando o endereÃ§o descoberto via ICE/STUN

	// AnÃ¡lise bÃ¡sica do header RTP
	if len(rtpData) >= 12 {
		version := (rtpData[0] >> 6) & 0x3
		payloadType := rtpData[1] & 0x7F
		sequenceNumber := (uint16(rtpData[2]) << 8) | uint16(rtpData[3])

		cm.logger.Infof("ðŸŽµðŸ” [RTP-ANALYSIS] Version=%d, PayloadType=%d, Seq=%d", version, payloadType, sequenceNumber)
	}

	cm.logger.Infof("ðŸŽµâœ… [RTP-FORWARD-SUCCESS] RTP real reenviado para dispositivo WhatsApp")
}

// sendCodecPreferences envia preferÃªncias de codec
func (cm *WhatsmeowCallManager) sendCodecPreferences(from types.JID, callID string) {
	cm.logger.Infof("âœ…ðŸ“‹ [CODEC-PREF-SENT] PreferÃªncias de codec enviadas")
}

// sendAdvancedRTPConfig envia configuraÃ§Ãµes RTP avanÃ§adas
func (cm *WhatsmeowCallManager) sendAdvancedRTPConfig(from types.JID, callID string) {
	cm.logger.Infof("âœ…âš™ï¸ [RTP-ADVANCED-SENT] ParÃ¢metros RTP configurados")
}

// sendQualityOfService estabelece QoS
func (cm *WhatsmeowCallManager) sendQualityOfService(from types.JID, callID string) {
	cm.logger.Infof("âœ…ðŸ“¶ [QOS-SENT] Quality of Service estabelecido")
}

// GetSIPProxy retorna a integraÃ§Ã£o SIP
func (cm *WhatsmeowCallManager) GetSIPProxy() *SIPProxyIntegration {
	return cm.sipIntegration
}

// SetSIPIntegration define a integraÃ§Ã£o SIP
func (cm *WhatsmeowCallManager) SetSIPIntegration(integration *SIPProxyIntegration) {
	cm.sipIntegration = integration
}

// RejectCall rejeita uma chamada
func (cm *WhatsmeowCallManager) RejectCall(from types.JID, callID string) error {
	cm.logger.Infof("âŒ Rejeitando chamada de %v (CallID: %s)", from, callID)
	// Implementar lÃ³gica de rejeiÃ§Ã£o se necessÃ¡rio
	return nil
}

// HandleCallTransport manipula dados de transporte da chamada
func (cm *WhatsmeowCallManager) HandleCallTransport(from types.JID, callID string, transportData interface{}) error {
	if envTruthy("QP_CALL_OBSERVE_ONLY") {
		cm.logger.Warnf("ðŸšš [CALL] Observe-only enabled (QP_CALL_OBSERVE_ONLY=1): ignoring transport (CallID=%s)", callID)
		return nil
	}
	cm.logger.Infof("ðŸšš Processando transporte de chamada de %v (CallID: %s)", from, callID)

	// Log detalhado do node recebido para depuraÃ§Ã£o de ausÃªncia de media
	if node, ok := transportData.(*binary.Node); ok {
		cm.markRemoteTransportReceived(callID)
		if envTruthy("QP_CALL_TRANSPORT_SUMMARY") {
			full := envTruthy("QP_CALL_TRANSPORT_SUMMARY_FULL")
			s := summarizeCallTransportNode(node, full)
			cm.logger.Infof(
				"ðŸ§Š [TRANSPORT-SUMMARY] medium=%s protocol=%s ice.ufrag=%s ice.pwd=%s candidates=%d fingerprints=%d secrets=%d (CallID=%s)",
				s.NetMedium,
				s.NetProtocol,
				s.ICEUfrag,
				s.ICEPwd,
				len(s.Candidates),
				len(s.Fingerprints),
				len(s.SecretsFound),
				callID,
			)
			if full {
				for _, c := range s.Candidates {
					cm.logger.Infof("ðŸ§Š [TRANSPORT-CANDIDATE] %s (CallID=%s)", c, callID)
				}
				for _, fp := range s.Fingerprints {
					cm.logger.Infof("ðŸ§Š [TRANSPORT-FINGERPRINT] %s (CallID=%s)", fp, callID)
				}
				for _, sec := range s.SecretsFound {
					cm.logger.Infof("ðŸ§Š [TRANSPORT-SECRET] %s (CallID=%s)", sec, callID)
				}
			} else if envTruthy("QP_CALL_TRANSPORT_SUMMARY_DEBUG") {
				// Optional: show a few snapshots to help add exact extraction rules without flooding logs.
				for i := 0; i < len(s.AttrSnapshots) && i < 5; i++ {
					cm.logger.Infof("ðŸ§Š [TRANSPORT-SNAPSHOT] %s (CallID=%s)", s.AttrSnapshots[i], callID)
				}
			}
		}
		if children, okc := node.Content.([]binary.Node); okc {
			cm.logger.Infof("ðŸ§ª [TRANSPORT-RAW] Tag=%s Attrs=%v ChildCount=%d", node.Tag, node.Attrs, len(children))
			for i, c := range children {
				var subCount int
				if sc, oksc := c.Content.([]binary.Node); oksc {
					subCount = len(sc)
				}
				cm.logger.Infof("ðŸ§ª [TRANSPORT-CHILD-%d] Tag=%s Attrs=%v SubChildren=%d", i, c.Tag, c.Attrs, subCount)
			}
		} else {
			cm.logger.Infof("ðŸ§ª [TRANSPORT-RAW] Tag=%s Attrs=%v (sem children slice tipado)", node.Tag, node.Attrs)
		}
	}

	// NOVO FLUXO: Quando recebemos transport, respondemos com ACCEPT (nÃ£o transport)
	// FLUXO CORRETO: PREACCEPT â†’ eles enviam TRANSPORT â†’ nÃ³s enviamos ACCEPT
	if node, ok := transportData.(*binary.Node); ok {
		return cm.sendAcceptResponseToTransport(from, callID, node)
	}

	// Implementar lÃ³gica de transporte se necessÃ¡rio
	return nil
}

// sendAcceptResponseToTransport envia ACCEPT como resposta ao transport recebido
func (cm *WhatsmeowCallManager) sendAcceptResponseToTransport(from types.JID, callID string, receivedTransport *binary.Node) error {
	cm.logger.Infof("ðŸŽµ [ACCEPT-RESPONSE] === ENVIANDO ACCEPT COMO RESPOSTA AO TRANSPORT ===")
	cm.logger.Infof("ðŸŽ¯ðŸ“‹ [ACCEPT-INFO] From: %v, CallID: %s", from, callID)

	// If we already sent ACCEPT (direct/snippet mode), WhatsApp often expects us to send
	// a TRANSPORT message next with our media endpoint details (where to send audio).
	// In that case, do NOT send ACCEPT again. Send transport info instead.
	cm.hsMutex.Lock()
	if st, ok := cm.handshakeStates[callID]; ok && st.AcceptSent {
		cm.hsMutex.Unlock()
		cm.logger.Warnf("ðŸ§ª [ACCEPT-SKIP] ACCEPT jÃ¡ enviado anteriormente (modo direct/snippet). Enviando TRANSPORT em resposta ao transport remoto.")
		if err := cm.sendTransportInfoResponse(from, callID, receivedTransport); err != nil {
			cm.logger.Warnf("âš ï¸ðŸŽµðŸš› [TRANSPORT-RESPONSE] Falha ao enviar TRANSPORT em resposta ao transport remoto: %v (CallID=%s)", err, callID)
			return err
		}
		cm.logger.Infof("âœ…ðŸŽµðŸš› [TRANSPORT-RESPONSE] TRANSPORT enviado em resposta ao transport remoto (CallID=%s)", callID)
		return nil
	}
	cm.hsMutex.Unlock()

	// Extrair informaÃ§Ãµes do transport recebido para usar no accept
	callCreator := ""
	transportType := ""
	transportMedium := ""

	if attrs := receivedTransport.Attrs; attrs != nil {
		if creator, exists := attrs["call-creator"]; exists {
			callCreator = fmt.Sprintf("%v", creator)
		}
		if msgType, exists := attrs["transport-message-type"]; exists {
			transportType = fmt.Sprintf("%v", msgType)
		}
	}
	// Extract medium from <net medium='X'> child if present
	if transportMedium == "" {
		if children, ok := receivedTransport.Content.([]binary.Node); ok {
			for _, ch := range children {
				if ch.Tag != "net" {
					continue
				}
				if ch.Attrs != nil {
					if m, ok2 := ch.Attrs["medium"]; ok2 {
						transportMedium = fmt.Sprintf("%v", m)
					}
				}
				break
			}
		}
	}
	if transportMedium == "" {
		transportMedium = cm.getNetMedium()
	}

	cm.logger.Infof("ðŸŽ¯ðŸ” [ACCEPT-EXTRACT] CallCreator: %s, MessageType: %s", callCreator, transportType)
	cm.logger.Infof("ðŸŽ¯ðŸ’¡ [ACCEPT-THEORY] Received transport type %s, responding with ACCEPT", transportType)

	ownID := cm.connection.Client.Store.ID
	if ownID == nil {
		return fmt.Errorf("own ID not available")
	}

	// CRITICAL: Enviar ACCEPT como resposta ao transport (ao invÃ©s de outro transport)
	// Este Ã© o handshake final: PREACCEPT â†’ TRANSPORT (deles) â†’ ACCEPT (nosso)
	//
	// Relay-only nuance: when medium=2 and the peer transport has no ICE details/candidates,
	// sending a host candidate inside a relay net can keep the caller UI stuck.
	// We support an env-gated mode to send an empty <net> (no candidates) for relay calls.
	localIP := cm.getLocalNetworkIP()
	if localIP == "" {
		return fmt.Errorf("failed to determine local IPv4 address")
	}
	lockedPort := cm.getCallMediaPort(callID)
	publicIP, publicPort := cm.getCallPublicMapping(callID)
	port := lockedPort
	if port <= 0 {
		if cm.disableSTUN() {
			port = 64006
			cm.logger.Infof("ðŸ§ª [STUN-DISABLED] ACCEPT response usando porta estÃ¡tica %d", port)
		} else {
			var errStun error
			var localPort int
			publicIP, publicPort, localPort, errStun = cm.performSTUNDiscovery()
			if errStun != nil || publicIP == "" {
				port = 64006
			} else if localPort > 0 {
				port = localPort
			}
			cm.setCallMediaPort(callID, port)
			cm.setCallPublicMapping(callID, publicIP, publicPort)
		}
	}
	includeSrflx := os.Getenv("QP_CALL_INCLUDE_SRFLX") == "1"
	candidates := cm.buildCandidates(localIP, port, publicIP, publicPort, includeSrflx)
	if envTruthy("QP_CALL_ACCEPT_RELAY_EMPTY_NET") && strings.TrimSpace(transportMedium) == "2" {
		cm.logger.Warnf("ðŸ§Š [ACCEPT-RELAY-EMPTY-NET] Sending ACCEPT with empty net (no candidates) for relay call (CallID=%s)", callID)
		candidates = nil
	}
	replyTo := cm.callReplyJID(from)
	if callCreator == "" {
		callCreator = replyTo.String()
	}
	acceptResponseNode := cm.buildAcceptNodeWithMedium(replyTo, ownID.ToNonAD().String(), callID, callCreator, candidates, transportMedium)

	err := cm.connection.Client.DangerousInternals().SendNode(context.Background(), acceptResponseNode)
	if err != nil {
		return fmt.Errorf("failed to send accept response to transport: %w", err)
	}

	// Dump outgoing ACCEPT (response to transport)
	if os.Getenv("QP_CALL_DUMP_ACCEPT") == "1" {
		if p, e := DumpAcceptSent(callID, from, *ownID, acceptResponseNode); e == nil {
			cm.logger.Infof("ðŸ’¾ [ACCEPT-RESPONSE-DUMP] Saved to %s", p)
		} else {
			cm.logger.Warnf("âš ï¸ [ACCEPT-RESPONSE-DUMP] Failed: %v", e)
		}
	}

	cm.logger.Infof("âœ…ðŸŽ¯ [ACCEPT-RESPONSE-SENT] ACCEPT enviado como resposta ao transport!")
	cm.logger.Infof("ðŸŽ¯ðŸ“‹ [ACCEPT-FLOW] PREACCEPT â†’ TRANSPORT (recebido) â†’ ACCEPT (enviado)")
	cm.logger.Infof("ðŸŽ¯ðŸŽ‰ [ACCEPT-COMPLETE] Handshake completo: transport type %s respondido com ACCEPT", transportType)

	// Iniciar fake RTP se em modo META-only (para exercitar fluxo pÃ³s-handshake)
	if os.Getenv("QP_CALL_META_ONLY") == "1" {
		stop := make(chan struct{})
		go StartFakeRTP(callID, "127.0.0.1", 50000, stop, cm.logger) // porta arbitrÃ¡ria local
		cm.logger.Infof("ðŸ§ªðŸŽµ [FAKE-RTP-START] Gerador RTP falso iniciado (porta 50000)")
	}

	return nil
}

// AcceptDirectCall envia apenas o nÃ³ ACCEPT imediatamente (sem PREACCEPT ou TRANSPORT local)
// Usado para o cenÃ¡rio pedido: responder oferta de chamada apenas com ACCEPT e depois logar TRANSPORT remoto
func (cm *WhatsmeowCallManager) AcceptDirectCall(from types.JID, callID string) error {
	cm.logger.Infof("ðŸ“žâš¡ [DIRECT-ACCEPT] Respondendo CallOffer diretamente com ACCEPT (sem PREACCEPT)")

	// Optional: delay direct/snippet ACCEPT to avoid immediately taking ownership and getting
	// an instant hangup before relay latency / probes can run.
	// Clamp to 0..5000ms to keep behavior predictable.
	delayMS := 0
	if raw := strings.TrimSpace(os.Getenv("QP_CALL_DIRECT_ACCEPT_DELAY_MS")); raw != "" {
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
		cm.logger.Warnf("ðŸ•°ï¸ðŸ“ž [DIRECT-ACCEPT-DELAY] Delaying ACCEPT by %dms (CallID=%s)", delayMS, callID)
		time.Sleep(time.Duration(delayMS) * time.Millisecond)
	}

	ownID := cm.connection.Client.Store.ID
	if ownID == nil {
		return fmt.Errorf("own ID not available")
	}

	// Use the peer JID format we want to reply to (raw LID when enabled).
	targetJID := cm.callReplyJID(from)
	cm.logger.Infof("ðŸ“ž [ACCEPT-TARGET] fromRaw=%s replyTo=%s useLID=%v", from.String(), targetJID.String(), envTruthy("QP_CALL_REPLY_USE_LID"))

	// Inicializar estado e marcar AcceptSent
	cm.initHandshakeState(callID)
	cm.hsMutex.Lock()
	if st, ok := cm.handshakeStates[callID]; ok {
		st.AcceptSent = true
	}
	cm.hsMutex.Unlock()

	// If user requested the exact blog snippet accept, send that structure (no top-level 'from', medium=3)
	if envTruthy("QP_CALL_USE_SNIPPET_ACCEPT") {
		selectedMedium := cm.getNetMediumForCall(callID)
		cm.logger.Warnf("ðŸ§ª [SNIPPET-ACCEPT] Sending snippet-accept (no top-level from, medium=%s)", selectedMedium)
		snippetNode := binary.Node{
			Tag: "call",
			Attrs: binary.Attrs{
				"to": targetJID,
				"id": cm.connection.Client.GenerateMessageID(),
			},
			Content: []binary.Node{{
				Tag:   "accept",
				Attrs: binary.Attrs{"call-id": callID, "call-creator": targetJID},
				Content: []binary.Node{
					{Tag: "audio", Attrs: binary.Attrs{"enc": "opus", "rate": "16000"}},
					{Tag: "audio", Attrs: binary.Attrs{"enc": "opus", "rate": "8000"}},
					{Tag: "net", Attrs: binary.Attrs{"medium": selectedMedium}},
					{Tag: "encopt", Attrs: binary.Attrs{"keygen": "2"}},
				},
			}},
		}

		if err := cm.connection.Client.DangerousInternals().SendNode(context.Background(), snippetNode); err != nil {
			return fmt.Errorf("failed to send snippet accept: %w", err)
		}
		if os.Getenv("QP_CALL_DUMP_ACCEPT") == "1" {
			if p, e := DumpAcceptSent(callID, targetJID, *ownID, snippetNode); e == nil {
				cm.logger.Infof("ðŸ’¾ [SNIPPET-ACCEPT-DUMP] Saved to %s", p)
			} else {
				cm.logger.Warnf("âš ï¸ [SNIPPET-ACCEPT-DUMP] Failed: %v", e)
			}
		}
		cm.logger.Infof("âœ…âš¡ [SNIPPET-ACCEPT-SENT] Snippet ACCEPT sent to=%s (awaiting remote TRANSPORT)", targetJID.String())
		cm.MaybeStartRelaySessionProbe(callID)
		return nil
	}

	// Construir candidatos mÃ­nimos (host). Para consistÃªncia reutilizamos lÃ³gica existente
	localIP := cm.getLocalNetworkIP()
	if localIP == "" {
		return fmt.Errorf("failed to determine local IPv4 address")
	}
	lockedPort := cm.getCallMediaPort(callID)
	publicIP, publicPort := cm.getCallPublicMapping(callID)
	port := lockedPort
	if port <= 0 {
		if cm.disableSTUN() {
			port = 64006
			cm.logger.Infof("ðŸ§ª [STUN-DISABLED] DIRECT ACCEPT usando porta estÃ¡tica %d", port)
		} else {
			var errStun error
			var localPort int
			publicIP, publicPort, localPort, errStun = cm.performSTUNDiscovery()
			if errStun != nil || publicIP == "" {
				port = 64006
			} else if localPort > 0 {
				port = localPort
			}
			cm.setCallMediaPort(callID, port)
			cm.setCallPublicMapping(callID, publicIP, publicPort)
		}
	}
	includeSrflx := envTruthy("QP_CALL_INCLUDE_SRFLX")
	candidates := cm.buildCandidates(localIP, port, publicIP, publicPort, includeSrflx)
	acceptNode := cm.buildAcceptNodeWithMedium(targetJID, ownID.ToNonAD().String(), callID, targetJID.String(), candidates, cm.getNetMediumForCall(callID))

	if err := cm.connection.Client.DangerousInternals().SendNode(context.Background(), acceptNode); err != nil {
		return fmt.Errorf("failed to send direct accept: %w", err)
	}

	// Dump ACCEPT sent for debugging (captures exact node we sent)
	if os.Getenv("QP_CALL_DUMP_ACCEPT") == "1" {
		if dumpPath, dumpErr := DumpAcceptSent(callID, targetJID, *ownID, acceptNode); dumpErr == nil {
			cm.logger.Infof("ðŸ’¾ [ACCEPT-DUMP] Saved to %s", dumpPath)
		} else {
			cm.logger.Warnf("âš ï¸ [ACCEPT-DUMP] Failed: %v", dumpErr)
		}
	}

	cm.logger.Infof("âœ…âš¡ [DIRECT-ACCEPT-SENT] ACCEPT sent to=%s (awaiting remote TRANSPORT)", targetJID.String())
	cm.MaybeStartRelaySessionProbe(callID)
	return nil
}

// buildAcceptNodeWithMedium constrÃ³i nÃ³ ACCEPT completo padronizado com medium explÃ­cito.
func (cm *WhatsmeowCallManager) buildAcceptNodeWithMedium(to types.JID, fromNonAD string, callID string, callCreator string, candidates []binary.Node, medium string) binary.Node {
	// Ajuste: ordem dos nÃ³s replicando preaccept (audio,audio,net,encopt) para consistÃªncia
	node := binary.Node{Tag: "call", Attrs: binary.Attrs{"to": to, "from": fromNonAD, "id": cm.connection.Client.GenerateMessageID()}, Content: []binary.Node{{
		Tag:   "accept",
		Attrs: binary.Attrs{"call-id": callID, "call-creator": callCreator},
		Content: []binary.Node{
			{Tag: "audio", Attrs: binary.Attrs{"enc": "opus", "rate": "16000"}},
			{Tag: "audio", Attrs: binary.Attrs{"enc": "opus", "rate": "8000"}},
			{Tag: "net", Attrs: binary.Attrs{"medium": medium, "protocol": "0"}, Content: candidates},
			{Tag: "encopt", Attrs: binary.Attrs{"keygen": "2"}},
		},
	}}}
	// Log estruturado EXTENSO para debug
	if content, ok := node.Content.([]binary.Node); ok && len(content) > 0 {
		for i, c := range content[0].Content.([]binary.Node) {
			cm.logger.Infof("ðŸ§ª [ACCEPT-NODE-PART-%d] Tag=%s Attrs=%v", i, c.Tag, c.Attrs)
			if c.Tag == "net" {
				if candChildren, ok2 := c.Content.([]binary.Node); ok2 {
					for j, cand := range candChildren {
						cm.logger.Infof("   ðŸŒ [ACCEPT-CANDIDATE-%d] %v", j, cand.Attrs)
					}
				}
			}
		}
		cm.logger.Infof("ðŸ“¦ [ACCEPT-NODE-FULL]\n%s", cm.debugFormatNode(node))
	}
	return node
}

// buildAcceptNode constrÃ³i nÃ³ ACCEPT completo padronizado usando o medium configurado por env.
func (cm *WhatsmeowCallManager) buildAcceptNode(to types.JID, fromNonAD string, callID string, callCreator string, candidates []binary.Node) binary.Node {
	return cm.buildAcceptNodeWithMedium(to, fromNonAD, callID, callCreator, candidates, cm.getNetMediumForCall(callID))
}

// debugFormatNode cria uma representaÃ§Ã£o hierÃ¡rquica do node para inspecionar diferenÃ§as sutis de ordem/atributos
func (cm *WhatsmeowCallManager) debugFormatNode(n binary.Node) string {
	var buf bytes.Buffer
	var walk func(binary.Node, int)
	walk = func(cur binary.Node, level int) {
		indent := strings.Repeat("  ", level)
		// ordenar attrs para determinismo
		keys := make([]string, 0, len(cur.Attrs))
		for k := range cur.Attrs {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		attrParts := make([]string, 0, len(keys))
		for _, k := range keys {
			attrParts = append(attrParts, fmt.Sprintf("%s='%v'", k, cur.Attrs[k]))
		}
		buf.WriteString(fmt.Sprintf("%s<%s %s>\n", indent, cur.Tag, strings.Join(attrParts, " ")))
		if children, ok := cur.Content.([]binary.Node); ok {
			for _, ch := range children {
				walk(ch, level+1)
			}
		} else if cur.Content != nil {
			buf.WriteString(fmt.Sprintf("%s  %v\n", indent, cur.Content))
		}
		buf.WriteString(fmt.Sprintf("%s</%s>\n", indent, cur.Tag))
	}
	walk(n, 0)
	return buf.String()
}

