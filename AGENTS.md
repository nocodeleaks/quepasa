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

# New Findings (2026-03-09, `proto.f10.f1 + te2` closure)
- TURN probe now exercised the `miKDF` family over `enc.plain.proto.f10.f1 + te2` with full budget coverage (`maxTry=120`, `mikdf=120` attempts executed in one call).
- Confirmed tested direct combinations over `alloc.preimage`:
  - `proto.f10.f1 + te2.prefix8+tail4`
  - `proto.f10.f1 + te2.tail4`
  - `proto.f10.f1 + te2.prefix8`
  - `proto.f10.f1 + te2.full18`
- Outcome unchanged across all tested combinations:
  - `450 Integrity failure: Hmac mismatch`
- Practical boundary:
  - current `HMAC/SHA1` heuristics over `proto.f10.f1 + te2` are exhausted enough to stop treating ranking/budget as the blocker
  - this line should be considered closed unless a new nontrivial KDF hypothesis appears

# New Findings (2026-03-09, relay token/auth shape)
- `token` and `auth_token` are stored once under the `relay` node; `te2` maps each relay to `token_id` / `auth_token_id`.
- Observed shape:
  - `token len=182`, binary prefix starts with `090f01...`
  - `auth_token len=70`, binary prefix starts with `0903...`
- Behavioral pattern across fresh calls:
  - `token` is relay-specific and changes per call
  - `auth_token` changes per call too, but can be shared by multiple relays in the same offer (for example `gru*` and `poa1c01`)
- Next active hypothesis:
  - test `proto.f10.f1 + authRaw/tokRaw` directly as `miKDF` seeds, instead of only `proto.f10.f1 + te2`

# New Findings (2026-03-10, token/auth envelope correlation)
- Current local sample:
  - `call_offer_received_20260310101919_ACE7D3E205963F5E2B0AE120BA359554.json`
- Relay-to-token/auth mapping in that offer:
  - `fcfc2c01 -> token_id=0, auth_id=0`
  - `gru1c02 -> token_id=2, auth_id=1`
  - `poa1c01 -> token_id=1, auth_id=1`
- So `auth_token id=1` is shared by `gru1c02` and `poa1c01` in the same call, while `token_id` remains relay-specific.
- Block-level correlation over trimmed envelopes:
  - `authTrim len=68` => `headLen=4` + `4 x 16-byte blocks`
  - `tokTrim len=179` => `headLen=3` + `11 x 16-byte blocks`
- Negative result:
  - no literal duplicate 16-byte block values were found across `auth_token` and `token` bodies in that offer
  - this weakens the hypothesis that the missing secret is exposed as a directly reused AES-sized block between shared `auth_token` and relay-specific `token`
- Practical boundary:
  - `auth_token` sharing across relays is real and still important
  - but first/last block reuse is not visible as a trivial literal block match in the current sample

# New Findings (2026-03-10, shared-auth token delta)
- In `call_offer_received_20260310101919_ACE7D3E205963F5E2B0AE120BA359554.json`:
  - `gru1c02` and `poa1c01` share `auth_id=1`, but use `token_id=2` and `token_id=1` respectively
- Negative result:
  - XOR/delta between the two relay-specific `token` envelopes does not show a simple low-entropy relation to the corresponding `te2` delta
  - example:
    - `token_head_xor = 001e6a`
    - `te2_xor_prefix8 = 00000000011100f7`
    - `te2_xor_tail4 = 000001c6`
  - no obvious direct mapping was observed between the token delta and relay-specific `te2` delta
- Practical boundary:
  - `auth_token` sharing is still semantically important
  - but the relay-specific `token` difference is not explained by a trivial XOR/update against `te2`

# New Findings (2026-03-10, `rte` structure)
- The top-level `rte` node in the offer is another 18-byte binary payload and should not be treated as opaque noise.
- Current sample:
  - raw hex: `280403b0950b6b0088bbb4ee024cd378c750`
- Observed layout:
  - `8 bytes` IPv6-like prefix: `2804:03b0:950b:6b00`
  - `4 bytes` middle field: `88bbb4ee`
  - `4 bytes` tail field: `024cd378`
  - `2 bytes` suffix: `c750`
- Practical note:
  - `rte` has the same length as `te2 len=18`, but a different internal pattern
  - this makes `rte` a separate structural artifact worth preserving in dumps and future correlation, not just another alias of `te2`

- Verified (2026-03-10): `proto.f10.f1 + auth/token` and `auth/token` raw/trim heuristics are exhausted
  - CallID `ACA90F0ECC895F6CDD85ED5E17EBB4D8`: top probe seeds were `enc.plain.proto.f10.f1+authRaw(...)`, then `...+tokRaw(...)`; all stayed `450`.
  - CallID `ACC3F5E43620DC6ADFC718E498E18251`: top probe seeds were `authTrim(...)`, `authTrim+tokTrim`, `tokTrim+authTrim`; all stayed `450`.
  - CallID `AC538643E9D64A52C5C01D7F02895343`: top probe used `authTrim(seed) + tokTrim(msg)` directly (`miKDF:msg=tokTrim(...)`); all stayed `450`.
  - Base/discovery Allocate remains `451 Integrity failure: Hmac missing`, with no `REALM/NONCE`.
  - Structural notes:
    - `token` is `182` bytes and consistently starts with `090f01...`; after trimming that fixed 3-byte header, body len is `179`.
    - `auth_token` is `70` bytes and consistently starts with `0903...`; after trimming the fixed 2-byte header, body len is `68`.
    - `auth_token` continues to be shared by `gru*` and `poa1c01` within the same offer, while `fcfc2c01` keeps its own auth token.
    - `authTrim` length `68` splits cleanly as `4 + 32 + 32`; parser now exposes this summary shape for future work (`Head4Hex`, `Left32Hex`, `Right32Hex`).

# New Findings (2026-03-10, auth HKDF and F1+afterF1 closure)
- Fresh call `AC0188B84E3B385CFC6A2A5C21DF2988` exercised HKDF-SHA256-derived seeds over structured `authTrim`:
  - `authHKDF(ikm=authTrimLeft32, salt=authTrimRight32, info=authTrimHead4)`
  - `authHKDF(ikm=authTrimRight32, salt=authTrimLeft32, info=authTrimHead4)`
  - `authHKDF(... salt=enc.plain.proto.f10.f1, info=authTrimHead4)`
- Outcome remained unchanged:
  - all top attempts returned `450 Integrity failure: Hmac mismatch`
- Fresh call `AC23C80BED1001A063068E26A4DE5363` exercised a more semantic split of raw relay token/auth blobs:
  - `authRawAfterF1(id=...)(miKDF:msg=authRawF1Fixed64(id=...))`
  - `tokRawAfterF1(id=...)(miKDF:msg=tokRawF1Fixed64(id=...))`
  - `enc.plain.proto.f10.f1+authRawF1Fixed64(id=...)`
  - `enc.plain.proto.f10.f1+tokRawF1Fixed64(id=...)`
- Outcome also unchanged:
  - `NON450 = 0`
  - the `fixed64 + afterF1` line should be considered closed under the current short-HMAC/SHA1 model.
- Practical conclusion:
  - `proto.f10.f1`
  - `te2`
  - `auth/token`
  - `authTrim(4+32+32)`
  - `authHKDF(...)`
  - `raw field1 fixed64 + afterF1`
  have all been exercised enough to stop treating ranking/budget as the blocker.
- Next useful work should be semantic/protocol interpretation of `token/auth_token` as opaque envelopes, not more small blob permutations.

# New Findings (2026-03-10, token/auth envelope shape)
- Current `RelayToken` parsing now treats trimmed token/auth payloads as opaque envelopes with:
  - `EnvelopeHeadLen = TrimLen % 16`
  - `EnvelopeHeadHex`
  - `Block16Count`
  - `Block16Hex[]`
- This matches the observed stable sizes much better than the previous ad hoc cuts:
  - `authTrim len=68` => `headLen=4` + `4 x 16-byte blocks`
  - `tokTrim len=179` => `headLen=3` + `11 x 16-byte blocks`
- Practical implication:
  - the useful next semantic hypothesis is block-oriented envelope interpretation, not more blind HMAC permutations over whole trimmed blobs.

# New Findings (2026-03-10, Desktop page attach `WA-MON v7`)
- The most useful Desktop monitor remains direct CDP attach to the `page:WhatsApp`, via `wa_cdp_page_attach.py`.
- Page-level `crypto.subtle` instrumentation (`encrypt/decrypt/importKey`) produced no observable call-related events in the tested Desktop-only calls.
- Current page-console capture is therefore still dominated by `WebSocket.send` binary frames.
- Stable framing remains:
  - first `3` bytes are big-endian length
  - `len_hdr = frame_len - 3`
- New outbound frame families observed cleanly in the page log:
  - `len=42`, prefix `000027`
  - `len=45`, prefix `00002a`
  - `len=46`, prefix `00002b`
  - `len=47`, prefix `00002c`
  - `len=65`, prefix `00003e`
  - `len=69`, prefix `000042`
  - `len=70`, prefix `000043`
  - `len=102`, prefix `000063`
  - `len=408`, prefix `000195`
  - `len=430`, prefix `0001ab`
  - `len=659`, prefix `000290`
  - `len=2697`, prefix `000a86`
- Practical interpretation from the current log:
  - `42/45/46` frames recur frequently and look more like keepalive/control/ack traffic than call-specific setup
  - `65/69/70/102` frames are rarer and more likely to belong to higher-level control or call-related application state
  - `408/430/659/2697` frames look like opaque app-sync/state blobs, not plain TURN/STUN or raw WebRTC SDP
- Current boundary:
  - page-level WebSocket framing is useful for timing and family classification
  - but it is still too opaque to expose the call secret/KDF path directly
  - if Desktop instrumentation continues, the next useful step is to classify these families by timing/motif, not to go back to generic Winsock hooks

# New Findings (2026-03-10, Desktop page attach rare-frame motifs)
- Rare page-WebSocket outbound families now have clearer local context in `wa_cdp_page_console.log`:
  - `69 / 000042`
    - observed next to `model-storage/message-info/*` updates and generic `wawc/logs`
    - currently looks more like message/state bookkeeping than direct media setup
  - `659 / 000290` + `70 / 000043`
    - observed next to `model-storage/participant/*`, `signal-storage/session-store/*`, then `message/callOutcome`
    - likely application/session sync around call state, not raw TURN/STUN
  - `102 / 000063`
    - observed next to `message/callOutcome` plus `Worker.postMessage {"command":{"operation":"consume"}}`
    - suggests post-call or state-consume path, not media-plane open
  - `430 / 0001ab` + `2697 / 000a86`
    - observed next to `wawc/wam_meta/seqNum` and `wawc/ps_tokens`
    - these are strong candidates for telemetry/state upload, not direct call signaling
- Practical conclusion:
  - page-level WebSocket remains useful for phase classification
  - but the currently identified rare families still cluster around storage/state/reporting paths
  - this lowers the value of further classifying page-WebSocket frames alone unless a family appears outside those motifs

# New Findings (2026-03-10, Desktop page WebSocket phase split)
- Direct page attach (`wa_cdp_page_attach.py`) is now the most useful Desktop-side monitor.
- The page-level binary WebSocket framing is confirmed as:
  - `3`-byte big-endian length prefix
  - for the captured families, `len_hdr = frame_len - 3`
- Stable generic family:
  - repeated `send len=41 prefix=000026` followed by `recv len=47 prefix=00002c`
  - this pair repeats continuously and currently looks more like generic session/ack traffic than call-specific control.
- More call-specific short motifs were captured with full inbound hex:
  - `send(60,000039)` -> `recv(52,000031)` -> `recv(52,000031)` -> `recv(110,00006b)` -> another `send(60,000039)` -> `recv(109,00006a)`
  - `send(37,000022)` -> `recv(39,000024)` -> `recv(51,000030)` -> `recv(41,000026)` -> `recv(138,000087)` -> `send(59,000038)`
  - `send(59,000038)` -> `recv(168,0000a5)`
- Phase interpretation from the same page-log window:
  - the `60/59/37` families occur adjacent to `BroadcastChannel.postMessage` events for:
    - `idb://model-storage/message/callOutcome`
    - `idb://wawc/wam_meta/seqNum`
    - `idb://wawc/ps_tokens/`
  - practical reading: these motifs are more likely post-call/reporting/persistence traffic than the earlier media-setup trigger.
- Practical next direction:
  - stop treating the `41 -> 47` pair as the primary call clue
  - treat `60/59/37` as post-call/reporting motifs
  - focus future Desktop-side decoding on earlier/larger setup bursts (for example `724`, `94`, `100`, `400`) if correlating with `offer/accept/transport`

# New Findings (2026-03-10, Desktop large WebSocket bursts)
- Large outbound page-WebSocket bursts are now isolated with `helpers/summarize-wa-cdp-setup-bursts.ps1`.
- Observed examples:
  - `send len=634 prefix=000277`
  - `send len=446 prefix=0001bb`
  - `send len=514 prefix=0001ff`
  - `send len=431 prefix=0001ac`
  - `send len=459 prefix=0001c8`
  - `send len=2683 prefix=000a78`
- Current contextual reading from the page log:
  - `634` sits next to `idb://wawc/wam/buffer` and participant/device-list updates
  - `446/514/431/459/2683` sit next to `idb://wawc/wam_meta/seqNum`, `idb://wawc/ps_tokens/`, `idb://model-storage/user-prefs/WARoutingInfo`, and device/participant updates
  - practical reading: these larger bursts still look like higher-level app sync/report/state-transfer blobs, not exposed TURN/STUN/plain WebRTC signaling.
- Practical implication:
  - page-level `WebSocket` capture is useful for phase correlation, but payload bodies still look opaque/encrypted at this layer
  - if Desktop-side reverse engineering continues, the next meaningful gain is likely before this page-level framing (worker/internal serializer) or by correlating exact timing with local `quepasa` `offer/accept/transport`

# New Findings (2026-03-10, Desktop page WebSocket pair family)
- Direct page attach through `wa_cdp_page_attach.py` is now the most productive Desktop-side monitor.
- In `wa_cdp_page_console.log`, a stable adjacent pair family is now confirmed:
  - outbound `send len=41 prefix=000026`
  - immediately followed by inbound `recv len=47 prefix=00002c`
- This pair repeated multiple times in the same Desktop call capture with small line gaps (`2-3` lines).
- Example pair:
  - send: `000026d72172e3021d7a8809ee1f0943fd770ff605748f6cdac4d9e42dd6c2aa06cf947c76ac5b7800`
  - recv: `00002c35b66e4674412edf85904075c8b27c7a3a729ec62674c04e001dd0415303e3b0b018d3c485bfd2ac96bbce17`
- Family-level byte analysis:
  - `send41` has only bytes `0..2` stable across captures (`00 00 26`)
  - `recv47` has only bytes `0..2` stable across captures (`00 00 2c`)
  - when paired, byte `2` differs by a constant XOR `0x0a` (`0x26 -> 0x2c`)
  - beyond the 3-byte framing header, payload bytes are effectively fully variable in current samples
- `wa_cdp_page_attach.py` now decodes binary CDP WebSocket frames more completely:
  - `prefix3_hex`
  - `len_hdr`
  - `len_hdr_matches`
  - `bin_full_hex` for frames `<= 256`
  - `bin_tail_hex` for larger frames
- Fresh inbound `bin_full_hex` now confirms additional short call motifs beyond the `41 -> 47` pair:
  - `send(37,000022)` followed by:
    - `recv(39,000024)`
    - `recv(51,000030)`
    - `recv(41,000026)`
    - `recv(138,000087)`
    - then `send(59,000038)`
  - `send(60,000039)` followed by:
    - `recv(52,000031)`
    - `recv(52,000031)`
    - `recv(110,00006b)`
    - then another `send(60,000039)` and `recv(109,00006a)`
  - `send(59,000038)` can also be followed by `recv(168,0000a5)`
- Local correlation from the same page log strongly suggests:
  - `41 -> 47` is likely generic session/ack traffic, because it repeats continuously across the capture
  - the `37/59/60` families are more call-specific, because they appear adjacent to:
    - `model-storage/message ... callOutcome`
    - `false_<jid>_<callid>` message persistence
    - `wam_meta`
    - `ps_tokens`
  - practical interpretation: `37/59/60` are currently the better candidates for post-call/reporting/control transitions than `41 -> 47`
- Practical implication:
  - the Desktop call path is clearly visible in the page WebSocket binary framing
  - the next useful Desktop-side work is family-level decoding/correlation of these `41 -> 47`, `37 -> 39/51/41/138`, and `59/60 -> 52/109/110/168` motifs
  - a second motif worth tracking is the short setup burst around `send(634,000277)`, typically followed by `recv(39,000024)`, `recv(108,000069)`, then `send(46,00002b)`
  - this is now a better direction than more generic Winsock or TURN-only probing

# New Findings (2026-03-10, Desktop CDP direct page attach)
- Browser-level auto-attach for CDP remained unreliable for the heavy `page WhatsApp` target, even though `service_worker` attach worked.
- A dedicated direct page monitor (`wa_cdp_page_attach.py`) works better:
  - attaches directly to the stable `page` target from `/json/list`
  - enables `Runtime`, `Page`, `Network`
  - injects `wa_cdp_inject.js`
  - receives both JS console events and native `Network.webSocketFrameSent/Received`
- Fresh successful capture on Desktop-only call produced both outbound and inbound binary WebSocket frames from `page:WhatsApp`:
  - outbound example:
    - `len=41`
    - `full_hex=000026d72172e3021d7a8809ee1f0943fd770ff605748f6cdac4d9e42dd6c2aa06cf947c76ac5b7800`
  - inbound example:
    - `len=47`
    - `full_hex=00002c35b66e4674412edf85904075c8b27c7a3a729ec62674c04e001dd0415303e3b0b018d3c485bfd2ac96bbce17`
- This confirms the most productive Desktop-side observability path is:
  - `wa_cdp_page_attach.py`
  - not generic Winsock hooks
  - not browser-level auto-attach alone
- Practical implication:
  - next Desktop-side reverse engineering should focus on grouping and decoding these page WebSocket binary frame families (`41/47`, `37`, `46`, `724`, etc.).

# New Findings (2026-03-10, Desktop CDP cold-start capture)
- Cold-start CDP monitoring finally captured the main page `WebSocket` from the beginning of the session.
- Relevant artifacts:
  - log: `.dist/wa_cdp_console.log`
  - summarizer: `helpers/summarize-wa-cdp-websocket.ps1`
- Important confirmation:
  - the useful internal call/signaling behavior is visible on the page `WebSocket`, not only via generic Winsock tracing
  - `RTCPeerConnection` still did not provide useful lifecycle output in this round
- Framing rule is now confirmed with a large sample:
  - binary `WebSocket` frames use a 3-byte big-endian length prefix
  - for almost all captured frames: `len_hdr == frame_len - 3`
  - sample counts from the cold-start capture:
    - `frame_count = 170`
    - `len_hdr_matches_true = 169`
    - `len_hdr_matches_false = 1`
  - the only mismatch is the initial bootstrap frame starting with `454400...`
- High-signal frame families observed:
  - repeated outbound `len=110 / prefix3=00006b` (`x45`)
  - repeated inbound `len=248 / prefix3=0000f5` (`x45`)
  - also repeated:
    - inbound `len=39 / prefix3=000024`
    - outbound `len=47 / prefix3=00002c`
    - inbound/outbound `len=60 / prefix3=000039`
    - outbound `len=59 / prefix3=000038`
    - outbound `len=37 / prefix3=000022`
  - larger bursts seen during the same capture:
    - outbound `218`, `268`, `380`, `393`, `424`, `720`, `1280`, `2572`
    - inbound `478`, `524`, `724`, `900`, `5223`
- Practical conclusion:
  - Desktop-side work should now focus on this binary `WebSocket` framing and frame families
  - generic Winsock / `bcrypt` hooks have lower value than decoding or correlating these framed WebSocket payloads
  - next useful step is to correlate `seq/ts` from `.dist/wa_cdp_console.log` against local branch call logs to identify which frame families correspond to `offer / accept / transport / terminate`

# New Findings (2026-03-10, CDP/WebSocket internal monitor)
- CDP injection through WebView2 is now working reliably enough on the `page:WhatsApp` target.
- Current monitor confirms:
  - `RTCPeerConnection` exists in the page context (`hasRTCPeerConnection=true`)
  - `WebSocket` exists and is actively used for internal signaling
  - `WebSocket:new` observed URLs:
    - `wss://web.whatsapp.com/ws/chat?ED=CAUIEggC`
    - `wss://web.whatsapp.com:5222/ws/chat?ED=CAUIEggC`
- Fresh `WebSocket.send/message` binary frames are now visible from the page target.
- Strong framing pattern observed in current binary frames:
  - after the initial `send len=218` frame with head starting `4544...`, subsequent binary frames consistently start with a 3-byte big-endian length field
  - for those frames, `prefix3_hex` / `len_hdr` matches `frame_len - 3`
  - examples:
    - `send len=46` starts `00002b...` (`0x2b = 43 = 46 - 3`)
    - `send len=75` starts `000048...` (`0x48 = 72 = 75 - 3`)
    - `message len=57` starts `000036...` (`0x36 = 54 = 57 - 3`)
    - `message len=108` starts `000069...` (`0x69 = 105 = 108 - 3`)
- Practical conclusion:
  - page-level WebSocket instrumentation is now a productive observability path
  - the useful next work is frame-family analysis / decoding of these binary WebSocket payloads, not more blind Winsock hooks
  - `RTCPeerConnection:*` events still have not appeared in the current monitor output, so the most visible internal call behavior is still in the WebSocket signaling channel

# New Findings (2026-03-10, CDP/WebSocket framing v3)
- The `v3` CDP monitor is now active on the `page:WhatsApp` target and emits `seq`, `ts`, `prefix3_hex`, `len_hdr`, and `len_hdr_matches`.
- Practical note:
  - duplicate `WebSocket.send` lines may appear when `v2` and `v3` wrappers coexist on the same already-open page
  - for analysis, only use frames carrying `seq`/`ts` from `v3`
- In the fresh monitored call, all useful binary frames after bootstrap matched the same framing rule:
  - first 3 bytes are a big-endian length field
  - `len_hdr == frame_len - 3`
- Example fresh sequence from `v3` page log:
  - `seq=28 send len=59 prefix3=000038`
  - `seq=29 send len=401 prefix3=00018e`
  - `message len=39`
  - `seq=30 send len=2587 prefix3=000a18`
  - `message len=39`
  - `message len=108`
  - `seq=31 send len=46 prefix3=00002b`
  - `message len=108`
  - `seq=32 send len=46 prefix3=00002b`
  - `message len=201`
  - `seq=33 send len=46 prefix3=00002b`
  - `seq=34 send len=41 prefix3=000026`
  - `message len=47`
- Practical conclusion:
  - page-level binary WebSocket signaling is currently the highest-value Desktop observability path
  - the next useful work is grouping/decoding these frame families, not more generic Winsock hooks

# New Findings (2026-03-10, CDP v5 / page WebSocket burst signature)
- `WA-MON version=5` is now active on the `page:WhatsApp` target.
- Practical result:
  - `WebSocket.send` continues to emit fully structured records with `seq`, `ts`, `prefix3_hex`, `len_hdr`, and `len_hdr_matches`
  - `len_hdr_matches` remains consistently `true` for all current `send` frames
  - this confirms the binary signaling framing rule is stable: first 3 bytes are a big-endian payload length equal to `frame_len - 3`
- Current limitation:
  - `WebSocket.message` is visible in the page log, but still appears without `seq/ts` when the page WebSocket already existed before reinjection
  - inference: JS monkeypatches remain useful, but true full-frame `message` sequencing likely requires attaching before socket creation or using a lower-level browser event source
- Fresh call burst signature from the page log (current run) is now clear enough to track by frame families:
  - many repeated small outbound frames around `len=42 / prefix3=000027`
  - repeated outbound frames around `len=46 / prefix3=00002b`
  - medium outbound control bursts around:
    - `len=59 / prefix3=000038`
    - `len=60 / prefix3=000039`
    - `len=61 / prefix3=00003a`
    - `len=95 / prefix3=00005c`
  - larger outbound bursts around:
    - `len=382 / prefix3=00017b`
    - `len=402 / prefix3=00018f`
    - `len=2616 / prefix3=000a35`
  - inbound frames observed interleaved with these outbound bursts include sizes such as:
    - `40`, `48`, `51`, `52`, `53`, `65`, `72`, `73`, `94`, `106`, `108`, `110`, `127`, `138`, `168`, `169`, `200`, `222`, `235`, `248`, `5749`
- Practical conclusion:
  - the non-official call signaling behavior is definitely visible on the binary page WebSocket
  - the next Desktop-side improvement should be a cold-start CDP capture where the monitor is active before the main WebSocket is created; this is more valuable than more generic Winsock work

# New Findings (2026-03-10, Desktop JS transport stack)
- Direct page attach via `wa_cdp_page_attach.py` is the most useful Desktop instrumentation path so far.
- Strong stack-trace evidence from the page bundle now identifies the outbound WebSocket transport pipeline:
  - `WANoiseSocket.sendFrame(...)` encrypts payloads with `crypto.subtle.encrypt({ name: "AES-GCM", iv: ..., additionalData: ... }, key, payload)`
  - `WAFrameSocket.sendFrame(...)` prepends the 3-byte big-endian frame length and forwards to transport
  - `WASocketTransport.requestSend()` ultimately calls browser `WebSocket.send(...)`
- Practical implication:
  - the binary `WebSocket` frames observed in the page are already post-encryption Noise frames
  - continued classification of frame sizes/prefixes alone has limited value without intercepting pre-encryption plaintext or crypto parameters
- Desktop-side next direction should prefer:
  - `SubtleCrypto.prototype.encrypt/decrypt/importKey` instrumentation on the page
  - or a more direct hook into `WANoiseSocket` / `WAFrameSocket`
  - rather than more Winsock/TURN blind tracing

# New Findings (2026-03-10, Desktop NoiseSocket plaintext/ciphertext)
- Direct page attach with `wa_cdp_page_attach.py` is now capturing `crypto.subtle` activity in the live `page:WhatsApp` target.
- Confirmed path from live runtime:
  - outbound `WebSocket.send` frames come from `WAFrameSocket.sendFrame(...)` (`...js:674`)
  - those frames are produced by `WANoiseSocket.sendFrame(...)` (`...js:698`) after `AES-GCM` encryption
  - call-related app payloads are visible just before encryption in `crypto.direct.encrypt` logs
  - inbound frames are visible just before decryption in `crypto.direct.decrypt` logs
- Concrete live examples observed:
  - outbound plaintext len `27` before encrypt:
    - `00f8071b11faff88189451108126848f7608ff0524709731731509`
    - then `WebSocket.send len=46` / prefix `00002b`
  - outbound plaintext len `22` before encrypt:
    - `00f8091908ff0712472b55864a140429165711fa0003`
    - then `WebSocket.send len=41` / prefix `000026`
  - inbound ciphertext len `105` before decrypt:
    - associated with `webSocketFrameReceived bin_len=108` / prefix `000069`
  - inbound ciphertext len `44` before decrypt:
    - associated with `webSocketFrameReceived bin_len=47` / prefix `00002c`
- Practical implication:
  - the Desktop monitor is finally above the NoiseSocket encryption layer for inputs to `encrypt/decrypt`
  - next useful step is to capture decrypt/encrypt results as well, to obtain the post-decrypt plaintext and confirm exact mapping to frame families

# New Findings (2026-03-10, Desktop pre-/post-Noise plaintext)
- `wa_cdp_page_attach.py` with direct page attach is now capturing both inputs and resolved outputs of `crypto.subtle.encrypt/decrypt` in `page:WhatsApp`.
- This is the first confirmed visibility point above the `WANoiseSocket` AES-GCM layer.
- Concrete live examples:
  - outbound plaintext before encrypt (len `22`):
    - `00f8091908ff0712472b55864a170429165711fa0003`
    - corresponding `crypto.result.encrypt` output len `38`
    - corresponding `WebSocket.send` frame len `41` / prefix `000026`
  - inbound decrypt result (len `28`):
    - `00f8091906fa0003041408ff0712472b55864a171aff051773173626`
    - produced from ciphertext len `44`
    - corresponding received frame len `47` / prefix `00002c`
  - outbound plaintext before encrypt (len `27`):
    - `00f8071b11faff88189451108126848f7608ff0524709731731509`
    - corresponding `WebSocket.send` frame len `46` / prefix `00002b`
- Strong implication:
  - the Desktop monitor is now seeing structured application payloads immediately before/after NoiseSocket AES-GCM.
  - the next useful work is offline decoding/correlation of these plaintexts, not more blind socket/TURN tracing.

# New Findings (2026-03-10, Desktop call stanzas decoded from page crypto)
- Direct Desktop page-crypto decoding now exposes actual call stanzas in plaintext before/after the page `WANoiseSocket` encryption layer.
- Concrete decoded Desktop call flow for:
  - `call-id=ACC44F3C5E6D5BD6816A00A85F5A3741`
- Observed inbound/outbound stanzas:
  - `preaccept` ack inbound
  - `relaylatency` acks inbound
  - `relaylatency` payloads inbound
  - `relaylatency` acks outbound
  - `accept` ack inbound
  - `transport` payload inbound
  - `transport` ack outbound
  - `mute_v2` inbound + ack outbound
  - `accept` receipts inbound from both `62556332941345@lid` and `189335563419811@lid`
  - receipt acks outbound
  - `terminate` ack inbound
  - `duration` ack inbound
  - `terminate` payload inbound + ack outbound
- Critical structural decoding:
  - Desktop `relaylatency` payloads carry `te` values that decode directly as `IPv4(4) + port(2)`:
    - `aa96ec230d96` -> `170.150.236.35:3478`
    - `3990b3360d96` -> `57.144.179.54:3478`
    - `9df0de3e0d96` -> `157.240.222.62:3478`
  - Desktop `transport` payload also carries 6-byte `IPv4 + port` endpoints:
    - `<te priority="96">c0a81f2ce758</te>` -> `192.168.31.44:59224`
    - `<te priority="32">b124bc1de758</te>` -> `177.36.188.29:59224`
    - `<rte>b124bc1dbf48</rte>` -> `177.36.188.29:48968`
- Practical implication:
  - `transport.te` and `transport.rte` are not opaque call tokens; they are compact endpoint encodings.
  - This matches the local `quepasa` `transport_sent` shape:
    - host candidate `192.168.31.202:65196`
    - srflx candidate `177.36.188.29:65196`
  - It also aligns with local `CallRelayLatency`, where relay endpoints are already decoded as IPv4 `:3478`.
  - Therefore:
    - `relaylatency.te` = chosen relay IPv4 endpoint
    - `transport.te` = compact media candidates (`host` / `srflx`)
    - `transport.rte` = additional compact endpoint, likely relay/reflexive-related, but still endpoint-shaped rather than secret-shaped
- Practical boundary:
  - these Desktop plaintext stanzas do not expose a new obvious media-plane secret
  - they reinforce that the remaining blocker is relay/media-plane auth, not signaling shape
- Code support updated:
  - `call_relaylatency_*` dumps now preserve `compact_hex` for relay endpoints (`IPv4+port` encoded as 6 bytes)
  - `call_transport_received_*` dumps now expose `compact_items` when `<te>` / `<rte>` are present
  - `call_transport_sent_*` dumps now expose `compact_candidates` derived from local host/srflx ICE candidates for direct comparison with Desktop compact endpoint encodings
