package srtp

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"errors"

	"github.com/nocodeleaks/quepasa/voip/calls/util"
	qplog "github.com/nocodeleaks/quepasa/qplog"
)

// errBadCallKeyLen is returned when the call key is not exactly 32 bytes, the only
// length the SFrame key derivation accepts.
var errBadCallKeyLen = errors.New("srtp: sframe call key must be exactly 32 bytes")

// SFrame KDF labels and lengths.
const (
	KDFLabelE2ESframe = "e2e sframe key"
	KDFLabelWarpAuth  = "warp auth key"
	gcmTagLen         = 16
	aesKeyLen         = 16
)

// FormatSframeParticipantID formats the participant id used as the SFrame HKDF info.
func FormatSframeParticipantID(jid string) string {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/sframe.rs#L33-L35
	return util.FormatParticipantID(jid)
}

// SframeInfoLabel builds the HKDF info label "e2e sframe key<participantID>".
func SframeInfoLabel(participantID string) string {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/sframe.rs#L37-L39
	return KDFLabelE2ESframe + participantID
}

// splitCallKey splits a 32-byte call key into (salt, ikm).
func splitCallKey(callKey []byte) (salt, ikm []byte, err error) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/sframe.rs#L22-L27
	if len(callKey) != 32 {
		return nil, nil, errBadCallKeyLen
	}
	return callKey[0:16], callKey[16:32], nil
}

// DeriveE2eSframeKeyForParticipant derives the 32-byte per-participant SFrame key
// from callKey (exactly 32B), salt = callKey[0:16], ikm = callKey[16:32].
func DeriveE2eSframeKeyForParticipant(callKey []byte, participantID string, log ...qplog.Logger) ([]byte, error) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/sframe.rs#L42-L54
	lg := pickLog(log)
	salt, ikm, err := splitCallKey(callKey)
	if err != nil {
		lg.DebugE().Err(err).Int("call_key_bytes", len(callKey)).Str("participant_id", participantID).Msg("sframe key derivation rejected: bad call key length")
		return nil, err
	}
	lg.DebugE().Str("participant_id", participantID).Msg("deriving e2e sframe key for participant")
	return util.HKDFSHA256(salt, ikm, []byte(SframeInfoLabel(participantID)), 32)
}

// DeriveWarpAuthKey derives the 32-byte WARP auth key from callKey (32B), empty
// salt, label "warp auth key".
func DeriveWarpAuthKey(callKey []byte, log ...qplog.Logger) ([]byte, error) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/sframe.rs#L56-L66
	lg := pickLog(log)
	if len(callKey) != 32 {
		lg.DebugE().Err(errBadCallKeyLen).Int("call_key_bytes", len(callKey)).Msg("warp auth key derivation rejected: bad call key length")
		return nil, errBadCallKeyLen
	}
	lg.DebugE().Msg("deriving warp auth key")
	return util.HKDFSHA256(nil, callKey, []byte(KDFLabelWarpAuth), 32)
}

// counterToIV builds the 16-byte GCM nonce: 8 zero bytes then counter as LE uint64.
func counterToIV(counter uint64) [16]byte {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/sframe.rs#L88-L92
	var iv [16]byte
	binary.LittleEndian.PutUint64(iv[8:16], counter)
	return iv
}

// buildSframeHeader encodes [varint counter || varint keyID || total-length byte].
// binary.AppendUvarint is the same unsigned LEB128 as the reference encode_varint.
func buildSframeHeader(counter, keyID uint64) []byte {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/sframe.rs#L94-L101
	h := binary.AppendUvarint(nil, counter)
	h = binary.AppendUvarint(h, keyID)
	return append(h, byte(len(h)+1))
}

// parseSframeHeader decodes the trailing header, validating the total-length byte.
func parseSframeHeader(header []byte) (counter, keyID uint64, ok bool) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/sframe.rs#L103-L115
	if len(header) < 2 {
		return 0, 0, false
	}
	if int(header[len(header)-1]) != len(header) {
		return 0, 0, false
	}
	body := header[:len(header)-1]
	counter, n := binary.Uvarint(body)
	if n <= 0 {
		return 0, 0, false
	}
	keyID, n2 := binary.Uvarint(body[n:])
	if n2 <= 0 {
		return 0, 0, false
	}
	return counter, keyID, true
}

// gcmEncrypt seals plaintext with AES-128-GCM under the non-standard 16-byte nonce.
func gcmEncrypt(key []byte, nonce16 [16]byte, plaintext []byte) ([]byte, error) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/sframe.rs#L117-L123
	gcm, err := newSframeGCM(key)
	if err != nil {
		return nil, err
	}
	return gcm.Seal(nil, nonce16[:], plaintext, nil), nil
}

// gcmDecrypt opens ciphertext+tag with AES-128-GCM; ok=false on any auth failure.
func gcmDecrypt(key []byte, nonce16 [16]byte, ciphertextWithTag []byte) ([]byte, bool) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/sframe.rs#L125-L129
	gcm, err := newSframeGCM(key)
	if err != nil {
		return nil, false
	}
	plain, err := gcm.Open(nil, nonce16[:], ciphertextWithTag, nil)
	if err != nil {
		return nil, false
	}
	return plain, true
}

// newSframeGCM builds AES-128-GCM with the non-standard 16-byte nonce size so the
// GHASH-derived J0 matches the reference.
func newSframeGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key[:aesKeyLen])
	if err != nil {
		return nil, err
	}
	return cipher.NewGCMWithNonceSize(block, 16)
}

// SframeSession holds the per-direction SFrame keys (encrypt for peer, decrypt for
// self) and the send-side counter.
type SframeSession struct {
	SelfParticipantID string
	PeerParticipantID string
	encryptKey        [16]byte
	decryptKey        [16]byte
	txCounter         uint64
	log               qplog.Logger
}

// NewSframeSession builds a session from callKey and the self/peer JIDs.
func NewSframeSession(callKey []byte, selfJID, peerJID string, opts ...Option) (*SframeSession, error) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/sframe.rs#L154-L172
	cfg := resolveConfig(opts)
	selfID := FormatSframeParticipantID(selfJID)
	peerID := FormatSframeParticipantID(peerJID)
	sendKey, err := DeriveE2eSframeKeyForParticipant(callKey, peerID, cfg.log)
	if err != nil {
		cfg.log.DebugE().Err(err).Str("peer_participant_id", peerID).Msg("sframe session send key derivation failed")
		return nil, err
	}
	recvKey, err := DeriveE2eSframeKeyForParticipant(callKey, selfID, cfg.log)
	if err != nil {
		cfg.log.DebugE().Err(err).Str("self_participant_id", selfID).Msg("sframe session recv key derivation failed")
		return nil, err
	}
	s := &SframeSession{SelfParticipantID: selfID, PeerParticipantID: peerID, log: cfg.log}
	copy(s.encryptKey[:], sendKey[:aesKeyLen])
	copy(s.decryptKey[:], recvKey[:aesKeyLen])
	cfg.log.DebugE().Str("self_participant_id", selfID).Str("peer_participant_id", peerID).Msg("sframe session established")
	return s, nil
}

// Encrypt seals one frame as [ciphertext || 16-byte tag || varint-header].
func (s *SframeSession) Encrypt(plaintext []byte) ([]byte, error) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/sframe.rs#L173-L187
	counter := s.txCounter
	s.txCounter++
	header := buildSframeHeader(counter, 0)
	iv := counterToIV(counter)
	encrypted, err := gcmEncrypt(s.encryptKey[:], iv, plaintext)
	if err != nil {
		s.log.DebugE().Err(err).Uint64("counter", counter).Msg("sframe encrypt failed")
		return nil, err
	}
	out := make([]byte, 0, len(encrypted)+len(header))
	out = append(out, encrypted...)
	out = append(out, header...)
	s.log.TraceE().Uint64("counter", counter).Int("plaintext_bytes", len(plaintext)).Int("header_bytes", len(header)).Int("frame_bytes", len(out)).Msg("sframe encrypted frame")
	return out, nil
}

// Decrypt classifies one frame. It returns (plaintext, true) when the trailing
// SFrame header parses and the GCM tag authenticates; otherwise (nil, false),
// meaning the frame is plain Opus the caller must use verbatim.
func (s *SframeSession) Decrypt(frame []byte) ([]byte, bool) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/sframe.rs#L188-L213
	if len(frame) < gcmTagLen+3 {
		s.log.DebugE().Int("frame_bytes", len(frame)).Msg("sframe decrypt: frame too short, treating as plain opus")
		return nil, false
	}
	headerLen := int(frame[len(frame)-1])
	if headerLen < 3 || headerLen > len(frame) {
		s.log.DebugE().Int("frame_bytes", len(frame)).Int("header_len", headerLen).Msg("sframe decrypt: bad header length, treating as plain opus")
		return nil, false
	}
	headerStart := len(frame) - headerLen
	header := frame[headerStart:]
	ciphertext := frame[:headerStart]
	if len(ciphertext) < gcmTagLen+1 {
		s.log.DebugE().Int("ciphertext_bytes", len(ciphertext)).Msg("sframe decrypt: ciphertext too short, treating as plain opus")
		return nil, false
	}
	counter, _, ok := parseSframeHeader(header)
	if !ok {
		s.log.DebugE().Int("header_bytes", len(header)).Msg("sframe decrypt: header parse failed, treating as plain opus")
		return nil, false
	}
	iv := counterToIV(counter)
	plain, ok := gcmDecrypt(s.decryptKey[:], iv, ciphertext)
	if !ok {
		s.log.DebugE().Uint64("counter", counter).Int("ciphertext_bytes", len(ciphertext)).Msg("sframe decrypt: gcm auth failed, treating as plain opus")
		return nil, false
	}
	s.log.TraceE().Uint64("counter", counter).Int("frame_bytes", len(frame)).Int("plaintext_bytes", len(plain)).Msg("sframe decrypted frame")
	return plain, true
}
