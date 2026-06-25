package srtp

import (
	"crypto/aes"
	"encoding/binary"
	"errors"

	"github.com/nocodeleaks/quepasa/voip/calls/util"
	qplog "github.com/nocodeleaks/quepasa/qplog"
)

// errBadHbhKeyLen is returned when the hop-by-hop key is not exactly 30 bytes,
// the only valid length the relay produces.
var errBadHbhKeyLen = errors.New("srtp: hbh key must be exactly 30 bytes")

// hbh session-key derivation labels (libsrtp srtp_kdf_generate).
const (
	labelRTPEncryption = 0x00
	labelRTPAuth       = 0x01
	labelRTPSalt       = 0x02
)

// SrtpKeyingMaterial is the 16-byte master key + 14-byte master salt split.
type SrtpKeyingMaterial struct {
	MasterKey  [16]byte
	MasterSalt [14]byte
}

// LibsrtpSessionKeys is the expanded per-session keying (AES_CM_128_HMAC_SHA1_80).
type LibsrtpSessionKeys struct {
	SessionKey  [16]byte
	SessionSalt [14]byte
	AuthKey     [20]byte
}

// keyingFromCryptoKey splits a 30-byte crypto key into master key (16) + salt (14).
func keyingFromCryptoKey(cryptoKey []byte) SrtpKeyingMaterial {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/hbh_srtp.rs#L34-L42
	var m SrtpKeyingMaterial
	copy(m.MasterKey[:], cryptoKey[0:16])
	copy(m.MasterSalt[:], cryptoKey[16:30])
	return m
}

// deriveHbhSrtpKeyWithLabels runs the two-stage WA-SFU KDF (HKDF-SHA256 with the
// literal label as info): stage 1 derives the srtcp salt, stage 2 the 30-byte key.
func deriveHbhSrtpKeyWithLabels(lg qplog.Logger, hbhKey []byte, saltLabel, keyLabel string) ([]byte, error) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/hbh_srtp.rs#L46-L64
	if len(hbhKey) != 30 {
		lg.DebugE().Err(errBadHbhKeyLen).Int("hbh_key_bytes", len(hbhKey)).Str("key_label", keyLabel).Msg("hbh key derivation rejected: bad key length")
		return nil, errBadHbhKeyLen
	}
	lg.DebugE().Int("hbh_key_bytes", len(hbhKey)).Str("salt_label", saltLabel).Str("key_label", keyLabel).Msg("deriving hbh srtp key")
	masterKey := hbhKey[0:16]
	masterSalt := hbhKey[16:30]
	srtcpSalt, err := util.HKDFSHA256(make([]byte, 32), masterSalt, []byte(saltLabel), 32)
	if err != nil {
		return nil, err
	}
	return util.HKDFSHA256(srtcpSalt, masterKey, []byte(keyLabel), 30)
}

// DeriveHbhSrtpKeyUplink derives the 30-byte uplink HBH SRTP key from hbhKey (30B).
func DeriveHbhSrtpKeyUplink(hbhKey []byte, log ...qplog.Logger) ([]byte, error) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/hbh_srtp.rs#L66-L68
	return deriveHbhSrtpKeyWithLabels(pickLog(log), hbhKey, "uplink hbh srtcp salt", "uplink hbh srtcp key")
}

// DeriveHbhSrtpKeyDownlink derives the 30-byte downlink HBH SRTP key from hbhKey (30B).
func DeriveHbhSrtpKeyDownlink(hbhKey []byte, log ...qplog.Logger) ([]byte, error) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/hbh_srtp.rs#L70-L72
	return deriveHbhSrtpKeyWithLabels(pickLog(log), hbhKey, "downlink hbh srtcp salt", "downlink hbh srtcp key")
}

// KeyingFromHbhKeyUplink derives the uplink key and splits it into keying material.
func KeyingFromHbhKeyUplink(hbhKey []byte, log ...qplog.Logger) (SrtpKeyingMaterial, error) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/hbh_srtp.rs#L74-L78
	k, err := DeriveHbhSrtpKeyUplink(hbhKey, pickLog(log))
	if err != nil {
		return SrtpKeyingMaterial{}, err
	}
	return keyingFromCryptoKey(k), nil
}

// KeyingFromHbhKeyDownlink derives the downlink key and splits it into keying material.
func KeyingFromHbhKeyDownlink(hbhKey []byte, log ...qplog.Logger) (SrtpKeyingMaterial, error) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/hbh_srtp.rs#L80-L84
	k, err := DeriveHbhSrtpKeyDownlink(hbhKey, pickLog(log))
	if err != nil {
		return SrtpKeyingMaterial{}, err
	}
	return keyingFromCryptoKey(k), nil
}

// aesICMKey30 concatenates a 16-byte AES key with a 14-byte salt into the 30-byte
// libsrtp AES-ICM key layout.
func aesICMKey30(aesKey, salt []byte) [30]byte {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/hbh_srtp.rs#L87-L92
	var out [30]byte
	copy(out[:16], aesKey[:16])
	copy(out[16:30], salt[:14])
	return out
}

// aesICMCrypt is libsrtp AES-ICM: counter = (salt padded to 16) XOR iv, keystream =
// AES(counter), counter increments byte 15 with a single carry into byte 14 (2-level,
// NOT a 128-bit CTR — this divergence is faithful to libsrtp and load-bearing).
func aesICMCrypt(key30, iv16, data []byte) ([]byte, error) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/hbh_srtp.rs#L96-L121
	aesKey := key30[:16]
	salt := key30[16:30]
	var counter [16]byte
	copy(counter[:14], salt)
	for i := range 16 {
		counter[i] ^= iv16[i]
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	out := append([]byte(nil), data...)
	var ks [16]byte
	for pos := 0; pos < len(out); {
		block.Encrypt(ks[:], counter[:])
		n := 16
		if rem := len(out) - pos; rem < 16 {
			n = rem
		}
		for i := range n {
			out[pos+i] ^= ks[i]
		}
		pos += n
		counter[15]++
		if counter[15] == 0 {
			counter[14]++
		}
	}
	return out, nil
}

// deriveSessionBytes is libsrtp srtp_kdf_generate: IV all-zero except byte 7 = label.
func deriveSessionBytes(masterKey, masterSalt []byte, label byte, length int) ([]byte, error) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/hbh_srtp.rs#L124-L129
	kdfKey := aesICMKey30(masterKey, masterSalt)
	var iv [16]byte
	iv[7] = label
	return aesICMCrypt(kdfKey[:], iv[:], make([]byte, length))
}

// ExpandLibsrtpSessionKeys runs the libsrtp session-key expansion (labels 0x00 enc,
// 0x01 auth, 0x02 salt).
func ExpandLibsrtpSessionKeys(keying *SrtpKeyingMaterial, log ...qplog.Logger) (LibsrtpSessionKeys, error) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/hbh_srtp.rs#L132-L157
	lg := pickLog(log)
	lg.DebugE().Msg("expanding libsrtp session keys")
	var out LibsrtpSessionKeys
	sessionKey, err := deriveSessionBytes(keying.MasterKey[:], keying.MasterSalt[:], labelRTPEncryption, 16)
	if err != nil {
		return LibsrtpSessionKeys{}, err
	}
	copy(out.SessionKey[:], sessionKey)
	sessionSalt, err := deriveSessionBytes(keying.MasterKey[:], keying.MasterSalt[:], labelRTPSalt, 14)
	if err != nil {
		return LibsrtpSessionKeys{}, err
	}
	copy(out.SessionSalt[:], sessionSalt)
	authKey, err := deriveSessionBytes(keying.MasterKey[:], keying.MasterSalt[:], labelRTPAuth, 20)
	if err != nil {
		return LibsrtpSessionKeys{}, err
	}
	copy(out.AuthKey[:], authKey)
	return out, nil
}

// BuildRtpICMNonce builds the RTP AES-ICM nonce: zero, SSRC at bytes 4-7 (BE),
// (packetIndex << 16) at bytes 8-15 (BE).
func BuildRtpICMNonce(ssrc uint32, packetIndex uint64) [16]byte {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/hbh_srtp.rs#L160-L165
	var iv [16]byte
	binary.BigEndian.PutUint32(iv[4:8], ssrc)
	binary.BigEndian.PutUint64(iv[8:16], packetIndex<<16)
	return iv
}

// CryptRtpPayload encrypts/decrypts an RTP payload with the expanded session key
// (symmetric).
func CryptRtpPayload(session *LibsrtpSessionKeys, ssrc uint32, packetIndex uint64, payload []byte, log ...qplog.Logger) ([]byte, error) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/hbh_srtp.rs#L168-L177
	lg := pickLog(log)
	icmKey := aesICMKey30(session.SessionKey[:], session.SessionSalt[:])
	nonce := BuildRtpICMNonce(ssrc, packetIndex)
	out, err := aesICMCrypt(icmKey[:], nonce[:], payload)
	if err != nil {
		lg.DebugE().Err(err).Uint32("ssrc", ssrc).Uint64("packet_index", packetIndex).Msg("hbh rtp crypt failed")
		return nil, err
	}
	lg.TraceE().Uint32("ssrc", ssrc).Uint64("packet_index", packetIndex).Int("payload_bytes", len(payload)).Msg("hbh crypt rtp payload")
	return out, nil
}
