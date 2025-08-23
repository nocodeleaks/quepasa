## Common Guidelines
* code comments should always be in English;
* response to user queries should be in IDE current language;
* avoid to change code that was not related to the query;
* when agent has to change a method and it change the async status, the agent should update the method callers too;
* for extensions methods use always "source" as default parameter name
* use one file for each class
* for #region tags: no blank lines between consecutive regions, but always add one blank line after region opening and one blank line before region closing
* do not try to build if you just changed the code comments or documentation files;
* **when making relevant code changes, always create or update internal documentation following the Internal Documentation Guidelines**;
* whenever creating an extension method, use 'source' as parameter name for the extended object;
* for class and structure names, e.g.: whatsmeow_group_manager.go => WhatsmeowGroupManager;

## Testing Guidelines
* **Follow official Go testing conventions** - use `*_test.go` files within the same package
* Test files should be named with `_test.go` suffix (e.g., `environment_test.go`)
* Test functions must start with `Test` prefix (e.g., `TestEnvironmentSettings`)
* Execute tests from project root where environment variables are available: `go test -v ./packagename`
* Use VS Code's integrated testing via F5 (Debug) to load `.env` files automatically
* For environment package: all 45 variables across 8 categories must be testable

## Build and Environment Guidelines
* `.env` file should be in project root for VS Code integration
* Environment file versioning uses `YYYYMMDDHHMMSS` timestamp format (no dots)

## Identifier Conventions
* JId: Whatsapp Jabber Identifier ("go.mau.fi/whatsmeow/types".JID)
* WId: Whatsapp String Identifier (string)
* LId: Whatsapp Local Identifier (new default Identifier, used to hide the phone number)

## Logging
* logrus "github.com/sirupsen/logrus"

## Calls Branch - Temporary META-Only Focus (SIP Disabled)
* **Current Branch**: `calls`
* **Temporary Scope Change**: All SIP proxy / `sipproxy` package usage must be ignored or bypassed. Assume SIP layer works; do NOT modify or depend on it for now.
* **New Primary Objective**: Achieve a correct META (WhatsApp) inbound call handshake (PREACCEPT → remote TRANSPORT → ACCEPT) using ONLY the `whatsmeow` client.
* **Out of Scope (Temporarily)**: Any SIP invite/answer / RTP bridging to external VoIP. Remove calls to `sipproxy.*` inside call flow code paths (guard with nil checks or feature flag) while we focus on raw WhatsApp signaling.
* **Reason**: Need to isolate why remote TRANSPORT is never received and validate ACCEPT formatting/order without side‑effects from SIP timing.

## Minimal Handshake Target
1. Receive `CallOffer`
2. Send `preaccept` node with: audio (rates 16000 + 8000), net(candidates), encopt(keygen=2)
3. Wait for remote `transport` (must log full node dump)
4. Respond with `accept` node (same structure order: audio,audio,net,encopt)
5. (Later) Exchange further transport/crypto if required

## Logging Requirements (MANDATORY)
* Tag each phase with consistent emojis/prefix: PREACCEPT, TRANSPORT-LOCAL, TRANSPORT-REMOTE, ACCEPT-SENT
* On receiving remote transport: dump hierarchy (tag, attrs, number of children, candidate list)
* Record handshake mode and env variables at start
* When skipping SIP: single log line `🛑 [SIP-DISABLED] Ignoring SIP integration (META-only mode)`

## Environment Flags (Provisional)
* QP_CALL_HANDSHAKE_MODE: preaccept-only | accept-early | accept-immediate | preaccept+transport(default)
* QP_CALL_INCLUDE_SRFLX: 0/1 include server-reflexive candidate
* QP_CALL_NET_MEDIUM: 1 => medium=1 else medium=3
* QP_CALL_STUN_FALLBACK: 1 enables multi-server STUN fallback
* QP_CALL_DISABLE_MONITOR: 1 disables re-send monitor
* QP_CALL_META_ONLY: 1 activates META-only mode (skip SIP code)

## Fake RTP Generator (Next Step)
Objective: After successful ACCEPT (handshake confirmed), start a lightweight goroutine that emits synthetic RTP-like UDP packets to our own candidate port (loopback) just to exercise code paths.

Implementation Sketch:
* New file: `whatsmeow/fake_rtp_generator.go`
* Function: `StartFakeRTP(callID string, targetIP string, targetPort int, stop <-chan struct{}, logger *log.Entry)`
  - Packet: 12-byte RTP header + 20 bytes dummy payload (increment sequence & timestamp)
  - PayloadType: 111 (Opus) or 96 dynamic
  - Interval: every 20ms (50pps) to simulate 20ms Opus frames
  - Uses `net.DialUDP` to targetIP:targetPort; if targetIP empty uses 127.0.0.1
  - Logs first 5 packets then every 100th
* Called ONLY after: remote transport received + ACCEPT sent.
* Provide Stop channel or context for cleanup on CallTerminate.

## Acceptance Success Indicators (Revised)
* Remote sends at least one `transport` node (log shows candidates)
* Our `accept` is sent exactly once afterward (unless experimental mode)
* No SIP logs appear while QP_CALL_META_ONLY=1
* (Optional) Fake RTP goroutine starts (log tag FAKE-RTP)

## Known Constraints (META-Only Phase)
* We still lack confirmed field ordering sensitivities—maintain identical order between preaccept & accept.
* STUN to Meta may fail; fallback list enabled when QP_CALL_STUN_FALLBACK=1.
* No attempt yet to forward real RTP toward WhatsApp (pending transport negotiation success).

## Next Milestones After Meta-Only Success
1. Integrate real RTP forwarding (replace fake generator) once remote transport appears
2. Re-enable SIP bridging behind feature flag
3. Add unit tests for candidate/net node building and handshake mode branching
4. Implement media encryption / key exchange if required by protocol traces

* **Primary STUN Server**: 157.240.226.62:3478 (Meta); fallbacks optional
* Files that you want to delete, just clean the contents, save the file, and delete it from the filesystem.

## Call Messages
* *Data*: ...