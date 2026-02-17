
# WhatsApp Call Implementation (Unofficial)

This repository contains an *unofficial* WhatsApp call handling flow built on top of `whatsmeow` plus a SIP proxy bridge.

The goal of this document is to describe what exists in the codebase today (signaling + SIP forwarding) and what is still required for real audio media bridging (RTP/SRTP/ICE).

## High-level flow

1. **Incoming call offer** is received via `whatsmeow` events (`events.CallOffer`).
2. The call manager chooses an **accept mode** and sends internal `call` nodes to WhatsApp Web.
3. In parallel, the SIP proxy integration sends a **SIP INVITE** to an external SIP server.
4. For real audio, the system must also handle **transport negotiation** and establish media (ICE + SRTP/RTP).

## Key code locations

- WhatsApp signaling (unofficial)
	- `src/whatsmeow/whatsmeow_call_manager.go`
	- `src/whatsmeow/whatsmeow_call_manager_accept.go`
	- `src/whatsmeow/whatsmeow_call_offer.go`
	- `src/whatsmeow/whatsmeow_handlers+call.go`
	- `src/whatsmeow/whatsmeow_handlers+calls.go`
- Transport decoding helpers
	- `src/whatsmeow/whatsmeow_call_transport.go`
- SIP proxy bridge
	- `src/sipproxy/`
	- `src/sipproxy/sipproxy_call_answer_manager.go`
	- `src/environment/sipproxy_settings.go`

## Accept/handshake modes

The call manager supports multiple modes controlled by environment variables:

- `QP_CALL_ACCEPT_MODE`
	- `handshake` (default): attempts a fuller handshake (preaccept + transport + accept)
	- `direct`: tries sending accept directly
- `QP_CALL_HANDSHAKE_MODE` (affects the handshake strategy)
	- Examples observed in code: `preaccept+transport` (default), `preaccept-only`, `accept-early`, `accept-immediate`

## Transport debugging (recommended)

Incoming `CallTransport` payloads are critical to understand why audio is not flowing.
You can enable safe, offline inspection by dumping a normalized JSON file per transport event:

- `QP_CALL_DUMP_TRANSPORT=1` enables file dumps.
- `QP_CALL_DUMP_OFFER=1` dumps normalized `CallOffer` payloads as well (useful for `voip_settings` and relay tokens).
- `QP_CALL_DUMP_DIR` sets the dump directory (default: `.dist/call_dumps`).
- `QP_CALL_LOG_TRANSPORT_JSON=1` logs the raw event JSON to stdout (not recommended for production; may be noisy and can include sensitive tokens).

The implementation builds and sends `binary.Node` stanzas with tags such as:

- `call` root
- `preaccept`
- `transport`
- `accept`

These nodes are sent via `Client.DangerousInternals().SendNode(context.Background(), node)`.

## Media reality check

Sending signaling nodes and forwarding to SIP is not sufficient by itself to get audio.
For two-way audio, the system still needs to:

- Parse remote `CallTransport` contents (ICE candidates, tokens, keys/material, etc.).
- Perform ICE negotiation (host/srflx/relay candidates, connectivity checks).
- Establish SRTP/RTP streams and bridge them to/from the SIP side.

See `src/docs/WHATSAPP_PUBLIC_API_CALLS_RESEARCH.md` for a focused note about official/public API exposure vs. unofficial VoIP requirements.

