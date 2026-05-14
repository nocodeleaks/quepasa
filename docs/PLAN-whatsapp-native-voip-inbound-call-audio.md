# Plan: Native WhatsApp VoIP Inbound Call Acceptance and Audio Playback

## Objective

Implement an initial native WhatsApp VoIP path in QuePasa focused on **incoming calls**.

The first practical target is:

1. receive an incoming WhatsApp call
2. accept the call natively
3. keep the call alive long enough for media exchange
4. play a deterministic prerecorded audio payload to the caller

This plan is intentionally focused on **voice directly inside the call**, not fallback text/audio messages in chat.

## Desired Done State

The slice is considered successful when all of the following are true:

- QuePasa receives a real incoming WhatsApp voice call event.
- QuePasa accepts the call without manual action.
- The caller hears a prerecorded test audio clip.
- The call stays connected for more than a minimal handshake window.
- The implementation is repeatable in a local Linux + Go environment without Docker.
- Debug artifacts exist for signaling, TURN, SRTP, RTP, and Opus framing.

## Scope

### In Scope

- inbound WhatsApp call handling
- native call accept path
- TURN relay negotiation
- SRTP key derivation and packet protection
- RTP packetization for voice media
- Opus framing compatible with WhatsApp call expectations
- automatic answer with prerecorded audio
- local observability and reproducible test workflow

### Out of Scope

- SIP/PBX bridging through `src/sipproxy/`
- remote telephony origination/forwarding
- outbound call origination as the first milestone
- video calls
- production-grade multi-call orchestration
- polished public API surface before the media pipeline is proven

## Technical Anchors From External Research

The current implementation strategy should explicitly anchor on the public findings already identified:

1. **Signaling**
   - WhatsApp call setup uses `<call>` IQ stanzas.
   - Inbound `offer` parsing is already partially exposed by `whatsmeow` through call events.
   - Outbound `accept` and related call nodes may require custom protobuf or binary node assembly.

2. **TURN relay**
   - Media relay is done through Meta TURN endpoints such as `relay-*.facebook.com:3478`.
   - The TURN HMAC-SHA1 secret must use the **raw base64 string itself**, not the base64-decoded bytes.
   - This rule must be validated with explicit unit/integration coverage because it is easy to implement incorrectly.

3. **SRTP**
   - Media is protected with SRTP.
   - Key material is derived from signaling data using HKDF.
   - Packet headers, rollover behavior, index handling, and auth tags must conform to RFC 3711 expectations.

4. **Audio framing risk**
   - The biggest known risk is not signaling or crypto.
   - The most likely blocker is Opus/RTP framing compatibility:
     - DTX/silence behavior
     - frame duration
     - application mode
     - channels/sample rate assumptions
     - RTP extension/header expectations

## Current Local Starting Point

The repository already gives a useful base but not the full pipeline:

- `src/whatsmeow/whatsmeow_event_router.go`
  - observes `CallOffer` and `CallOfferNotice`
- `src/whatsmeow/whatsmeow_handlers.go`
  - converts call events into internal message flow
  - rejects calls when call handling is disabled
- `src/whatsmeow/whatsmeow_handlers+call.go`
  - resolves identifiers and LID/phone mapping context for call-related events
- local Linux service installation is already available through:
  - `helpers/install.sh`
  - `helpers/quepasa.service`

This means the plan should extend the existing WhatsApp integration layer rather than route the first milestone through `sipproxy`.

## Phase 0 — Lab Preparation and Evidence Capture

### Goal

Create a controlled local validation loop before changing behavior.

### Checklist

- [ ] Confirm one dedicated WhatsApp account for the QuePasa linked device.
- [ ] Confirm one independent caller account for manual test calls.
- [ ] Enable call handling in local runtime configuration.
- [ ] Create a repeatable local `.env` profile for VoIP experiments.
- [ ] Install local packet capture tooling (`tcpdump`, `tshark`, or Wireshark-compatible capture flow).
- [ ] Prepare a short prerecorded source file for testing.
- [ ] Prepare an Opus-focused media conversion toolchain (`ffmpeg` and validation commands).
- [ ] Capture at least one baseline real-app-to-real-app call for comparison.

### Deliverables

- local runbook for starting capture
- baseline packet captures
- reproducible media test asset set

## Immediate Execution Order

This section converts the plan into the first executable work queue.

### Step 0.1 — Local VoIP lab profile

- [ ] Create a dedicated local `.env` profile for VoIP experiments.
- [ ] Turn `CALLS=true` on explicitly.
- [ ] Keep experimental VoIP flags off by default until signaling instrumentation lands.
- [ ] Define a stable local database path and media fixture directory.

### Step 0.2 — Capture tooling and assets

- [ ] Install `tcpdump`.
- [ ] Install `tshark` or confirm equivalent PCAP workflow.
- [ ] Install `ffmpeg`.
- [ ] Create `extra/voip-fixtures/` or equivalent fixture directory.
- [ ] Prepare one short WAV source and one target Opus experiment source.

### Step 0.3 — Signaling-first implementation slice

- [ ] Add dedicated call-session state structures.
- [ ] Add structured call-event logs by `CallID`.
- [ ] Persist or dump enough offer payload detail for offline diffing.
- [ ] Add experimental feature flags for `observe`, `auto-answer`, and `media-debug`.

### Step 0.4 — First success target

- [ ] Accept the call without media playback first.
- [ ] Verify that the caller sees the answer.
- [ ] Verify that the session reaches relay negotiation.
- [ ] Only then start TURN/SRTP/audio work.

## Phase 1 — Call State and Signaling Observability

### Goal

Stop treating incoming calls as generic call messages only and build a real call-state model.

### Checklist

- [ ] Add dedicated internal call-state structures for active VoIP sessions.
- [ ] Track call lifecycle by `CallID`.
- [ ] Expand inbound handling to explicitly observe:
  - [ ] offer
  - [ ] offer notice
  - [ ] accept
  - [ ] terminate
  - [ ] relay/latency-related events if exposed
- [ ] Persist structured debug logs for signaling payloads.
- [ ] Add safe binary/protobuf dump helpers for call payload inspection.
- [ ] Add feature flags so experimental accept/media behavior can be toggled per environment.

### Suggested File Areas

- `src/whatsmeow/whatsmeow_event_router.go`
- `src/whatsmeow/whatsmeow_handlers.go`
- new `src/whatsmeow/whatsmeow_call_*.go` files
- optional runtime boundary helpers in `src/runtime/`

## Phase 2 — Native Call Accept Path

### Goal

Implement the first true inbound answer path instead of only observing or rejecting calls.

### Checklist

- [ ] Identify the exact `accept` node structure required for successful call pickup.
- [ ] Implement call accept assembly as an isolated helper, not inline inside the generic handler.
- [ ] Wire auto-answer logic behind an explicit feature flag.
- [ ] Preserve a manual mode where calls are only observed.
- [ ] Validate that the remote caller sees the call as answered.
- [ ] Validate that the session transitions into relay/media setup instead of immediate drop.

### Acceptance Criteria

- caller sees the call answered
- QuePasa reaches the media setup stage
- signaling logs show a complete offer → accept transition

## Phase 3 — TURN Credential Handling and Relay Session Setup

### Goal

Establish the relay path correctly against Meta TURN infrastructure.

### Checklist

- [ ] Identify where TURN credentials appear in inbound signaling payloads.
- [ ] Extract relay host, username, password/key, and token-related metadata.
- [ ] Implement TURN client setup with the verified HMAC rule:
  - [ ] use raw base64 string as HMAC secret
  - [ ] do not decode before HMAC
- [ ] Implement `Allocate` request flow.
- [ ] Implement channel/binding flow required for media transport.
- [ ] Log relay negotiation in a structured way without leaking sensitive material unnecessarily.
- [ ] Add focused tests for TURN auth derivation.

### Acceptance Criteria

- successful TURN allocation
- successful binding/channel setup
- capture evidence proving bidirectional relay traffic

## Phase 4 — SRTP Session Derivation and Packet Protection

### Goal

Derive working SRTP session keys and build a reliable packet protection layer.

### Checklist

- [ ] Decode the signaling-side encryption payload structure used for call media.
- [ ] Implement HKDF derivation for SRTP keying material.
- [ ] Model local and remote crypto contexts separately.
- [ ] Implement SRTP packet protection and authentication.
- [ ] Implement inbound SRTP packet parsing for debug verification.
- [ ] Add deterministic tests using captured fixtures.

### Acceptance Criteria

- local SRTP packets match expected structure
- decrypt/verify flow works against captured fixture material
- packet indexes and auth tags are stable under repeated tests

## Phase 5 — RTP and Opus Media Engine

### Goal

Build the media pipeline that can actually survive inside a real WhatsApp voice call.

### Checklist

- [ ] Define the exact audio profile to target first:
  - [ ] sample rate
  - [ ] mono/stereo
  - [ ] frame duration
  - [ ] bitrate policy
  - [ ] DTX behavior
- [ ] Implement a minimal RTP packetizer for outbound audio.
- [ ] Add timestamp and sequence-number management.
- [ ] Add silence frame behavior and explicit DTX experiments.
- [ ] Support a prerecorded audio source converted to the required Opus framing.
- [ ] Add a local media debug mode to save generated RTP payload sequences for comparison.
- [ ] Compare generated packets against known-good captures.

### Important Note

A plain chat audio file is **not** enough here.

The call path needs correct live-call-compatible Opus/RTP/SRTP framing, which is a different problem from sending an audio attachment in a WhatsApp chat.

### Acceptance Criteria

- caller hears deterministic audio
- call remains alive long enough to complete the clip or a meaningful portion of it
- disconnect is no longer caused by obvious framing incompatibility

## Phase 6 — Automatic Answer Bot Behavior

### Goal

Turn the low-level pipeline into a controllable QuePasa runtime feature.

### Checklist

- [ ] Add a runtime option for auto-answer enable/disable.
- [ ] Add configuration for prerecorded test audio file path.
- [ ] Add timeout safeguards for unanswered or partially-negotiated calls.
- [ ] Ensure only one experimental call handler owns a given `CallID`.
- [ ] Add a safe fallback path when media setup fails.
- [ ] Expose enough status internally for operations and debugging.

### Optional Later Enhancements

- [ ] multiple test audio assets
- [ ] programmable call scripts
- [ ] dynamic TTS-to-Opus pipeline
- [ ] explicit public API endpoints for manual answer/hangup

## Phase 7 — Validation, Capture Diffing, and Hardening

### Goal

Prove that the implementation works against reality and not only against assumptions.

### Checklist

- [ ] Capture a full successful test run from inbound ring to audio playback.
- [ ] Compare the signaling sequence with a known-good real call.
- [ ] Compare TURN traffic with a known-good real call.
- [ ] Compare RTP payload cadence and timing with a known-good real call.
- [ ] Validate behavior across repeated calls, not only one run.
- [ ] Validate failure behavior on hangup, timeout, and network jitter.
- [ ] Add focused tests for all deterministic pieces.
- [ ] Document any remaining unknown fields or heuristics.

## Proposed Milestones

### Milestone A — Answer Without Media

Done when:

- call is observed
- call is accepted
- relay setup begins
- no useful audio yet

### Milestone B — Relay and Crypto Working

Done when:

- TURN allocation works
- SRTP packets are produced correctly
- packet traces look structurally valid

### Milestone C — Audio Heard by Caller

Done when:

- prerecorded audio is heard by caller
- call does not immediately collapse due to framing issues

### Milestone D — Repeatable Experimental Feature

Done when:

- feature can be enabled locally by config
- repeatable incoming-call tests pass
- logs/captures are sufficient for future debugging

## Suggested File Layout Direction

The implementation should avoid one giant `call.go` file.

A better split is:

- `src/whatsmeow/whatsmeow_call_state.go`
- `src/whatsmeow/whatsmeow_call_signaling.go`
- `src/whatsmeow/whatsmeow_call_accept.go`
- `src/whatsmeow/whatsmeow_call_turn.go`
- `src/whatsmeow/whatsmeow_call_srtp.go`
- `src/whatsmeow/whatsmeow_call_rtp.go`
- `src/whatsmeow/whatsmeow_call_opus.go`
- `src/whatsmeow/whatsmeow_call_runtime.go`
- `src/whatsmeow/whatsmeow_call_*_test.go`

If configuration or orchestration boundaries are needed, keep them explicit in `src/runtime/` instead of leaking everything into generic helpers.

## Verification Checklist

- [ ] Local service starts normally with VoIP feature disabled.
- [ ] Local service starts normally with VoIP feature enabled.
- [ ] Incoming call is detected and tracked by `CallID`.
- [ ] Native accept path executes successfully.
- [ ] TURN allocation succeeds.
- [ ] SRTP key derivation passes fixture validation.
- [ ] RTP packetizer produces stable output.
- [ ] Caller hears prerecorded audio.
- [ ] Call lasts longer than the current failure window.
- [ ] Repeated test calls behave consistently.
- [ ] Failure diagnostics are good enough for iterative debugging.

## Risks and Decision Gates

### High-Risk Areas

- undocumented WhatsApp signaling details
- TURN auth subtleties
- SRTP derivation mistakes
- Opus frame cadence and DTX behavior
- hidden RTP header/extension expectations

### Decision Gates

1. **Gate 1 — Signaling**
   - If inbound offer cannot be answered reliably, do not proceed to media work.

2. **Gate 2 — TURN**
   - If relay allocation/bind fails, pause and fix transport before SRTP work.

3. **Gate 3 — SRTP**
   - If packet protection is not deterministic, do not trust audio-layer experiments.

4. **Gate 4 — Audio framing**
   - If the remote hears nothing or the call drops after decryption, focus on Opus/RTP framing diffing before expanding scope.

## Final Implementation Principle

Build this as an **evidence-driven VoIP experiment** first, then evolve it into a stable QuePasa feature.

The fastest path is not broad API design.
The fastest path is:

- capture
- diff
- implement one layer at a time
- validate with real calls
- only then expose higher-level controls
