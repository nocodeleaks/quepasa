# ISSUE: G.711 μ-law companding is inverted (SIP PCMU audio collapses)

Status: Open — found 2026-06-28 while building P3.1 codec regression tests
Severity: High (affects the primary SIP audio path)
Area: `src/voip/voip_codec.go`

## Symptom

The hand-written G.711 μ-law encode/decode in `voip_codec.go` has an **inverted
level curve**. A round-trip (`UlawEncode` → `UlawDecode`) preserves only very
quiet samples; normal speech levels are attenuated by 10×–250× and some levels
collapse to silence.

Measured round-trip of a constant input frame (before any fix):

| input amplitude | decoded amplitude | expected |
|-----------------|-------------------|----------|
| 0.05            | 0.0508            | ~0.05 ✓  |
| 0.10            | 0.0098            | ~0.10 ✗  |
| 0.30            | 0.0059            | ~0.30 ✗  |
| 0.50            | 0.0020            | ~0.50 ✗  |
| 0.90            | 0.0017            | ~0.90 ✗  |

Louder input → quieter output. For SIP endpoints negotiating **G.711 μ-law
(PCMU, RTP PT 0)** — the codec most SIP providers default to — call audio is
severely degraded or near-silent.

## Root cause (two independent defects)

1. **Encoder exponent scan is backwards** (`muLawEncodeSample`):
   ```go
   exp := byte(0)
   for mask := 0x4000; sample < int(mask) && exp < 7; exp++ { mask >>= 1 }
   ```
   This maps large magnitudes to `exp 0` and small magnitudes to `exp 7` — the
   inverse of G.711 segments.

2. **Decoder reconstruction subtracts the wrong bias** (`muLawDecodeSample`):
   ```go
   sample := int(((int(mantissa) << 3) + ulawBias) << int(exp))
   sample -= ulawBias << int(exp)   // ITU-T subtracts BIAS (0x84), not BIAS<<exp
   ```

The implementation also companders on a 15-bit magnitude (`ulawClip = 32635`)
whereas reference G.711 works on 14-bit (`pcm_val >> 2`), so encode/decode and
scaling are mutually inconsistent.

## Why it was not caught

`voip/voip_codec.go` had **no tests** (the module was the least-covered in the
repo). The L16 path (`L16Encode`/`L16Decode`, RTP PT 118) is near-lossless and
likely masked the issue when that codec was negotiated instead of PCMU.

## Guardrails added (this session)

`src/voip/voip_codec_characterization_test.go` now covers the SIP bridge:
frame-size contracts, μ-law silence-is-constant, **L16 near-lossless round-trip**,
resampler shape, **RTP build/parse round-trip + parser guards**, and a μ-law
golden hash. `TestUlawRoundTripKnownLevelBug` deliberately asserts the *inverted*
curve so the bug stays visible; it will FAIL the moment μ-law is corrected,
prompting conversion into a real fidelity test.

## Recommended fix (needs decision)

Do **not** hand-tweak the exponent loop in isolation — a partial change made the
output differently wrong (amp 0.5 → pure silence). Options:

1. **Replace with canonical ITU-T G.711** μ-law (and A-law, since SIP providers
   also use it) `linear2ulaw`/`ulaw2linear`, validated against official G.711
   reference vectors. Lowest risk of recurrence.
2. **Use a vetted Go G.711 library** for PCMU/PCMA and keep `voip_codec.go` only
   for resampling/RTP.

Either way: add A-law (`PCMA`) support — the user reports most SIP providers use
μ-law **and** A-law, and A-law is currently absent.

## Acceptance

- Round-trip of speech-level tones preserves energy within a lossy-but-monotonic
  bound (loud > quiet).
- Output matches official G.711 reference vectors for both μ-law and A-law.
- `TestUlawRoundTripKnownLevelBug` flipped into a fidelity assertion.
