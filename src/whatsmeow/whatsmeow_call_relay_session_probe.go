package whatsmeow

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pion/stun"
)

type turnIntegrityCandidate struct {
	label    string
	username []byte
	key      []byte
	realm    []byte
	nonce    []byte
	longTerm bool
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

	go cm.runRelaySessionProbe(callID, best, endpoints)
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

func isSTUNPacket(b []byte) bool {
	if len(b) < 8 {
		return false
	}
	// STUN magic cookie (RFC 5389)
	return b[4] == 0x21 && b[5] == 0x12 && b[6] == 0xA4 && b[7] == 0x42
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
	baseReq := stun.MustBuild(stun.TransactionID)
	baseReq.Type = stun.NewType(stun.Method(0x003), stun.ClassRequest) // Allocate
	baseReq.Add(stun.AttrType(0x0019), []byte{17, 0, 0, 0})            // REQUESTED-TRANSPORT: UDP
	baseAllocateTxID = stunTxIDHex(baseReq)
	baseReq.Encode()
	if envTruthyDefault("QP_CALL_RELAY_TURN_INCLUDE_FINGERPRINT", true) {
		_ = stun.Fingerprint.AddTo(baseReq)
	}
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
	if baseAllocateTxID != "" {
		respTxID := stunTxIDHex(&resp)
		if respTxID != "" && respTxID != baseAllocateTxID {
			cm.logger.Warnf("⚠️📡 [RELAY-TURN] Base Allocate txid mismatch: req=%s resp=%s relay=%s endpoint=%s (CallID=%s)", baseAllocateTxID, respTxID, ep.RelayName, endpoint, callID)
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

	if !envTruthy("QP_CALL_RELAY_TURN_ALLOCATE_TRY_INTEGRITY") {
		return nil
	}

	rb := cm.getRelayBlockForCall(callID)
	if rb == nil {
		cm.logger.Warnf("⚠️📡 [RELAY-TURN] No RelayBlock available for integrity attempts (CallID=%s)", callID)
		return nil
	}

	// If the relay uses long-term auth, it should provide REALM and NONCE.
	// Sometimes servers only return these when USERNAME is present, so we do one discovery request.
	ltNonce := baseNonceB
	ltRealm := baseRealmB
	if (len(ltNonce) == 0 || len(ltRealm) == 0) && len(rb.TE2) > 0 {
		u := strings.TrimSpace(rb.TE2[0].TokenID)
		if u == "" {
			u = strings.TrimSpace(rb.TE2[0].AuthTokenID)
		}
		if u != "" {
			discUser = u
			disc := stun.MustBuild(stun.TransactionID)
			disc.Type = stun.NewType(stun.Method(0x003), stun.ClassRequest)
			disc.Add(stun.AttrType(0x0019), []byte{17, 0, 0, 0})
			disc.Add(stun.AttrUsername, []byte(u))
			discTxID = stunTxIDHex(disc)
			disc.Encode()
			if envTruthyDefault("QP_CALL_RELAY_TURN_INCLUDE_FINGERPRINT", true) {
				_ = stun.Fingerprint.AddTo(disc)
			}
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
	if encBlock != nil && len(encBlock.Raw) > 0 {
		encRaw = encBlock.Raw
		// Stable, fixed-size seed (useful if relays hash the offer payload first).
		encHash = sha256Sum(encRaw)
	}
	// For relay keys, try using relay UUID as USERNAME to help server-side key selection.
	defaultUser := []byte(strings.TrimSpace(rb.UUID))
	userVariants := buildRelayUsernameVariants(rb)
	if len(userVariants) == 0 {
		userVariants = []labeledBytes{{label: "uuid", data: defaultUser}}
	}
	// Offer enc (pkmsg/msg) candidates: try using enc bytes directly as MI key.
	// Rationale: relay short-term integrity secret may be derived from or equal to offer enc payload.
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
	deriveHMAC := envTruthyDefault("QP_CALL_RELAY_TURN_DERIVE_HMAC", true)
	var relayKeyDecoded []byte
	var relayKeyDecoded2 []byte
	var hbhKeyDecoded []byte
	var hbhKeyDecoded2 []byte
	// Relay keys: try both decoded bytes and the original base64 text itself.
	if rb.Key != "" {
		if b1, ok := decodeMaybeBase64(rb.Key); ok {
			relayKeyDecoded = b1
			for _, u := range userVariants {
				appendKeyVariantsUserEncodings(&cands, fmt.Sprintf("relay.key(dec1_b64,len=%d):u=%s", len(b1), u.label), u.data, b1)
				appendMD5RealmEmptyCandidates(&cands, fmt.Sprintf("relay.key(dec1_b64,len=%d):u=%s", len(b1), u.label), u.data, b1)
			}
			appendKeyVariantsNoUser(&cands, fmt.Sprintf("relay.key(dec1_b64,len=%d)", len(b1)), b1)
			if b2, depth, ok2 := decodeMaybeBase64Recursive(rb.Key, 2); ok2 && depth >= 2 && len(b2) > 0 && !bytes.Equal(b2, b1) {
				relayKeyDecoded2 = b2
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
			for _, u := range userVariants {
				appendKeyVariantsUserEncodings(&cands, fmt.Sprintf("relay.hbh_key(dec1_b64,len=%d):u=%s", len(b1), u.label), u.data, b1)
				appendMD5RealmEmptyCandidates(&cands, fmt.Sprintf("relay.hbh_key(dec1_b64,len=%d):u=%s", len(b1), u.label), u.data, b1)
			}
			appendKeyVariantsNoUser(&cands, fmt.Sprintf("relay.hbh_key(dec1_b64,len=%d)", len(b1)), b1)
			if b2, depth, ok2 := decodeMaybeBase64Recursive(rb.HBHKey, 2); ok2 && depth >= 2 && len(b2) > 0 && !bytes.Equal(b2, b1) {
				hbhKeyDecoded2 = b2
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
	if len(cands) == 0 {
		cm.logger.Warnf("⚠️📡 [RELAY-TURN] No integrity candidates available (CallID=%s)", callID)
		return nil
	}

	// Category counts to understand what will be tried under maxTry.
	countLT := 0
	countDrv := 0
	countTE2 := 0
	countRelayKey := 0
	countRelayHBH := 0
	countTokOnly := 0
	countAuthOnly := 0
	countREST := 0
	cands = dedupeTurnIntegrityCandidates(cands)
	for _, c := range cands {
		l := c.label
		switch {
		case strings.HasPrefix(l, "lt "):
			countLT++
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
	cm.logger.Warnf("📡 [RELAY-TURN] Candidate buckets: lt=%d rest=%d relay.key=%d relay.hbh=%d te2=%d drv=%d tokenOnly=%d authOnly=%d (CallID=%s)", countLT, countREST, countRelayKey, countRelayHBH, countTE2, countDrv, countTokOnly, countAuthOnly, callID)

	if deriveHMAC {
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
				case strings.Contains(s, "u=uuid(bin)"):
					return 0
				case strings.Contains(s, "u=uuid"):
					return 1
				case strings.Contains(s, "u=uuid:self") || strings.Contains(s, "u=self:uuid"):
					return 2
				case strings.Contains(s, "u=uuid:peer") || strings.Contains(s, "u=peer:uuid"):
					return 3
				case strings.Contains(s, "u=self:peer") || strings.Contains(s, "u=peer:self"):
					return 4
				case strings.Contains(s, "u=self"):
					return 5
				case strings.Contains(s, "u=peer"):
					return 6
				case strings.Contains(s, "accept.te"):
					return 7
				default:
					return 9
				}
			}

			encodingRank := func(label string) int {
				// Prefer raw USERNAME/key encodings first; then base64/hex text variants.
				s := strings.ToLower(label)
				switch {
				case strings.Contains(s, "uenc=raw") || strings.Contains(s, "key=raw"):
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
				// Offer-derived enc candidates are high-signal and should be tried early.
				if strings.HasPrefix(label, "enc.") {
					return 1
				}
				// REST-like candidates are derived from relays/TE2 material and must be tried early,
				// otherwise relay.key() alone can consume the whole maxTry window.
				if strings.HasPrefix(label, "rest ") {
					return 2
				}
				if strings.HasPrefix(label, "relay.key(") {
					return 3
				}
				if strings.HasPrefix(label, "relay.hbh_key(") {
					return 4
				}
				// TE2 mapping (token_id/auth_token_id) is often used as TURN username/password.
				// However, we've observed repeated HMAC mismatches; try direct relay keys first.
				if strings.HasPrefix(label, "te2 ") || strings.HasPrefix(label, "te2(") {
					return 5
				}
				if strings.HasPrefix(label, "drv ") || strings.Contains(label, ":drv=") {
					return 6
				}
				if strings.HasPrefix(label, "token-only(") {
					return 7
				}
				if strings.HasPrefix(label, "auth-only(") {
					return 8
				}
				return 9
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
		"rest":       clampInt(envInt("QP_CALL_RELAY_TURN_REST_BUDGET", 20), 0, maxTry),
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
		req.Add(stun.AttrType(0x0019), []byte{17, 0, 0, 0})
		if len(c.username) > 0 {
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
			Index:     curAttemptIndex,
			TxID:      txid,
			Cand:      c.label,
			Algo:      algo,
			UserLen:   len(c.username),
			KeyLen:    len(c.key),
			Code:      code,
			Reason:    reason,
			NonceLen:  nonceLen,
			RealmLen:  realmLen,
			Success:   resp2.Type.Class == stun.ClassSuccessResponse,
			LongTerm:  c.longTerm,
			TryMI256:  tryMI256,
			Endpoint:  endpoint,
			RelayName: ep.RelayName,
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
		if resp2.Type.Class == stun.ClassSuccessResponse {
			cm.logger.Warnf("✅📡 [RELAY-TURN] Allocate success candidate: cand=%s (CallID=%s)", c.label, callID)
			break
		}
		if code == int(stun.CodeUnauthorized) || code == int(stun.CodeStaleNonce) || nonceLen > 0 || realmLen > 0 {
			cm.logger.Warnf("🔐📡 [RELAY-TURN] Server requested long-term auth: cand=%s code=%d (CallID=%s)", c.label, code, callID)
			break
		}
	}

	if envTruthy("QP_CALL_DUMP_TURN_PROBE") {
		dump := callTurnProbeDump{
			Kind:               "TurnProbe",
			Captured:           time.Now().UTC().Format(time.RFC3339Nano),
			CallID:             callID,
			RelayName:          ep.RelayName,
			Endpoint:           endpoint,
			LocalAddr:          local,
			BaseAllocateTxID:   baseAllocateTxID,
			BaseAllocateCode:   baseAllocateCode,
			BaseAllocateReason: baseAllocateReason,
			BaseNonceLen:       baseNonceLen,
			BaseRealmLen:       baseRealmLen,
			DiscoveryTxID:      discTxID,
			DiscoveryUser:      discUser,
			DiscoveryCode:      discCode,
			DiscoveryReason:    discReason,
			DiscoveryNonceLen:  discNonceLen,
			DiscoveryRealmLen:  discRealmLen,
			RelayUUID:          strings.TrimSpace(rb.UUID),
			SelfPID:            strings.TrimSpace(rb.SelfPID),
			PeerPID:            strings.TrimSpace(rb.PeerPID),
			HasKey:             strings.TrimSpace(rb.Key) != "",
			HasHBHKey:          strings.TrimSpace(rb.HBHKey) != "",
			Buckets: map[string]int{
				"lt":        countLT,
				"relay_key": countRelayKey,
				"relay_hbh": countRelayHBH,
				"te2":       countTE2,
				"drv":       countDrv,
				"tokenOnly": countTokOnly,
				"authOnly":  countAuthOnly,
			},
			MaxTry:   maxTry,
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
