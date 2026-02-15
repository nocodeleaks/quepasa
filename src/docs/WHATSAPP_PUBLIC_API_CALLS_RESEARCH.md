# WhatsApp Public API vs Calls (Research Note)

This note is meant to answer:

- What does the **official/public** WhatsApp Business Platform / Cloud API expose for **voice calls**?
- Why does an **unofficial Web/MD** stack (like whatsmeow) still need extra work to get **audio media**?

## What we could verify from official docs (via fetch)

In this environment, the Meta docs pages for WhatsApp webhook reference root and `calls` were intermittently unavailable (showing a generic “page not available at the moment” error).

We *could* fetch the messages reference successfully:

- https://developers.facebook.com/documentation/business-messaging/whatsapp/webhooks/reference/messages

That page documents message webhooks (incoming `messages` array and outgoing `statuses` array) and does not describe voice call media transport.

Because the dedicated `calls` reference page could not be retrieved reliably here, any call-specific fields/events must be validated by opening the docs in a real browser session.

## Practical implication

Even if the official platform provides a **call event** webhook, the official APIs generally focus on **business messaging**, not delivering raw VoIP media streams to third-party servers.

So, if your goal is:

- “answer the WhatsApp call and bridge audio to SIP”,

then you should assume you need an **unofficial VoIP implementation** that handles signaling + transport + media (ICE + SRTP/RTP), because the messaging APIs do not provide that media plane.

## How this repo currently handles calls

This codebase implements an **unofficial call signaling flow** via whatsmeow and forwards the call into a SIP domain:

- Signaling: `src/whatsmeow/whatsmeow_call_manager.go`
  - Sends internal `call` stanzas (`preaccept`, `accept`, etc.) using `DangerousInternals().SendNode(ctx, node)`.
  - Runs STUN discovery unless disabled (`QP_CALL_DISABLE_STUN=1`).
  - Builds ICE-like candidates and wraps them under `net`.
- Transport parsing helper: `src/whatsmeow/whatsmeow_call_transport.go`
  - Normalizes the `events.CallTransport` “Data” payload into a tree for later interpretation.
- SIP forwarding: `src/sipproxy/sipproxy_call_answer_manager.go`
  - Converts call intent into a SIP INVITE to the configured SIP server.

## Why “accept” can stop ringing but still have no audio

Stopping the ringing / marking the call as “accepted” is mainly **signaling**.
Audio requires the **media plane** to be up:

- Remote transport must be received and interpreted (candidates, keys, tokens).
- Connectivity must be established (ICE checks / NAT traversal, possibly TURN).
- SRTP/RTP streams must be created and bridged to SIP RTP.

If the current code only sends signaling and does not complete media negotiation + SRTP bridging, the result is typically:

- Call appears accepted (ring stops on other devices), but
- No actual audio flows.

## Suggested next debugging checkpoints (repo-local)

1. Capture and persist a real `events.CallTransport` payload and inspect it through `WhatsmeowCallTransport`.
2. Confirm whether the remote transport provides:
   - ICE candidates (host/srflx/relay)
   - key material needed for SRTP
3. Decide where SRTP/RTP bridging will live:
   - In `src/sipproxy/` (RTP proxy component), or
   - In a dedicated VoIP media package (still within this repo).

## Notes about sources

- This note intentionally avoids asserting exact official capabilities for “calls” webhooks, because the `calls` reference page could not be fetched reliably in this session.
- Validate the current official state by opening the docs pages directly and/or checking the WhatsApp Business Platform changelog.
