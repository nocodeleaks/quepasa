# QuePasa + 3CX Bridge Setup

This guide configures QuePasa as a SIP bridge/trunk peer to 3CX.

## Goal

- Receive WhatsApp incoming call events in QuePasa.
- Forward each call as SIP INVITE to 3CX.
- Let 3CX decide final destination through inbound routing (queue/ring group/extensions).

## Prerequisites

- QuePasa running with `CALLS=true`.
- 3CX reachable from QuePasa host on SIP port (typically UDP 5060).
- One SIP authentication identity in 3CX for this bridge/trunk integration.

## 1) Configure QuePasa Environment

Edit `.env` and configure the SIP block:

```env
SIPPROXY_HOST=<3CX_IP_OR_FQDN>
SIPPROXY_PORT=5060
SIPPROXY_PROTOCOL=UDP
SIPPROXY_LOCALPORT=5061

SIPPROXY_AUTHUSERNAME=<3CX_AUTH_ID>
SIPPROXY_AUTHPASSWORD=<3CX_AUTH_PASSWORD>

# Keep empty to preserve real WhatsApp caller on SIP identity headers
# SIPPROXY_FROMUSER=
SIPPROXY_TOUSER=<3CX_DESTINATION>

SIPPROXY_PUBLICIP=
SIPPROXY_STUNSERVER=stun.l.google.com:19302
SIPPROXY_USEUPNP=true
SIPPROXY_MEDIAPORTS=10000-20000
SIPPROXY_CODECS=PCMU,PCMA,G729
SIPPROXY_TIMEOUT=30
SIPPROXY_RETRIES=3
```

Note: SIP bridge mode is enabled automatically when `SIPPROXY_HOST` is set.

### Recommended values for bridge mode

- `SIPPROXY_FROMUSER`: keep empty to preserve caller number from WhatsApp.
- `SIPPROXY_TOUSER`: queue/ring-group/route-point number in 3CX, for example `800`.

### Caller ID policy

- Default and recommended: do not set `SIPPROXY_FROMUSER`, so 3CX receives dynamic caller identity from the incoming WhatsApp call.
- Optional fixed identity mode: set `SIPPROXY_FROMUSER` only if your PBX policy requires a static caller identity for all bridged calls.
- `SIPPROXY_AUTHUSERNAME`/`SIPPROXY_AUTHPASSWORD` are authentication credentials only and should not be used as caller identity.

## 2) Configure 3CX (Bridge/Trunk Style)

Use 3CX admin console and create or adapt a SIP endpoint that will authenticate QuePasa.

1. Create a dedicated SIP auth identity for QuePasa bridge.
2. Ensure transport and port match QuePasa (`UDP`/`5060` by default).
3. Set inbound handling to route calls to your desired destination logic.
4. Use destination that can fan out (Queue/Ring Group) instead of a single extension.
5. If 3CX requires source filtering, allow QuePasa server IP.
6. Ensure codec overlap includes `PCMU` or `PCMA`.

## 3) Start/Restart QuePasa

From `src/`:

```bash
go run main.go
```

## 4) Validate Signaling

On incoming WhatsApp call, expected progression in logs:

1. CallOffer received.
2. SIP bridge forwarding started.
3. SIP INVITE sent to 3CX.
4. No final `407 Proxy Authentication Required`.
5. Final SIP status should evolve to `180 Ringing` and/or `200 OK`.

## 5) Troubleshooting Map

- `407 Proxy Authentication Required`: wrong/missing `SIPPROXY_AUTHUSERNAME` or `SIPPROXY_AUTHPASSWORD`.
- `404 Not Found`: invalid `SIPPROXY_TOUSER` in 3CX.
- `403 Forbidden`: source permission or policy issue in 3CX.
- `486 Busy Here`: destination endpoint is busy.
- Rings but no caller number in 3CX: ensure `SIPPROXY_FROMUSER` is empty and verify 3CX trust/usage of identity headers.

## 6) SIP Headers Sent by QuePasa Bridge

For each forwarded call, QuePasa adds identity and trace headers on the SIP INVITE:

- `P-Asserted-Identity`: asserted SIP identity for caller display/routing.
- `Remote-Party-ID`: caller identity hint for PBX compatibility.
- `X-QuePasa-WA-Caller`: original WhatsApp caller number.
- `X-QuePasa-WA-Called`: WhatsApp number that received the call in QuePasa.
- `X-QuePasa-Session`: session identifier (currently aligned with called WhatsApp number).
- `X-QuePasa-CallID`: shared call identifier across WhatsApp and SIP flows.

These headers are useful for multi-session operations, auditing, CDR enrichment, and PBX-side routing rules.

## 7) Test Plan (End-to-End)

1. Confirm QuePasa is connected to WhatsApp and `CALLS=true`.
2. Trigger an incoming WhatsApp call.
3. Confirm SIP INVITE reaches 3CX.
4. Confirm 3CX routes to queue/ring group.
5. Answer from one target extension and check call acceptance behavior.
6. Hang up from either side and verify termination is propagated.

## 8) Multi-Session Operating Model

When multiple WhatsApp sessions are connected to the same QuePasa instance:

- Keep one bridge destination (`SIPPROXY_TOUSER`) if you want a single 3CX entry point.
- Use `X-QuePasa-WA-Called` and `X-QuePasa-Session` in 3CX logs/CDR to identify which WhatsApp session received each call.
- For advanced routing, create PBX rules that inspect custom headers and dispatch to different queues by WhatsApp session.

## 9) Example 3CX Header-Based Routing

Use this pattern when QuePasa forwards calls from multiple WhatsApp sessions into a single 3CX entry point.

Example target mapping:

- WhatsApp session `5511999990001` -> Queue `801`
- WhatsApp session `5511999990002` -> Queue `802`
- Any other session -> Default Queue `800`

Recommended rule design in 3CX inbound processing:

1. Create rule `qp-wa-session-0001`
2. Condition: custom SIP header `X-QuePasa-WA-Called` equals `5511999990001`
3. Action: route to Queue/Ring Group `801`
4. Create rule `qp-wa-session-0002`
5. Condition: custom SIP header `X-QuePasa-WA-Called` equals `5511999990002`
6. Action: route to Queue/Ring Group `802`
7. Keep a final fallback/default rule to Queue/Ring Group `800`

Operational tips:

- Keep `SIPPROXY_TOUSER` as the single trunk entry on QuePasa side.
- In 3CX logs, correlate `X-QuePasa-CallID` with WhatsApp/SIP events.
- If your 3CX version does not expose custom headers in GUI routing, keep default route and use headers for CDR/log traceability.

## Notes

- Current implementation is SIP bridge/origination.
- It is not a full SIP registrar for extension registration management.
- Use 3CX routing rules for distribution and business logic.
