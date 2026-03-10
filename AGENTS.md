# Task Objective
- Stabilize `calls` branch for non-official WhatsApp voice call receive flow with `whatsmeow`.

# Ground Truth
- Incoming call signaling is now consistent with `@lid` end-to-end for `snippet ACCEPT`, `ACCEPT`, and `TRANSPORT`.
- Current call flow reaches `connecting`, receives remote `CallTransport`, and replies with local `TRANSPORT`.
- The remaining blocker is relay/TURN auth for media-plane establishment.

# New Findings (2026-03-09)
- Real WhatsApp Desktop TURN capture still shows:
  - `Allocate Request (0x0003)`
  - attrs `0x4000`, `0x4024`, `0x0016`
  - `MESSAGE-INTEGRITY`
  - no `USERNAME`, `REALM`, or `NONCE`
- The `CallOffer` `<enc type="pkmsg">` can be decrypted using the normal Signal pipeline already present in `whatsmeow`.
- Successful local decrypt was observed for:
  - `CallID=AC1EFD39CD6D3A12D5505328A46D7397`
  - via `189335563419811@lid`
  - plaintext length `79`
- This confirms the useful secret/material is not the raw `pkmsg` blob alone; the decrypted plaintext must be treated as a first-class TURN probe seed.

# Current Debug Direction
- Prioritize `enc.plain(...)` candidates above `enc.pkmsg(...)` in TURN probe sorting.
- Dump decrypted `CallOffer` plaintext locally for inspection:
  - `call_offer_enc_plain_*.json`
- Compare the `79`-byte plaintext against:
  - Desktop-like attrs `0x4000`, `0x4024`, `0x0016`
  - relay metadata `uuid`, `self_pid`, `peer_pid`
  - TURN `alloc.preimage`

# New Findings (2026-03-09, later)
- Parsed structure of the decrypted `79`-byte `enc.plain` now consistently shows:
  - `proto.f10` len `34`
  - `proto.f10.f1` len `32`
  - `proto.f35` len `40`
  - `proto.f35.f1` len `36`
  - `proto.f35.f2.varint = 2`
- `proto.f10.f1` is the strongest local candidate discovered so far, but current MI heuristics over it are still failing.
- Negative results already confirmed with fresh call tests:
  - raw `proto.f10.f1` + relay context still `450 Hmac mismatch`
  - `ECDH(identity, proto.f10.f1)` + relay context still `450`
    - example `CallID=AC061B47E3FB060A1E3225F64E7B6519`
  - `ECDH(signedprekey, proto.f10.f1)` + relay context still `450`
    - example `CallID=AC9A299D2F925410A0F07C154693964F`
- Practical boundary:
  - the current `HMAC/SHA1` candidate family over `enc.plain`, `proto.f10.f1`, and simple ECDH-derived secrets has low remaining value
  - do not keep spending calls on more blind MI permutations without a new semantic hypothesis

# Recommended Next Direction
- Stop expanding TURN MI brute-force permutations for now.
- Treat `proto.f10` / `proto.f35` as structured payloads that need semantic interpretation.
- Determine whether `proto.f10.f1` is:
  - actual key material
  - a remote ephemeral pubkey
  - a wrapped token/envelope
  - or only one field of a larger KDF input
- Only resume TURN candidate experiments after a new semantic hypothesis exists.

# Expected Signals In Next Test
- Log line:
  - `Offer enc decrypted: ... plain_len=...`
- TURN probe first candidates should start with:
  - `enc.plain...`
- Progress criteria:
  - any change from `450 Integrity failure: Hmac mismatch`
  - or first `Allocate Success (0x0103)`

# Notes
- Keep `AGENTS.md` local/custom only.
- File remains ignored by Git.

# New Findings (2026-03-09, Desktop instrumentation round)
- Relevant Desktop process:
  - the most useful child process observed for call/network work is a `msedgewebview2.exe` instance (example PID `21328`), not only `WhatsApp.Root.exe`
- Frida/process mapping:
  - direct attach to renderer PIDs is unreliable because many WebView2 children terminate during injection
  - a polling loader (`wa_frida_loader.py`) successfully attaches tracing scripts to surviving `msedgewebview2.exe` and `WhatsApp.Root.exe` processes
- Winsock tracing:
  - `msedgewebview2.exe:21328` opens UDP/IPv6 sockets via `WSASocketW` with `af=23`, `type=2`, `proto=17`
  - `WhatsApp.Root.exe` also opens sockets, but observed patterns look more like bootstrap/control
  - hooks on `sendto/send/WSASend*/recv*/connect/WSAConnect/WSAIoctl/ConnectEx` provided only limited visibility and mostly emitted events near call teardown
  - `GetAddrInfoW`, `GetAddrInfoExW`, and `WSAConnectByNameW` produced no useful call-path output
- Crypto hooks:
  - `bcrypt.dll` is loaded in the relevant process
  - hooks on `BCryptHashData`, `BCryptFinishHash`, `BCryptOpenAlgorithmProvider`, `BCryptCreateHash`, and `BCryptGenerateSymmetricKey` produced no useful call-related output during the tested call
  - current inference: the relevant call/media-plane crypto is not going through the simple `bcrypt` path we attempted to intercept
- pktmon round (2026-03-09):
  - fresh capture from `C:\Windows\System32\PktMon.etl` showed only TCP/TLS traffic on port `443` during the tested call window
  - no explicit `3478` TURN/STUN traffic was observed in this capture
  - repeated example flow:
    - `192.168.31.202:55013 -> 140.82.114.21:443`
    - additional `443` flow also seen to `20.42.65.84:443`
- Practical conclusion:
  - Desktop instrumentation confirmed the likely network/media child process and that UDP/IPv6 sockets exist
  - however, this round did not expose the TURN Allocate/auth path directly
  - generic Winsock/`bcrypt` hooks now have clearly diminishing returns
  - if Desktop-side work continues, the next step should be heavier Chromium/WebRTC internal tracing or ETW, not more blind socket hooks
