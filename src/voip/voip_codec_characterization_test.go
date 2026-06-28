package voip

import (
	"crypto/sha256"
	"encoding/hex"
	"math"
	"testing"
)

// Characterization tests for the VoIP SIP bridge codec (G.711 μ-law, L16,
// sample-rate conversion, RTP framing). This is the transcoding chain between
// the WhatsApp call audio contract (16 kHz float32) and SIP endpoints
// (G.711 μ-law / RTP). These tests freeze the current behaviour so a refactor
// of the audio bridge cannot silently corrupt call audio (PLAN P3.1).

const codecFrameSamples = 960 // 60 ms @ 16 kHz

func sine16k(freqHz, amp float64, n int) []float32 {
	const sampleRate = 16000.0
	pcm := make([]float32, n)
	for i := range pcm {
		pcm[i] = float32(amp * math.Sin(2*math.Pi*freqHz*float64(i)/sampleRate))
	}
	return pcm
}

func energy(x []float32) float64 {
	var e float64
	for _, s := range x {
		e += float64(s) * float64(s)
	}
	return e
}

// TestUlawEncodeDecodeFrameSizes locks the rate-conversion contract: μ-law
// encode halves the sample count (16→8 kHz, 1 byte/sample); decode doubles it
// back (8→16 kHz). A 960-sample frame must survive as 480 bytes then 960 floats.
func TestUlawEncodeDecodeFrameSizes(t *testing.T) {
	frame := sine16k(440, 0.5, codecFrameSamples)

	enc := UlawEncode(frame, nil)
	if len(enc) != codecFrameSamples/2 {
		t.Fatalf("ulaw encoded length: got %d, want %d", len(enc), codecFrameSamples/2)
	}

	dec := UlawDecode(enc, nil)
	if len(dec) != codecFrameSamples {
		t.Fatalf("ulaw decoded length: got %d, want %d", len(dec), codecFrameSamples)
	}
}

// TestUlawSilenceIsConstant verifies that a silent frame companders to a single
// repeated byte (the μ-law representation of zero). A non-constant result would
// mean silence leaks noise onto the wire.
func TestUlawSilenceIsConstant(t *testing.T) {
	enc := UlawEncode(make([]float32, codecFrameSamples), nil)
	if len(enc) == 0 {
		t.Fatal("empty μ-law output for silence")
	}
	for i, b := range enc {
		if b != enc[0] {
			t.Fatalf("silence produced non-constant μ-law byte at %d: %#x vs %#x", i, b, enc[0])
		}
	}
}

// TestUlawRoundTripKnownLevelBug DOCUMENTS A KNOWN BUG, it does not assert
// correctness. The hand-written G.711 μ-law companding in voip_codec.go is
// non-standard: muLawEncodeSample scans the exponent in the wrong direction and
// muLawDecodeSample subtracts (bias<<exp) instead of bias. The net effect is an
// inverted level curve — loud samples round-trip to LESS energy than quiet ones,
// collapsing normal speech levels toward silence on the PCMU wire.
//
// This test freezes that inversion as a witness: when μ-law is corrected to
// ITU-T G.711 (ideally against official reference vectors), this assertion will
// FAIL — at which point flip it into a real fidelity check. See PLAN P3.1.
func TestUlawRoundTripKnownLevelBug(t *testing.T) {
	quiet := UlawDecode(UlawEncode(sine16k(440, 0.05, codecFrameSamples), nil), nil)
	loud := UlawDecode(UlawEncode(sine16k(440, 0.90, codecFrameSamples), nil), nil)

	// Correct G.711 would make `loud` carry far more energy than `quiet`.
	// Current (buggy) behaviour: the opposite. Asserting the bug keeps it visible.
	if energy(loud) >= energy(quiet) {
		t.Fatalf("μ-law level curve now monotonic (loud=%.4f >= quiet=%.4f): the "+
			"G.711 bug appears FIXED — convert this into a fidelity test",
			energy(loud), energy(quiet))
	}
}

// TestL16RoundTripNearLossless confirms L16 (full-rate 16-bit PCM) is
// effectively lossless within int16 quantization. This path preserves fidelity
// end-to-end and lets the SIP server transcode.
func TestL16RoundTripNearLossless(t *testing.T) {
	in := sine16k(440, 0.8, codecFrameSamples)

	enc := L16Encode(in, nil)
	if len(enc) != codecFrameSamples*2 {
		t.Fatalf("L16 encoded length: got %d, want %d", len(enc), codecFrameSamples*2)
	}

	out := L16Decode(enc, nil)
	if len(out) != codecFrameSamples {
		t.Fatalf("L16 decoded length: got %d, want %d", len(out), codecFrameSamples)
	}

	var maxErr float64
	for i := range in {
		if d := math.Abs(float64(in[i] - out[i])); d > maxErr {
			maxErr = d
		}
	}
	if maxErr > 1e-3 {
		t.Fatalf("L16 round-trip max error too high: %g", maxErr)
	}
}

// TestResampleRoundTripFrameSize locks the standalone resampler shape:
// 16→8→16 kHz restores the original sample count.
func TestResampleRoundTripFrameSize(t *testing.T) {
	in := sine16k(300, 0.5, codecFrameSamples)
	out := Interpolate8to16(Decimate16to8(in))
	if len(out) != codecFrameSamples {
		t.Fatalf("resample round-trip length: got %d, want %d", len(out), codecFrameSamples)
	}
}

// TestRTPBuildParseRoundTrip verifies the RTP framing contract: a built packet
// parses back to the same header fields and payload, and the sequence/timestamp
// counters advance correctly across packets (μ-law: 1 byte = 1 sample).
func TestRTPBuildParseRoundTrip(t *testing.T) {
	const ssrc = 0xDEADBEEF
	const initialSeq = 1000
	builder := NewVoipRTPBuilder(ssrc, initialSeq)

	payload1 := UlawEncode(sine16k(440, 0.5, codecFrameSamples), nil)

	// Build returns a buffer reused on the next call — copy before building again.
	pkt1 := append([]byte(nil), builder.Build(payload1, true)...)

	info, err := ParseRTP(pkt1)
	if err != nil {
		t.Fatalf("ParseRTP failed: %v", err)
	}
	if info.SSRC != ssrc {
		t.Fatalf("SSRC: got %#x, want %#x", info.SSRC, uint32(ssrc))
	}
	if info.Seq != initialSeq {
		t.Fatalf("Seq: got %d, want %d", info.Seq, initialSeq)
	}
	if !info.Marker {
		t.Fatal("Marker bit lost")
	}
	if info.PayloadType != RTPPayloadType {
		t.Fatalf("PayloadType: got %d, want %d", info.PayloadType, RTPPayloadType)
	}
	if info.Timestamp != 0 {
		t.Fatalf("first packet timestamp: got %d, want 0", info.Timestamp)
	}
	if len(info.Payload) != len(payload1) {
		t.Fatalf("payload length: got %d, want %d", len(info.Payload), len(payload1))
	}
	for i := range payload1 {
		if info.Payload[i] != payload1[i] {
			t.Fatalf("payload byte %d differs", i)
		}
	}

	// Second packet: seq advances by 1, timestamp advances by sample count.
	pkt2 := builder.Build(payload1, false)
	info2, err := ParseRTP(pkt2)
	if err != nil {
		t.Fatalf("ParseRTP second failed: %v", err)
	}
	if info2.Seq != initialSeq+1 {
		t.Fatalf("Seq did not advance: got %d, want %d", info2.Seq, initialSeq+1)
	}
	if info2.Timestamp != uint32(len(payload1)) {
		t.Fatalf("Timestamp advance: got %d, want %d", info2.Timestamp, len(payload1))
	}
}

// TestParseRTPRejectsInvalid locks the parser guards: short datagrams and
// non-v2 packets must error rather than read out of bounds.
func TestParseRTPRejectsInvalid(t *testing.T) {
	if _, err := ParseRTP(make([]byte, RTPHeaderLen-1)); err != ErrShortRTP {
		t.Fatalf("short packet: got %v, want ErrShortRTP", err)
	}

	bad := make([]byte, RTPHeaderLen)
	bad[0] = 0x40 // version 1
	if _, err := ParseRTP(bad); err != ErrBadRTPVersion {
		t.Fatalf("bad version: got %v, want ErrBadRTPVersion", err)
	}
}

// TestUlawGoldenHash freezes the exact μ-law bytes for a known tone. A change
// here means the companding/decimation output moved — confirm it was intended,
// then update the constants.
func TestUlawGoldenHash(t *testing.T) {
	const goldenHash = "35671d139404b30787df75c022e0d36e21f4374a9e645937848f67773987447b"

	out := UlawEncode(sine16k(440, 0.3, codecFrameSamples), nil)
	sum := sha256.Sum256(out)
	got := hex.EncodeToString(sum[:])
	if goldenHash == "REPLACE_ME" {
		t.Fatalf("CAPTURE GOLDEN: len=%d sha256=%s", len(out), got)
	}
	if got != goldenHash {
		t.Fatalf("μ-law bitstream changed:\n  got  %s\n  want %s", got, goldenHash)
	}
}
