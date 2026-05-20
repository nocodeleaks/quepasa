# Investigation: WhatsApp Send Suspensions and Destination Validation

Date: 2026-05-15
Status: In progress (no code changes applied)
Scope: Investigate whether recent message-send flow changes introduced destination existence checks that can increase suspension risk.

## Context

Reported behavior:
- Before 2026-05-13: each account could send ~40-50 messages.
- After 2026-05-13: each account can send only ~3-5 messages before suspension/rate collapse.

Primary concern:
- A recent update may have introduced destination validation ("is this number on WhatsApp?") in the hot path before sending messages.
- At high volume, this can resemble number enumeration and increase anti-abuse risk.

## Objective

Determine whether the production send path now performs remote recipient-existence checks before message send.

## Investigation Constraints

- No functional code changes before investigation results are confirmed.
- Evidence-first approach: file/symbol tracing, then behavior classification.
- Distinguish local resolution from remote enumeration.

## Confirmed Findings (Code + Git)

### 1) Send hot path does call recipient-existence validation (conditional)

Confirmed in `src/models/server_messaging.go` (`QpWhatsappServer.SendMessage`):
- The send path has a conditional block guarded by `ENV.ShouldNormalizeBRPhone()`.
- Inside this block, it calls `contactManager.IsOnWhatsApp(phone, variant)` before `conn.Send(msg)`.

Meaning:
- If `NORMALIZE_BR_PHONE=true` (or legacy `REMOVEDIGIT9=true`), outbound sends can perform pre-send recipient registration checks.
- This is in the core send path, not only in a dedicated admin endpoint.

### 2) `IsOnWhatsApp` is a live remote WhatsApp query

Confirmed in `src/whatsmeow/whatsmeow_contact_manager.go`:
- `IsOnWhatsApp(...)` calls `cm.Client.IsOnWhatsApp(context.Background(), uncached)` for uncached phones.
- There is an explicit warning in code comments that frequent calls may trigger anti-abuse and banning.

Confirmed in `src/whatsmeow/whatsmeow_contact_maps.go`:
- Cache exists (`GetIsOnWhatsAppCache` / `SetIsOnWhatsAppCache`) and stores both positive and negative results.
- Warning comment explicitly states anti-abuse risk for frequent `IsOnWhatsApp` usage.

Operational interpretation:
- Cache reduces repeated lookups for the same number.
- High-volume sends to many unique numbers still produce many remote lookups and remain high risk.

### 3) `FormatEndpoint` itself is local formatting only

Confirmed in `src/whatsapp/whatsapp_extensions.go`:
- `FormatEndpoint` only normalizes/parses chat identifiers and suffixes.
- No remote existence call occurs inside `FormatEndpoint`.

### 4) Additional remote lookup found in contact vCard generation path

Confirmed in `src/whatsmeow/whatsmeow_connection.go` (`generateVCardForContact`):
- Calls `source.Client.GetUserInfo(...)` to infer WhatsApp/business status for contact message vCards.
- This is not the generic text-send path, but it is another recipient-info remote query path.

### 5) Timeline correlation (commits)

Confirmed by Git history:
- 2026-05-06 commit `9d723b2`: introduced BR normalization logic in send path and `ShouldNormalizeBRPhone()` environment behavior.
  - Files: `src/models/server_messaging.go`, `src/models/qp_env.go`.
- 2026-05-13 commit `d422433`: updated `go.mau.fi/whatsmeow` dependency to `v0.0.0-20260513140310-c551a4055c0f`.
  - Files: `src/go.mod`, `src/whatsmeow/go.mod`.

Interpretation:
- There are two plausible contributors close to the reported incident window:
  1. local project send-path behavior (conditional `IsOnWhatsApp` pre-check)
  2. dependency update on 2026-05-13

Neither alone proves causality yet, but both are temporally relevant.

### 6) Environment default does not enable BR normalization automatically

Confirmed in `src/.env.example`:
- `REMOVEDIGIT9=false` by default.

Interpretation:
- The risky pre-send `IsOnWhatsApp` path in `SendMessage` is conditional.
- It is only active if `NORMALIZE_BR_PHONE=true` (or legacy `REMOVEDIGIT9=true`) in runtime environment.

### 8) Runtime service environment (current host) has normalization enabled

Confirmed from service wiring + env file:
- Service template (`helpers/quepasa.service`) loads `EnvironmentFile=-/opt/quepasa/.env`.
- Current host file `/opt/quepasa/.env` contains `REMOVEDIGIT9=true`.

Interpretation:
- On this host/service profile, the conditional pre-send `IsOnWhatsApp` path is effectively active.
- This strengthens the likelihood that high-volume sends can trigger large recipient-existence query volume before actual message send.

### 7) Additional remote lookup path exists inside current `whatsmeow` send behavior

Confirmed in local module source (`go.mau.fi/whatsmeow@v0.0.0-20260513140310-c551a4055c0f`):
- In `send.go`, when destination is PN (`@s.whatsapp.net`), `LIDMigrationTimestamp > 0`, and LID mapping is missing in store, send path performs `GetUserInfo(...)` to fetch LID before send.
- In `message.go`, `LIDMigrationTimestamp` can be set from sync/global settings payload (`storeLIDSyncMessage`, `storeGlobalSettings`).

Operational interpretation:
- Even with BR normalization disabled, this library-level path may still perform recipient-info lookups for unmapped recipients in migrated contexts.
- If many recipients are new/unmapped, lookup volume can increase significantly.

## High-Risk Patterns to Confirm/Reject

If present per-destination in send hot path, these are high risk:
- `IsOnWhatsApp(...)`
- `GetUserInfo(...)`
- `GetUserDevices(...)`
- `usync` lookup calls
- Any remote lookup fallback used to "confirm recipient exists" before send

## Low-Risk Patterns (Generally)

- Local `@lid` to `@s.whatsapp.net` mapping from local store/cache.
- Pure formatting/normalization without remote contact checks.

## Classification Matrix (Current)

- `SendAnyWithServer` / `SendWithMessageType` (`src/api/api_handlers+SendController.go`): local formatting + routing, no direct `IsOnWhatsApp` call there.
- `FormatEndpoint` (`src/whatsapp/whatsapp_extensions.go`): local-only.
- `QpWhatsappServer.SendMessage` (`src/models/server_messaging.go`): conditional remote lookup via `IsOnWhatsApp` when BR normalization is enabled.
- `WhatsmeowContactManager.IsOnWhatsApp` (`src/whatsmeow/whatsmeow_contact_manager.go`): remote lookup with per-session cache.
- `WhatsmeowConnection.generateVCardForContact` (`src/whatsmeow/whatsmeow_connection.go`): remote `GetUserInfo` lookup for contact cards.
- `whatsmeow send.go` (module `c551a4055c0f`): conditional remote `GetUserInfo` lookup for PN->LID resolution when `LIDMigrationTimestamp > 0` and mapping miss occurs.

## Audit Plan (Remaining)

1. Confirm runtime environment values for `NORMALIZE_BR_PHONE` / `REMOVEDIGIT9` across all affected hosts (one host already confirmed enabled).
2. Quantify cache hit/miss behavior for `IsOnWhatsApp` during high-volume send campaigns (unique recipients vs repeated recipients).
3. Quantify PN->LID store miss rate during send (to estimate `GetUserInfo` calls triggered by LID migration).
4. Compare suspension behavior with BR normalization flag off vs on (controlled test cohort).
5. If needed, compare send behavior changes between whatsmeow `eb05d94` and `c551a40` in a controlled environment.
6. Improve observability for lookup volume (current journal probe did not expose lookup-specific entries in available logs).

## Evidence Log

- 2026-05-15: Investigation document created.
- 2026-05-15: No code behavior changed yet.
- 2026-05-15: Confirmed `SendMessage` contains conditional `IsOnWhatsApp` pre-send lookup (`NORMALIZE_BR_PHONE` / `REMOVEDIGIT9` gated).
- 2026-05-15: Confirmed `IsOnWhatsApp` implementation performs live remote queries for uncached recipients and warns about anti-abuse risk.
- 2026-05-15: Confirmed `FormatEndpoint` is local-only parsing/normalization.
- 2026-05-15: Confirmed commit `9d723b2` (2026-05-06) added BR normalization + pre-send lookup logic.
- 2026-05-15: Confirmed commit `d422433` (2026-05-13) updated whatsmeow to `c551a4055c0f`.
- 2026-05-15: Confirmed `REMOVEDIGIT9=false` default in `src/.env.example` (runtime activation still environment-dependent).
- 2026-05-15: Confirmed current whatsmeow send path can call `GetUserInfo` for PN->LID resolution under LID migration conditions.
- 2026-05-15: Confirmed service env file `/opt/quepasa/.env` has `REMOVEDIGIT9=true` on current host; service template loads this file.
- 2026-05-15: Runtime log probe (`/var/log/quepasa` + `journalctl -u quepasa`) found no lookup-specific traces with current patterns/log level.

## Working Conclusion (Updated)

There is now explicit evidence that the send path can perform recipient-existence checks before sending, when BR normalization is enabled. This behavior is conditional, cached, and intended for 8/9-digit BR disambiguation, but at scale (many unique numbers) it can still produce high volumes of remote existence probing.

Given the reported date window, both the pre-send validation path and the 2026-05-13 whatsmeow bump are relevant suspects and should be validated with runtime flag state and controlled A/B sending tests.

Additionally, current whatsmeow internals indicate a second potential lookup amplifier (PN->LID `GetUserInfo` fallback under migration). This should be measured before final attribution.

Current confidence increased for the pre-send validation hypothesis on this host because runtime configuration is explicitly enabling the path.
