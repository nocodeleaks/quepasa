package util

import (
	"crypto/hkdf"
	"crypto/sha256"

	qplog "github.com/nocodeleaks/quepasa/qplog"
)

// HKDFSHA256 derives length bytes of key material from ikm using HKDF-SHA256
// (RFC 5869): an HMAC-SHA256 extract keyed by salt, then expand with info. Every
// VoIP key schedule (SRTP session keys, SFrame keys, the WARP auth key) reduces
// to this one shape. It errors when length exceeds the HKDF bound (255*32 = 8160
// bytes for SHA-256) rather than aborting the caller.
func HKDFSHA256(salt, ikm, info []byte, length int, log ...qplog.Logger) ([]byte, error) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/mod.rs#L32-L39
	lg := pickLog(log)
	okm, err := hkdf.Key(sha256.New, ikm, salt, string(info), length)
	if err != nil {
		lg.DebugE().Err(err).Int("ikm_bytes", len(ikm)).Int("salt_bytes", len(salt)).Int("info_bytes", len(info)).Int("okm_bytes", length).Msg("hkdf-sha256 derive failed")
		return nil, err
	}
	lg.TraceE().Int("ikm_bytes", len(ikm)).Int("salt_bytes", len(salt)).Int("info_bytes", len(info)).Int("okm_bytes", len(okm)).Msg("derived hkdf-sha256 key material")
	return okm, nil
}
