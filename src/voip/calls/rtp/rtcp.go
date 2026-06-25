package rtp

import (
	"encoding/binary"

	qplog "github.com/nocodeleaks/quepasa/qplog"
)

// RTCP: WhatsApp compact reports (PT 208/209) and a Sender Report (PT 200). The
// SR's NTP timestamp is taken as a nowMs argument so this stays pure/no-clock.

const (
	RtcpPtSr         uint8 = 200
	RtcpPtWaCompact  uint8 = 208
	RtcpPtWaCompact2 uint8 = 209
	RtcpHeaderLen    int   = 8
	SrtcpTrailerLen  int   = 14

	ntpUnixOffsetSecs uint64 = 2208988800
)

// IsRtcpPacket reports whether data is an RTCP packet (vs a WhatsApp RTP packet).
func IsRtcpPacket(data []byte) bool {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/rtcp.rs#L16-L28
	if len(data) < RtcpHeaderLen+SrtcpTrailerLen {
		return false
	}
	if (data[0]>>6)&0x03 != 2 {
		return false
	}
	// WhatsApp RTP uses X=1 (byte0 0x90) and a 7-bit PT in byte1; RTCP uses the full byte1 as PT.
	if data[0]&0x10 != 0 && data[1]&0x7f == RtpPayloadTypeOpus {
		return false
	}
	return data[1] >= 64
}

// RtcpPayloadType returns the RTCP payload type; ok=false if not an RTCP packet.
func RtcpPayloadType(data []byte) (uint8, bool) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/rtcp.rs#L30-L32
	if !IsRtcpPacket(data) {
		return 0, false
	}
	return data[1], true
}

// ParseRtcpSenderSsrc returns the sender SSRC (bytes 4-7); ok=false if malformed.
func ParseRtcpSenderSsrc(data []byte, log ...qplog.Logger) (uint32, bool) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/rtcp.rs#L34-L39
	lg := pickLog(log)
	if len(data) < 8 || (data[0]>>6)&0x03 != 2 {
		lg.TraceE().Int("packet_bytes", len(data)).Msg("parse rtcp sender ssrc: malformed packet")
		return 0, false
	}
	ssrc := binary.BigEndian.Uint32(data[4:8])
	lg.TraceE().Uint32("ssrc", ssrc).Msg("parsed rtcp sender ssrc")
	return ssrc, true
}

// RtcpSenderStats are the Sender Report counters.
type RtcpSenderStats struct {
	PacketsSent  uint32
	OctetsSent   uint32
	RtpTimestamp uint32
}

// BuildCompactRtcp208 builds the 12-byte compact RTCP (PT 208, RC=1).
func BuildCompactRtcp208(localSsrc, remoteSsrc uint32, log ...qplog.Logger) [12]byte {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/rtcp.rs#L49-L58
	var buf [12]byte
	buf[0] = 0x81 // V=2, P=0, RC=1
	buf[1] = RtcpPtWaCompact
	buf[3] = 2 // (2+1)*4 = 12 bytes
	binary.BigEndian.PutUint32(buf[4:8], localSsrc)
	binary.BigEndian.PutUint32(buf[8:12], remoteSsrc)
	lg := pickLog(log)
	lg.TraceE().Uint32("local_ssrc", localSsrc).Uint32("remote_ssrc", remoteSsrc).Uint8("payload_type", RtcpPtWaCompact).Msg("built compact rtcp 208")
	return buf
}

// BuildCompactRtcp209 builds the 8-byte compact RTCP (PT 209, RC=1).
func BuildCompactRtcp209(localSsrc uint32, log ...qplog.Logger) [8]byte {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/rtcp.rs#L61-L69
	var buf [8]byte
	buf[0] = 0x81
	buf[1] = RtcpPtWaCompact2
	buf[3] = 1 // (1+1)*4 = 8 bytes
	binary.BigEndian.PutUint32(buf[4:8], localSsrc)
	lg := pickLog(log)
	lg.TraceE().Uint32("local_ssrc", localSsrc).Uint8("payload_type", RtcpPtWaCompact2).Msg("built compact rtcp 209")
	return buf
}

// BuildSenderReport builds the 28-byte Sender Report (PT 200, RC=0); nowMs is wall-clock ms.
func BuildSenderReport(localSsrc uint32, stats *RtcpSenderStats, nowMs uint64, log ...qplog.Logger) [28]byte {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/rtcp.rs#L72-L89
	var buf [28]byte
	buf[0] = 0x80 // V=2, RC=0
	buf[1] = RtcpPtSr
	buf[3] = 6 // (6+1)*4 = 28 bytes
	binary.BigEndian.PutUint32(buf[4:8], localSsrc)
	// NTP timestamp: seconds (upper 32) since 1900, fraction (lower 32). Both truncate
	// to u32 (wrapping), matching the reference encoder.
	ntpSec := uint32((nowMs / 1000) + ntpUnixOffsetSecs)
	ntpFrac := uint32(float64(nowMs%1000) / 1000.0 * 4294967296.0)
	binary.BigEndian.PutUint32(buf[8:12], ntpSec)
	binary.BigEndian.PutUint32(buf[12:16], ntpFrac)
	binary.BigEndian.PutUint32(buf[16:20], stats.RtpTimestamp)
	binary.BigEndian.PutUint32(buf[20:24], stats.PacketsSent)
	binary.BigEndian.PutUint32(buf[24:28], stats.OctetsSent)
	lg := pickLog(log)
	lg.TraceE().Uint32("local_ssrc", localSsrc).Uint32("rtp_timestamp", stats.RtpTimestamp).Uint32("packets_sent", stats.PacketsSent).Uint32("octets_sent", stats.OctetsSent).Msg("built rtcp sender report")
	return buf
}
