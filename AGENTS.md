# Task Objective
- Stabilize `calls` branch for WhatsApp voice call handling (offer/accept/transport) and SIP proxy bridging.
- Track competitor landscape for non-official WhatsApp API call support.

# Ground Truth (Decided)
- `CallAccept` (answered) events are observed using `@lid` JIDs (e.g. `from_raw=...@lid`).
- `CallOffer` often includes `caller_pn=...@s.whatsapp.net` even when `from` is `@lid`.
- Taking ownership (stop ringing on other devices) requires the exact snippet ACCEPT shape (`QP_CALL_USE_SNIPPET_ACCEPT=1`).
- For experiments, replies should use the same peer JID form received; if peer is `@lid`, reply using `@lid` (`QP_CALL_REPLY_USE_LID=1`).

# Mandatory Checklist
- Keep `AGENTS.md` on custom branches only (do not merge into `develop`/`main`).
- Preserve VoIP/SIP code in `src/sipproxy` and call-related code in `src/whatsmeow`.
- Validate build on Windows before publishing.
- Before EVERY server call test (apoint-voip), reset noise: clear `/opt/quepasa/.dist/call_dumps` + `/opt/quepasa/.dist/pcaps`, vacuum journald, restart service.

## Server Test Hygiene (apoint-voip)
- Always run this before starting a new test call to avoid mixing CallIDs/logs:
  - `rm -f /opt/quepasa/.dist/call_dumps/* /opt/quepasa/.dist/pcaps/* || true`
  - `journalctl --rotate || true; journalctl --vacuum-time=1s || true`
  - `systemctl restart quepasa; sleep 1; journalctl -u quepasa --no-pager -n 20`

# Current Status
- `calls` branch builds successfully (warnings from `go-sqlite3` may appear but do not fail the build).
- Whatsmeow API signature changes were aligned in `calls` to restore build/debug.
- Added call documentation notes under `src/docs/`:
  - `src/docs/WHATSAPP_CALL_IMPLEMENTATION.md` (repo-specific overview)
  - `src/docs/WHATSAPP_PUBLIC_API_CALLS_RESEARCH.md` (public API vs unofficial VoIP)
- Fixed ICE candidate construction for ACCEPT/TRANSPORT: host candidate now uses the real local UDP socket port, and optional srflx uses the STUN XOR-MAPPED public IP:port (prevents invalid combos like local IP + mapped port).
- Enabled `QP_CALL_INCLUDE_SRFLX=1` in `src/.env` to advertise srflx candidate by default.
- Prepared server deployment assets for `apoint-voip.sufficit.com.br`:
  - `src/.env.apoint-voip` (Linux-focused env)
  - `helpers/deploy-apoint-voip.ps1` (uploads `src/` + activates env + builds on server)
- Deployed to `apoint-voip.sufficit.com.br` and installed Go 1.24.2 under `/usr/local/go`.
- Fixed Linux migrations path bug (`file://opt/...` invalid) so the service can start on Linux.
- Service was deployed and is running on the server (effective port depends on the deployed `.env`; current template uses `WEBSERVER_PORT=31000` and requires service restart to apply).
- Systemd unit template `helpers/quepasa.service` was updated to run the deployed binary `/opt/quepasa/quepasa` with `WorkingDirectory=/opt/quepasa/src` and `EnvironmentFile=/opt/quepasa/src/.env`.
- Default server port for `apoint-voip` env was set to `WEBSERVER_PORT=31000`.

- Media-port consistency fix deployed (2026-02-15): WhatsApp call flow now locks a single media UDP port per `CallID` and reuses it across PREACCEPT/TRANSPORT/ACCEPT and the RTP bridge listener.
  - Evidence: logs show `🎵🔒 [MEDIA-PORT] Locked media port=47993` and `✅🎵 [UDP-SUCCESS] Listener UDP ativo na porta 47993` for the same call.
  - Result: the previous bug (announcing one port but listening on another) is resolved.

- RTP still not reaching QuePasa listeners:
  - SIP-side RTP monitor reserved/advertised a port (example: `:10278`) but did not log `First RTP packet`.
  - Server tcpdump + Asterisk RTP debug show inbound RTP hitting a different local port (example: `143.208.224.21:10250`) and Asterisk logs `Got RTP packet from 177.36.188.29:10030`.
  - Interpretation: media is currently flowing to Asterisk’s negotiated RTP port (from its SDP answer), not to QuePasa’s SIP RTP monitor nor to the WhatsApp RTP bridge socket.

# Next Steps
- Confirm `snippet ACCEPT` behavior is stable across multiple calls and accounts.
- Capture and dump the first `CallTransport` received after the state flips to `connecting`.
- Implement the next step: respond to `CallTransport` with the required media endpoint/relay details (so the peer knows where to send audio).
- Push TURN relay auth discovery further:
  - Try `relay.key` / `relay.hbh_key` double-base64 decode (base64(base64(bytes))) as Allocate integrity candidates.
  - Try `CallAccept` `<te>` payload values as additional TURN `USERNAME` variants (budget-limited) to match relay expectations.
  - Re-run a fresh call and inspect `call_turn_probe_*.json` for the first successful Allocate (or new error codes).
- Decide call scope: reject-only vs accept + media bridge vs full RTP/SRTP integration.
- Validate call event ingestion end-to-end (offer -> SIP proxy -> termination).
- Capture real `CallTransport` payloads for media debugging (enable `QP_CALL_DUMP_TRANSPORT=1`, optional `QP_CALL_DUMP_DIR`).
- Optionally capture `CallOffer` payloads too (enable `QP_CALL_DUMP_OFFER=1`) to inspect `voip_settings`/relay tokens.
- Run a fresh call and confirm logs show consistent candidates:
  - `host`: `localIP:localPort` (localPort from the STUN socket)
  - `srflx` (when enabled): `publicIP:publicPort` (XOR-MAPPED)
  - No more `host` candidate using the mapped public port.
- If WhatsApp still stays "connecting" after this fix, focus shifts back to relay/SRTP media plane (relay-only calls with `disable_p2p=1` likely require real relay/ICE/SRTP handling beyond signaling).

- Add SIP 200 OK SDP parsing on the sipgo client side and log the negotiated remote media `IP:port` from `DialogClientSession.InviteResponse.Body()`.
- Optional: implement an env-gated RTP "probe" sender (e.g. `SIPPROXY_RTP_PROBE=1`) to send a few RTP packets from the reserved socket to the negotiated remote media to help Strict RTP/NAT learning.

- Implemented (2026-02-15): sipgo 200 OK SDP parsing + optional RTP probe
  - `parseSDPAudioEndpoint()` extracts `c=` + `m=audio` from `InviteResponse.Body()`.
  - `SIPPROXY_RTP_PROBE=1` sends a short RTP probe burst from the reserved local RTP socket to the negotiated remote media endpoint.
  - `SIPPROXY_LOG_SDP_200OK=1` logs full SDP body (noisy; keep off by default).

- Breakthrough (2026-02-15): SIP-side RTP monitor is now receiving RTP
  - Evidence (CallID `A5EB75E9C9599452EB6796DA55180E98`): `🎵 [RTP-MONITOR] First RTP packet` from `177.36.188.29:*` + continuous RTP stats.
  - WhatsApp-side RTP bridge reception is still unconfirmed; existing bridge logs were missing `CallID` in `UDP-SUCCESS`/`RTP-RECEIVED`, so correlation was ambiguous.
  - Next change: include `CallID` in RTP bridge logs and optionally dump the first packet (`QP_CALL_RTP_DUMP_FIRST=1`) to identify packet format (RTP vs SRTP).

- Implemented (2026-02-15): WhatsApp-side candidate + RTP bridge debug hardening
  - Removed hardcoded local IP fallback (`192.168.31.202`) in PREACCEPT/ACCEPT/TRANSPORT flows (now fails fast if local IPv4 cannot be determined).
  - `srflx` candidates now include `rel-addr`/`rel-port` pointing to the host candidate.
  - RTP bridge logs now include `CallID` consistently (`UDP-SUCCESS`, `RTP-RECEIVED`, `RTP-TIMEOUT`, read errors).
  - Optional first packet dump added: set `QP_CALL_RTP_DUMP_FIRST=1` to log first 32 bytes of the first received packet per call.

- Implemented (2026-02-15): SIP-side RTP mirror to WhatsApp bridge (debug)
  - When the WhatsApp call manager locks `LocalMediaPort` for a `CallID`, it configures the SIP RTP monitor to mirror RTP packets to `127.0.0.1:<LocalMediaPort>`.
  - Goal: force `RTP-BRIDGE` to log `RTP-RECEIVED`/`RTP-DUMP-FIRST` and confirm local UDP path + packet shape even before real WhatsApp SRTP/ICE send is implemented.

- Verified (2026-02-15): RTP mirror reaches WhatsApp RTP bridge
  - Evidence (CallID `A599E56FCD208713E6B33D7CEFF89530`): `🎵 [RTP-MIRROR] Mirroring RTP stream to 127.0.0.1:38239` + WhatsApp-side `🎵📥 [RTP-RECEIVED] 172 bytes de 127.0.0.1:*`.
  - Conclusion: local UDP listener path is working; remaining gap is real WhatsApp media plane/SRTP send (not implemented yet).

- Implemented (2026-02-15): make `QP_CALL_RTP_DUMP_FIRST` robust on Linux/systemd env files
  - `RTP-DUMP-FIRST` gating now uses `strings.TrimSpace()` and accepts `true` as well as `1`.
  - Rationale: `.env` deployed from Windows may have CRLF, making the env value effectively `"1\r"` and preventing the dump from triggering.

- Observed (2026-02-15): first RTP packet dump confirms SIP-side stream format
  - Evidence (CallID `A520AE09B6BE509032DA188A363EF8EE`): `RTP-DUMP-FIRST ... first=8080001000000a035ab6115f...` from `127.0.0.1:*`, packet size `172` bytes.
  - Interpretation: starts with `0x80` (RTP v2) and `0x80` second byte implies `M=1` and `PT=0` (PCMU/G.711u) in standard RTP payload type mapping; 172 bytes suggests `12-byte RTP header + 160-byte audio payload` (~20ms @ 8kHz).

- Verified (2026-02-15): enhanced `RTP-DUMP-FIRST` decoding + mirror/port alignment
  - Evidence (CallID `A59767328C5DD98ED6D7BD0B6527FD35`): SIP monitor reserved `:10184`, WhatsApp bridge listener on `:51176`, mirror target `127.0.0.1:51176`, and dump shows `v=2 pt=0 payload=160`.

- Implemented (2026-02-15): RTP-to-WAV dump for SIP audio validation
  - When `QP_CALL_RTP_DUMP_WAV=1`, the WhatsApp-side RTP bridge decodes RTP payload type `0` (PCMU/G.711u) into PCM16 @ 8kHz and writes a short `.wav` to `QP_CALL_DUMP_DIR`.
  - `QP_CALL_RTP_DUMP_WAV_SECONDS` limits capture duration (clamped 1..60; default 10).
  - Goal: validate that received RTP contains real audio before attempting WhatsApp media-plane/SRTP/ICE send.

- Verified (2026-02-15): WAV dumps generated on server
  - Evidence: `/opt/quepasa/.dist/call_dumps/rtp_A5F5E6370D65EBC0048E0EE0DD059335_20260215_020717.wav` (~116KB) and `rtp_A541E52897EA9B7DF49103867BA0DA3D_20260215_020713.wav` (~9.6KB).
  - Conclusion: mirrored RTP stream contains decodable PCMU audio; next gap remains WhatsApp media-plane/SRTP send.

- Observed (2026-02-15): relay-only call signaling profile (no ICE details in peer transport)
  - Evidence (CallID `A5B23B21A1255061387DC0CE99C69702`): `OFFER-SUMMARY offer.medium=3 disable_p2p=true relay_candidates=true relays=[fcfc2c01 gru2c02 poa1c01] relay_tokens=3`.
  - Evidence (same CallID): `TRANSPORT-SUMMARY medium=2 protocol=0 ... candidates=0 fingerprints=0` and `TRANSPORT-CHILD-0 Tag=net Attrs=map[medium:2 protocol:0]`.
  - Interpretation: peer transport is relay-only (`medium=2`) and does not include P2P ICE ufrag/pwd/candidates in `CallTransport`; media-plane work likely depends on relay tokens/enc payload and SRTP/relay implementation.

- Verified (2026-02-15): enhanced `OFFER-SUMMARY` shows relay material needed for media-plane work
  - Evidence (CallID `A5087E33A5A6F85E68C39FB092153DDA`): `OFFER-SUMMARY offer.medium=3 disable_p2p=true relay_candidates=true relays=[ffln5c01 gru2c02 poa1c01] relay_tokens=3 sample_tokens_b64=[...]`.
  - Evidence (same CallID): relay metadata present: `relay.uuid=... relay.self_pid=3 relay.peer_pid=1 relay.te2=10 relay.protocols=[1] relay.token_nodes=3 relay.auth_token_nodes=2 relay.key=... relay.hbh_key=...`.
  - Evidence (same CallID): `TRANSPORT-SUMMARY medium=2 protocol=0 candidates=0 fingerprints=0` (still relay-only; no ICE ufrag/pwd/candidates).

- Implemented (2026-02-17): snippet ACCEPT now respects `QP_CALL_NET_MEDIUM=auto`
  - `src/whatsmeow/whatsmeow_call_manager.go`: snippet ACCEPT `net.medium` is no longer hardcoded to `3`; it uses `getNetMediumForCall(callID)`.
  - Goal: for relay-only offers (peer transport `medium=2`), send snippet ACCEPT with `medium=2` to avoid medium mismatch during the early handshake.

- Implemented (2026-02-15): structured relay parsing (prep for SRTP/relay media-plane)
  - `src/whatsmeow/whatsmeow_call_offer.go`: added `RelayBlock` + `ExtractRelayBlock()` and cached access via `GetRelayBlock()`.
  - Relay-related helpers (`GetRelayTokens()`, `HasRelayCandidatesCached()`, `RelayNamesCached()`) now reuse the single relay parse instead of re-unmarshalling relay nodes.
  - `src/whatsmeow/whatsmeow_handlers+calls.go`: `OFFER-SUMMARY` now pulls `relay.uuid/self_pid/peer_pid`, token/auth_token counts, `te2` count, `protocols`, `key` and `hbh_key` from `GetRelayBlock()` (still redacted by default).
  - `src/whatsmeow/whatsmeow_call_manager.go`: handshake state (`CallHandshakeState`) now stores the parsed `RelayBlock` per `CallID` for future media-plane/SRTP logic.

- Implemented (2026-02-15): `QP_CALL_NET_MEDIUM=auto` (relay-aware)
  - `src/whatsmeow/whatsmeow_call_manager.go`: added `getNetMediumForCall(callID)`; when env is `auto` and offer had relay candidates (`RelayBlock.TE2`), we pick `medium=2`.
  - Updated server env `/opt/quepasa/src/.env`: set `QP_CALL_NET_MEDIUM=auto` (keeps behavior fixed if env is 1/2/3).
  - Updated template `src/.env.apoint-voip`: set `QP_CALL_NET_MEDIUM=auto` to avoid reverting on next deploy.

- Implemented (2026-02-15): `QP_CALL_INCLUDE_SRFLX_ALWAYS=1` (signaling experiment)
  - `src/whatsmeow/whatsmeow_call_manager.go`: can advertise an `srflx` candidate even when `publicIP == localIP` (previously suppressed), to raise `candidates` count in PREACCEPT/ACCEPT.
  - Updated template `src/.env.apoint-voip`: enable `QP_CALL_INCLUDE_SRFLX_ALWAYS=1`.

- Implemented (2026-02-15): relay-only PREACCEPT empty-net experiment
  - Goal: for relay-only offers (`disable_p2p=true`, peer transport `medium=2`), try sending PREACCEPT with `net medium=2` and `candidates=0` to trigger peer `CallTransport` delivery (previously peer often terminated without sending transport even when we advertised host/srflx candidates).
  - `src/.env.apoint-voip`: set `QP_CALL_NET_MEDIUM=auto` and enable `QP_CALL_PREACCEPT_RELAY_EMPTY_NET=1` (gate only applies when computed medium is `2`).
  - `src/whatsmeow/whatsmeow_call_manager.go`: switched `QP_CALL_INCLUDE_SRFLX*` parsing to `envTruthy()` (TrimSpace + accepts `true/yes`) to be robust with CRLF env files.

- Implemented (2026-02-15): relay endpoint extraction from `CallRelayLatency`
  - `src/whatsmeow/whatsmeow_handlers+calls.go`: decodes `<relaylatency><te>` payload bytes into `relay_name -> ip:port` and logs `RelayLatency decoded`.
  - Added `QP_CALL_DUMP_RELAY_LATENCY=1` to dump decoded endpoints to `call_relaylatency_<CallID>_*.json` under `QP_CALL_DUMP_DIR`.
  - Observed example (CallID `A592DBBE2DFBD41F4DC8FCE710ACCDEB`): `ffln5c01=170.150.237.35:3478`, `poa1c01=57.144.179.54:3478`, `gru1c02=157.240.226.62:3478`.

- Implemented (2026-02-15): hard-disable SIP forwarding on BasicCallMeta path
  - `src/whatsmeow/whatsmeow_handlers.go`: `CallMessage` and `CallTerminateMessage` now respect `QP_CALL_DISABLE_SIP_FORWARDING=1` (no SIP forwarding and no SIP BYE/CANCEL forwarding via this alternate handler path).

- Implemented (2026-02-15): explicit call/SIP config log on startup
  - `src/main.go`: logs `[CALL-CONFIG] QP_CALL_ACCEPT_MODE=... QP_CALL_DISABLE_SIP_FORWARDING=... SIPPROXY_HOST=...` at boot to make runtime mode unambiguous.

- Implemented (2026-02-15): minimal relay UDP/STUN session probe (first media-plane step)
  - Purpose: validate outbound UDP reachability to Meta relay endpoints (`CallRelayLatency` decoded `ip:port`) and capture first inbound packet bytes.
  - `QP_CALL_RELAY_SESSION_PROBE=1`: after legacy `<accept>`, dials UDP to the best relay endpoint (lowest latency when parseable), loops STUN Binding Requests, and logs `RELAY-SESSION-STUN` mapped address or `RELAY-SESSION-PACKET` first non-STUN packet.
  - Normalized relay endpoints to `net.JoinHostPort()` to avoid confusing `::` formatting.

- Implemented (2026-02-17): TURN probe txid correlation for PCAP
  - `src/whatsmeow/whatsmeow_call_relay_session_probe.go`: logs `txid` (transaction ID hex) for base Allocate, auth discovery, and each integrity attempt.
  - `call_turn_probe_*.json` now includes `base_allocate_txid`, `discovery_txid`, and per-attempt `txid` to match request/response pairs in captured PCAPs.

- Verified (2026-02-17): txid dump ↔ PCAP correlation works (apoint-voip)
  - CallID: `A50E15353F23E4A43F7C598C7D5779EB`
  - Server artifacts:
    - `/opt/quepasa/.dist/call_dumps/call_turn_probe_20260217104546_A50E15353F23E4A43F7C598C7D5779EB.json`
    - `/opt/quepasa/.dist/pcaps/wa_turn_20260217_104504.pcap`
  - Correlation:
    - Dump `base_allocate_txid` matches PCAP Allocate request/response txid.
    - Dump per-attempt `txid` matches PCAP Allocate request/response txid.
  - Outcome remains blocked:
    - Base Allocate returns `451 Integrity failure: Hmac missing`.
    - All integrity attempts return `450 Integrity failure: Hmac mismatch`.
    - Relay responses still provide no `REALM`/`NONCE`.

- Implemented (2026-02-17): try TURN USERNAME variants using binary relay UUID
  - Observation: `relay.uuid` in offers looks base64-ish (example `dUWd1wpo7PJZp5tR`) and decodes cleanly to 12 bytes.
  - Change: `src/whatsmeow/whatsmeow_call_relay_session_probe.go`
    - `buildRelayUsernameVariants()` now adds `uuid(bin)` variants (raw bytes) and combinations with `self/peer`.
    - Candidate sorting now prioritizes `u=uuid(bin)` before ASCII `u=uuid` so it fits within per-family budgets.
  - Goal: cover relays that expect STUN USERNAME as raw UUID bytes instead of UTF-8 text.

- Fixed (2026-02-17): relay.uuid may be base64url ("_"/"-") so uuid(bin) decode must support URL alphabet
  - Evidence (CallID `A50FEA29DCF60C2E4656B9C3B210D07A`): `relay.uuid=j_aJr_IHSgOlPVwi` contains `_` (base64url).
  - Symptom: no `u=uuid(bin)` candidates appeared; TURN probe stayed on ASCII `u=uuid...` and kept failing `450/451`.
  - Change: `src/whatsmeow/whatsmeow_call_relay_session_probe.go`
    - Updated UUID decode helper to try `RawURLEncoding`/`URLEncoding` as well as std base64, with optional padding.
  - Next: re-run call to confirm first candidates include `u=uuid(bin)` and compare Allocate error codes.

- Verified (2026-02-17): `u=uuid(bin)` is now attempted first (and visible on-wire)
  - CallID: `A5673F94D21A7DA7DBF8D627FB0BB88B`
  - Dump evidence: first 40 TURN integrity attempts are `...:u=uuid(bin):...` (relay.key + relay.hbh_key variants).
  - PCAP evidence: Allocate requests include non-UTF8 USERNAME bytes (now identifiable as `hex:...` in the decoder output).
  - Outcome unchanged: all integrity attempts still return `450 Integrity failure: Hmac mismatch` (112/112).

- Implemented (2026-02-17): enable TURN REST candidate attempts by default
  - `src/whatsmeow/whatsmeow_call_relay_session_probe.go`: `QP_CALL_RELAY_TURN_TRY_REST` now defaults to enabled (still disable-able via env).
  - Rationale: adds a distinct auth family (password = base64(hmac-sha1(secret, username))) within existing budgets/maxTry.

- Implemented (2026-02-17): use CallOffer `<enc>` (pkmsg/msg) as TURN probe seed/secret
  - `src/whatsmeow/whatsmeow_call_offer.go`: added `ExtractEncBlock()` (decodes wrapped base64 bytes or loose base64 strings).
  - `src/whatsmeow/whatsmeow_handlers+calls.go` + `src/whatsmeow/whatsmeow_call_manager.go`: store offer `EncBlock` in `CallHandshakeState` per `CallID`.
  - `src/whatsmeow/whatsmeow_call_relay_session_probe.go`: adds REST candidates where `secret=enc.pkmsg` (and `enc.sha256`) and derived candidates where `seedMsg=enc`.
  - Logging remains safe: only `enc.type/v/len/kind` is logged; raw bytes only exist in dumps/state.

- Implemented (2026-02-17): helper to fetch clean per-call artifacts from apoint-voip
  - `helpers/fetch-apoint-voip-call-artifacts.ps1 -Latest` downloads `/opt/quepasa/.dist/call_dumps/*<CallID>*` + filtered `journalctl` into `.dist/server_artifacts/<CallID>/`.

- Verified (2026-02-17): TURN Allocate error responses expose no REALM/NONCE (even with Fingerprint)
  - CallID: `A56975CB20CAB9EEC85E8D1AF39E4C55` (pcap `wa_turn_20260217_113832.pcap`)
  - Requests include `FINGERPRINT` (`fp=1` in decoder output).
  - Error responses consistently include only `ERROR-CODE (0x0009)` and `XOR-MAPPED-ADDRESS (0x0020)`.
  - `realmLen=0 nonceLen=0` for all Allocate error responses; no long-term challenge is ever offered.
  - Conclusion: relay uses short-term integrity, but the correct key/username pair is still unknown.
- Implemented (2026-02-15): TURN integrity candidate prioritization
  - `src/whatsmeow/whatsmeow_call_relay_session_probe.go`: when `QP_CALL_RELAY_TURN_DERIVE_HMAC=1`, the probe now sorts integrity candidates to attempt `drv` (derived) candidates first.
  - Increased `QP_CALL_RELAY_TURN_MAX_CANDIDATES` clamp to allow larger values (up to 120) and added a log line `Integrity candidates: total=... maxTry=...`.
  - `src/.env.apoint-voip`: set `QP_CALL_RELAY_TURN_MAX_CANDIDATES=60` to ensure derived candidates are actually attempted.
- Implemented (2026-02-15): expand TURN key-derivation search space
  - `src/whatsmeow/whatsmeow_call_relay_session_probe.go`: tries additional derived-key seed message variants (`auth+tok`, `uuid+tok+auth`, separators like `:` and `\x00`).
  - Token/auth values are also tried as base64-decoded bytes when they look like base64 (adds `drv(dec)` candidates).
  - Added a safe log line `First candidates: ...` (labels only) to confirm ordering without leaking secrets.
- Implemented (2026-02-15): TURN long-term auth discovery + Fingerprint
  - `src/whatsmeow/whatsmeow_call_relay_session_probe.go`: parses `REALM`/`NONCE` from Allocate error responses and performs a second discovery Allocate with `USERNAME` (no MI) to try to elicit `REALM/NONCE`.
  - When `REALM/NONCE` is available, probe attempts standard TURN long-term MI with key `MD5(username:realm:password)` and includes `USERNAME/REALM/NONCE` in the request.
  - Adds `stun.Fingerprint` to Allocate requests (base + attempts) to match common TURN server expectations.
- Fixed (2026-02-15): TURN Fingerprint generation order
  - Relay returned `453 "Invalid Fingerprint"` when Fingerprint was added before `Encode()`.
  - Fingerprint is now added after `Encode()` as the last attribute and the message is not re-encoded afterwards.
- Fixed (2026-02-15): TURN MESSAGE-INTEGRITY invalidated by re-Encode
  - Root cause: integrity attempts computed `MESSAGE-INTEGRITY`, then called `Encode()` which re-serialized the message and invalidated the HMAC (causing guaranteed `450 Hmac mismatch`).
  - TURN requests now follow: add base attrs -> `Encode()` -> `MESSAGE-INTEGRITY.AddTo()` -> `FINGERPRINT.AddTo()` with no re-encode.
- Adjusted (2026-02-15): candidate ordering to avoid `drv` starvation
  - `src/whatsmeow/whatsmeow_call_relay_session_probe.go`: integrity candidate sort now prioritizes direct `relay.key`/`relay.hbh_key` and `te2` variants before the large `drv` bucket.
  - Added `Candidate buckets: ...` log to confirm which categories exist for a given call and how maxTry may truncate them.
- Implemented (2026-02-15): expand TURN candidate families (MD5 lt0)
  - Added `lt0=md5(user::pass)` candidates (MD5 of `username ":" "" ":" password`) to try long-term-style key derivation even when the relay never provides `REALM/NONCE`.
  - Increased `QP_CALL_RELAY_TURN_MAX_CANDIDATES` in `src/.env.apoint-voip` to `120` to ensure these candidates are not truncated.
- Implemented (2026-02-15): expand TURN USERNAME variants (SelfPID/PeerPID)
  - `src/whatsmeow/whatsmeow_call_relay_session_probe.go`: adds `USERNAME` variants derived from `relay.uuid`, `relay.self_pid`, `relay.peer_pid` and combinations like `self:peer`, `uuid:self`, etc.
  - These variants are applied to `relay.key`/`relay.hbh_key` candidates (and `lt0`) to match TURN relays that select the HMAC secret based on participant IDs.
- Implemented (2026-02-15): TURN MI-SHA256 probing
  - Added optional `MESSAGE-INTEGRITY-SHA256` attempts (attr `0x001C`) gated by `QP_CALL_RELAY_TURN_TRY_MI_SHA256=1`.
  - Updated candidate labels to include non-sensitive `USERNAME` variant tags (`u=uuid/self/peer/...`) to reduce duplicate-looking logs.
- Observed (2026-02-15): MI-SHA256 not accepted by relay
  - Relay returned `456 "Failed to decode allocate request stun message"` when `MESSAGE-INTEGRITY-SHA256` was used.
  - Probe now auto-disables MI-SHA256 on `456/420` or decode/unknown errors and falls back to standard MI-SHA1.
- Observed (2026-02-15): disabling TURN FINGERPRINT did not help
  - With `QP_CALL_RELAY_TURN_INCLUDE_FINGERPRINT=0`, Allocate still returns `451 Hmac missing` and all integrity attempts still return `450 Hmac mismatch`.

- Implemented (2026-02-16): observe-only mode for multi-device experiments (non-interfering)
  - `QP_CALL_OBSERVE_ONLY=1` makes the bot dump/log call telemetry but skip sending WhatsApp call signaling (no ACCEPT/PREACCEPT/TRANSPORT), and also skips RelayLatency echo and relay probes.
  - Guarded in `HandleCallOffer`, `HandleCallTransport`, `HandleCallRelayLatency`, BasicCallMeta path, and `WhatsmeowCallManager` flow entrypoints.

- Implemented (2026-02-16): TURN long-term auth probe hardening
  - Fixed bug where long-term candidates computed with REALM/NONCE did not include `NONCE` in the actual Allocate request (making long-term attempts invalid).
  - When the relay reveals REALM/NONCE only after an integrity attempt, the probe now dynamically appends and retries long-term candidates immediately.
- If `CallTransport` stops arriving from the peer after PREACCEPT/TRANSPORT changes, test `QP_CALL_ACCEPT_MODE=minimal` to send a minimal `<accept>` (legacy behavior) to try to trigger remote transport delivery again.
- If `QP_CALL_ACCEPT_MODE=minimal` triggers remote `CallTransport`, try a hybrid flow: minimal `<accept>` first (to trigger transport) and then send `PREACCEPT` (preaccept-only) to better match WA-JS signaling.
- Decision (2026-02-17): stay on WhatsApp consumer protocol via `whatsmeow` (no Cloud API migration).
- Decision (2026-02-17): physical gateway / device audio bridge approach is out of scope (needs scalable server-side handling: multiple WA accounts + concurrent calls).
- If publishing, push `calls` branch to remote and open PR if desired.

## CRITICAL ISSUE IDENTIFIED (2026-02-16)
**Problem**: WhatsApp is NOT stopping ringing on other devices when we send PREACCEPT/ACCEPT.
- **Expected behavior**: When we send PREACCEPT, WhatsApp should stop ringing on all other devices and route the call only to our device.
- **Actual behavior**: Call continues ringing on ALL devices even after we send PREACCEPT and ACCEPT messages.
- **Impact**: This is the fundamental first step - without this working, we cannot proceed with media bridging.

### Investigation Findings (2026-02-16)
1. **Official WhatsApp Documentation**:
   - Meta's official docs mention "Ligações" (Calls) as a feature in the Cloud API
   - Quote: "API de Nuvem do WhatsApp permite que você envie mensagens e faça ligações de forma programática no WhatsApp"
   - Specific calls documentation page was unavailable (technical error)
   - Action: Need to check https://developers.facebook.com/docs/whatsapp/cloud-api/calls or find working documentation URL

2. **Historical Code Analysis**:
   - Git history shows PREACCEPT code existed before Feb 13, 2026
   - Commit 657c979 (2026-02-14) only added `context.Background()` to API calls - no logic changes
   - State transition logic (`"chamando" → "conectando"`) is preserved and correct in current code
   - Files checked: `whatsmeow_call_manager.go`, `whatsmeow_call_manager_accept.go`, `whatsmeow_handlers+calls.go`

3. **Testing Results**:
   - Multiple test calls performed with different configurations:
     * Legacy mode (minimal PREACCEPT→ACCEPT)
     * Handshake complete mode (PREACCEPT→TRANSPORT→ACCEPT)
     * Minimal mode
   - ALL tests show same result: calls continue ringing on all devices
   - CallOffer arrives correctly (dumps created)
   - PREACCEPT and ACCEPT sent correctly (with ToNonAD() fixes applied)
   - NO CallTransport received from WhatsApp server (zero `call_transport_*.json` dumps)

4. **Code Fixes Applied**:
   - Applied ToNonAD() conversion to 4 locations in `AcceptCallLegacy`:
     * preacceptNode `call-creator`: `from` → `from.ToNonAD()`
     * preacceptCallAttrs `to`: `from` → `from.ToNonAD()`
     * acceptNode `call-creator`: `from` → `from.ToNonAD()`
     * acceptCallAttrs `to`: `from` → `from.ToNonAD()`
   - Rationale: convert LID format (`174560875933824@lid`) to phone number format (`557138388109@s.whatsapp.net`)
   - Status: Applied and tested, but WhatsApp still not responding

### Root Cause Hypotheses
1. **WhatsApp Protocol Change**: Server-side protocol may have changed requirements for PREACCEPT acceptance
2. **Session/Authentication Issue**: Current session may need reconnection or re-authentication
3. **Missing Required Field**: PREACCEPT may require additional fields we're not sending
4. **Message Format Issue**: Binary node structure may not match current WhatsApp expectations
5. **Offer-derived key missing**: relay/TURN short-term integrity secret is likely derived from `CallOffer` (e.g., `enc`/pkmsg + relay tokens/auth tokens/uuid), not obtained via TURN long-term challenge (REALM/NONCE).
6. **Official Cloud API calling**: separate product/protocol; not applicable to current `whatsmeow` consumer call flow (out of scope for this branch).

### Next Actions (Priority Order)
1. **Capture + analyze `enc` in CallOffer dumps**: treat `<enc>` (pkmsg/msg) as the prime suspect for relay/TURN short-term key material.
2. **~~Compare Protocol~~**: Attempted - removed `from` attr with no effect
3. **Test Session Reconnection**: ⚠️ **URGENT** - Force WhatsApp disconnect/reconnect to get fresh session
4. **Verify Protocol Breaking Change**: Check if WhatsApp completely changed call acceptance protocol in recent updates
5. **Test with Different Account**: Try with a fresh WhatsApp account to rule out account-specific issues
6. **Capture Network Traffic**: Use Wireshark/tcpdump to see actual server responses to PREACCEPT
7. **Check Capabilities**: Verify if we need to send additional capability flags for call handling

### Critical Questions to Answer
- ❓ Did WhatsApp deprecate PREACCEPT/ACCEPT protocol entirely?
- ❓ Is there a new authentication/token required for call acceptance?
- ❓ Does call acceptance require specific WhatsApp account permissions/flags?
- ❓ Is the current session too old/invalid for call handling?

### Test CallIDs from Today (2026-02-16)
- Multiple calls tested, all showing same pattern:
  * CallOffer received: ✓
  * PREACCEPT sent: ✓
  * ACCEPT sent: ✓
  * CallTransport received: ✗
  * Ringing stops on other devices: ✗

### NEW TEST - QP_CALL_LEGACY_WAJS_ATTRS Flag (2026-02-16)
**Discovery**: Found configuration mismatch between local and server environments
- **Server** (`src/.env.apoint-voip`): Has `QP_CALL_LEGACY_WAJS_ATTRS=1`
- **Local** (`src/.env`): Did NOT have this flag

**What this flag does**:
- `QP_CALL_LEGACY_WAJS_ATTRS=1` removes the `from` attribute from the `<call>` node
- Simulates older WA-JS implementations that only send `{to, id}` at the top-level

**Binary Node Difference**:
```xml
<!-- WITHOUT flag (before): -->
<call to="557138388109@s.whatsapp.net" from="YOUR_ID@s.whatsapp.net" id="...">
  <preaccept call-id="..." call-creator="557138388109@s.whatsapp.net"/>
</call>

<!-- WITH flag (after): -->
<call to="557138388109@s.whatsapp.net" id="...">
  <preaccept call-id="..." call-creator="557138388109@s.whatsapp.net"/>
</call>
```

**Action taken**:
1. ✅ Added `QP_CALL_LEGACY_WAJS_ATTRS=1` to `src/.env`
2. ✅ Recompiled `debug.exe`
3. ✅ **USER TESTED**: Made test call with flag enabled
4. ❌ **RESULT**: Call STILL rings on all devices - NO CHANGE

**Conclusion**: Removing `from` attribute from `<call>` node did NOT solve the problem. WhatsApp continues ignoring our PREACCEPT/ACCEPT messages regardless of this structural change.

**Latest Test CallID**: `A52D10D9CC77EC80134098A84CA111BF` (timestamp: 1771248647772148800)
- CallOffer received: ✓ (dump created)
- PREACCEPT sent without `from` attr: ✓
- ACCEPT sent without `from` attr: ✓
- CallTransport received: ✗ (still zero transport dumps)
- Ringing stops: ✗ (still rings on ALL devices)

**Hypothesis REJECTED**: The `from` attribute is not the blocking issue.

### **CRITICAL TEST - DIRECT MODE (ACCEPT Only, No PREACCEPT) - 2026-02-16**

**Based on user memory**: "No passado funcionava enviando só o ACCEPT, não precisava PREACCEPT"

**Test Configuration**:
- Set `QP_CALL_ACCEPT_MODE=direct` (sends ONLY ACCEPT, skips PREACCEPT entirely)
- Method used: `AcceptDirectCall()` - legacy code that worked in the past

**Test CallID**: `A553A6991851FEDA8E4B762DA0329C60`

**Timeline**:
```
10:37:58 - CallOffer received ✓
10:37:58 - [DIRECT-ACCEPT] Sending ACCEPT directly (NO PREACCEPT) ✓
10:37:58 - ACCEPT sent successfully with candidates ✓
10:37:58 - RelayLatency received (3 relay endpoints) ✓
10:38:05 - CallTerminate (call timed out after 7 seconds) ✗
```

**Result**: ❌ **FAILED - Same behavior as before**
- ACCEPT sent successfully
- NO CallTransport received from WhatsApp
- Call continues ringing on ALL devices
- Call terminates after timeout (~7 seconds)

**Hypothesis REJECTED**: Sending only ACCEPT (without PREACCEPT) does NOT solve the problem.

# BREAKTHROUGH (2026-02-16)
**Fixed the core UX requirement**: WhatsApp now stops ringing on other devices and flips the call UI state from **"calling"** to **"connecting"** automatically.

## What worked
- Using the exact **snippet ACCEPT** structure (older blog snippet) on incoming `CallOffer`.
- Observed behavior:
  - Ringing stops on other devices.
  - The call routes to this device only.
  - UI changes from **"calling"** → **"connecting"**.

## Exact signaling shape (snippet)
Top-level `<call>` attrs contain only `{to, id}` (no `from`). Inside it, `<accept>` contains `{call-id, call-creator}` and:
- `audio` (opus@16000)
- `audio` (opus@8000)
- `net` with `medium="3"` (snippet form)
- `encopt keygen="2"`

## Runtime config used
- `QP_CALL_ACCEPT_MODE=direct`
- `QP_CALL_USE_SNIPPET_ACCEPT=1`
- Dumps enabled to capture full in/out call signaling:
  - `QP_CALL_DUMP_OFFER=1`
  - `QP_CALL_DUMP_ACCEPT=1`
  - `QP_CALL_DUMP_TRANSPORT=1`
  - `QP_CALL_DUMP_DIR=../.dist/call_dumps`

## Why this matters
This solves the **first mandatory step** before any media/SRTP/relay bridging work: we can reliably take ownership of the incoming call at the signaling/UI level.

## Next diagnostic step
Now that the peer is expected to send `CallTransport` after our snippet ACCEPT, capture:
- `call_transport_received_*.json`
and implement the follow-up response that contains where to send audio (likely relay/SRTP/ICE/transport details).

# Immutable Constraints Discovered During Execution
- WhatsApp call handling is sensitive to upstream protocol/API changes; upgrades commonly require refactors in internal send/receive primitives.
- This repository treats LIDs as opaque identifiers; never derive phone numbers from `@lid`.
- Meta/WhatsApp docs pages for webhook reference `calls` were intermittently unavailable via fetch in this environment; validate call webhook details in a real browser session.
- ICE candidate correctness matters even before media bridging: sending `localIP` paired with a STUN-mapped `publicPort` is internally inconsistent and can keep the WhatsApp UI stuck in "connecting".
- Sending a minimal `<accept>` on `CallOffer` (without PREACCEPT/TRANSPORT negotiation and candidates) may not advance the WhatsApp call state; a WA-JS-like handshake flow is required for better state progression.
- The `net medium` used in our `<accept>` should match the peer's `CallTransport` `<net medium='...'>` (e.g., `2` for relay-only calls). A mismatch can keep the UI in "connecting" even when signaling messages are sent.
- **CRITICAL (2026-02-16)**: The fundamental PREACCEPT function (stopping ring on other devices) is NOT working. This is the first mandatory step before any media bridging work. If PREACCEPT doesn't stop the ring on other devices and route the call to our device, all subsequent work (media plane, RTP, SRTP) is blocked. WhatsApp server is not responding with CallTransport events and continues ringing all devices despite our PREACCEPT/ACCEPT messages.
