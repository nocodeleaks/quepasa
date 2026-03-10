package whatsmeow

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/pion/stun"
	"golang.org/x/crypto/curve25519"
)

type turnIntegrityCandidate struct {
	label    string
	username []byte
	key      []byte
	realm    []byte
	nonce    []byte
	longTerm bool
}

func summarizeRelayTE2IPv6(rb *RelayBlock) []string {
	if rb == nil || len(rb.TE2) == 0 {
		return nil
	}
	out := make([]string, 0, len(rb.TE2))
	seen := make(map[string]struct{}, len(rb.TE2))
	for _, te := range rb.TE2 {
		if te.PayloadLen != 18 || strings.TrimSpace(te.IPv6Prefix) == "" {
			continue
		}
		item := te.RelayName + "=" + te.IPv6Prefix
		if te.RelayTailHex != "" {
			item += "|tail=" + te.RelayTailHex
		}
		if te.Protocol != "" {
			item += "|p=" + te.Protocol
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	sort.Strings(out)
	return out
}

func buildRelayTE2MsgVariants(rb *RelayBlock) map[string][]byte {
	out := map[string][]byte{}
	if rb == nil || len(rb.TE2) == 0 {
		return out
	}
	add := func(label string, data []byte) {
		if len(data) == 0 {
			return
		}
		cp := append([]byte(nil), data...)
		out[label] = cp
	}
	for _, te := range rb.TE2 {
		if te.PayloadLen != 18 || len(te.Payload) != 18 {
			continue
		}
		base := "te2." + strings.TrimSpace(te.RelayName)
		add(base+".full18", te.Payload)
		add(base+".prefix8", te.Payload[:8])
		add(base+".marker4", te.Payload[8:12])
		add(base+".tail4", te.Payload[12:16])
		add(base+".suffix2", te.Payload[16:18])
		add(base+".prefix8+tail4", appendParts(te.Payload[:8], te.Payload[12:16]))
	}
	return out
}

func dedupeTurnIntegrityCandidates(in []turnIntegrityCandidate) []turnIntegrityCandidate {
	if len(in) == 0 {
		return in
	}
	// Dedupe by a stable hash of all auth-relevant bytes. This avoids repeating identical
	// Allocate attempts when relay TE2 blocks contain duplicates.
	seen := make(map[string]struct{}, len(in))
	out := make([]turnIntegrityCandidate, 0, len(in))
	for _, c := range in {
		h := sha1.New()
		_, _ = h.Write([]byte(c.label))
		_, _ = h.Write([]byte{0})
		_, _ = h.Write(c.username)
		_, _ = h.Write([]byte{0})
		_, _ = h.Write(c.key)
		_, _ = h.Write([]byte{0})
		_, _ = h.Write(c.realm)
		_, _ = h.Write([]byte{0})
		_, _ = h.Write(c.nonce)
		_, _ = h.Write([]byte{0})
		if c.longTerm {
			_, _ = h.Write([]byte{1})
		} else {
			_, _ = h.Write([]byte{0})
		}
		k := hex.EncodeToString(h.Sum(nil))
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, c)
	}
	return out
}

func addMessageIntegritySHA256(m *stun.Message, key []byte) error {
	// Minimal MESSAGE-INTEGRITY-SHA256 support (RFC 8489).
	// Similar to pion/stun MessageIntegrity.AddTo (SHA1) but using SHA256 and AttrMessageIntegritySHA256.
	if m == nil || len(key) == 0 {
		return fmt.Errorf("invalid message or key")
	}
	for _, a := range m.Attributes {
		if a.Type == stun.AttrFingerprint {
			return fmt.Errorf("fingerprint before integrity")
		}
	}
	const (
		miSHA256Size = 32
		attrHdrSize  = 4
	)
	length := m.Length
	m.Length += miSHA256Size + attrHdrSize
	m.WriteLength()
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write(m.Raw)
	v := mac.Sum(nil)
	m.Length = length
	m.WriteLength()
	m.Add(stun.AttrMessageIntegritySHA256, v)
	return nil
}

type seedMsgVariant struct {
	label string
	data  []byte
}

type labeledSeed struct {
	label string
	data  []byte
}

func hmacSHA1(key []byte, msg []byte) []byte {
	h := hmac.New(sha1.New, key)
	_, _ = h.Write(msg)
	return h.Sum(nil)
}

func hmacSHA256(key []byte, msg []byte) []byte {
	h := hmac.New(sha256.New, key)
	_, _ = h.Write(msg)
	return h.Sum(nil)
}

func sha1Sum(msg []byte) []byte {
	s := sha1.Sum(msg)
	return s[:]
}

func appendParts(parts ...[]byte) []byte {
	total := 0
	for _, p := range parts {
		total += len(p)
	}
	if total == 0 {
		return nil
	}
	out := make([]byte, 0, total)
	for _, p := range parts {
		if len(p) == 0 {
			continue
		}
		out = append(out, p...)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func sha256Sum(msg []byte) []byte {
	s := sha256.Sum256(msg)
	return s[:]
}

func tryAppendTurnRESTCandidates(cands *[]turnIntegrityCandidate, labelPrefix string, username []byte, secret []byte) {
	// TURN REST (common pattern): password = base64(hmac-sha1(secret, username))
	// and MESSAGE-INTEGRITY uses the password bytes as key.
	if len(username) == 0 || len(secret) == 0 {
		return
	}
	h := hmacSHA1(secret, username)
	passB64 := []byte(base64.StdEncoding.EncodeToString(h))
	appendCand(cands, labelPrefix+":pass=hmacb64", username, passB64)
	// Also try raw hmac bytes (some deployments may not base64-encode the derived password).
	appendCand(cands, labelPrefix+":pass=hmacraw", username, h)
}

func md5Sum(msg []byte) []byte {
	s := md5.Sum(msg)
	return s[:]
}

func turnLongTermKey(username []byte, realm []byte, password []byte) []byte {
	// TURN long-term credential key is MD5(username ":" realm ":" password)
	// (RFC 5389 / RFC 5766)
	buf := make([]byte, 0, len(username)+len(realm)+len(password)+2)
	buf = append(buf, username...)
	buf = append(buf, ':')
	buf = append(buf, realm...)
	buf = append(buf, ':')
	buf = append(buf, password...)
	return md5Sum(buf)
}

func appendMD5RealmEmptyCandidates(cands *[]turnIntegrityCandidate, labelPrefix string, username []byte, password []byte) {
	// Some TURN deployments may derive the short-term key via MD5(username:realm:password)
	// but still behave like short-term auth (no REALM/NONCE attributes exposed).
	if len(username) == 0 || len(password) == 0 {
		return
	}
	key := turnLongTermKey(username, []byte(""), password)
	appendCand(cands, labelPrefix+":lt0=md5(user::pass)", username, key)
}

type labeledBytes struct {
	label string
	data  []byte
}

func buildRelayUsernameVariants(rb *RelayBlock) []labeledBytes {
	if rb == nil {
		return nil
	}
	uuid := strings.TrimSpace(rb.UUID)
	self := strings.TrimSpace(rb.SelfPID)
	peer := strings.TrimSpace(rb.PeerPID)

	variants := make([]labeledBytes, 0, 18)
	add := func(label string, s string) {
		s = strings.TrimSpace(s)
		if s == "" {
			return
		}
		variants = append(variants, labeledBytes{label: label, data: []byte(s)})
	}
	addBytes := func(label string, b []byte) {
		if len(b) == 0 {
			return
		}
		bb := make([]byte, len(b))
		copy(bb, b)
		variants = append(variants, labeledBytes{label: label, data: bb})
	}
	joinBytes := func(a []byte, sep string, b []byte) []byte {
		if len(a) == 0 {
			return nil
		}
		out := make([]byte, 0, len(a)+len(sep)+len(b))
		out = append(out, a...)
		out = append(out, []byte(sep)...)
		out = append(out, b...)
		return out
	}
	decodeUUIDBase64 := func(s string) []byte {
		s = strings.TrimSpace(s)
		if s == "" {
			return nil
		}
		// Offers have shown UUID strings that look like base64url (e.g. containing '_' or '-').
		// Try both std and URL alphabets, with and without padding.
		encsRaw := []*base64.Encoding{base64.RawStdEncoding, base64.RawURLEncoding}
		for _, enc := range encsRaw {
			if b, err := enc.DecodeString(s); err == nil && len(b) > 0 {
				return b
			}
		}
		pad := ""
		switch len(s) % 4 {
		case 2:
			pad = "=="
		case 3:
			pad = "="
		case 0:
			pad = ""
		default:
			pad = ""
		}
		sp := s + pad
		encs := []*base64.Encoding{base64.StdEncoding, base64.URLEncoding}
		for _, enc := range encs {
			if b, err := enc.DecodeString(sp); err == nil && len(b) > 0 {
				return b
			}
		}
		return nil
	}

	add("uuid", uuid)
	add("self", self)
	add("peer", peer)
	if uuid != "" {
		if uuidBin := decodeUUIDBase64(uuid); len(uuidBin) > 0 {
			addBytes("uuid(bin)", uuidBin)
			if self != "" {
				addBytes("uuid(bin):self", joinBytes(uuidBin, ":", []byte(self)))
				addBytes("self:uuid(bin)", joinBytes([]byte(self), ":", uuidBin))
			}
			if peer != "" {
				addBytes("uuid(bin):peer", joinBytes(uuidBin, ":", []byte(peer)))
				addBytes("peer:uuid(bin)", joinBytes([]byte(peer), ":", uuidBin))
			}
		}
	}
	if self != "" && peer != "" {
		add("self:peer", self+":"+peer)
		add("peer:self", peer+":"+self)
	}
	if uuid != "" && self != "" {
		add("uuid:self", uuid+":"+self)
		add("self:uuid", self+":"+uuid)
	}
	if uuid != "" && peer != "" {
		add("uuid:peer", uuid+":"+peer)
		add("peer:uuid", peer+":"+uuid)
	}

	// Deduplicate.
	seen := map[string]bool{}
	out := make([]labeledBytes, 0, len(variants))
	for _, v := range variants {
		k := string(v.data)
		if seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, v)
	}
	return out
}

// MaybeStartRelaySessionProbe starts a minimal relay session probe toward relay endpoints learned
// from CallRelayLatency. This is intended as the first step toward relay/SRTP media-plane work:
// validate egress UDP reachability and capture any inbound packets.
//
// Env flags:
//   - QP_CALL_RELAY_SESSION_PROBE=1 enables
//   - QP_CALL_RELAY_SESSION_STUN_TIMEOUT_MS (default 900, clamp 100..5000) // reused as UDP read timeout
//   - QP_CALL_RELAY_TURN_ALLOCATE_TRY_INTEGRITY=1 enables MESSAGE-INTEGRITY attempts
func (cm *WhatsmeowCallManager) MaybeStartRelaySessionProbe(callID string) {
	if cm == nil || callID == "" {
		return
	}
	if !envTruthy("QP_CALL_RELAY_SESSION_PROBE") {
		return
	}
	endpoints := cm.getRelayEndpoints(callID)
	if len(endpoints) == 0 {
		// RelayLatency events may arrive slightly after ACCEPT; wait a bit once per call.
		if !cm.markRelaySessionProbePending(callID) {
			return
		}
		cm.logger.Infof("⏳📡 [RELAY-SESSION] Waiting for relay endpoints... (CallID=%s)", callID)
		go func() {
			deadline := time.Now().Add(6 * time.Second)
			for time.Now().Before(deadline) {
				if !cm.hasCallState(callID) {
					return
				}
				if eps := cm.getRelayEndpoints(callID); len(eps) > 0 {
					cm.MaybeStartRelaySessionProbe(callID)
					return
				}
				time.Sleep(250 * time.Millisecond)
			}
			if !cm.hasCallState(callID) {
				return
			}
			if fallback := relaySessionFallbackEndpointsFromEnv(); len(fallback) > 0 {
				cm.logger.Warnf("⚠️📡 [RELAY-SESSION] No relay endpoints after wait window; using env fallback endpoints=%d (CallID=%s)", len(fallback), callID)
				cm.forceStartRelaySessionProbe(callID, fallback)
				return
			}
			cm.logger.Warnf("⚠️📡 [RELAY-SESSION] No relay endpoints after wait window (CallID=%s)", callID)
		}()
		return
	}

	best := pickBestRelayEndpoint(endpoints)
	if best.Endpoint == "" {
		cm.logger.Warnf("⚠️📡 [RELAY-SESSION] Best relay endpoint empty (CallID=%s)", callID)
		return
	}
	if !cm.markRelaySessionProbeStarted(callID, best.Endpoint) {
		return
	}
	if te2 := summarizeRelayTE2IPv6(cm.getRelayBlock(callID)); len(te2) > 0 {
		cm.logger.Infof("📡 [RELAY-SESSION] TE2 IPv6 hints: callID=%s best=%s hints=%v", callID, best.Endpoint, te2)
	}

	go cm.runRelaySessionProbe(callID, best, endpoints)
}

func (cm *WhatsmeowCallManager) forceStartRelaySessionProbe(callID string, endpoints []RelayEndpoint) {
	if cm == nil || callID == "" {
		return
	}
	if len(endpoints) == 0 {
		return
	}
	best := pickBestRelayEndpoint(endpoints)
	if best.Endpoint == "" {
		cm.logger.Warnf("⚠️📡 [RELAY-SESSION] Best relay endpoint empty (fallback) (CallID=%s)", callID)
		return
	}
	if !cm.markRelaySessionProbeStarted(callID, best.Endpoint) {
		return
	}
	if te2 := summarizeRelayTE2IPv6(cm.getRelayBlock(callID)); len(te2) > 0 {
		cm.logger.Infof("📡 [RELAY-SESSION] TE2 IPv6 hints (fallback): callID=%s best=%s hints=%v", callID, best.Endpoint, te2)
	}
	go cm.runRelaySessionProbe(callID, best, endpoints)
}

func relaySessionFallbackEndpointsFromEnv() []RelayEndpoint {
	raw := strings.TrimSpace(os.Getenv("QP_CALL_RELAY_SESSION_FALLBACK_ENDPOINTS"))
	if raw == "" {
		return nil
	}
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';' || r == ' ' || r == '\n' || r == '\r' || r == '\t'
	})
	out := make([]RelayEndpoint, 0, len(parts))
	seen := map[string]bool{}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		host, portStr, err := net.SplitHostPort(p)
		if err != nil {
			continue
		}
		port, err := strconv.Atoi(portStr)
		if err != nil || port <= 0 || port > 65535 {
			continue
		}
		ep := net.JoinHostPort(host, strconv.Itoa(port))
		if seen[ep] {
			continue
		}
		seen[ep] = true
		out = append(out, RelayEndpoint{RelayName: "env", Endpoint: ep, LatencyRaw: ""})
	}
	return out
}

func pickBestRelayEndpoint(endpoints []RelayEndpoint) RelayEndpoint {
	if len(endpoints) == 0 {
		return RelayEndpoint{}
	}
	copyList := make([]RelayEndpoint, 0, len(endpoints))
	copyList = append(copyList, endpoints...)
	// Prefer lower latency when parseable.
	sort.SliceStable(copyList, func(i, j int) bool {
		li, okI := parseLatencyMs(copyList[i].LatencyRaw)
		lj, okJ := parseLatencyMs(copyList[j].LatencyRaw)
		if okI && okJ {
			return li < lj
		}
		if okI != okJ {
			return okI
		}
		// Otherwise deterministic: by relay name then endpoint.
		if copyList[i].RelayName != copyList[j].RelayName {
			return copyList[i].RelayName < copyList[j].RelayName
		}
		return copyList[i].Endpoint < copyList[j].Endpoint
	})
	return copyList[0]
}

func parseLatencyMs(raw string) (int, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, false
	}
	// Accept "25", "25ms", "25.0".
	raw = strings.TrimSuffix(raw, "ms")
	if strings.Contains(raw, ".") {
		parts := strings.SplitN(raw, ".", 2)
		raw = parts[0]
	}
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, false
	}
	if v < 0 {
		v = 0
	}
	return v, true
}

func clampInt(v, minV, maxV int) int {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}

func envInt(name string, def int) int {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return v
}

func envBytes(name string) ([]byte, bool) {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return nil, false
	}
	// Allow explicit prefixes to avoid ambiguity.
	// Note: hex-only strings are also valid base64 alphabet, so we MUST prefer hex when it matches.
	lower := strings.ToLower(raw)
	switch {
	case strings.HasPrefix(lower, "hex:"):
		b, ok := decodeMaybeHex(strings.TrimSpace(raw[4:]))
		return b, ok
	case strings.HasPrefix(lower, "b64:"):
		b, ok := decodeMaybeBase64(strings.TrimSpace(raw[4:]))
		return b, ok
	case strings.HasPrefix(lower, "raw:"):
		v := strings.TrimSpace(raw[4:])
		if v == "" {
			return nil, false
		}
		return []byte(v), true
	}

	// Prefer hex when it matches (otherwise hex would be mis-decoded as base64).
	if b, ok := decodeMaybeHex(raw); ok {
		return b, true
	}
	if b, ok := decodeMaybeBase64(raw); ok {
		return b, true
	}
	// Fallback: treat as raw ASCII bytes.
	return []byte(raw), true
}

type turnExtraAttrs struct {
	a4000           []byte
	a4024           []byte
	a0016           []byte
	a4000ByEndpoint map[string][]byte
}

func decodeFlexibleBytes(raw string) ([]byte, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, false
	}
	lower := strings.ToLower(raw)
	switch {
	case strings.HasPrefix(lower, "hex:"):
		return decodeMaybeHex(strings.TrimSpace(raw[4:]))
	case strings.HasPrefix(lower, "b64:"):
		return decodeMaybeBase64(strings.TrimSpace(raw[4:]))
	case strings.HasPrefix(lower, "raw:"):
		v := strings.TrimSpace(raw[4:])
		if v == "" {
			return nil, false
		}
		return []byte(v), true
	}
	// Same preference as envBytes(): hex first, then base64, else raw.
	if b, ok := decodeMaybeHex(raw); ok {
		return b, true
	}
	if b, ok := decodeMaybeBase64(raw); ok {
		return b, true
	}
	return []byte(raw), true
}

func parseEndpointBytesMapEnv(name string) map[string][]byte {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return nil
	}

	// Preferred: JSON object mapping endpoint -> value
	// Example: {"170.150.237.35:3478":"hex:...","57.144.179.54:3478":"hex:..."}
	var js map[string]string
	if err := json.Unmarshal([]byte(raw), &js); err == nil && len(js) > 0 {
		out := make(map[string][]byte, len(js))
		for ep, v := range js {
			ep = strings.TrimSpace(ep)
			if ep == "" {
				continue
			}
			if b, ok := decodeFlexibleBytes(v); ok && len(b) > 0 {
				out[ep] = b
			}
		}
		if len(out) == 0 {
			return nil
		}
		return out
	}

	// Fallback: semicolon-separated list: endpoint=value;endpoint=value
	parts := strings.Split(raw, ";")
	out := make(map[string][]byte, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		kv := strings.SplitN(p, "=", 2)
		if len(kv) != 2 {
			continue
		}
		ep := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])
		if ep == "" || val == "" {
			continue
		}
		if b, ok := decodeFlexibleBytes(val); ok && len(b) > 0 {
			out[ep] = b
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func buildTurnXorAddrAttr0016(remote *net.UDPAddr) ([]byte, bool) {
	if remote == nil {
		return nil, false
	}
	ip4 := remote.IP.To4()
	if ip4 == nil {
		return nil, false
	}
	if remote.Port <= 0 || remote.Port > 65535 {
		return nil, false
	}

	// Observed WhatsApp Desktop Allocate packets include attr 0x0016 as an XOR-ADDRESS-like IPv4 tuple.
	// Format (8 bytes): 0x00 0x01 | X-Port(2) | X-Addr(4)
	// X-Port = port ^ 0x2112 ; X-Addr = ipv4_u32 ^ 0x2112A442
	b := make([]byte, 8)
	b[0] = 0
	b[1] = 0x01
	xPort := uint16(remote.Port) ^ 0x2112
	binary.BigEndian.PutUint16(b[2:4], xPort)
	ipU32 := binary.BigEndian.Uint32(ip4)
	xAddr := ipU32 ^ 0x2112A442
	binary.BigEndian.PutUint32(b[4:8], xAddr)
	return b, true
}

func readTurnExtraAttrsFromEnv() turnExtraAttrs {
	var out turnExtraAttrs
	if b, ok := envBytes("QP_CALL_RELAY_TURN_ATTR_4000"); ok {
		out.a4000 = b
	}
	if b, ok := envBytes("QP_CALL_RELAY_TURN_ATTR_4024"); ok {
		out.a4024 = b
	}
	if b, ok := envBytes("QP_CALL_RELAY_TURN_ATTR_0016"); ok {
		out.a0016 = b
	}
	out.a4000ByEndpoint = parseEndpointBytesMapEnv("QP_CALL_RELAY_TURN_ATTR_4000_BY_ENDPOINT")
	return out
}

func addTurnExtraAttrs(req *stun.Message, extra turnExtraAttrs) {
	if req == nil {
		return
	}
	if len(extra.a4000) > 0 {
		req.Add(stun.AttrType(0x4000), extra.a4000)
	}
	if len(extra.a4024) > 0 {
		req.Add(stun.AttrType(0x4024), extra.a4024)
	}
	if len(extra.a0016) > 0 {
		req.Add(stun.AttrType(0x0016), extra.a0016)
	}
}

func addTurnExtraAttrsForEndpoint(req *stun.Message, extra turnExtraAttrs, remote *net.UDPAddr) {
	if req == nil {
		return
	}
	if len(extra.a4000ByEndpoint) > 0 && remote != nil {
		key := net.JoinHostPort(remote.IP.String(), strconv.Itoa(remote.Port))
		if b, ok := extra.a4000ByEndpoint[key]; ok && len(b) > 0 {
			req.Add(stun.AttrType(0x4000), b)
		} else if len(extra.a4000) > 0 {
			req.Add(stun.AttrType(0x4000), extra.a4000)
		}
	} else if len(extra.a4000) > 0 {
		req.Add(stun.AttrType(0x4000), extra.a4000)
	}
	if len(extra.a4024) > 0 {
		req.Add(stun.AttrType(0x4024), extra.a4024)
	}
	if len(extra.a0016) > 0 {
		req.Add(stun.AttrType(0x0016), extra.a0016)
		return
	}
	if envTruthy("QP_CALL_RELAY_TURN_AUTO_ATTR_0016") {
		if b, ok := buildTurnXorAddrAttr0016(remote); ok {
			req.Add(stun.AttrType(0x0016), b)
		}
	}
}

func resolveTurnExtraAttrsForEndpoint(extra turnExtraAttrs, remote *net.UDPAddr) (a4000, a4024, a0016 []byte) {
	if len(extra.a4000ByEndpoint) > 0 && remote != nil {
		key := net.JoinHostPort(remote.IP.String(), strconv.Itoa(remote.Port))
		if b, ok := extra.a4000ByEndpoint[key]; ok && len(b) > 0 {
			a4000 = b
		} else if len(extra.a4000) > 0 {
			a4000 = extra.a4000
		}
	} else if len(extra.a4000) > 0 {
		a4000 = extra.a4000
	}
	if len(extra.a4024) > 0 {
		a4024 = extra.a4024
	}
	if len(extra.a0016) > 0 {
		a0016 = extra.a0016
	} else if envTruthy("QP_CALL_RELAY_TURN_AUTO_ATTR_0016") {
		if b, ok := buildTurnXorAddrAttr0016(remote); ok {
			a0016 = b
		}
	}
	return a4000, a4024, a0016
}

func stunTLVBytes(attrType uint16, value []byte) []byte {
	if len(value) == 0 {
		return nil
	}
	b := make([]byte, 0, 4+len(value)+3)
	hdr := make([]byte, 4)
	binary.BigEndian.PutUint16(hdr[0:2], attrType)
	binary.BigEndian.PutUint16(hdr[2:4], uint16(len(value)))
	b = append(b, hdr...)
	b = append(b, value...)
	if pad := (4 - (len(value) % 4)) % 4; pad != 0 {
		b = append(b, make([]byte, pad)...)
	}
	return b
}

func buildExtraAttrsMsgVariants(a4000, a4024, a0016 []byte) map[string][]byte {
	out := map[string][]byte{}
	if len(a4024) > 0 {
		out["extra4024"] = a4024
		out["tlv4024"] = stunTLVBytes(0x4024, a4024)
	}
	// Values concatenation in a stable order.
	valuesAll := make([]byte, 0, len(a4000)+len(a4024)+len(a0016))
	if len(a4000) > 0 {
		valuesAll = append(valuesAll, a4000...)
		out["extra4000"] = a4000
		out["tlv4000"] = stunTLVBytes(0x4000, a4000)
	}
	if len(a4024) > 0 {
		valuesAll = append(valuesAll, a4024...)
	}
	if len(a0016) > 0 {
		valuesAll = append(valuesAll, a0016...)
		out["extra0016"] = a0016
		out["tlv0016"] = stunTLVBytes(0x0016, a0016)
	}
	if len(valuesAll) > 0 {
		out["extra4000+4024+0016"] = valuesAll
	}

	// TLV concatenation in the same stable order.
	tlvsAll := make([]byte, 0, 0)
	if len(a4000) > 0 {
		tlvsAll = append(tlvsAll, stunTLVBytes(0x4000, a4000)...)
	}
	if len(a4024) > 0 {
		tlvsAll = append(tlvsAll, stunTLVBytes(0x4024, a4024)...)
	}
	if len(a0016) > 0 {
		tlvsAll = append(tlvsAll, stunTLVBytes(0x0016, a0016)...)
	}
	if len(tlvsAll) > 0 {
		out["tlv4000+4024+0016"] = tlvsAll
	}

	return out
}

func buildAllocatePreimageMsgVariants(extra turnExtraAttrs, remote *net.UDPAddr) map[string][]byte {
	// Allocate preimage = request bytes after Encode() and before MESSAGE-INTEGRITY/FINGERPRINT.
	// Desktop uses MI without USERNAME/REALM/NONCE, so the relay may derive the short-term key from
	// the exact Allocate payload shape.
	out := map[string][]byte{}
	if remote == nil {
		return out
	}

	req := stun.MustBuild(stun.TransactionID)
	req.Type = stun.NewType(stun.Method(0x003), stun.ClassRequest)
	if !envTruthy("QP_CALL_RELAY_TURN_OMIT_REQUESTED_TRANSPORT") {
		req.Add(stun.AttrType(0x0019), []byte{17, 0, 0, 0})
	}
	addTurnExtraAttrsForEndpoint(req, extra, remote)
	req.Encode()
	if len(req.Raw) == 0 {
		return out
	}
	pre := append([]byte(nil), req.Raw...)
	out["alloc.preimage"] = pre
	out["alloc.preimage_sha1"] = sha1Sum(pre)
	out["alloc.preimage_sha256"] = sha256Sum(pre)

	// Variant: patch header length as if MI (24 bytes) was appended.
	// This matches Desktop's Allocate header length-field when MI is present.
	if len(pre) >= 4 {
		patched := append([]byte(nil), pre...)
		baseLen := int(binary.BigEndian.Uint16(patched[2:4]))
		patchedLen := baseLen + 24
		if patchedLen <= 0xffff {
			binary.BigEndian.PutUint16(patched[2:4], uint16(patchedLen))
			out["alloc.preimage_lenplus24"] = patched
			out["alloc.preimage_lenplus24_sha1"] = sha1Sum(patched)
			out["alloc.preimage_lenplus24_sha256"] = sha256Sum(patched)
		}
	}

	return out
}

func buildRelayContextMsgVariants(rb *RelayBlock, extraMsgs map[string][]byte) map[string][]byte {
	out := map[string][]byte{}
	if rb == nil {
		return out
	}

	uuidText := strings.TrimSpace(rb.UUID)
	selfText := strings.TrimSpace(rb.SelfPID)
	peerText := strings.TrimSpace(rb.PeerPID)

	var uuidBin []byte
	for _, u := range buildRelayUsernameVariants(rb) {
		if u.label == "uuid(bin)" {
			uuidBin = append([]byte(nil), u.data...)
			break
		}
	}

	add := func(label string, parts ...[]byte) {
		total := 0
		for _, p := range parts {
			total += len(p)
		}
		if total == 0 {
			return
		}
		buf := make([]byte, 0, total)
		for _, p := range parts {
			if len(p) == 0 {
				continue
			}
			buf = append(buf, p...)
		}
		if len(buf) == 0 {
			return
		}
		out[label] = buf
	}

	uuidTextBytes := []byte(uuidText)
	selfTextBytes := []byte(selfText)
	peerTextBytes := []byte(peerText)

	add("ctx.uuid_text", uuidTextBytes)
	add("ctx.uuid_bin", uuidBin)
	add("ctx.self_text", selfTextBytes)
	add("ctx.peer_text", peerTextBytes)
	add("ctx.uuid_bin+self+peer", uuidBin, selfTextBytes, peerTextBytes)
	add("ctx.uuid_text+self+peer", uuidTextBytes, selfTextBytes, peerTextBytes)
	add("ctx.self+peer", selfTextBytes, peerTextBytes)
	add("ctx.peer+self", peerTextBytes, selfTextBytes)

	for msgLabel, msg := range extraMsgs {
		if len(msg) == 0 {
			continue
		}
		add("ctx.uuid_bin+"+msgLabel, uuidBin, msg)
		add("ctx.uuid_text+"+msgLabel, uuidTextBytes, msg)
		add("ctx.uuid_bin+self+peer+"+msgLabel, uuidBin, selfTextBytes, peerTextBytes, msg)
		add("ctx.uuid_text+self+peer+"+msgLabel, uuidTextBytes, selfTextBytes, peerTextBytes, msg)
		add("ctx.self+peer+"+msgLabel, selfTextBytes, peerTextBytes, msg)
	}

	return out
}

func parseSimpleProtoFields(raw []byte) map[string][]byte {
	return parseSimpleProtoFieldsWithPrefix(raw, "proto", 0)
}

func deriveX25519SharedSecret(priv *[32]byte, peerPub []byte) []byte {
	if priv == nil || len(peerPub) != 32 {
		return nil
	}
	shared, err := curve25519.X25519(priv[:], peerPub)
	if err != nil || len(shared) == 0 {
		return nil
	}
	return append([]byte(nil), shared...)
}

func parseSimpleProtoFieldsWithPrefix(raw []byte, prefix string, depth int) map[string][]byte {
	out := map[string][]byte{}
	if len(raw) == 0 || depth > 2 {
		return out
	}
	readVarint := func(data []byte, idx int) (uint64, int, bool) {
		var v uint64
		var shift uint
		for i := 0; i < 10 && idx < len(data); i++ {
			b := data[idx]
			idx++
			v |= uint64(b&0x7f) << shift
			if b < 0x80 {
				return v, idx, true
			}
			shift += 7
		}
		return 0, idx, false
	}
	idx := 0
	for idx < len(raw) {
		key, next, ok := readVarint(raw, idx)
		if !ok {
			break
		}
		idx = next
		fieldNum := int(key >> 3)
		wireType := int(key & 0x7)
		if fieldNum <= 0 {
			break
		}
		base := fmt.Sprintf("%s.f%d", prefix, fieldNum)
		switch wireType {
		case 0:
			val, ni, ok := readVarint(raw, idx)
			if !ok {
				return out
			}
			idx = ni
			out[base+".varint"] = []byte(strconv.FormatUint(val, 10))
		case 2:
			n, ni, ok := readVarint(raw, idx)
			if !ok {
				return out
			}
			idx = ni
			if int(n) < 0 || idx+int(n) > len(raw) {
				return out
			}
			b := append([]byte(nil), raw[idx:idx+int(n)]...)
			idx += int(n)
			out[base] = b
			if len(b) == 32 {
				out[base+".sha256"] = sha256Sum(b)
			}
			if utf8.Valid(b) {
				out[base+".text"] = b
			}
			for k, v := range parseSimpleProtoFieldsWithPrefix(b, base, depth+1) {
				out[k] = v
			}
		default:
			return out
		}
	}
	return out
}

func isSTUNPacket(b []byte) bool {
	if len(b) < 8 {
		return false
	}
	// STUN magic cookie (RFC 5389)
	return b[4] == 0x21 && b[5] == 0x12 && b[6] == 0xA4 && b[7] == 0x42
}

func stunMsgTypeHexFromMessage(m *stun.Message) string {
	if m == nil {
		return ""
	}
	return stunMsgTypeHexFromRaw(m.Raw)
}

func extractMappedEndpointAnd4002(resp *stun.Message) (mapped string, a4002Hex string) {
	if resp == nil {
		return "", ""
	}
	var xorAddr stun.XORMappedAddress
	if err := xorAddr.GetFrom(resp); err == nil {
		mapped = net.JoinHostPort(xorAddr.IP.String(), strconv.Itoa(xorAddr.Port))
	}
	for _, a := range resp.Attributes {
		if uint16(a.Type) == 0x4002 && len(a.Value) > 0 {
			a4002Hex = strings.ToLower(hex.EncodeToString(a.Value))
			break
		}
	}
	return mapped, a4002Hex
}

func stunMsgTypeHexFromRaw(raw []byte) string {
	if len(raw) < 2 {
		return ""
	}
	return fmt.Sprintf("0x%02x%02x", raw[0], raw[1])
}

func stunTxIDHexFromRaw(raw []byte) string {
	// STUN header is 20 bytes, TXID at bytes 8..20.
	if len(raw) < 20 {
		return ""
	}
	return strings.ToLower(hex.EncodeToString(raw[8:20]))
}

func captureTurnRequestDump(stage string, raw []byte, preimage []byte) callTurnProbeRequestDump {
	out := callTurnProbeRequestDump{Stage: strings.TrimSpace(stage)}
	out.Len = len(raw)
	out.TxID = stunTxIDHexFromRaw(raw)
	out.MsgType = stunMsgTypeHexFromRaw(raw)
	if len(preimage) > 0 {
		out.PreimageHex = strings.ToLower(hex.EncodeToString(preimage))
	}
	if len(raw) == 0 {
		return out
	}
	out.RawHex = strings.ToLower(hex.EncodeToString(raw))

	// Decode attrs from raw (safer than relying on internal state after Encode/AddTo).
	var decoded stun.Message
	decoded.Raw = raw
	if err := decoded.Decode(); err == nil {
		attrs := make([]callTurnProbeRequestAttr, 0, len(decoded.Attributes))
		for _, a := range decoded.Attributes {
			attrs = append(attrs, callTurnProbeRequestAttr{
				Type: fmt.Sprintf("0x%04x", uint16(a.Type)),
				Len:  len(a.Value),
				Hex:  strings.ToLower(hex.EncodeToString(a.Value)),
			})
		}
		out.Attrs = attrs
	}

	return out
}

func (cm *WhatsmeowCallManager) runRelaySessionProbe(callID string, best RelayEndpoint, endpoints []RelayEndpoint) {
	stunTimeoutMS := clampInt(envInt("QP_CALL_RELAY_SESSION_STUN_TIMEOUT_MS", 900), 100, 5000)
	readTimeout := time.Duration(stunTimeoutMS) * time.Millisecond

	ordered := make([]RelayEndpoint, 0, len(endpoints))
	ordered = append(ordered, endpoints...)
	// Try best first, then others.
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Endpoint == best.Endpoint {
			return true
		}
		if ordered[j].Endpoint == best.Endpoint {
			return false
		}
		return ordered[i].Endpoint < ordered[j].Endpoint
	})

	cm.logger.Infof("📡 [RELAY-SESSION] Starting relay session probe (TURN Allocate): endpoints=%d readTimeoutMS=%d (CallID=%s)", len(ordered), stunTimeoutMS, callID)

	for _, ep := range ordered {
		if strings.TrimSpace(ep.Endpoint) == "" {
			continue
		}
		if err := cm.probeRelayEndpointTURNAllocate(callID, ep, readTimeout); err != nil {
			cm.logger.Warnf("⚠️📡 [RELAY-SESSION] relay=%s endpoint=%s probe failed: %v (CallID=%s)", ep.RelayName, ep.Endpoint, err, callID)
			continue
		}
		// If we got at least one response, don't spam all endpoints by default.
		break
	}
}

func looksLikeASCIIBase64(b []byte) bool {
	if len(b) < 8 {
		return false
	}
	for _, c := range b {
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			continue
		}
		switch c {
		case '+', '/', '=', '-', '_':
			continue
		default:
			return false
		}
	}
	return true
}

// decodeMaybeBase64Recursive tries to decode base64 multiple times.
// This is useful when the relay key is base64 of a base64 string (double-encoded).
func decodeMaybeBase64Recursive(value string, maxDepth int) ([]byte, int, bool) {
	cur := strings.TrimSpace(value)
	if cur == "" {
		return nil, 0, false
	}
	b, ok := decodeMaybeBase64(cur)
	if !ok {
		return nil, 0, false
	}
	depth := 1
	for depth < maxDepth {
		if !looksLikeASCIIBase64(b) {
			break
		}
		b2, ok2 := decodeMaybeBase64(string(b))
		if !ok2 {
			break
		}
		b = b2
		depth++
	}
	return b, depth, true
}

func decodeMaybeBase64(value string) ([]byte, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, false
	}
	// Try common base64 encodings:
	// - StdEncoding / RawStdEncoding
	// - URLEncoding / RawURLEncoding
	// Also tolerate missing padding.
	tryDecode := func(enc *base64.Encoding, s string) ([]byte, bool) {
		b, err := enc.DecodeString(s)
		if err != nil || len(b) == 0 {
			return nil, false
		}
		return b, true
	}

	encs := []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	}

	// First try as-is.
	for _, enc := range encs {
		if b, ok := tryDecode(enc, value); ok {
			return b, true
		}
	}

	// Try adding padding for encodings that expect it.
	padTo4 := func(s string) string {
		if m := len(s) % 4; m != 0 {
			return s + strings.Repeat("=", 4-m)
		}
		return s
	}
	valuePadded := padTo4(value)
	for _, enc := range encs {
		if b, ok := tryDecode(enc, valuePadded); ok {
			return b, true
		}
	}

	return nil, false
}

func decodeMaybeHex(value string) ([]byte, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, false
	}
	// Fast reject: must be even length and only hex chars.
	if len(value)%2 != 0 {
		return nil, false
	}
	for _, r := range value {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') {
			continue
		}
		return nil, false
	}
	b, err := hex.DecodeString(value)
	if err != nil || len(b) == 0 {
		return nil, false
	}
	return b, true
}

func appendCand(cands *[]turnIntegrityCandidate, label string, username []byte, key []byte) {
	if cands == nil {
		return
	}
	*cands = append(*cands, turnIntegrityCandidate{label: label, username: username, key: key})
}

func appendKeyVariants(cands *[]turnIntegrityCandidate, baseLabel string, username []byte, rawKey []byte) {
	// 1) raw bytes
	appendCand(cands, baseLabel+":key=raw", username, rawKey)
	// 2) base64 text of raw bytes
	appendCand(cands, baseLabel+":key=b64(text)", username, []byte(base64.StdEncoding.EncodeToString(rawKey)))
	// 3) hex text of raw bytes
	appendCand(cands, baseLabel+":key=hex(text)", username, []byte(strings.ToLower(hex.EncodeToString(rawKey))))
}

func appendKeyVariantsUserEncodings(cands *[]turnIntegrityCandidate, baseLabel string, rawUser []byte, rawKey []byte) {
	// Try USERNAME in several encodings (some relays may require ASCII-safe USERNAME).
	appendKeyVariants(cands, baseLabel+":uenc=raw", rawUser, rawKey)
	appendKeyVariants(cands, baseLabel+":uenc=b64(text)", []byte(base64.StdEncoding.EncodeToString(rawUser)), rawKey)
	appendKeyVariants(cands, baseLabel+":uenc=hex(text)", []byte(strings.ToLower(hex.EncodeToString(rawUser))), rawKey)
}

func appendKeyVariantsNoUser(cands *[]turnIntegrityCandidate, baseLabel string, rawKey []byte) {
	// Some servers may accept short-term integrity without USERNAME, or may select the key by other means.
	appendCand(cands, baseLabel+":user=none:key=raw", nil, rawKey)
	appendCand(cands, baseLabel+":user=none:key=b64(text)", nil, []byte(base64.StdEncoding.EncodeToString(rawKey)))
	appendCand(cands, baseLabel+":user=none:key=hex(text)", nil, []byte(strings.ToLower(hex.EncodeToString(rawKey))))
}

func appendUserKeyVariants(cands *[]turnIntegrityCandidate, baseLabel string, rawUser []byte, rawKey []byte) {
	// Try raw user/key
	appendCand(cands, baseLabel+":user=raw:key=raw", rawUser, rawKey)
	// Try base64 user/key
	appendCand(cands, baseLabel+":user=b64:key=b64", []byte(base64.StdEncoding.EncodeToString(rawUser)), []byte(base64.StdEncoding.EncodeToString(rawKey)))
	// Try hex user/key
	appendCand(cands, baseLabel+":user=hex:key=hex", []byte(strings.ToLower(hex.EncodeToString(rawUser))), []byte(strings.ToLower(hex.EncodeToString(rawKey))))
}

func appendDerivedKeyCandidates(cands *[]turnIntegrityCandidate, labelPrefix string, username []byte, seedKey []byte, seedMsg []byte) {
	if len(seedKey) == 0 || len(seedMsg) == 0 {
		return
	}
	appendCand(cands, labelPrefix+":drv=hmac-sha1", username, hmacSHA1(seedKey, seedMsg))
	appendCand(cands, labelPrefix+":drv=hmac-sha256", username, hmacSHA256(seedKey, seedMsg))
	appendCand(cands, labelPrefix+":drv=sha1(msg)", username, sha1Sum(seedMsg))
	appendCand(cands, labelPrefix+":drv=sha256(msg)", username, sha256Sum(seedMsg))
}

func appendLongTermCandidate(cands *[]turnIntegrityCandidate, label string, username []byte, realm []byte, nonce []byte, password []byte) {
	if len(username) == 0 || len(realm) == 0 || len(nonce) == 0 || len(password) == 0 {
		return
	}
	*cands = append(*cands, turnIntegrityCandidate{
		label:    label,
		username: username,
		key:      turnLongTermKey(username, realm, password),
		realm:    realm,
		nonce:    nonce,
		longTerm: true,
	})
}

func appendLongTermCandidatesFromRelayBlock(cands *[]turnIntegrityCandidate, rb *RelayBlock, realm []byte, nonce []byte) int {
	if cands == nil || rb == nil || len(realm) == 0 || len(nonce) == 0 {
		return 0
	}
	added := 0
	add := func(label string, username []byte, password []byte) {
		before := len(*cands)
		appendLongTermCandidate(cands, label, username, realm, nonce, password)
		if len(*cands) > before {
			added++
		}
	}

	// Primary hypothesis: TE2 maps TURN USERNAME (token_id) to TURN password (auth_token value).
	// Keep this small and avoid logging raw secret bytes.
	for _, te := range rb.TE2 {
		tokID := strings.TrimSpace(te.TokenID)
		authID := strings.TrimSpace(te.AuthTokenID)
		if tokID == "" || authID == "" {
			continue
		}
		var tokBytes []byte
		var authBytes []byte
		for _, t := range rb.Tokens {
			if t.ID == te.TokenID {
				tokBytes = t.Bytes()
				break
			}
		}
		for _, a := range rb.Auth {
			if a.ID == te.AuthTokenID {
				authBytes = a.Bytes()
				break
			}
		}
		if len(tokBytes) > 0 && len(authBytes) > 0 {
			add(fmt.Sprintf("lt te2 user=tokID(%s) pass=authVal(%s)", tokID, authID), []byte(tokID), authBytes)
			add(fmt.Sprintf("lt te2 user=authID(%s) pass=tokVal(%s)", authID, tokID), []byte(authID), tokBytes)
			// Some servers might use raw token/auth values as USERNAME.
			add(fmt.Sprintf("lt te2 user=tokVal(%s) pass=authVal(%s)", tokID, authID), tokBytes, authBytes)
			add(fmt.Sprintf("lt te2 user=authVal(%s) pass=tokVal(%s)", authID, tokID), authBytes, tokBytes)
		}
	}

	// Secondary hypothesis: relay.key/hbh_key may act as TURN password.
	userVariants := buildRelayUsernameVariants(rb)
	if len(userVariants) == 0 {
		u := strings.TrimSpace(rb.UUID)
		if u != "" {
			userVariants = []labeledBytes{{label: "uuid", data: []byte(u)}}
		}
	}
	if len(userVariants) > 0 {
		if b, ok := decodeMaybeBase64(rb.Key); ok {
			for _, u := range userVariants {
				add(fmt.Sprintf("lt relay.key(dec) u=%s", u.label), u.data, b)
			}
		}
		if b, ok := decodeMaybeBase64(rb.HBHKey); ok {
			for _, u := range userVariants {
				add(fmt.Sprintf("lt relay.hbh_key(dec) u=%s", u.label), u.data, b)
			}
		}
	}

	return added
}

func buildSeedMessages(tok []byte, auth []byte, uuid []byte) []seedMsgVariant {
	seeds := make([]seedMsgVariant, 0, 10)
	if len(tok) > 0 && len(auth) > 0 {
		seeds = append(seeds,
			seedMsgVariant{label: "tok+auth", data: append(append([]byte{}, tok...), auth...)},
			seedMsgVariant{label: "auth+tok", data: append(append([]byte{}, auth...), tok...)},
			seedMsgVariant{label: "tok:auth", data: append(append(append([]byte{}, tok...), ':'), auth...)},
			seedMsgVariant{label: "tok\\x00auth", data: append(append(append([]byte{}, tok...), 0), auth...)},
		)
		if len(uuid) > 0 {
			seeds = append(seeds,
				seedMsgVariant{label: "uuid+tok+auth", data: append(append(append([]byte{}, uuid...), tok...), auth...)},
				seedMsgVariant{label: "tok+auth+uuid", data: append(append(append([]byte{}, tok...), auth...), uuid...)},
			)
		}
	}
	if len(tok) > 0 {
		seeds = append(seeds, seedMsgVariant{label: "tok", data: append([]byte{}, tok...)})
	}
	if len(auth) > 0 {
		seeds = append(seeds, seedMsgVariant{label: "auth", data: append([]byte{}, auth...)})
	}
	if len(uuid) > 0 {
		seeds = append(seeds, seedMsgVariant{label: "uuid", data: append([]byte{}, uuid...)})
	}
	return seeds
}

func firstNonEmpty(prefer []byte, fallback []byte) []byte {
	if len(prefer) > 0 {
		return prefer
	}
	return fallback
}

func stunTxIDHex(m *stun.Message) string {
	if m == nil {
		return ""
	}
	return strings.ToLower(hex.EncodeToString(m.TransactionID[:]))
}

func (cm *WhatsmeowCallManager) getRelayBlockForCall(callID string) *RelayBlock {
	if cm == nil || callID == "" {
		return nil
	}
	cm.hsMutex.Lock()
	st := cm.handshakeStates[callID]
	cm.hsMutex.Unlock()
	if st == nil {
		return nil
	}
	return st.Relay
}

func (cm *WhatsmeowCallManager) getOfferEncForCall(callID string) *EncBlock {
	if cm == nil || callID == "" {
		return nil
	}
	cm.hsMutex.Lock()
	st := cm.handshakeStates[callID]
	cm.hsMutex.Unlock()
	if st == nil {
		return nil
	}
	return st.OfferEnc
}

func (cm *WhatsmeowCallManager) getOfferEncDecryptedForCall(callID string) ([]byte, string) {
	if cm == nil || callID == "" {
		return nil, ""
	}
	cm.hsMutex.Lock()
	st := cm.handshakeStates[callID]
	cm.hsMutex.Unlock()
	if st == nil || len(st.OfferEncDecrypted) == 0 {
		return nil, ""
	}
	return append([]byte(nil), st.OfferEncDecrypted...), strings.TrimSpace(st.OfferEncDecryptedSource)
}

func (cm *WhatsmeowCallManager) probeRelayEndpointTURNAllocate(callID string, ep RelayEndpoint, readTimeout time.Duration) error {
	endpoint := strings.TrimSpace(ep.Endpoint)
	if endpoint == "" {
		return fmt.Errorf("empty endpoint")
	}

	// Resolve as UDP4 first (observed relay endpoints are IPv4), fallback to generic.
	remote, err := net.ResolveUDPAddr("udp4", endpoint)
	if err != nil {
		remote, err = net.ResolveUDPAddr("udp", endpoint)
		if err != nil {
			return fmt.Errorf("resolve udp addr: %w", err)
		}
	}

	conn, err := net.DialUDP("udp", nil, remote)
	if err != nil {
		return fmt.Errorf("dial udp: %w", err)
	}
	defer conn.Close()

	local := conn.LocalAddr().String()
	cm.logger.Infof("📡 [RELAY-SESSION] UDP dialed: local=%s remote=%s relay=%s latency=%s (CallID=%s)", local, remote.String(), ep.RelayName, strings.TrimSpace(ep.LatencyRaw), callID)

	// Collect summary for optional dump (no secrets).
	baseAllocateTxID := ""
	baseAllocateSuccess := false
	baseRespMsgType := ""
	baseMappedEndpoint := ""
	baseExtra4002Hex := ""
	baseAllocateCode := 0
	baseAllocateReason := ""
	baseNonceLen := 0
	baseRealmLen := 0
	discTxID := ""
	discUser := ""
	discCode := 0
	discReason := ""
	discNonceLen := 0
	discRealmLen := 0

	// TURN Allocate request (without integrity) should elicit an error response from the relay.
	// We use this to confirm the endpoint is alive and to learn what kind of auth is expected.
	extra := readTurnExtraAttrsFromEnv()
	if len(extra.a4000) > 0 || len(extra.a4024) > 0 || len(extra.a0016) > 0 || len(extra.a4000ByEndpoint) > 0 {
		cm.logger.Warnf("📡 [RELAY-TURN] Using extra Allocate attrs: 0x4000=%d 0x4000_map=%d 0x4024=%d 0x0016=%d (CallID=%s)", len(extra.a4000), len(extra.a4000ByEndpoint), len(extra.a4024), len(extra.a0016), callID)
	}
	if len(extra.a0016) == 0 && envTruthy("QP_CALL_RELAY_TURN_AUTO_ATTR_0016") {
		cm.logger.Warnf("📡 [RELAY-TURN] Auto-building Allocate attr 0x0016 from remote endpoint (CallID=%s)", callID)
	}
	captureReqs := envTruthy("QP_CALL_DUMP_TURN_REQUESTS") && envTruthy("QP_CALL_DUMP_TURN_PROBE")
	capMax := clampInt(envInt("QP_CALL_DUMP_TURN_REQUESTS_MAX", 20), 0, 200)
	requestDumps := make([]callTurnProbeRequestDump, 0, 8)
	maybeCap := func(stage string, m *stun.Message, preimage []byte) {
		if !captureReqs {
			return
		}
		if capMax == 0 {
			return
		}
		if len(requestDumps) >= capMax {
			return
		}
		if m == nil {
			return
		}
		requestDumps = append(requestDumps, captureTurnRequestDump(stage, m.Raw, preimage))
	}

	baseReq := stun.MustBuild(stun.TransactionID)
	baseReq.Type = stun.NewType(stun.Method(0x003), stun.ClassRequest) // Allocate
	if !envTruthy("QP_CALL_RELAY_TURN_OMIT_REQUESTED_TRANSPORT") {
		baseReq.Add(stun.AttrType(0x0019), []byte{17, 0, 0, 0}) // REQUESTED-TRANSPORT: UDP
	}
	addTurnExtraAttrsForEndpoint(baseReq, extra, remote)
	baseAllocateTxID = stunTxIDHex(baseReq)
	baseReq.Encode()
	basePreimage := append([]byte(nil), baseReq.Raw...)
	if envTruthyDefault("QP_CALL_RELAY_TURN_INCLUDE_FINGERPRINT", true) {
		_ = stun.Fingerprint.AddTo(baseReq)
	}
	maybeCap("base", baseReq, basePreimage)
	if _, err := conn.Write(baseReq.Raw); err != nil {
		return fmt.Errorf("turn allocate write: %w", err)
	}

	buf := make([]byte, 1500)
	_ = conn.SetReadDeadline(time.Now().Add(readTimeout))
	n, err := conn.Read(buf)
	if err != nil {
		return fmt.Errorf("turn allocate read: %w", err)
	}
	if n <= 0 {
		return fmt.Errorf("turn allocate empty read")
	}

	respRaw := buf[:n]
	if !isSTUNPacket(respRaw) {
		return fmt.Errorf("unexpected non-STUN response bytes=%d", n)
	}
	var resp stun.Message
	resp.Raw = respRaw
	if err := resp.Decode(); err != nil {
		return fmt.Errorf("turn allocate decode: %w", err)
	}
	baseRespMsgType = stunMsgTypeHexFromMessage(&resp)
	if baseAllocateTxID != "" {
		respTxID := stunTxIDHex(&resp)
		if respTxID != "" && respTxID != baseAllocateTxID {
			cm.logger.Warnf("⚠️📡 [RELAY-TURN] Base Allocate txid mismatch: req=%s resp=%s relay=%s endpoint=%s (CallID=%s)", baseAllocateTxID, respTxID, ep.RelayName, endpoint, callID)
		}
	}
	if resp.Type.Method == stun.Method(0x003) && resp.Type.Class == stun.ClassSuccessResponse {
		baseAllocateSuccess = true
		baseMappedEndpoint, baseExtra4002Hex = extractMappedEndpointAnd4002(&resp)
		cm.logger.Warnf("✅📡 [RELAY-TURN] Allocate SUCCESS (base): txid=%s relay=%s endpoint=%s mapped=%s a4002=%s (CallID=%s)", baseAllocateTxID, ep.RelayName, endpoint, baseMappedEndpoint, baseExtra4002Hex, callID)
		if envTruthy("QP_CALL_DUMP_TURN_PROBE") {
			// Dump still gets written at the end; skip integrity attempts.
			// Fall through to dump section by returning nil from here is not possible
			// because we want to preserve the dump fields. We'll short-circuit later.
		}
		// No need to attempt integrity candidates.
		if !envTruthy("QP_CALL_DUMP_TURN_PROBE") {
			return nil
		}
	}

	var errCode stun.ErrorCodeAttribute
	if err2 := errCode.GetFrom(&resp); err2 == nil {
		baseAllocateCode = int(errCode.Code)
		baseAllocateReason = strings.TrimSpace(string(errCode.Reason))
	}
	if baseAllocateTxID != "" {
		cm.logger.Warnf("📡 [RELAY-TURN] Allocate response: txid=%s relay=%s endpoint=%s code=%d reason=%q (CallID=%s)", baseAllocateTxID, ep.RelayName, endpoint, baseAllocateCode, baseAllocateReason, callID)
	} else {
		cm.logger.Warnf("📡 [RELAY-TURN] Allocate response: relay=%s endpoint=%s code=%d reason=%q (CallID=%s)", ep.RelayName, endpoint, baseAllocateCode, baseAllocateReason, callID)
	}

	// Extract nonce/realm (if server offers long-term auth on the initial error response).
	var baseNonce stun.Nonce
	var baseRealm stun.Realm
	_ = resp.Parse(&baseNonce, &baseRealm)
	baseNonceB := []byte(baseNonce)
	baseRealmB := []byte(baseRealm)
	baseNonceLen = len(baseNonceB)
	baseRealmLen = len(baseRealmB)
	if len(baseNonceB) > 0 || len(baseRealmB) > 0 {
		cm.logger.Warnf("🔐📡 [RELAY-TURN] Base response auth hints: nonceLen=%d realmLen=%d (CallID=%s)", len(baseNonceB), len(baseRealmB), callID)
	}

	tryIntegrity := envTruthy("QP_CALL_RELAY_TURN_ALLOCATE_TRY_INTEGRITY")
	if baseAllocateSuccess {
		tryIntegrity = false
	}

	rb := cm.getRelayBlockForCall(callID)
	if rb == nil {
		rb = &RelayBlock{}
	}

	if !tryIntegrity {
		if envTruthy("QP_CALL_DUMP_TURN_PROBE") {
			// fall through to dump section
		} else {
			return nil
		}
	}
	if rb == nil {
		cm.logger.Warnf("⚠️📡 [RELAY-TURN] No RelayBlock available for integrity attempts (CallID=%s)", callID)
		return nil
	}

	// If the relay uses long-term auth, it should provide REALM and NONCE.
	// Sometimes servers only return these when USERNAME is present, so we do one discovery request.
	ltNonce := baseNonceB
	ltRealm := baseRealmB
	forceNoUsername := envTruthy("QP_CALL_RELAY_TURN_FORCE_NO_USERNAME")
	if (len(ltNonce) == 0 || len(ltRealm) == 0) && len(rb.TE2) > 0 {
		u := strings.TrimSpace(rb.TE2[0].TokenID)
		if u == "" {
			u = strings.TrimSpace(rb.TE2[0].AuthTokenID)
		}
		if u != "" {
			discUser = u
			disc := stun.MustBuild(stun.TransactionID)
			disc.Type = stun.NewType(stun.Method(0x003), stun.ClassRequest)
			if !envTruthy("QP_CALL_RELAY_TURN_OMIT_REQUESTED_TRANSPORT") {
				disc.Add(stun.AttrType(0x0019), []byte{17, 0, 0, 0})
			}
			addTurnExtraAttrsForEndpoint(disc, extra, remote)
			if !forceNoUsername {
				disc.Add(stun.AttrUsername, []byte(u))
			} else {
				discUser = ""
			}
			discTxID = stunTxIDHex(disc)
			disc.Encode()
			discPreimage := append([]byte(nil), disc.Raw...)
			if envTruthyDefault("QP_CALL_RELAY_TURN_INCLUDE_FINGERPRINT", true) {
				_ = stun.Fingerprint.AddTo(disc)
			}
			maybeCap("discovery", disc, discPreimage)
			if _, err := conn.Write(disc.Raw); err == nil {
				_ = conn.SetReadDeadline(time.Now().Add(readTimeout))
				if n3, err := conn.Read(buf); err == nil && n3 > 0 {
					raw3 := buf[:n3]
					if isSTUNPacket(raw3) {
						var r3 stun.Message
						r3.Raw = raw3
						if err := r3.Decode(); err == nil {
							if discTxID != "" {
								respTxID := stunTxIDHex(&r3)
								if respTxID != "" && respTxID != discTxID {
									cm.logger.Warnf("⚠️📡 [RELAY-TURN] Auth discovery txid mismatch: req=%s resp=%s relay=%s endpoint=%s (CallID=%s)", discTxID, respTxID, ep.RelayName, endpoint, callID)
								}
							}
							var ec3 stun.ErrorCodeAttribute
							code3 := 0
							reason3 := ""
							if err := ec3.GetFrom(&r3); err == nil {
								code3 = int(ec3.Code)
								reason3 = strings.TrimSpace(string(ec3.Reason))
							}
							var n3a stun.Nonce
							var r3a stun.Realm
							_ = r3.Parse(&n3a, &r3a)
							ltNonce = []byte(n3a)
							ltRealm = []byte(r3a)
							discCode = code3
							discReason = reason3
							discNonceLen = len(ltNonce)
							discRealmLen = len(ltRealm)
							if discTxID != "" {
								cm.logger.Warnf("🔐📡 [RELAY-TURN] Auth discovery response: txid=%s user=%q code=%d reason=%q nonceLen=%d realmLen=%d (CallID=%s)", discTxID, u, code3, reason3, len(ltNonce), len(ltRealm), callID)
							} else {
								cm.logger.Warnf("🔐📡 [RELAY-TURN] Auth discovery response: user=%q code=%d reason=%q nonceLen=%d realmLen=%d (CallID=%s)", u, code3, reason3, len(ltNonce), len(ltRealm), callID)
							}
						}
					}
				}
			}
		}
	}

	// Candidate integrity keys (redacted):
	//  - relay.key / relay.hbh_key (base64 in offers)
	//  - token/auth_token values (binary-like strings)
	cands := make([]turnIntegrityCandidate, 0, 8)
	encBlock := cm.getOfferEncForCall(callID)
	var encRaw []byte
	var encHash []byte
	var encPlain []byte
	var encPlainHash []byte
	encPlainFields := map[string][]byte(nil)
	encPlain, encPlainSource := cm.getOfferEncDecryptedForCall(callID)
	if encBlock != nil && len(encBlock.Raw) > 0 {
		encRaw = encBlock.Raw
		// Stable, fixed-size seed (useful if relays hash the offer payload first).
		encHash = sha256Sum(encRaw)
	}
	if len(encPlain) > 0 {
		encPlainHash = sha256Sum(encPlain)
		encPlainFields = parseSimpleProtoFields(encPlain)
	}
	// Extra attrs used in Desktop-like Allocate flow (if provided). Used as KDF message material.
	a4000Used, a4024Used, a0016Used := resolveTurnExtraAttrsForEndpoint(extra, remote)
	extraMsgs := buildExtraAttrsMsgVariants(a4000Used, a4024Used, a0016Used)
	if envTruthy("QP_CALL_RELAY_TURN_TRY_MIKDF_ALLOC_PREIMAGE") {
		for k, v := range buildAllocatePreimageMsgVariants(extra, remote) {
			if len(v) == 0 {
				continue
			}
			extraMsgs[k] = v
		}
	}
	if envTruthyDefault("QP_CALL_RELAY_TURN_TRY_MIKDF_CONTEXT", true) {
		for k, v := range buildRelayContextMsgVariants(rb, extraMsgs) {
			if len(v) == 0 {
				continue
			}
			extraMsgs[k] = v
		}
	}
	for k, v := range buildRelayTE2MsgVariants(rb) {
		if len(v) == 0 {
			continue
		}
		extraMsgs[k] = v
	}
	if len(encPlainFields) > 0 {
		for k, v := range encPlainFields {
			if len(v) == 0 {
				continue
			}
			extraMsgs["encplain."+k] = v
		}
	}
	// Env-gated experiment: try using the extra attrs themselves as MESSAGE-INTEGRITY keys.
	// Rationale: Desktop Allocate succeeds without USERNAME/REALM/NONCE; the short-term key may be derived from
	// or equal to these proprietary blobs (or their hashes) rather than offer secrets.
	if envTruthy("QP_CALL_RELAY_TURN_TRY_EXTRA_ATTRS_AS_KEY") {
		selected := []struct {
			label string
			data  []byte
		}{
			{label: "extra4000", data: a4000Used},
			{label: "extra4024", data: a4024Used},
			{label: "extra0016", data: a0016Used},
			{label: "extra4000+4024+0016", data: extraMsgs["extra4000+4024+0016"]},
			{label: "tlv4000+4024+0016", data: extraMsgs["tlv4000+4024+0016"]},
		}
		for _, it := range selected {
			if len(it.data) == 0 {
				continue
			}
			appendCand(&cands, fmt.Sprintf("extra-key(%s):user=none:key=raw", it.label), nil, it.data)
			appendCand(&cands, fmt.Sprintf("extra-key(%s):user=none:key=sha1", it.label), nil, sha1Sum(it.data))
			appendCand(&cands, fmt.Sprintf("extra-key(%s):user=none:key=sha256", it.label), nil, sha256Sum(it.data))
		}
	}
	appendMiKDF := func(labelPrefix string, seedKey []byte) {
		if len(seedKey) == 0 {
			return
		}
		// Derive MI key candidates using only HMAC-SHA1 (Desktop packets use MI-SHA1 attr 0x0008).
		// Keep this space intentionally small: relay probe has a global maxTry budget.
		seedSHA1 := sha1Sum(seedKey)
		for msgLabel, msg := range extraMsgs {
			if len(msg) == 0 {
				continue
			}
			msgSHA1 := sha1Sum(msg)

			// Baseline: HMAC(seedKey, msg)
			appendCand(&cands, labelPrefix+"(miKDF:msg="+msgLabel+"):user=none:key=hmac1", nil, hmacSHA1(seedKey, msg))
			// Reverse: HMAC(msg, seedKey)
			appendCand(&cands, labelPrefix+"(miKDF:msg="+msgLabel+"):user=none:key=hmac1_rev", nil, hmacSHA1(msg, seedKey))
			// Hashed seed: HMAC(sha1(seedKey), msg)
			appendCand(&cands, labelPrefix+"(miKDF:msg="+msgLabel+"):user=none:key=hmac1_seedSHA1", nil, hmacSHA1(seedSHA1, msg))
			// Hashed msg: HMAC(seedKey, sha1(msg))
			appendCand(&cands, labelPrefix+"(miKDF:msg="+msgLabel+"):user=none:key=hmac1_msgSHA1", nil, hmacSHA1(seedKey, msgSHA1))
			// Desktop-like context variant: hash the candidate key material once first.
			appendCand(&cands, labelPrefix+"(miKDF:msg="+msgLabel+"):user=none:key=sha1(seed||msg)", nil, sha1Sum(append(append([]byte(nil), seedKey...), msg...)))
			appendCand(&cands, labelPrefix+"(miKDF:msg="+msgLabel+"):user=none:key=sha1(msg||seed)", nil, sha1Sum(append(append([]byte(nil), msg...), seedKey...)))
		}
	}
	// Env-gated experiment: derive MI keys from offer secrets + Desktop-like extra attrs.
	if envTruthyDefault("QP_CALL_RELAY_TURN_TRY_MIKDF_EXTRA_ATTRS", true) {
		if len(extraMsgs) > 0 {
			var uuidBin []byte
			for _, u := range buildRelayUsernameVariants(rb) {
				if u.label == "uuid(bin)" {
					uuidBin = append([]byte(nil), u.data...)
					break
				}
			}
			selfBytes := []byte(strings.TrimSpace(rb.SelfPID))
			peerBytes := []byte(strings.TrimSpace(rb.PeerPID))
			te2Msgs := buildRelayTE2MsgVariants(rb)

			seeds := make([]labeledSeed, 0, 8)
			if len(encPlain) > 0 {
				label := "enc.plain"
				if encPlainSource != "" {
					label += "." + encPlainSource
				}
				label += fmt.Sprintf("(len=%d)", len(encPlain))
				seeds = append(seeds, labeledSeed{label: label, data: encPlain})
				if b := appendParts(encPlain, uuidBin); len(b) > 0 {
					seeds = append(seeds, labeledSeed{label: "enc.plain+uuid(bin)", data: b})
				}
				if b := appendParts(encPlain, selfBytes, peerBytes); len(b) > 0 {
					seeds = append(seeds, labeledSeed{label: "enc.plain+self+peer", data: b})
				}
				if b := appendParts(encPlain, uuidBin, selfBytes, peerBytes); len(b) > 0 {
					seeds = append(seeds, labeledSeed{label: "enc.plain+uuid(bin)+self+peer", data: b})
				}
				for k, v := range encPlainFields {
					if len(v) == 0 {
						continue
					}
					seeds = append(seeds, labeledSeed{label: "enc.plain." + k, data: v})
					if strings.Contains(k, "proto.f10.f1") {
						if b := appendParts(v, uuidBin); len(b) > 0 {
							seeds = append(seeds, labeledSeed{label: "enc.plain." + k + "+uuid(bin)", data: b})
						}
						if b := appendParts(v, a4000Used); len(b) > 0 {
							seeds = append(seeds, labeledSeed{label: "enc.plain." + k + "+a4000", data: b})
						}
						if b := appendParts(v, a4024Used); len(b) > 0 {
							seeds = append(seeds, labeledSeed{label: "enc.plain." + k + "+a4024", data: b})
						}
						if b := appendParts(v, a0016Used); len(b) > 0 {
							seeds = append(seeds, labeledSeed{label: "enc.plain." + k + "+a0016", data: b})
						}
						if b := appendParts(v, a4000Used, a4024Used, a0016Used); len(b) > 0 {
							seeds = append(seeds, labeledSeed{label: "enc.plain." + k + "+a4000+a4024+a0016", data: b})
						}
						if b := appendParts(v, uuidBin, a4000Used, a4024Used, a0016Used, selfBytes, peerBytes); len(b) > 0 {
							seeds = append(seeds, labeledSeed{label: "enc.plain." + k + "+uuid(bin)+attrs+self+peer", data: b})
						}
					}
					if b := appendParts(v, selfBytes, peerBytes); len(b) > 0 {
						seeds = append(seeds, labeledSeed{label: "enc.plain." + k + "+self+peer", data: b})
					}
					if b := appendParts(v, uuidBin, selfBytes, peerBytes); len(b) > 0 {
						seeds = append(seeds, labeledSeed{label: "enc.plain." + k + "+uuid(bin)+self+peer", data: b})
					}
				}
				if remotePub := encPlainFields["proto.f10.f1"]; len(remotePub) == 32 && cm.connection != nil && cm.connection.Client != nil && cm.connection.Client.Store != nil {
					for te2Label, te2Data := range te2Msgs {
						if len(te2Data) == 0 {
							continue
						}
						if b := appendParts(remotePub, te2Data); len(b) > 0 {
							seeds = append(seeds, labeledSeed{label: "enc.plain.proto.f10.f1+" + te2Label, data: b})
						}
						if b := appendParts(te2Data, remotePub); len(b) > 0 {
							seeds = append(seeds, labeledSeed{label: te2Label + "+enc.plain.proto.f10.f1", data: b})
						}
					}
					if id := cm.connection.Client.Store.IdentityKey; id != nil && id.Priv != nil {
						if shared := deriveX25519SharedSecret(id.Priv, remotePub); len(shared) > 0 {
							seeds = append(seeds, labeledSeed{label: "enc.plain.proto.f10.f1.ecdh(identity)", data: shared})
							if b := appendParts(shared, uuidBin); len(b) > 0 {
								seeds = append(seeds, labeledSeed{label: "enc.plain.proto.f10.f1.ecdh(identity)+uuid(bin)", data: b})
							}
							if b := appendParts(shared, a4000Used, a4024Used, a0016Used); len(b) > 0 {
								seeds = append(seeds, labeledSeed{label: "enc.plain.proto.f10.f1.ecdh(identity)+attrs", data: b})
							}
							if b := appendParts(shared, uuidBin, a4000Used, a4024Used, a0016Used, selfBytes, peerBytes); len(b) > 0 {
								seeds = append(seeds, labeledSeed{label: "enc.plain.proto.f10.f1.ecdh(identity)+uuid(bin)+attrs+self+peer", data: b})
							}
						}
					}
					if spk := cm.connection.Client.Store.SignedPreKey; spk != nil && spk.Priv != nil {
						if shared := deriveX25519SharedSecret(spk.Priv, remotePub); len(shared) > 0 {
							seeds = append(seeds, labeledSeed{label: "enc.plain.proto.f10.f1.ecdh(signedprekey)", data: shared})
							if b := appendParts(shared, uuidBin); len(b) > 0 {
								seeds = append(seeds, labeledSeed{label: "enc.plain.proto.f10.f1.ecdh(signedprekey)+uuid(bin)", data: b})
							}
							if b := appendParts(shared, a4000Used, a4024Used, a0016Used); len(b) > 0 {
								seeds = append(seeds, labeledSeed{label: "enc.plain.proto.f10.f1.ecdh(signedprekey)+attrs", data: b})
							}
							if b := appendParts(shared, uuidBin, a4000Used, a4024Used, a0016Used, selfBytes, peerBytes); len(b) > 0 {
								seeds = append(seeds, labeledSeed{label: "enc.plain.proto.f10.f1.ecdh(signedprekey)+uuid(bin)+attrs+self+peer", data: b})
							}
						}
					}
				}
			}
			if len(encPlainHash) > 0 {
				seeds = append(seeds, labeledSeed{label: "enc.plain.sha256", data: encPlainHash})
				if b := appendParts(encPlainHash, uuidBin); len(b) > 0 {
					seeds = append(seeds, labeledSeed{label: "enc.plain.sha256+uuid(bin)", data: b})
				}
				if b := appendParts(encPlainHash, selfBytes, peerBytes); len(b) > 0 {
					seeds = append(seeds, labeledSeed{label: "enc.plain.sha256+self+peer", data: b})
				}
				if b := appendParts(encPlainHash, uuidBin, selfBytes, peerBytes); len(b) > 0 {
					seeds = append(seeds, labeledSeed{label: "enc.plain.sha256+uuid(bin)+self+peer", data: b})
				}
			}
			if len(encRaw) > 0 {
				seeds = append(seeds, labeledSeed{label: "enc." + strings.TrimSpace(encBlock.Type) + fmt.Sprintf("(raw,len=%d)", len(encRaw)), data: encRaw})
				if b := appendParts(encRaw, uuidBin); len(b) > 0 {
					seeds = append(seeds, labeledSeed{label: "enc.raw+uuid(bin)", data: b})
				}
				if b := appendParts(encRaw, selfBytes, peerBytes); len(b) > 0 {
					seeds = append(seeds, labeledSeed{label: "enc.raw+self+peer", data: b})
				}
				if b := appendParts(encRaw, uuidBin, selfBytes, peerBytes); len(b) > 0 {
					seeds = append(seeds, labeledSeed{label: "enc.raw+uuid(bin)+self+peer", data: b})
				}
			}
			if len(encHash) > 0 {
				seeds = append(seeds, labeledSeed{label: "enc.sha256", data: encHash})
				if b := appendParts(encHash, uuidBin); len(b) > 0 {
					seeds = append(seeds, labeledSeed{label: "enc.sha256+uuid(bin)", data: b})
				}
				if b := appendParts(encHash, selfBytes, peerBytes); len(b) > 0 {
					seeds = append(seeds, labeledSeed{label: "enc.sha256+self+peer", data: b})
				}
				if b := appendParts(encHash, uuidBin, selfBytes, peerBytes); len(b) > 0 {
					seeds = append(seeds, labeledSeed{label: "enc.sha256+uuid(bin)+self+peer", data: b})
				}
			}
			for te2Label, te2Data := range te2Msgs {
				if len(te2Data) == 0 {
					continue
				}
				seeds = append(seeds, labeledSeed{label: te2Label, data: te2Data})
				if b := appendParts(te2Data, uuidBin); len(b) > 0 {
					seeds = append(seeds, labeledSeed{label: te2Label + "+uuid(bin)", data: b})
				}
				if b := appendParts(te2Data, selfBytes, peerBytes); len(b) > 0 {
					seeds = append(seeds, labeledSeed{label: te2Label + "+self+peer", data: b})
				}
			}
			for _, s := range seeds {
				if len(s.data) == 0 {
					continue
				}
				appendMiKDF(s.label, s.data)
			}
		}
	}
	// For relay keys, try using relay UUID as USERNAME to help server-side key selection.
	defaultUser := []byte(strings.TrimSpace(rb.UUID))
	userVariants := buildRelayUsernameVariants(rb)
	if len(userVariants) == 0 {
		userVariants = []labeledBytes{{label: "uuid", data: defaultUser}}
	}
	// Offer enc (pkmsg/msg) candidates: try using enc bytes directly as MI key.
	// Rationale: relay short-term integrity secret may be derived from or equal to offer enc payload.
	if len(encPlain) > 0 {
		encPlainLabel := "enc.plain"
		if encPlainSource != "" {
			encPlainLabel += "." + encPlainSource
		}
		encPlainLabel += fmt.Sprintf("(len=%d)", len(encPlain))
		for _, u := range userVariants {
			appendKeyVariantsUserEncodings(&cands, encPlainLabel+":u="+u.label, u.data, encPlain)
			appendMD5RealmEmptyCandidates(&cands, encPlainLabel+":u="+u.label, u.data, encPlain)
		}
		appendKeyVariantsNoUser(&cands, encPlainLabel, encPlain)
	}
	if len(encPlainHash) > 0 {
		for _, u := range userVariants {
			appendKeyVariantsUserEncodings(&cands, "enc.plain.sha256:u="+u.label, u.data, encPlainHash)
			appendMD5RealmEmptyCandidates(&cands, "enc.plain.sha256:u="+u.label, u.data, encPlainHash)
		}
		appendKeyVariantsNoUser(&cands, "enc.plain.sha256", encPlainHash)
	}
	if len(encRaw) > 0 {
		encLabel := fmt.Sprintf("enc.%s(raw,len=%d)", strings.TrimSpace(encBlock.Type), len(encRaw))
		for _, u := range userVariants {
			appendKeyVariantsUserEncodings(&cands, encLabel+":u="+u.label, u.data, encRaw)
			appendMD5RealmEmptyCandidates(&cands, encLabel+":u="+u.label, u.data, encRaw)
		}
		appendKeyVariantsNoUser(&cands, encLabel, encRaw)
	}
	if len(encHash) > 0 {
		for _, u := range userVariants {
			appendKeyVariantsUserEncodings(&cands, "enc.sha256:u="+u.label, u.data, encHash)
			appendMD5RealmEmptyCandidates(&cands, "enc.sha256:u="+u.label, u.data, encHash)
		}
		appendKeyVariantsNoUser(&cands, "enc.sha256", encHash)
	}
	// Add <te> payloads observed in CallAccept as additional candidate TURN USERNAMEs.
	// Budget-limited to avoid blowing up the candidate space.
	if te := cm.getCallAcceptTE(callID); len(te) > 0 {
		max := 6
		if len(te) < max {
			max = len(te)
		}
		for i := 0; i < max; i++ {
			v := strings.TrimSpace(te[i])
			if v == "" {
				continue
			}
			userVariants = append(userVariants, labeledBytes{label: fmt.Sprintf("accept.te[%d]", i), data: []byte(v)})
		}
	}
	// TE2 payload candidates: treat relay <te2> payload bytes as potential short-term integrity secrets.
	// Rationale: offers include multiple small opaque TE2 payloads (len=6/18) that may seed relay auth.
	// Keep this VERY small to avoid exploding the candidate space.
	te2PayloadMax := envInt("QP_CALL_RELAY_TURN_TE2_PAYLOAD_MAX", 4)
	if te2PayloadMax > 0 {
		if te2PayloadMax > 10 {
			te2PayloadMax = 10
		}
		added := 0
		seenPayload := make(map[string]struct{}, te2PayloadMax)
		pickUsers := func() []labeledBytes {
			picked := make([]labeledBytes, 0, 6)
			want := map[string]bool{
				"uuid(bin)":      true,
				"uuid":           true,
				"uuid(bin):peer": true,
				"uuid(bin):self": true,
				"peer":           true,
				"self":           true,
			}
			for _, u := range userVariants {
				if want[u.label] {
					picked = append(picked, u)
				}
				if len(picked) >= 6 {
					break
				}
			}
			if len(picked) == 0 && len(userVariants) > 0 {
				picked = append(picked, userVariants[0])
			}
			return picked
		}
		for _, te := range rb.TE2 {
			if added >= te2PayloadMax {
				break
			}
			if len(te.Payload) == 0 {
				continue
			}
			// Dedupe by payload bytes (as hex) to avoid repeating identical payloads.
			ph := strings.ToLower(hex.EncodeToString(te.Payload))
			if _, ok := seenPayload[ph]; ok {
				continue
			}
			seenPayload[ph] = struct{}{}
			users := pickUsers()
			if len(users) == 0 {
				continue
			}
			keys := []labeledBytes{
				{label: "key=payload", data: te.Payload},
				{label: "key=sha1(payload)", data: sha1Sum(te.Payload)},
				{label: "key=sha256(payload)", data: sha256Sum(te.Payload)},
			}
			for _, u := range users {
				for _, k := range keys {
					appendCand(&cands, fmt.Sprintf("te2 payload(relay=%s,len=%d):u=%s:%s", te.RelayName, len(te.Payload), u.label, k.label), u.data, k.data)
				}
			}
			// Also try TE2 payload as the TURN USERNAME (raw bytes), with a few key options.
			for _, k := range keys {
				appendCand(&cands, fmt.Sprintf("te2 payload(relay=%s,len=%d):u=payload:%s", te.RelayName, len(te.Payload), k.label), te.Payload, k.data)
			}
			for _, k := range keys {
				appendCand(&cands, fmt.Sprintf("te2 payload(relay=%s,len=%d):user=none:%s", te.RelayName, len(te.Payload), k.label), nil, k.data)
			}
			added++
		}
	}
	deriveHMAC := envTruthyDefault("QP_CALL_RELAY_TURN_DERIVE_HMAC", true)
	var relayKeyDecoded []byte
	var relayKeyDecoded2 []byte
	var hbhKeyDecoded []byte
	var hbhKeyDecoded2 []byte
	// Relay keys: try both decoded bytes and the original base64 text itself.
	if rb.Key != "" {
		if b1, ok := decodeMaybeBase64(rb.Key); ok {
			relayKeyDecoded = b1
			if envTruthyDefault("QP_CALL_RELAY_TURN_TRY_MIKDF_EXTRA_ATTRS", true) {
				appendMiKDF(fmt.Sprintf("relay.key(dec1_b64,len=%d)", len(b1)), b1)
			}
			for _, u := range userVariants {
				appendKeyVariantsUserEncodings(&cands, fmt.Sprintf("relay.key(dec1_b64,len=%d):u=%s", len(b1), u.label), u.data, b1)
				appendMD5RealmEmptyCandidates(&cands, fmt.Sprintf("relay.key(dec1_b64,len=%d):u=%s", len(b1), u.label), u.data, b1)
			}
			appendKeyVariantsNoUser(&cands, fmt.Sprintf("relay.key(dec1_b64,len=%d)", len(b1)), b1)
			if b2, depth, ok2 := decodeMaybeBase64Recursive(rb.Key, 2); ok2 && depth >= 2 && len(b2) > 0 && !bytes.Equal(b2, b1) {
				relayKeyDecoded2 = b2
				if envTruthyDefault("QP_CALL_RELAY_TURN_TRY_MIKDF_EXTRA_ATTRS", true) {
					appendMiKDF(fmt.Sprintf("relay.key(dec2_b64,len=%d)", len(b2)), b2)
				}
				for _, u := range userVariants {
					appendKeyVariantsUserEncodings(&cands, fmt.Sprintf("relay.key(dec2_b64,len=%d):u=%s", len(b2), u.label), u.data, b2)
					appendMD5RealmEmptyCandidates(&cands, fmt.Sprintf("relay.key(dec2_b64,len=%d):u=%s", len(b2), u.label), u.data, b2)
				}
				appendKeyVariantsNoUser(&cands, fmt.Sprintf("relay.key(dec2_b64,len=%d)", len(b2)), b2)
			}
		}
		for _, u := range userVariants {
			appendKeyVariantsUserEncodings(&cands, fmt.Sprintf("relay.key(text_b64,len=%d):u=%s", len(strings.TrimSpace(rb.Key)), u.label), u.data, []byte(strings.TrimSpace(rb.Key)))
			appendMD5RealmEmptyCandidates(&cands, fmt.Sprintf("relay.key(text_b64,len=%d):u=%s", len(strings.TrimSpace(rb.Key)), u.label), u.data, []byte(strings.TrimSpace(rb.Key)))
		}
		appendKeyVariantsNoUser(&cands, fmt.Sprintf("relay.key(text_b64,len=%d)", len(strings.TrimSpace(rb.Key))), []byte(strings.TrimSpace(rb.Key)))
	}
	if rb.HBHKey != "" {
		if b1, ok := decodeMaybeBase64(rb.HBHKey); ok {
			hbhKeyDecoded = b1
			if envTruthyDefault("QP_CALL_RELAY_TURN_TRY_MIKDF_EXTRA_ATTRS", true) {
				appendMiKDF(fmt.Sprintf("relay.hbh_key(dec1_b64,len=%d)", len(b1)), b1)
			}
			for _, u := range userVariants {
				appendKeyVariantsUserEncodings(&cands, fmt.Sprintf("relay.hbh_key(dec1_b64,len=%d):u=%s", len(b1), u.label), u.data, b1)
				appendMD5RealmEmptyCandidates(&cands, fmt.Sprintf("relay.hbh_key(dec1_b64,len=%d):u=%s", len(b1), u.label), u.data, b1)
			}
			appendKeyVariantsNoUser(&cands, fmt.Sprintf("relay.hbh_key(dec1_b64,len=%d)", len(b1)), b1)
			if b2, depth, ok2 := decodeMaybeBase64Recursive(rb.HBHKey, 2); ok2 && depth >= 2 && len(b2) > 0 && !bytes.Equal(b2, b1) {
				hbhKeyDecoded2 = b2
				if envTruthyDefault("QP_CALL_RELAY_TURN_TRY_MIKDF_EXTRA_ATTRS", true) {
					appendMiKDF(fmt.Sprintf("relay.hbh_key(dec2_b64,len=%d)", len(b2)), b2)
				}
				for _, u := range userVariants {
					appendKeyVariantsUserEncodings(&cands, fmt.Sprintf("relay.hbh_key(dec2_b64,len=%d):u=%s", len(b2), u.label), u.data, b2)
					appendMD5RealmEmptyCandidates(&cands, fmt.Sprintf("relay.hbh_key(dec2_b64,len=%d):u=%s", len(b2), u.label), u.data, b2)
				}
				appendKeyVariantsNoUser(&cands, fmt.Sprintf("relay.hbh_key(dec2_b64,len=%d)", len(b2)), b2)
			}
		}
		for _, u := range userVariants {
			appendKeyVariantsUserEncodings(&cands, fmt.Sprintf("relay.hbh_key(text_b64,len=%d):u=%s", len(strings.TrimSpace(rb.HBHKey)), u.label), u.data, []byte(strings.TrimSpace(rb.HBHKey)))
			appendMD5RealmEmptyCandidates(&cands, fmt.Sprintf("relay.hbh_key(text_b64,len=%d):u=%s", len(strings.TrimSpace(rb.HBHKey)), u.label), u.data, []byte(strings.TrimSpace(rb.HBHKey)))
		}
		appendKeyVariantsNoUser(&cands, fmt.Sprintf("relay.hbh_key(text_b64,len=%d)", len(strings.TrimSpace(rb.HBHKey))), []byte(strings.TrimSpace(rb.HBHKey)))
	}

	// Raw token/auth_token values are binary-like strings; use their bytes directly.
	// We try both "username=token, key=auth" and "username=auth, key=token".
	// Prefer mapping via TE2 if available.
	derive := envTruthyDefault("QP_CALL_RELAY_TURN_DERIVE_CANDIDATES", true)
	tryREST := envTruthyDefault("QP_CALL_RELAY_TURN_TRY_REST", true)
	if derive && len(rb.TE2) > 0 {
		// If we have long-term auth material, prioritize correct TURN-style attempts first.
		if len(ltNonce) > 0 && len(ltRealm) > 0 {
			for _, te := range rb.TE2 {
				tokID := strings.TrimSpace(te.TokenID)
				authID := strings.TrimSpace(te.AuthTokenID)
				if tokID == "" || authID == "" {
					continue
				}
				tokVal := ""
				authVal := ""
				for _, t := range rb.Tokens {
					if strings.TrimSpace(t.ID) == tokID {
						tokVal = t.Value
						break
					}
				}
				for _, a := range rb.Auth {
					if strings.TrimSpace(a.ID) == authID {
						authVal = a.Value
						break
					}
				}
				if tokVal == "" || authVal == "" {
					continue
				}
				appendLongTermCandidate(&cands, fmt.Sprintf("lt user=tokID(%s) pass=auth(%s)", tokID, authID), []byte(tokID), ltRealm, ltNonce, []byte(authVal))
				appendLongTermCandidate(&cands, fmt.Sprintf("lt user=authID(%s) pass=tok(%s)", authID, tokID), []byte(authID), ltRealm, ltNonce, []byte(tokVal))
				// Some servers may use the raw token/auth values as USERNAME.
				appendLongTermCandidate(&cands, fmt.Sprintf("lt user=tokVal(%s) pass=auth(%s)", tokID, authID), []byte(tokVal), ltRealm, ltNonce, []byte(authVal))
			}
		}

		seenPairs := map[string]bool{}
		payloadUserAdded := 0
		for _, te := range rb.TE2 {
			tokID := strings.TrimSpace(te.TokenID)
			authID := strings.TrimSpace(te.AuthTokenID)
			if tokID == "" || authID == "" {
				continue
			}
			k := tokID + ":" + authID
			if seenPairs[k] {
				continue
			}
			seenPairs[k] = true
			tok := ""
			auth := ""
			var tokRawBytes []byte
			var authRawBytes []byte
			for _, t := range rb.Tokens {
				if strings.TrimSpace(t.ID) == tokID {
					tok = t.Value
					tokRawBytes = t.Bytes()
					break
				}
			}
			for _, a := range rb.Auth {
				if strings.TrimSpace(a.ID) == authID {
					auth = a.Value
					authRawBytes = a.Bytes()
					break
				}
			}
			if tok != "" && auth != "" {
				// Try TE2 payload as USERNAME with key=token/auth raw bytes (budget-limited).
				// This tests the mapping: payload selects/derives the auth key rather than UUID.
				if te2PayloadMax > 0 && payloadUserAdded < te2PayloadMax && len(te.Payload) > 0 {
					appendCand(&cands, fmt.Sprintf("te2 payload-user(relay=%s,len=%d) tok(%s) auth(%s):u=payload:key=authRaw", te.RelayName, len(te.Payload), tokID, authID), te.Payload, authRawBytes)
					appendCand(&cands, fmt.Sprintf("te2 payload-user(relay=%s,len=%d) tok(%s) auth(%s):u=payload:key=tokRaw", te.RelayName, len(te.Payload), tokID, authID), te.Payload, tokRawBytes)
					// Also try the original text form (often base64) as the key.
					// Some short-term schemes use ASCII secrets directly instead of decoded bytes.
					appendCand(&cands, fmt.Sprintf("te2 payload-user(relay=%s,len=%d) tok(%s) auth(%s):u=payload:key=authText", te.RelayName, len(te.Payload), tokID, authID), te.Payload, []byte(strings.TrimSpace(auth)))
					appendCand(&cands, fmt.Sprintf("te2 payload-user(relay=%s,len=%d) tok(%s) auth(%s):u=payload:key=tokText", te.RelayName, len(te.Payload), tokID, authID), te.Payload, []byte(strings.TrimSpace(tok)))
					payloadUserAdded++
				}

				// Optional: TURN REST-style derived password.
				// NOTE: In observed offers, te2 payload length is 6/18 and looks like endpoint bytes,
				// so it is likely NOT a TURN username. Instead, try small stable usernames: tokID/authID
				// and relay UUID/PID variants.
				if tryREST {
					usernames := make([]labeledBytes, 0, 16)
					usernames = append(usernames, labeledBytes{label: fmt.Sprintf("tokID(%s)", tokID), data: []byte(tokID)})
					usernames = append(usernames, labeledBytes{label: fmt.Sprintf("authID(%s)", authID), data: []byte(authID)})
					for _, uv := range userVariants {
						usernames = append(usernames, uv)
					}

					secrets := make([]labeledBytes, 0, 8)
					// Offer-derived enc/pkmsg is a strong candidate for the short-term secret.
					// Keep it early so it survives truncation.
					if len(encRaw) > 0 {
						secrets = append(secrets, labeledBytes{label: fmt.Sprintf("enc.%s(raw,len=%d)", strings.TrimSpace(encBlock.Type), len(encRaw)), data: encRaw})
					}
					if len(encHash) > 0 {
						secrets = append(secrets, labeledBytes{label: "enc.sha256", data: encHash})
					}
					if len(relayKeyDecoded) > 0 {
						secrets = append(secrets, labeledBytes{label: "relay.key", data: relayKeyDecoded})
					}
					if len(relayKeyDecoded2) > 0 {
						secrets = append(secrets, labeledBytes{label: "relay.key2", data: relayKeyDecoded2})
					}
					// Also try the relay key as textual bytes (the base64 string itself).
					// Some TURN REST deployments use a shared secret as ASCII, not decoded bytes.
					if s := strings.TrimSpace(rb.Key); s != "" {
						secrets = append(secrets, labeledBytes{label: "relay.key(text)", data: []byte(s)})
					}
					if len(hbhKeyDecoded) > 0 {
						secrets = append(secrets, labeledBytes{label: "relay.hbh", data: hbhKeyDecoded})
					}
					if len(hbhKeyDecoded2) > 0 {
						secrets = append(secrets, labeledBytes{label: "relay.hbh2", data: hbhKeyDecoded2})
					}
					if s := strings.TrimSpace(rb.HBHKey); s != "" {
						secrets = append(secrets, labeledBytes{label: "relay.hbh(text)", data: []byte(s)})
					}
					if len(tokRawBytes) > 0 {
						secrets = append(secrets, labeledBytes{label: "tokRaw", data: tokRawBytes})
					}
					if len(authRawBytes) > 0 {
						secrets = append(secrets, labeledBytes{label: "authRaw", data: authRawBytes})
					}
					// Keep bounded to avoid runaway.
					if len(usernames) > 10 {
						usernames = usernames[:10]
					}
					if len(secrets) > 8 {
						secrets = secrets[:8]
					}
					for _, u := range usernames {
						for _, s := range secrets {
							tryAppendTurnRESTCandidates(&cands, fmt.Sprintf("rest %s u=%s", s.label, u.label), u.data, s.data)
						}
					}
				}

				// tok/auth are base64 strings; use Bytes() for raw token/auth bytes.
				tokRaw := tokRawBytes
				authRaw := authRawBytes
				var tokDec []byte
				var authDec []byte
				var tokHex []byte
				var authHex []byte
				if b, ok := decodeMaybeBase64(tok); ok {
					tokDec = b
				} else if b, ok := decodeMaybeHex(tok); ok {
					tokHex = b
				}
				if b, ok := decodeMaybeBase64(auth); ok {
					authDec = b
				} else if b, ok := decodeMaybeHex(auth); ok {
					authHex = b
				}

				pairs := []struct {
					label string
					u     []byte
					k     []byte
				}{
					{label: "raw", u: tokRaw, k: authRaw},
					{label: "raw(rev)", u: authRaw, k: tokRaw},
				}
				if len(tokDec) > 0 {
					pairs = append(pairs, struct {
						label string
						u     []byte
						k     []byte
					}{label: "tok=dec", u: tokDec, k: authRaw})
					pairs = append(pairs, struct {
						label string
						u     []byte
						k     []byte
					}{label: "tok=dec(rev)", u: authRaw, k: tokDec})
				}
				if len(authDec) > 0 {
					pairs = append(pairs, struct {
						label string
						u     []byte
						k     []byte
					}{label: "auth=dec", u: tokRaw, k: authDec})
					pairs = append(pairs, struct {
						label string
						u     []byte
						k     []byte
					}{label: "auth=dec(rev)", u: authDec, k: tokRaw})
				}
				if len(tokDec) > 0 && len(authDec) > 0 {
					pairs = append(pairs, struct {
						label string
						u     []byte
						k     []byte
					}{label: "tok+auth=dec", u: tokDec, k: authDec})
					pairs = append(pairs, struct {
						label string
						u     []byte
						k     []byte
					}{label: "tok+auth=dec(rev)", u: authDec, k: tokDec})
				}

				if len(tokHex) > 0 {
					pairs = append(pairs, struct {
						label string
						u     []byte
						k     []byte
					}{label: "tok=hex", u: tokHex, k: authRaw})
					pairs = append(pairs, struct {
						label string
						u     []byte
						k     []byte
					}{label: "tok=hex(rev)", u: authRaw, k: tokHex})
				}
				if len(authHex) > 0 {
					pairs = append(pairs, struct {
						label string
						u     []byte
						k     []byte
					}{label: "auth=hex", u: tokRaw, k: authHex})
					pairs = append(pairs, struct {
						label string
						u     []byte
						k     []byte
					}{label: "auth=hex(rev)", u: authHex, k: tokRaw})
				}
				if len(tokHex) > 0 && len(authHex) > 0 {
					pairs = append(pairs, struct {
						label string
						u     []byte
						k     []byte
					}{label: "tok+auth=hex", u: tokHex, k: authHex})
					pairs = append(pairs, struct {
						label string
						u     []byte
						k     []byte
					}{label: "tok+auth=hex(rev)", u: authHex, k: tokHex})
				}

				for _, p := range pairs {
					appendUserKeyVariants(&cands, fmt.Sprintf("te2(%s) tok(%s,%d) auth(%s,%d)", p.label, te.TokenID, len(tok), te.AuthTokenID, len(auth)), p.u, p.k)
				}

				// Also try TURN-style selection with short USERNAME=tokenID/authTokenID strings.
				tokID := []byte(strings.TrimSpace(te.TokenID))
				authID := []byte(strings.TrimSpace(te.AuthTokenID))
				if len(tokID) > 0 {
					appendKeyVariants(&cands, fmt.Sprintf("te2 user=tokID(%s)", te.TokenID), tokID, authRaw)
					appendKeyVariants(&cands, fmt.Sprintf("te2 user=tokID(%s)", te.TokenID), tokID, tokRaw)
					appendMD5RealmEmptyCandidates(&cands, fmt.Sprintf("te2 user=tokID(%s)", te.TokenID), tokID, authRaw)
					appendMD5RealmEmptyCandidates(&cands, fmt.Sprintf("te2 user=tokID(%s)", te.TokenID), tokID, tokRaw)
				}
				if len(authID) > 0 {
					appendKeyVariants(&cands, fmt.Sprintf("te2 user=authID(%s)", te.AuthTokenID), authID, tokRaw)
					appendKeyVariants(&cands, fmt.Sprintf("te2 user=authID(%s)", te.AuthTokenID), authID, authRaw)
					appendMD5RealmEmptyCandidates(&cands, fmt.Sprintf("te2 user=authID(%s)", te.AuthTokenID), authID, tokRaw)
					appendMD5RealmEmptyCandidates(&cands, fmt.Sprintf("te2 user=authID(%s)", te.AuthTokenID), authID, authRaw)
				}

				if deriveHMAC {
					uuidBytes := []byte(strings.TrimSpace(rb.UUID))
					msgSeeds := buildSeedMessages(tokRaw, authRaw, uuidBytes)
					if len(encRaw) > 0 {
						msgSeeds = append(msgSeeds, seedMsgVariant{label: fmt.Sprintf("0enc.%s(raw,len=%d)", strings.TrimSpace(encBlock.Type), len(encRaw)), data: append([]byte{}, encRaw...)})
					}
					if len(encHash) > 0 {
						msgSeeds = append(msgSeeds, seedMsgVariant{label: "0enc.sha256", data: append([]byte{}, encHash...)})
					}
					for _, s := range msgSeeds {
						appendDerivedKeyCandidates(&cands, fmt.Sprintf("drv relay.key %s tok(%s) auth(%s)", s.label, te.TokenID, te.AuthTokenID), defaultUser, relayKeyDecoded, s.data)
						appendDerivedKeyCandidates(&cands, fmt.Sprintf("drv relay.hbh %s tok(%s) auth(%s)", s.label, te.TokenID, te.AuthTokenID), defaultUser, hbhKeyDecoded, s.data)
					}

					// Also try decoded tokens/auth as seed messages (base64url/raw, or hex).
					if len(tokDec) > 0 || len(authDec) > 0 || len(tokHex) > 0 || len(authHex) > 0 {
						seedTok := firstNonEmpty(firstNonEmpty(tokDec, tokHex), tokRaw)
						seedAuth := firstNonEmpty(firstNonEmpty(authDec, authHex), authRaw)
						msgSeedsDec := buildSeedMessages(seedTok, seedAuth, uuidBytes)
						if len(encRaw) > 0 {
							msgSeedsDec = append(msgSeedsDec, seedMsgVariant{label: fmt.Sprintf("0enc.%s(raw,len=%d)", strings.TrimSpace(encBlock.Type), len(encRaw)), data: append([]byte{}, encRaw...)})
						}
						if len(encHash) > 0 {
							msgSeedsDec = append(msgSeedsDec, seedMsgVariant{label: "0enc.sha256", data: append([]byte{}, encHash...)})
						}
						for _, s := range msgSeedsDec {
							appendDerivedKeyCandidates(&cands, fmt.Sprintf("drv(dec) relay.key %s tok(%s) auth(%s)", s.label, te.TokenID, te.AuthTokenID), defaultUser, relayKeyDecoded, s.data)
							appendDerivedKeyCandidates(&cands, fmt.Sprintf("drv(dec) relay.hbh %s tok(%s) auth(%s)", s.label, te.TokenID, te.AuthTokenID), defaultUser, hbhKeyDecoded, s.data)
						}
					}
				}
			}
		}
	}

	// Fallback: brute a few direct token/auth values.
	if derive {
		for i := 0; i < len(rb.Tokens) && i < 3; i++ {
			t := rb.Tokens[i]
			appendKeyVariants(&cands, fmt.Sprintf("token-only(id=%s,len=%d)", t.ID, len(t.Value)), []byte(t.Value), []byte(t.Value))
		}
		for i := 0; i < len(rb.Auth) && i < 2; i++ {
			t := rb.Auth[i]
			appendKeyVariants(&cands, fmt.Sprintf("auth-only(id=%s,len=%d)", t.ID, len(t.Value)), []byte(t.Value), []byte(t.Value))
		}
	}
	if tryIntegrity {
		if len(cands) == 0 {
			cm.logger.Warnf("⚠️📡 [RELAY-TURN] No integrity candidates available (CallID=%s)", callID)
			// Still dump if requested.
			if !envTruthy("QP_CALL_DUMP_TURN_PROBE") {
				return nil
			}
			tryIntegrity = false
		}
	}

	// Category counts to understand what will be tried under maxTry.
	countLT := 0
	countMiKDF := 0
	countDrv := 0
	countTE2 := 0
	countRelayKey := 0
	countRelayHBH := 0
	countTokOnly := 0
	countAuthOnly := 0
	countREST := 0
	if tryIntegrity {
		cands = dedupeTurnIntegrityCandidates(cands)
	}
	for _, c := range cands {
		l := c.label
		switch {
		case strings.HasPrefix(l, "lt "):
			countLT++
		case strings.Contains(l, "miKDF"):
			countMiKDF++
		case strings.HasPrefix(l, "rest "):
			countREST++
		case strings.HasPrefix(l, "drv ") || strings.Contains(l, ":drv="):
			countDrv++
		case strings.HasPrefix(l, "te2 ") || strings.HasPrefix(l, "te2("):
			countTE2++
		case strings.HasPrefix(l, "relay.key("):
			countRelayKey++
		case strings.HasPrefix(l, "relay.hbh_key("):
			countRelayHBH++
		case strings.HasPrefix(l, "token-only("):
			countTokOnly++
		case strings.HasPrefix(l, "auth-only("):
			countAuthOnly++
		}
	}
	if tryIntegrity {
		cm.logger.Warnf("📡 [RELAY-TURN] Candidate buckets: lt=%d mikdf=%d rest=%d relay.key=%d relay.hbh=%d te2=%d drv=%d tokenOnly=%d authOnly=%d (CallID=%s)", countLT, countMiKDF, countREST, countRelayKey, countRelayHBH, countTE2, countDrv, countTokOnly, countAuthOnly, callID)
	}

	if tryIntegrity && deriveHMAC {
		sort.SliceStable(cands, func(i, j int) bool {
			pi := 10
			pj := 10

			li := cands[i].label
			lj := cands[j].label

			usernameRank := func(label string) int {
				// Prefer UUID-based usernames first. Using peer/self ids is convenient but may be rejected
				// if the relay key selection expects UUID (or UUID-derived) usernames.
				s := label
				s = strings.ReplaceAll(s, " ", "")
				s = strings.ToLower(s)
				s = strings.ReplaceAll(s, "u=", "u=")
				switch {
				case strings.Contains(s, "payload-user("):
					// Force TE2 payload-as-USERNAME mapping candidates (token/auth-keyed) early.
					// These can otherwise be starved under the default per-family TE2 budget.
					return 0
				case strings.Contains(s, "u=uuid(bin)"):
					return 1
				case strings.Contains(s, "u=uuid"):
					return 2
				case strings.Contains(s, "u=uuid:self") || strings.Contains(s, "u=self:uuid"):
					return 3
				case strings.Contains(s, "u=uuid:peer") || strings.Contains(s, "u=peer:uuid"):
					return 4
				case strings.Contains(s, "u=self:peer") || strings.Contains(s, "u=peer:self"):
					return 5
				case strings.Contains(s, "u=self"):
					return 6
				case strings.Contains(s, "u=peer"):
					return 7
				case strings.Contains(s, "accept.te"):
					return 8
				default:
					return 9
				}
			}

			encodingRank := func(label string) int {
				// Prefer raw USERNAME/key encodings first; then base64/hex text variants.
				s := strings.ToLower(label)
				switch {
				case strings.Contains(s, "uenc=raw") || strings.Contains(s, "key=raw") || strings.Contains(s, "key=authraw") || strings.Contains(s, "key=tokraw"):
					return 0
				case strings.Contains(s, "uenc=b64") || strings.Contains(s, "key=b64"):
					return 1
				case strings.Contains(s, "uenc=hex") || strings.Contains(s, "key=hex"):
					return 2
				default:
					return 5
				}
			}

			prio := func(label string) int {
				// Keep this intentionally simple: try likely real TURN secrets first.
				// Derived candidates can be very numerous; keep them later so direct keys are not starved by maxTry.
				if strings.HasPrefix(label, "lt ") {
					return 0
				}
				if strings.Contains(label, "enc.plain.proto.f10.f1+te2.") && strings.Contains(label, "miKDF:msg=alloc.preimage") {
					return 1
				}
				if strings.Contains(label, "te2.") && strings.Contains(label, "+enc.plain.proto.f10.f1") && strings.Contains(label, "miKDF:msg=alloc.preimage") {
					return 2
				}
				if strings.Contains(label, "enc.plain.proto.f10.f1.ecdh(signedprekey)") && strings.Contains(label, "miKDF:msg=te2.") && strings.Contains(label, "prefix8+tail4") {
					return 3
				}
				if strings.Contains(label, "enc.plain.proto.f10.f1.ecdh(signedprekey)") && strings.Contains(label, "miKDF:msg=te2.") && strings.Contains(label, "full18") {
					return 4
				}
				if strings.Contains(label, "enc.plain.proto.f10.f1.ecdh(identity)") && strings.Contains(label, "miKDF:msg=te2.") && strings.Contains(label, "prefix8+tail4") {
					return 5
				}
				if strings.Contains(label, "enc.plain.proto.f10.f1.ecdh(identity)") && strings.Contains(label, "miKDF:msg=te2.") && strings.Contains(label, "full18") {
					return 6
				}
				if strings.Contains(label, "enc.plain.proto.f10.f1") && strings.Contains(label, "miKDF:msg=te2.") && strings.Contains(label, "prefix8+tail4") {
					return 7
				}
				if strings.Contains(label, "enc.plain.proto.f10.f1") && strings.Contains(label, "miKDF:msg=te2.") && strings.Contains(label, "full18") {
					return 8
				}
				if strings.Contains(label, "enc.plain.proto.f10.f1") && strings.Contains(label, "miKDF:msg=te2.") && strings.Contains(label, "prefix8") {
					return 9
				}
				if strings.Contains(label, "enc.plain.proto.f10.f1") && strings.Contains(label, "miKDF:msg=te2.") && strings.Contains(label, "tail4") {
					return 10
				}
				if strings.Contains(label, "enc.plain.proto.f10.f1.ecdh(signedprekey)") && strings.Contains(label, "+uuid(bin)+attrs+self+peer") && strings.Contains(label, "miKDF") {
					return 9
				}
				if strings.Contains(label, "enc.plain.proto.f10.f1.ecdh(signedprekey)") && strings.Contains(label, "+attrs") && strings.Contains(label, "miKDF") {
					return 10
				}
				if strings.Contains(label, "enc.plain.proto.f10.f1.ecdh(signedprekey)") && strings.Contains(label, "+uuid(bin)") && strings.Contains(label, "miKDF") {
					return 11
				}
				if strings.Contains(label, "enc.plain.proto.f10.f1.ecdh(signedprekey)") && strings.Contains(label, "miKDF") {
					return 12
				}
				if strings.Contains(label, "enc.plain.proto.f10.f1.ecdh(identity)") && strings.Contains(label, "+uuid(bin)+attrs+self+peer") && strings.Contains(label, "miKDF") {
					return 13
				}
				if strings.Contains(label, "enc.plain.proto.f10.f1.ecdh(identity)") && strings.Contains(label, "+attrs") && strings.Contains(label, "miKDF") {
					return 14
				}
				if strings.Contains(label, "enc.plain.proto.f10.f1.ecdh(identity)") && strings.Contains(label, "+uuid(bin)") && strings.Contains(label, "miKDF") {
					return 15
				}
				if strings.Contains(label, "enc.plain.proto.f10.f1.ecdh(identity)") && strings.Contains(label, "miKDF") {
					return 16
				}
				if strings.Contains(label, "enc.plain.proto.f10.f1+uuid(bin)+attrs+self+peer") && strings.Contains(label, "miKDF") {
					return 17
				}
				if strings.Contains(label, "enc.plain.proto.f10.f1+a4000+a4024+a0016") && strings.Contains(label, "miKDF") {
					return 18
				}
				if strings.Contains(label, "enc.plain.proto.f10.f1+a4000") && strings.Contains(label, "miKDF") {
					return 19
				}
				if strings.Contains(label, "enc.plain.proto.f10.f1+a4024") && strings.Contains(label, "miKDF") {
					return 20
				}
				if strings.Contains(label, "enc.plain.proto.f10.f1+a0016") && strings.Contains(label, "miKDF") {
					return 21
				}
				if strings.Contains(label, "enc.plain.proto.f10.f1+uuid(bin)") && strings.Contains(label, "miKDF") {
					return 22
				}
				if strings.Contains(label, "enc.plain.proto.f10.f1+self+peer") && strings.Contains(label, "miKDF") {
					return 23
				}
				if strings.Contains(label, "enc.plain.proto.f10.f1") && strings.Contains(label, "miKDF") {
					return 24
				}
				if strings.Contains(label, "enc.plain.proto.f10") && strings.Contains(label, "miKDF") {
					return 25
				}
				if strings.Contains(label, "enc.plain.proto.") && strings.Contains(label, ".f1") && strings.Contains(label, "miKDF") {
					return 26
				}
				if strings.HasPrefix(label, "enc.plain") {
					if strings.Contains(label, "miKDF") && strings.Contains(label, "alloc.preimage") {
						return 27
					}
					if strings.Contains(label, "miKDF") {
						return 28
					}
					return 29
				}
				if strings.Contains(label, "miKDF") {
					if strings.Contains(label, "alloc.preimage") {
						return 30
					}
					if strings.Contains(label, "uuid(bin)+self+peer") || strings.Contains(label, "self+peer") {
						return 31
					}
					return 32
				}
				// Direct extra-attrs-as-key attempts: small but high-signal.
				if strings.HasPrefix(label, "extra-key(") {
					return 29
				}
				// Offer-derived enc candidates are high-signal and should be tried early.
				if strings.HasPrefix(label, "enc.") {
					return 30
				}
				// REST-like candidates are numerous and low-signal at this stage.
				// Keep them behind the decrypted-offer/ECDH families so they don't consume maxTry.
				if strings.HasPrefix(label, "rest ") {
					return 27
				}
				// TE2 mapping (token_id/auth_token_id) is often used as TURN username/password.
				// In practice, relay.key()/hbh_key variants are so numerous that TE2 can be starved under maxTry.
				// Prioritize TE2 earlier so we get broader TE2 coverage per call.
				if strings.HasPrefix(label, "te2 ") || strings.HasPrefix(label, "te2(") {
					return 31
				}
				if strings.HasPrefix(label, "relay.key(") {
					return 32
				}
				if strings.HasPrefix(label, "relay.hbh_key(") {
					return 33
				}
				if strings.HasPrefix(label, "drv ") || strings.Contains(label, ":drv=") {
					return 34
				}
				if strings.HasPrefix(label, "token-only(") {
					return 35
				}
				if strings.HasPrefix(label, "auth-only(") {
					return 36
				}
				return 36
			}

			pi = prio(li)
			pj = prio(lj)
			if pi != pj {
				return pi < pj
			}

			ui := usernameRank(li)
			uj := usernameRank(lj)
			if ui != uj {
				return ui < uj
			}

			ei := encodingRank(li)
			ej := encodingRank(lj)
			if ei != ej {
				return ei < ej
			}

			return li < lj
		})
	}

	// Try up to a few candidates, each with its own transaction ID.
	// NOTE: We don't log any raw token/auth bytes here.
	maxTry := 10
	if v := envInt("QP_CALL_RELAY_TURN_MAX_CANDIDATES", 10); v > 0 {
		maxTry = clampInt(v, 1, 120)
	}
	cm.logger.Warnf("📡 [RELAY-TURN] Integrity candidates: total=%d maxTry=%d deriveHMAC=%v (CallID=%s)", len(cands), maxTry, deriveHMAC, callID)
	if maxTry > 0 {
		labels := make([]string, 0, 6)
		for i := 0; i < len(cands) && i < 6 && i < maxTry; i++ {
			labels = append(labels, cands[i].label)
		}
		cm.logger.Warnf("📡 [RELAY-TURN] First candidates: %s (CallID=%s)", strings.Join(labels, " | "), callID)
	}
	tryMI256 := envTruthy("QP_CALL_RELAY_TURN_TRY_MI_SHA256")
	ltAppended := false
	// Collect attempt results for optional dump (labels only, no secrets).
	attemptResults := make([]callTurnProbeAttemptResult, 0, maxTry)
	familyOf := func(label string) string {
		if strings.Contains(label, "miKDF") {
			return "mikdf"
		}
		if strings.HasPrefix(label, "lt ") {
			return "lt"
		}
		if strings.HasPrefix(label, "enc.") {
			return "enc"
		}
		if strings.HasPrefix(label, "rest ") {
			return "rest"
		}
		if strings.HasPrefix(label, "relay.key(") {
			return "relay.key"
		}
		if strings.HasPrefix(label, "relay.hbh_key(") {
			return "relay.hbh"
		}
		if strings.HasPrefix(label, "te2 ") || strings.HasPrefix(label, "te2(") {
			return "te2"
		}
		if strings.HasPrefix(label, "drv ") || strings.Contains(label, ":drv=") {
			return "drv"
		}
		if strings.HasPrefix(label, "extra-key(") {
			return "extra-key"
		}
		if strings.HasPrefix(label, "token-only(") {
			return "token-only"
		}
		if strings.HasPrefix(label, "auth-only(") {
			return "auth-only"
		}
		return "other"
	}

	// Per-family budgets (caps). These defaults are tuned for diversity within maxTry.
	budgets := map[string]int{
		"enc":        clampInt(envInt("QP_CALL_RELAY_TURN_ENC_BUDGET", 20), 0, maxTry),
		"mikdf":      clampInt(envInt("QP_CALL_RELAY_TURN_MIKDF_BUDGET", 20), 0, maxTry),
		"rest":       clampInt(envInt("QP_CALL_RELAY_TURN_REST_BUDGET", 20), 0, maxTry),
		"extra-key":  clampInt(envInt("QP_CALL_RELAY_TURN_EXTRA_KEY_BUDGET", 15), 0, maxTry),
		"relay.key":  clampInt(envInt("QP_CALL_RELAY_TURN_RELAY_KEY_BUDGET", 30), 0, maxTry),
		"relay.hbh":  clampInt(envInt("QP_CALL_RELAY_TURN_RELAY_HBH_BUDGET", 30), 0, maxTry),
		"te2":        clampInt(envInt("QP_CALL_RELAY_TURN_TE2_BUDGET", 25), 0, maxTry),
		"drv":        clampInt(envInt("QP_CALL_RELAY_TURN_DRV_BUDGET", 15), 0, maxTry),
		"token-only": clampInt(envInt("QP_CALL_RELAY_TURN_TOKEN_ONLY_BUDGET", 10), 0, maxTry),
		"auth-only":  clampInt(envInt("QP_CALL_RELAY_TURN_AUTH_ONLY_BUDGET", 10), 0, maxTry),
		"lt":         clampInt(envInt("QP_CALL_RELAY_TURN_LT_BUDGET", 5), 0, maxTry),
		"other":      clampInt(envInt("QP_CALL_RELAY_TURN_OTHER_BUDGET", 10), 0, maxTry),
	}
	triedByFamily := map[string]int{}
	attemptIndex := 0
	if tryIntegrity {
		for i := 0; i < len(cands) && attemptIndex < maxTry; i++ {
			c := cands[i]
			fam := familyOf(c.label)
			capV, ok := budgets[fam]
			if ok {
				if capV == 0 || triedByFamily[fam] >= capV {
					continue
				}
			}
			triedByFamily[fam] = triedByFamily[fam] + 1
			curAttemptIndex := attemptIndex
			attemptIndex++
			req := stun.MustBuild(stun.TransactionID)
			txid := stunTxIDHex(req)
			req.Type = stun.NewType(stun.Method(0x003), stun.ClassRequest)
			if !envTruthy("QP_CALL_RELAY_TURN_OMIT_REQUESTED_TRANSPORT") {
				req.Add(stun.AttrType(0x0019), []byte{17, 0, 0, 0})
			}
			addTurnExtraAttrsForEndpoint(req, extra, remote)
			if !forceNoUsername && len(c.username) > 0 {
				req.Add(stun.AttrUsername, c.username)
			}
			if c.longTerm {
				if len(c.realm) > 0 {
					_ = stun.NewRealm(string(c.realm)).AddTo(req)
				}
				if len(c.nonce) > 0 {
					_ = stun.NewNonce(string(c.nonce)).AddTo(req)
				}
			}

			// IMPORTANT: Encode BEFORE adding MESSAGE-INTEGRITY/FINGERPRINT.
			// Re-encoding after MESSAGE-INTEGRITY will invalidate the computed HMAC.
			req.Encode()
			attemptPreimage := append([]byte(nil), req.Raw...)
			if len(c.key) > 0 {
				if tryMI256 {
					if err := addMessageIntegritySHA256(req, c.key); err != nil {
						cm.logger.Warnf("⚠️📡 [RELAY-TURN] MI-SHA256 add failed, falling back to SHA1: cand=%s err=%v (CallID=%s)", c.label, err, callID)
						tryMI256 = false
						_ = stun.MessageIntegrity(c.key).AddTo(req)
					}
				} else {
					_ = stun.MessageIntegrity(c.key).AddTo(req)
				}
			}
			if envTruthyDefault("QP_CALL_RELAY_TURN_INCLUDE_FINGERPRINT", true) {
				_ = stun.Fingerprint.AddTo(req)
			}
			if captureReqs && len(requestDumps) < capMax {
				maybeCap(fmt.Sprintf("attempt_%d", curAttemptIndex), req, attemptPreimage)
			}
			if _, err := conn.Write(req.Raw); err != nil {
				if txid != "" {
					cm.logger.Warnf("⚠️📡 [RELAY-TURN] Integrity attempt write failed: txid=%s cand=%s err=%v (CallID=%s)", txid, c.label, err, callID)
				} else {
					cm.logger.Warnf("⚠️📡 [RELAY-TURN] Integrity attempt write failed: cand=%s err=%v (CallID=%s)", c.label, err, callID)
				}
				continue
			}
			_ = conn.SetReadDeadline(time.Now().Add(readTimeout))
			n2, err := conn.Read(buf)
			if err != nil {
				if txid != "" {
					cm.logger.Warnf("⚠️📡 [RELAY-TURN] Integrity attempt read failed: txid=%s cand=%s err=%v (CallID=%s)", txid, c.label, err, callID)
				} else {
					cm.logger.Warnf("⚠️📡 [RELAY-TURN] Integrity attempt read failed: cand=%s err=%v (CallID=%s)", c.label, err, callID)
				}
				continue
			}
			if n2 <= 0 {
				if txid != "" {
					cm.logger.Warnf("⚠️📡 [RELAY-TURN] Integrity attempt empty read: txid=%s cand=%s (CallID=%s)", txid, c.label, callID)
				} else {
					cm.logger.Warnf("⚠️📡 [RELAY-TURN] Integrity attempt empty read: cand=%s (CallID=%s)", c.label, callID)
				}
				continue
			}
			resp2Raw := buf[:n2]
			if !isSTUNPacket(resp2Raw) {
				if txid != "" {
					cm.logger.Warnf("⚠️📡 [RELAY-TURN] Integrity attempt non-STUN response: txid=%s cand=%s bytes=%d (CallID=%s)", txid, c.label, n2, callID)
				} else {
					cm.logger.Warnf("⚠️📡 [RELAY-TURN] Integrity attempt non-STUN response: cand=%s bytes=%d (CallID=%s)", c.label, n2, callID)
				}
				continue
			}
			var resp2 stun.Message
			resp2.Raw = resp2Raw
			if err := resp2.Decode(); err != nil {
				if txid != "" {
					cm.logger.Warnf("⚠️📡 [RELAY-TURN] Integrity attempt decode failed: txid=%s cand=%s err=%v (CallID=%s)", txid, c.label, err, callID)
				} else {
					cm.logger.Warnf("⚠️📡 [RELAY-TURN] Integrity attempt decode failed: cand=%s err=%v (CallID=%s)", c.label, err, callID)
				}
				continue
			}
			if txid != "" {
				respTxID := stunTxIDHex(&resp2)
				if respTxID != "" && respTxID != txid {
					cm.logger.Warnf("⚠️📡 [RELAY-TURN] Integrity attempt txid mismatch: req=%s resp=%s cand=%s relay=%s endpoint=%s (CallID=%s)", txid, respTxID, c.label, ep.RelayName, endpoint, callID)
				}
			}

			// Extract STUN error info if present.
			var ec stun.ErrorCodeAttribute
			code := 0
			reason := ""
			if err := ec.GetFrom(&resp2); err == nil {
				code = int(ec.Code)
				reason = strings.TrimSpace(string(ec.Reason))
			}

			respMsgType := stunMsgTypeHexFromMessage(&resp2)
			mappedEndpoint := ""
			extra4002Hex := ""
			if resp2.Type.Method == stun.Method(0x003) && resp2.Type.Class == stun.ClassSuccessResponse {
				mappedEndpoint, extra4002Hex = extractMappedEndpointAnd4002(&resp2)
			}

			// Optional: extract nonce/realm if server switches to long-term auth.
			var nonce stun.Nonce
			var realm stun.Realm
			_ = resp2.Parse(&nonce, &realm)
			nonceLen := len([]byte(nonce))
			realmLen := len([]byte(realm))

			algo := "sha1"
			if tryMI256 {
				algo = "sha256"
			}
			if txid != "" {
				cm.logger.Warnf("📡 [RELAY-TURN] Integrity attempt result: txid=%s cand=%s algo=%s userLen=%d keyLen=%d code=%d reason=%q nonceLen=%d realmLen=%d (CallID=%s)", txid, c.label, algo, len(c.username), len(c.key), code, reason, nonceLen, realmLen, callID)
			} else {
				cm.logger.Warnf("📡 [RELAY-TURN] Integrity attempt result: cand=%s algo=%s userLen=%d keyLen=%d code=%d reason=%q nonceLen=%d realmLen=%d (CallID=%s)", c.label, algo, len(c.username), len(c.key), code, reason, nonceLen, realmLen, callID)
			}
			attemptResults = append(attemptResults, callTurnProbeAttemptResult{
				Index:          curAttemptIndex,
				TxID:           txid,
				Cand:           c.label,
				Algo:           algo,
				UserLen:        len(c.username),
				KeyLen:         len(c.key),
				Code:           code,
				Reason:         reason,
				NonceLen:       nonceLen,
				RealmLen:       realmLen,
				Success:        resp2.Type.Method == stun.Method(0x003) && resp2.Type.Class == stun.ClassSuccessResponse,
				LongTerm:       c.longTerm,
				TryMI256:       tryMI256,
				Endpoint:       endpoint,
				RelayName:      ep.RelayName,
				RespMsgType:    respMsgType,
				MappedEndpoint: mappedEndpoint,
				Extra4002Hex:   extra4002Hex,
			})

			// If the server reveals REALM/NONCE only after integrity attempts, immediately add and try long-term candidates.
			if !ltAppended && realmLen > 0 && nonceLen > 0 {
				ltRealm = []byte(realm)
				ltNonce = []byte(nonce)
				if added := appendLongTermCandidatesFromRelayBlock(&cands, rb, ltRealm, ltNonce); added > 0 {
					ltAppended = true
					cm.logger.Warnf("🔐📡 [RELAY-TURN] Learned REALM/NONCE during attempts; appended long-term candidates: added=%d total=%d (CallID=%s)", added, len(cands), callID)
					if deriveHMAC {
						sort.SliceStable(cands, func(i2, j2 int) bool {
							li := cands[i2].label
							lj := cands[j2].label
							prio := func(label string) int {
								if strings.HasPrefix(label, "lt ") {
									return 0
								}
								if strings.HasPrefix(label, "rest ") {
									return 1
								}
								if strings.HasPrefix(label, "relay.key(") {
									return 2
								}
								if strings.HasPrefix(label, "relay.hbh_key(") {
									return 3
								}
								if strings.HasPrefix(label, "te2 ") || strings.HasPrefix(label, "te2(") {
									return 4
								}
								if strings.HasPrefix(label, "drv ") || strings.Contains(label, ":drv=") {
									return 5
								}
								if strings.HasPrefix(label, "token-only(") {
									return 6
								}
								if strings.HasPrefix(label, "auth-only(") {
									return 7
								}
								return 9
							}
							pi := prio(li)
							pj := prio(lj)
							if pi != pj {
								return pi < pj
							}
							return li < lj
						})
					}
					// Restart the loop so newly-added lt candidates are attempted early.
					i = -1
					continue
				}
			}

			// If relay can't decode requests with MI-SHA256, disable it for subsequent attempts.
			if tryMI256 {
				if code == 456 || code == 420 || strings.Contains(strings.ToLower(reason), "decode") || strings.Contains(strings.ToLower(reason), "unknown") {
					cm.logger.Warnf("⚠️📡 [RELAY-TURN] Disabling MI-SHA256 after server error: code=%d reason=%q (CallID=%s)", code, reason, callID)
					tryMI256 = false
				}
			}

			// Stop early only on meaningful state changes.
			if resp2.Type.Method == stun.Method(0x003) && resp2.Type.Class == stun.ClassSuccessResponse {
				cm.logger.Warnf("✅📡 [RELAY-TURN] Allocate SUCCESS candidate: cand=%s mapped=%s a4002=%s (CallID=%s)", c.label, mappedEndpoint, extra4002Hex, callID)
				break
			}
			if code == int(stun.CodeUnauthorized) || code == int(stun.CodeStaleNonce) || nonceLen > 0 || realmLen > 0 {
				cm.logger.Warnf("🔐📡 [RELAY-TURN] Server requested long-term auth: cand=%s code=%d (CallID=%s)", c.label, code, callID)
				break
			}
		}
	}

	if envTruthy("QP_CALL_DUMP_TURN_PROBE") {
		dump := callTurnProbeDump{
			Kind:                "TurnProbe",
			Captured:            time.Now().UTC().Format(time.RFC3339Nano),
			CallID:              callID,
			RelayName:           ep.RelayName,
			Endpoint:            endpoint,
			LocalAddr:           local,
			BaseAllocateTxID:    baseAllocateTxID,
			BaseAllocateSuccess: baseAllocateSuccess,
			BaseRespMsgType:     baseRespMsgType,
			BaseMappedEndpoint:  baseMappedEndpoint,
			BaseExtra4002Hex:    baseExtra4002Hex,
			BaseAllocateCode:    baseAllocateCode,
			BaseAllocateReason:  baseAllocateReason,
			BaseNonceLen:        baseNonceLen,
			BaseRealmLen:        baseRealmLen,
			DiscoveryTxID:       discTxID,
			DiscoveryUser:       discUser,
			DiscoveryCode:       discCode,
			DiscoveryReason:     discReason,
			DiscoveryNonceLen:   discNonceLen,
			DiscoveryRealmLen:   discRealmLen,
			RelayUUID:           strings.TrimSpace(rb.UUID),
			SelfPID:             strings.TrimSpace(rb.SelfPID),
			PeerPID:             strings.TrimSpace(rb.PeerPID),
			HasKey:              strings.TrimSpace(rb.Key) != "",
			HasHBHKey:           strings.TrimSpace(rb.HBHKey) != "",
			Buckets: map[string]int{
				"lt":        countLT,
				"mikdf":     countMiKDF,
				"relay_key": countRelayKey,
				"relay_hbh": countRelayHBH,
				"te2":       countTE2,
				"drv":       countDrv,
				"tokenOnly": countTokOnly,
				"authOnly":  countAuthOnly,
			},
			MaxTry:   maxTry,
			Requests: requestDumps,
			Attempts: attemptResults,
		}
		if p, e := DumpCallTurnProbeSummary(callID, dump); e == nil {
			cm.logger.Infof("💾 [TURN-PROBE-DUMP] Saved to %s (CallID=%s)", p, callID)
		} else {
			cm.logger.Warnf("⚠️ [TURN-PROBE-DUMP] Failed: %v (CallID=%s)", e, callID)
		}
	}

	return nil
}
