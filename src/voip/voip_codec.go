package voip

// VoIP Bridge Codec: G.711 μ-law encoder/decoder, sample-rate conversion, and
// RTP packet builder/parser.
//
// The calls module audio contract is 16 kHz mono float32 PCM in 960-sample
// (60 ms) frames. SIP Phase 1 uses G.711 μ-law at 8 kHz, 8 bits/sample, carried
// in RTP. This file provides:
//
//   - UlawEncode / UlawDecode — pure-Go μ-law conversion (float32 ↔ uint8)
//   - Decimate16to8 / Interpolate8to16 — linear sample-rate conversion
//   - VoipRTPBuilder — incrementing RTP header + payload packetizer
//   - ParseRTP — RTP header validation + payload extraction
//
// All code is allocation-conscious: hot paths reuse pre-allocated buffers.

import (
	"encoding/binary"
	"errors"
	"math"
)

// ---------------------------------------------------------------------------
// G.711 μ-law encode / decode
// ---------------------------------------------------------------------------

// ulawMax is the amplitude clip level for μ-law encoding (clamped to 32767 s15).
const ulawMax = 32635.0

// ulawBias is the decoder bias removed after expansion (15-bit range).
const ulawBias = 0x84

// ulawClip is the maximum 15-bit magnitude used by the encoder.
const ulawClip = 32635

// UlawEncode converts one 16 kHz float32 PCM frame into:
//  1. an 8 kHz decimated int16 frame (half the samples), and
//  2. the μ-law bytes of that decimated frame.
//
// The returned []byte length is len(frame)/2. The caller should reuse the
// scratch buffer on each call to avoid allocations.
func UlawEncode(frame []float32, scratch []byte) []byte {
	half := len(frame) / 2
	if cap(scratch) < half {
		scratch = make([]byte, half)
	}
	out := scratch[:half]

	for i := 0; i < half; i++ {
		// Decimate 16 kHz → 8 kHz with a 3-tap triangular low-pass
		// (0.25, 0.5, 0.25) centered on the kept sample. This attenuates
		// content above 4 kHz before downsampling, removing the aliasing that
		// plain "take every other sample" produced (harsh/garbled audio).
		c := i * 2
		s := 0.5 * frame[c]
		if c > 0 {
			s += 0.25 * frame[c-1]
		} else {
			s += 0.25 * frame[c]
		}
		if c+1 < len(frame) {
			s += 0.25 * frame[c+1]
		} else {
			s += 0.25 * frame[c]
		}

		// Scale float32 [-1,1] to int15 magnitude.
		v := int(math.Round(float64(s) * ulawMax))
		if v > ulawClip {
			v = ulawClip
		} else if v < -ulawClip {
			v = -ulawClip
		}
		out[i] = muLawEncodeSample(v)
	}
	return out
}

// muLawEncodeSample encodes a signed 16-bit sample (actually 15-bit range) to
// a single μ-law byte.  Algorithm per ITU-T G.711 Table 11a.
func muLawEncodeSample(sample int) byte {
	sign := byte(0)
	if sample < 0 {
		sample = -sample
		sign = 0x80
	}
	if sample > ulawClip {
		sample = ulawClip
	}
	sample += ulawBias

	// Exponent: position of the highest set bit above bit 7.
	exp := byte(0)
	for mask := 0x4000; sample < int(mask) && exp < 7; exp++ {
		mask >>= 1
	}

	// Mantissa: next 3 significant bits after the exponent.
	mantissa := byte((sample >> (int(exp) + 3)) & 0x0F)

	// Combine: ~sign | (exp << 4) | mantissa
	return ^(sign | (exp << 4) | mantissa)
}

// UlawDecode converts μ-law bytes to a 16 kHz float32 PCM frame.  Each input
// byte produces two float32 samples (interpolated by linear repetition) so the
// output length is 2 × len(data).
//
// The caller should pre-allocate out with capacity 2*len(data).
func UlawDecode(data []byte, out []float32) []float32 {
	n := len(data) * 2
	if cap(out) < n {
		out = make([]float32, n)
	}
	out = out[:n]

	for i, b := range data {
		s := muLawDecodeSample(b)

		// Scale int16 [-32768, 32767] to float32 [-1, 1].
		f := float32(s) / 32768.0

		// Upsample 8 kHz → 16 kHz with linear interpolation instead of a
		// zero-order hold (sample duplication). The kept sample sits on even
		// indices; odd indices are the midpoint to the next sample, which
		// suppresses the imaging the duplication introduced.
		out[i*2] = f
		if i+1 < len(data) {
			next := float32(muLawDecodeSample(data[i+1])) / 32768.0
			out[i*2+1] = 0.5 * (f + next)
		} else {
			out[i*2+1] = f
		}
	}
	return out
}

// muLawDecodeSample decodes a μ-law byte to a signed 16-bit PCM sample.
func muLawDecodeSample(b byte) int16 {
	// Invert all bits (μ-law is transmitted inverted).
	b = ^b

	sign := b & 0x80
	exp := (b >> 4) & 0x07
	mantissa := b & 0x0F

	// Reconstruct the 13-bit magnitude.
	sample := int(((int(mantissa) << 3) + ulawBias) << int(exp))
	sample -= ulawBias << int(exp)

	if sign != 0 {
		sample = -sample
	}
	if sample > 32767 {
		sample = 32767
	} else if sample < -32768 {
		sample = -32768
	}
	return int16(sample)
}

// ---------------------------------------------------------------------------
// Sample-rate conversion helpers (standalone, for testing)
// ---------------------------------------------------------------------------

// Decimate16to8 performs simple 2:1 decimation (take every other sample).
// Input: 16 kHz float32. Output: 8 kHz float32 (half length).
func Decimate16to8(in []float32) []float32 {
	half := len(in) / 2
	out := make([]float32, half)
	for i := 0; i < half; i++ {
		out[i] = in[i*2]
	}
	return out
}

// Interpolate8to16 performs zero-order-hold upsampling (sample duplication).
// Input: 8 kHz float32. Output: 16 kHz float32 (double length).
func Interpolate8to16(in []float32) []float32 {
	n := len(in) * 2
	out := make([]float32, n)
	for i, v := range in {
		out[i*2] = v
		out[i*2+1] = v
	}
	return out
}

// ---------------------------------------------------------------------------
// RTP packet builder
// ---------------------------------------------------------------------------

const (
	RTPVersion     = 2
	RTPHeaderLen   = 12
	RTPPayloadType = 0   // PCMU (G.711 μ-law)
	L16PayloadType = 118 // dynamic PT for L16/16000 (network-order 16-bit PCM)
)

// L16Encode converts a 16 kHz mono float32 frame to network-order (big-endian)
// signed 16-bit PCM — RTP "L16/16000". No decimation or companding: the audio
// stays full-rate 16 kHz and the SIP server (asterisk) does any transcoding to
// the endpoint's codec, preserving fidelity end to end.
func L16Encode(frame []float32, scratch []byte) []byte {
	need := len(frame) * 2
	if cap(scratch) < need {
		scratch = make([]byte, need)
	}
	out := scratch[:need]
	for i, s := range frame {
		v := int32(math.Round(float64(s) * 32767.0))
		if v > 32767 {
			v = 32767
		} else if v < -32768 {
			v = -32768
		}
		binary.BigEndian.PutUint16(out[i*2:], uint16(int16(v)))
	}
	return out
}

// L16Decode converts network-order signed 16-bit PCM (L16/16000) to 16 kHz mono
// float32. No upsampling needed — the data is already 16 kHz.
func L16Decode(data []byte, out []float32) []float32 {
	n := len(data) / 2
	if cap(out) < n {
		out = make([]float32, n)
	}
	out = out[:n]
	for i := 0; i < n; i++ {
		s := int16(binary.BigEndian.Uint16(data[i*2:]))
		out[i] = float32(s) / 32768.0
	}
	return out
}

// VoipRTPBuilder builds sequential RTP packets for one direction of a call.
// It is NOT safe for concurrent use without external synchronization.
type VoipRTPBuilder struct {
	seq            uint16
	ts             uint32
	ssrc           uint32
	payloadType    byte
	bytesPerSample int // RTP timestamp advances by len(payload)/bytesPerSample
	buf            []byte
}

// NewVoipRTPBuilder returns a μ-law (PCMU, PT 0, 8 kHz) RTP builder with the given
// SSRC and initial sequence. One byte = one sample.
func NewVoipRTPBuilder(ssrc uint32, initialSeq uint16) *VoipRTPBuilder {
	return NewVoipRTPBuilderPT(ssrc, initialSeq, RTPPayloadType, 1)
}

// NewVoipRTPBuilderPT returns an RTP builder for an arbitrary payload type, where
// bytesPerSample relates payload bytes to RTP-clock samples (1 for G.711, 2 for
// 16-bit linear L16).
func NewVoipRTPBuilderPT(ssrc uint32, initialSeq uint16, payloadType byte, bytesPerSample int) *VoipRTPBuilder {
	if bytesPerSample < 1 {
		bytesPerSample = 1
	}
	return &VoipRTPBuilder{
		seq:            initialSeq,
		ts:             0,
		ssrc:           ssrc,
		payloadType:    payloadType,
		bytesPerSample: bytesPerSample,
		buf:            make([]byte, 0, 256),
	}
}

// Build appends a fixed 12-byte RTP header to the payload and advances internal
// sequence/timestamp counters. The returned slice is owned by the caller (the
// builder's internal buffer is reused across calls, so copy if needed).
func (b *VoipRTPBuilder) Build(payload []byte, marker bool) []byte {
	totalLen := RTPHeaderLen + len(payload)
	if cap(b.buf) < totalLen {
		b.buf = make([]byte, 0, totalLen+64)
	}
	b.buf = b.buf[:totalLen]

	// V=2, P=0, X=0, CC=0, M=marker, PT=payloadType
	mBit := byte(0)
	if marker {
		mBit = 0x80
	}
	b.buf[0] = 0x80 // V=2, P=0, X=0, CC=0
	b.buf[1] = mBit | b.payloadType

	binary.BigEndian.PutUint16(b.buf[2:4], b.seq)
	binary.BigEndian.PutUint32(b.buf[4:8], b.ts)
	binary.BigEndian.PutUint32(b.buf[8:12], b.ssrc)

	copy(b.buf[12:], payload)

	b.seq++
	b.ts += uint32(len(payload) / b.bytesPerSample) // advance by samples
	return b.buf
}

// ---------------------------------------------------------------------------
// RTP packet parser
// ---------------------------------------------------------------------------

// ErrShortRTP is returned when a datagram is too short to contain a valid RTP header.
var ErrShortRTP = errors.New("voip bridge: RTP packet shorter than header")

// ErrBadRTPVersion is returned when the RTP version field is not 2.
var ErrBadRTPVersion = errors.New("voip bridge: unsupported RTP version")

// VoipRTPInfo carries extracted header fields and the payload offset.
type VoipRTPInfo struct {
	Seq         uint16
	Timestamp   uint32
	SSRC        uint32
	PayloadType byte
	Marker      bool
	Payload     []byte
}

// ParseRTP extracts the RTP header fields and payload from a datagram.
// It does NOT validate the payload type; callers should check PayloadType
// against RTPPayloadType (PCMU) if they care.
func ParseRTP(data []byte) (*VoipRTPInfo, error) {
	if len(data) < RTPHeaderLen {
		return nil, ErrShortRTP
	}
	v := data[0] >> 6
	if v != RTPVersion {
		return nil, ErrBadRTPVersion
	}

	cc := int(data[0] & 0x0F)
	hdrLen := RTPHeaderLen + cc*4 // CSRC extensions
	if len(data) < hdrLen {
		return nil, ErrShortRTP
	}

	info := &VoipRTPInfo{
		Marker:      data[1]&0x80 != 0,
		PayloadType: data[1] & 0x7F,
		Seq:         binary.BigEndian.Uint16(data[2:4]),
		Timestamp:   binary.BigEndian.Uint32(data[4:8]),
		SSRC:        binary.BigEndian.Uint32(data[8:12]),
		Payload:     data[hdrLen:],
	}
	return info, nil
}
