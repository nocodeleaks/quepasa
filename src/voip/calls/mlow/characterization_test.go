package mlow

import (
	"crypto/sha256"
	"encoding/hex"
	"math"
	"testing"
)

// These are characterization (golden) tests for the hand-written MLow codec.
//
// The codec is the highest-complexity, highest-value DSP in the repo (it carries
// WhatsApp call audio). It has no upstream reference suite here, so these tests
// freeze the CURRENT behaviour: any change that alters the encoded bitstream,
// the frame contract, or the round-trip shape will fail loudly. They are a
// regression net for refactors (PLAN P3.1), not a correctness proof of the DSP.

const frameSamples = 960 // 60 ms @ 16 kHz, must match opusFrameSamps

// sineFrame builds a deterministic single 60 ms frame: a pure tone at the given
// frequency and amplitude, sampled at 16 kHz. Deterministic input is required so
// the golden hash below stays stable.
func sineFrame(freqHz, amp float64) []float32 {
	const sampleRate = 16000.0
	pcm := make([]float32, frameSamples)
	for i := range pcm {
		pcm[i] = float32(amp * math.Sin(2*math.Pi*freqHz*float64(i)/sampleRate))
	}
	return pcm
}

// TestMlowEncodeIsDeterministic guarantees that two fresh encoders fed identical
// PCM produce byte-identical output. Non-determinism here would make every other
// golden assertion meaningless and break wire reproducibility.
func TestMlowEncodeIsDeterministic(t *testing.T) {
	pcm := sineFrame(440, 0.3)

	a, err := NewMlowEncoder().Encode(pcm)
	if err != nil {
		t.Fatalf("first encode failed: %v", err)
	}
	b, err := NewMlowEncoder().Encode(pcm)
	if err != nil {
		t.Fatalf("second encode failed: %v", err)
	}

	if len(a) == 0 {
		t.Fatal("encoder produced empty output for a non-silent frame")
	}
	if hex.EncodeToString(a) != hex.EncodeToString(b) {
		t.Fatalf("encode not deterministic:\n  a=%x\n  b=%x", a, b)
	}
}

// TestMlowEncodeRejectsWrongFrameSize locks the frame contract: exactly 960
// samples. Callers (the call engine) rely on this hard boundary.
func TestMlowEncodeRejectsWrongFrameSize(t *testing.T) {
	for _, n := range []int{0, 480, 959, 961, 1920} {
		if _, err := NewMlowEncoder().Encode(make([]float32, n)); err == nil {
			t.Fatalf("expected error for frame of %d samples, got nil", n)
		}
	}
}

// TestMlowEncodeGoldenHash freezes the exact bitstream for a known tone. If this
// fails after a code change, the change altered the encoded output — confirm it
// was intentional, then update goldenSineHash and goldenSineLen.
func TestMlowEncodeGoldenHash(t *testing.T) {
	const goldenSineHash = "ef4d5def10087f0a99f4b5af46d1fc1ab16023cee971be55860d887cc780d38e"
	const goldenSineLen = 69

	out, err := NewMlowEncoder().Encode(sineFrame(440, 0.3))
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	sum := sha256.Sum256(out)
	got := hex.EncodeToString(sum[:])
	if goldenSineHash == "REPLACE_ME" {
		t.Fatalf("CAPTURE GOLDEN: len=%d sha256=%s", len(out), got)
	}
	if len(out) != goldenSineLen {
		t.Fatalf("encoded length changed: got %d, want %d", len(out), goldenSineLen)
	}
	if got != goldenSineHash {
		t.Fatalf("encoded bitstream changed:\n  got  %s\n  want %s", got, goldenSineHash)
	}
}

// TestMlowRoundTripShape verifies the encode→decode contract: a decoded frame is
// always exactly 960 finite samples, a non-silent input yields non-trivial
// energy out, and pure silence stays quiet. Thresholds are deliberately loose —
// this guards the pipeline shape, not codec fidelity.
func TestMlowRoundTripShape(t *testing.T) {
	enc := NewMlowEncoder()
	dec := NewMlowDecoder()

	payload, err := enc.Encode(sineFrame(440, 0.5))
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	out := dec.Decode(payload)
	if len(out) != frameSamples {
		t.Fatalf("decoded frame size: got %d, want %d", len(out), frameSamples)
	}

	var energy float64
	for _, s := range out {
		if math.IsNaN(float64(s)) || math.IsInf(float64(s), 0) {
			t.Fatalf("decoded sample is not finite: %v", s)
		}
		energy += float64(s) * float64(s)
	}
	if energy == 0 {
		t.Fatal("decoded a non-silent frame to pure zeros")
	}

	// Silence in must not blow up into loud noise out.
	silence := dec.Decode(nil)
	if len(silence) != frameSamples {
		t.Fatalf("decoded silence size: got %d, want %d", len(silence), frameSamples)
	}
	var silenceEnergy float64
	for _, s := range silence {
		silenceEnergy += float64(s) * float64(s)
	}
	if silenceEnergy > float64(frameSamples) {
		t.Fatalf("empty payload decoded to excessive energy: %v", silenceEnergy)
	}
}
