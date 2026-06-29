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

// TestG711RoundTripPreservesSignal asserts that the corrected ITU-T G.711 codecs
// (μ-law and A-law) carry speech-level signal faithfully: energy is preserved
// within companding loss, the level curve is monotonic (loud > quiet), and the
// decoded waveform correlates strongly with the input. This replaces the former
// known-bug witness once the inverted companding was fixed.
func TestG711RoundTripPreservesSignal(t *testing.T) {
	codecs := map[string]struct {
		enc func([]float32, []byte) []byte
		dec func([]byte, []float32) []float32
	}{
		"ulaw": {UlawEncode, UlawDecode},
		"alaw": {AlawEncode, AlawDecode},
	}

	for name, c := range codecs {
		t.Run(name, func(t *testing.T) {
			in := sine16k(440, 0.5, codecFrameSamples)
			out := c.dec(c.enc(in, nil), nil)

			eIn, eOut := energy(in), energy(out)
			if ratio := eOut / eIn; ratio < 0.7 || ratio > 1.3 {
				t.Fatalf("%s energy ratio out of range: %.3f", name, ratio)
			}

			var dot float64
			for i := range in {
				dot += float64(in[i]) * float64(out[i])
			}
			if corr := dot / math.Sqrt(eIn*eOut); corr < 0.95 {
				t.Fatalf("%s correlation too low: %.3f", name, corr)
			}

			// Monotonic level curve: louder in → louder out.
			quiet := c.dec(c.enc(sine16k(440, 0.05, codecFrameSamples), nil), nil)
			loud := c.dec(c.enc(sine16k(440, 0.90, codecFrameSamples), nil), nil)
			if energy(loud) <= energy(quiet) {
				t.Fatalf("%s level curve not monotonic: loud=%.4f <= quiet=%.4f",
					name, energy(loud), energy(quiet))
			}
		})
	}
}

// TestG711SilenceReferenceBytes locks the canonical encodings of digital silence:
// μ-law 0 → 0xFF, A-law 0 → 0xD5 (ITU-T G.711). These are fixed wire constants.
func TestG711SilenceReferenceBytes(t *testing.T) {
	if got := UlawEncode(make([]float32, 2), nil)[0]; got != 0xFF {
		t.Fatalf("μ-law silence byte: got %#x, want 0xFF", got)
	}
	if got := AlawEncode(make([]float32, 2), nil)[0]; got != 0xD5 {
		t.Fatalf("A-law silence byte: got %#x, want 0xD5", got)
	}
}

// TestAlawFrameSizes locks the A-law rate-conversion contract (16→8 kHz on
// encode, 8→16 kHz on decode), mirroring μ-law.
func TestAlawFrameSizes(t *testing.T) {
	enc := AlawEncode(sine16k(440, 0.5, codecFrameSamples), nil)
	if len(enc) != codecFrameSamples/2 {
		t.Fatalf("A-law encoded length: got %d, want %d", len(enc), codecFrameSamples/2)
	}
	if dec := AlawDecode(enc, nil); len(dec) != codecFrameSamples {
		t.Fatalf("A-law decoded length: got %d, want %d", len(dec), codecFrameSamples)
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
	const goldenHash = "d035180ea8d49367bcf7535850180ccb47de1f9500ae2f14d63c3e8067b53e43"

	out := UlawEncode(sine16k(440, 0.3, codecFrameSamples), nil)
	sum := sha256.Sum256(out)
	if got := hex.EncodeToString(sum[:]); got != goldenHash {
		t.Fatalf("μ-law bitstream changed:\n  got  %s\n  want %s", got, goldenHash)
	}
}

// TestAlawGoldenHash freezes the exact A-law bytes for a known tone.
func TestAlawGoldenHash(t *testing.T) {
	const goldenHash = "64bed2a8c97fa4e1b51db80c5f1eb97a290b7dba65eed269e529904f47c633e7"

	out := AlawEncode(sine16k(440, 0.3, codecFrameSamples), nil)
	sum := sha256.Sum256(out)
	if got := hex.EncodeToString(sum[:]); got != goldenHash {
		t.Fatalf("A-law bitstream changed:\n  got  %s\n  want %s", got, goldenHash)
	}
}
