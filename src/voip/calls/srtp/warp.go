package srtp

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/binary"

	qplog "github.com/nocodeleaks/quepasa/qplog"
)

// WARP RTP extension constants and the WARP MESSAGE-INTEGRITY tag (HMAC-SHA1
// appended to protected packets).

const WarpExtProfile uint16 = 0xdebe

const (
	WarpMITagLen             = 4
	WarpPiggybackStartPacket = 2
)

// WarpAudioPiggybackExt is the audio piggyback extension word (big-endian bytes).
var WarpAudioPiggybackExt = [4]byte{0x30, 0x01, 0x00, 0x00}

// AudioPiggybackExtensionFor returns the audio piggyback extension word for
// packetIndex, or nil for the first packets / when disabled.
func AudioPiggybackExtensionFor(packetIndex int, enabled bool, startPacket int, log ...qplog.Logger) *uint32 {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/warp.rs#L15-L24
	lg := pickLog(log)
	if !enabled || packetIndex < startPacket {
		lg.TraceE().Int("packet_index", packetIndex).Bool("enabled", enabled).Int("start_packet", startPacket).Msg("warp audio piggyback skipped")
		return nil
	}
	w := binary.BigEndian.Uint32(WarpAudioPiggybackExt[:])
	lg.TraceE().Int("packet_index", packetIndex).Msg("warp audio piggyback extension attached")
	return &w
}

// ComputeWarpMITag is the WARP MI tag: the first tagLen bytes of
// HMAC-SHA1(authKey, packetWithoutTag || roc_be32).
func ComputeWarpMITag(authKey, packetWithoutTag []byte, roc uint32, tagLen int, log ...qplog.Logger) []byte {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/warp.rs#L27-L38
	lg := pickLog(log)
	mac := hmac.New(sha1.New, authKey)
	mac.Write(packetWithoutTag)
	var rocBE [4]byte
	binary.BigEndian.PutUint32(rocBE[:], roc)
	mac.Write(rocBE[:])
	lg.TraceE().Int("packet_bytes", len(packetWithoutTag)).Uint32("roc", roc).Int("tag_len", tagLen).Msg("computed warp mi tag")
	return mac.Sum(nil)[:tagLen]
}

// AppendWarpMITag appends the WARP MI tag to a protected packet.
func AppendWarpMITag(authKey, packetWithoutTag []byte, roc uint32, tagLen int, log ...qplog.Logger) []byte {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/warp.rs#L41-L52
	lg := pickLog(log)
	tag := ComputeWarpMITag(authKey, packetWithoutTag, roc, tagLen, lg)
	out := make([]byte, 0, len(packetWithoutTag)+len(tag))
	out = append(out, packetWithoutTag...)
	out = append(out, tag...)
	lg.TraceE().Int("packet_bytes", len(packetWithoutTag)).Uint32("roc", roc).Int("tag_len", tagLen).Int("total_bytes", len(out)).Msg("appended warp mi tag")
	return out
}
