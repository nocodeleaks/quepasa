# Task Objective
- Stabilize `calls` branch for WhatsApp voice call handling (offer/accept/transport) and SIP proxy bridging.
- Track competitor landscape for non-official WhatsApp API call support.

# Mandatory Checklist
- Keep `AGENTS.md` on custom branches only (do not merge into `develop`/`main`).
- Preserve VoIP/SIP code in `src/sipproxy` and call-related code in `src/whatsmeow`.
- Validate build on Windows before publishing.

# Current Status
- `calls` branch builds successfully (warnings from `go-sqlite3` may appear but do not fail the build).
- Whatsmeow API signature changes were aligned in `calls` to restore build/debug.
- Added call documentation notes under `src/docs/`:
  - `src/docs/WHATSAPP_CALL_IMPLEMENTATION.md` (repo-specific overview)
  - `src/docs/WHATSAPP_PUBLIC_API_CALLS_RESEARCH.md` (public API vs unofficial VoIP)

# Next Steps
- Decide call scope: reject-only vs accept + media bridge vs full RTP/SRTP integration.
- Validate call event ingestion end-to-end (offer -> SIP proxy -> termination).
- If publishing, push `calls` branch to remote and open PR if desired.

# Immutable Constraints Discovered During Execution
- WhatsApp call handling is sensitive to upstream protocol/API changes; upgrades commonly require refactors in internal send/receive primitives.
- This repository treats LIDs as opaque identifiers; never derive phone numbers from `@lid`.
- Meta/WhatsApp docs pages for webhook reference `calls` were intermittently unavailable via fetch in this environment; validate call webhook details in a real browser session.

# Competitor Notes (Public Sources)
- Evolution API (EvolutionAPI/evolution-api)
  - Has call-related settings such as `rejectCall`, `msgCall`, and `wavoipToken`.
  - Contains a `voiceCalls` integration (`useVoiceCallsBaileys.ts`) that connects to a wavoip websocket endpoint and proxies call signaling/events.
  - Includes a call controller (`offerCall`) but code indicates it may be stubbed/disabled in some paths.
  - Conclusion: has active work around calls, but “ready to answer + bridge media” depends on their wavoip integration and is not guaranteed plug-and-play.
-
- WPPConnect Server (wppconnect-team/wppconnect-server)
  - Exposes incoming call events (`onIncomingCall`) and includes a reject-call endpoint (`rejectCall`).
  - Conclusion: supports call detection and rejection; no clear full media/bridge implementation as a general feature.
-
- whatsapp-web.js (pedroslopez/whatsapp-web.js)
  - Exposes call events and supports rejecting calls.
  - Also supports generating WhatsApp call links.
  - Conclusion: call event + reject is supported; full call answering/media bridging is not presented as a productized feature.
-
- Baileys (WhiskeySockets/Baileys)
  - Emits `call` events and provides primitives such as `rejectCall`.
  - Conclusion: SDK-level primitives exist; full call answering/media transport is still custom work.
-
- “Ability/Abeility”
  - Not identified reliably via GitHub repository search by name; may be closed-source or spelled differently.
  - Action: if an exact URL/name is provided, re-check for call support.
