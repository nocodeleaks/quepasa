# Error 463 — NackCallerReachoutTimelocked

> **Status update (2026-06-02):** whatsmeow now ships a native privacy-token lifecycle
> (`tctoken` + `cstoken`). The version we pin —
> `go.mau.fi/whatsmeow v0.0.0-20260525123251-933deb5f2ee9` — **already includes it**.
> The earlier conclusion that "no proactive token mechanism exists" is obsolete.
> A truly cold *first* send can still be rejected, but the client now attaches the same
> tokens the official WhatsApp client does, which mitigates the rate-limit accumulation.

## What is it?

WhatsApp error 463 (`NackCallerReachoutTimelocked`) is a server-side rejection applied when a message is sent to a **cold contact** — a recipient who has never initiated a conversation with the sending number via the WhatsApp Web/multidevice protocol.

WhatsApp enforces a **reach-out time-lock**: a time-based rate limit on messaging unknown contacts. The server expects the outgoing message node to carry the same privacy tokens the official client sends (`tctoken` and/or `cstoken`). Messages that arrive **without** these tokens are counted more aggressively as unsolicited outreach, accelerating the lock and producing 463.

## The two tokens

The send path now attaches **one of two** tokens to a direct-message node (see _whatsmeow send flow_ below).

### 1. `tctoken` (trusted-contact / privacy token)

A per-pair cryptographic token stored in:

```sql
whatsmeow_privacy_tokens (our_jid, their_jid, token, timestamp, sender_timestamp)
```

**Key properties:**

- Strictly per `our_jid ↔ their_jid` pair — a token held by number A for contact X cannot be used by number B.
- Populated only when the recipient sends us a message (server pushes a `privacy_token` notification) **or** when we issue one to them (see below).
- Rolling **~28-day** window: 4 buckets of 7 days each (`tcTokenBucketDuration = 604800s`, `tcTokenNumBuckets = 4`).
- The client **does not** proactively fetch a tctoken for a contact it has no relationship with. `ensureTCToken` only reads the store and validates expiration — if absent, it returns nothing and the send falls back to `cstoken`.

### 2. `cstoken` (NCT — "New Chat Token") — *new*

A token **derived locally** for new conversations, requiring **no prior relationship**:

```
cstoken = HMAC-SHA256(NCTSalt, recipientLID)
```

This is the genuinely new anti-463 mechanism, matching official-client behavior on new chats. It is attached **only when** there is no valid tctoken. Requirements:

- The account must have an **`NCTSalt`** in the store (server-provisioned — see below).
- The recipient must have a **resolvable LID**. For a `@s.whatsapp.net` target, whatsmeow resolves the LID via `Store.LIDs.GetLIDForPN`; if no LID is known, no cstoken is produced.

If either prerequisite is missing, `generateCsToken` returns `nil`, the node ships with no token, and a cold contact still gets 463.

### Where the `NCTSalt` comes from

The salt is **never generated locally** — it is pushed by the WhatsApp server and stored per account in `whatsmeow_nct_salt (our_jid, salt)`:

1. **History sync (bootstrap at login):** the `HistorySync` protobuf carries `nctSalt` (field 19); stored on receipt — `whatsmeow/message.go` `storeNCTSalt(historySync.GetNctSalt())`.
2. **App state sync (ongoing):** an `nct_salt_sync` (`NCT_SALT_SYNC_ACTION`) mutation — `SET` stores the salt, `REMOVE` clears it — `whatsmeow/appstate.go`.

**Implication:** a freshly paired session has no salt until the initial history sync completes. During that window, cold-contact sends will 463 even on the current whatsmeow. Treat "NCTSalt present" as a readiness gate before driving cold outbound — `Store.NCTSalt.GetNCTSalt(ctx)`.

## whatsmeow send flow (current, `send.go`)

```go
// 1. Attach a token: prefer a stored tctoken, else fall back to a derived cstoken.
tcTokenBytes, _ := cli.ensureTCToken(ctx, to)
if len(tcTokenBytes) > 0 {
    node.Content = append(node.GetChildren(), waBinary.Node{Tag: "tctoken", Content: tcTokenBytes})
} else if csToken := cli.generateCsToken(ctx, to); len(csToken) > 0 {
    node.Content = append(node.GetChildren(), waBinary.Node{Tag: "cstoken", Content: csToken})
}

// 2. Send the node.
data, err := cli.sendNodeAndGetData(ctx, *node)

// 3. Fire-and-forget: on a new 7-day bucket boundary, issue a trusted_contact
//    token TO the recipient (privacy IQ), mirroring the official client. This is
//    what lowers reach-out-timelock accumulation over time.
storageJID := cli.resolveTCTokenStorageLID(ctx, to)
if shouldSendTCTokenInChatAction(to) && shouldSendNewTCToken(cli.getTCTokenSenderTS(storageJID)) {
    go cli.issuePrivacyTokenAndSave(storageJID, time.Now())
}
```

`issuePrivacyToken` sends an IQ in the `privacy` namespace with a `trusted_contact` token entry to `ServerJID`. `ensureTCToken` also opportunistically prunes expired tokens (`deleteExpiredPrivacyTokens`, throttled to once per 24h).

## Server-pushed enforcement notification: `NotifyAccountReachoutTimelock`

In addition to the per-message 463 NACK, whatsmeow also exposes a **proactive, account-level notification** that the server pushes when reach-out-timelock enforcement is (de)activated for the account — independent of any specific send. It arrives as a `mex` notification (`xwa2_notify_account_reachout_timelock`) and is dispatched as the typed event `events.NotifyAccountReachoutTimelock` (`whatsmeow/notification.go`, `whatsmeow/types/events/events.go:635`):

```go
type NotifyAccountReachoutTimelock struct {
    Mex                 MexNotificationData `json:"-"`
    EnforcementType     string              `json:"enforcement_type,omitempty"`
    IsActive            bool                `json:"is_active,omitempty"`
    TimeEnforcementEnds jsontime.UnixString `json:"time_enforcement_ends,omitzero"`
}
```

**Observed example (2026-06-30):**

```json
{
  "enforcement_type": "RESTRICT_ALL_COMPANIONS",
  "is_active": true,
  "time_enforcement_ends": "1783343326"
}
```

`time_enforcement_ends` is a Unix timestamp — `1783343326` = **2026-07-06 13:08:46 UTC**, i.e. the enforcement was set for roughly a week from the time the event fired.

`enforcement_type: RESTRICT_ALL_COMPANIONS` indicates the lock applies to **all linked devices (companions)** on the account, not just a single send/contact — this is the account-wide signal that 463s are about to (or are already) being handed out broadly, as opposed to the per-recipient cold-contact case covered by the rest of this doc.

**Variant: empty payload (observed 2026-06-30):**

```json
{}
```

i.e. `enforcement_type`, `is_active`, and `time_enforcement_ends` are all absent — every field on `NotifyAccountReachoutTimelock` has `omitempty`/`omitzero`, so this is the server sending the notification with all-zero values (`EnforcementType: ""`, `IsActive: false`, `TimeEnforcementEnds: zero`) rather than a parse failure on our side. Read literally, an empty payload with `is_active` defaulting to `false` would mean "enforcement lifted / not active" — but since the fields are indistinguishable from "not sent," this can't be relied on as a clear signal. Anything that ends up handling this event needs to treat the all-empty case as **unknown/no-op**, not as an authoritative "timelock cleared."

**Current status:** this event is **not handled** by our code — it currently surfaces only via the generic `unhandled` event logger (`[Type: *events.NotifyAccountReachoutTimelock] ...`). There is no listener wired up in `src/whatsmeow/whatsmeow_connection.go`.

**Potential value of handling it:**
- It's a direct, authoritative signal from WhatsApp that the account is (or stops being) under
  reach-out enforcement — could be used to short-circuit outbound sends to cold contacts while
  `is_active` is true, instead of discovering it per-message via 463s.
- `time_enforcement_ends` gives an expected lift time, useful for backoff/alerting.
- Could feed into the NCTSalt-readiness/cold-outbound gating discussed in the Open TODOs below.

## Production investigation (2026-06-01)

### What was confirmed

**Scenario A — truly cold contact (`55xxxxxxxxxx`, LID `25xxxxxxxxxxxxx@lid`):**

```
11:37:34 → SendMessage to 55xxxxxxxxxx@s.whatsapp.net → 463
11:37:34 → no privacy token for 55xxxxxxxxxx@s.whatsapp.net
11:37:34 → retry via LID 25xxxxxxxxxxxxx@lid → 463
11:37:37 → [3s wait] → signal state reset → retry → 463
```

All retries fail. The token never appears.

**Scenario B — contact becomes warm (sent us a message):**

```
11:31:27 → received message from 11xxxxxxxxxxxxx@lid (contact messaged us)
11:31:35 → SendMessage → send success ✅ (no retry needed)
```

After the contact messages us, the token is issued and the very next send succeeds immediately.

**Scenario C — LID known via another connected number's group:**

- The token for that LID in the sender's store was absent — tokens from other numbers are not shared.
- Pre-warming via `SubscribePresence` from that sender was ignored by the server.

### Key findings

1. **SubscribePresence does not create tokens for cold contacts.** The server logs `"without privacy token"` and ignores the subscribe. Confirmed in multiple cases, and consistent with upstream: there is no presence-based token-fetch path in whatsmeow.
2. **Tokens are per sender, not global.** A token discovered via a shared group on another connected number is useless for a different sender — the DB key is `(our_jid, their_jid)`.
3. **Pre-warming proactively before send is ineffective** for the same reason.
4. **Once warm, sends succeed without retries.**
5. **The 3-second SubscribePresence wait was wasting time** for cold contacts.

## Our implementation (`whatsmeow_connection.go`, region `error 463`)

The send path runs `Client.SendMessage`; on a 463 it applies a retry ladder:

```
SendMessage(phone@s.whatsapp.net)
    └── 463
        └── resolveLIDRetryJID(phone) → LID found?
                ├── NO → return error
                └── YES
                    └── hasPrivacyToken(LID)?
                            ├── NO (cold contact) → return error immediately
                            └── YES (token exists but may be stale)
                                └── SendMessage(LID)
                                    └── 463
                                        └── hasPrivacyToken(finalJID)?
                                                ├── NO → return error
                                                └── YES
                                                    └── subscribePresencePreWarm (3s wait)
                                                        └── resetSignalStateForTargets
                                                            └── SendMessage(finalJID) → final attempt
```

`hasPrivacyToken` does a dual lookup (LID JID and the phone equivalent) because the token may be stored under either key depending on how the contact first interacted with us.

### ⚠️ This custom ladder is now partly obsolete

Given the native whatsmeow lifecycle, parts of our retry code no longer earn their keep:

- **`subscribePresencePreWarm` + 3s wait + `resetSignalStateForTargets`** — the doc and upstream both confirm presence pre-warm cannot mint a token for a cold contact, and the signal reset is unrelated to the token mechanism. This path costs ~3s per failed cold send for no benefit.
- **`hasPrivacyToken` only checks the stored `tctoken`** — it is blind to the `cstoken` (which is derived on the fly, never stored). The short-circuit *"skip all retries: no privacy token (cold contact)"* can therefore bail in cases where whatsmeow would still attach a valid cstoken on its own send. Whether that send would actually succeed is contact-dependent, but our pre-check is no longer an accurate proxy for "the server will reject this."

**Recommendation (not yet applied):** drop the pre-warm + signal-reset stage, and either remove the cold-contact short-circuit or rename it to reflect that it gates only on tctoken presence. Keep the cheap LID retry. Track via the TODO below.

## What works vs. what doesn't

| Approach                              | Works? | Notes                                                                 |
| ------------------------------------- | ------ | --------------------------------------------------------------------- |
| `tctoken` lifecycle (native)          | ✅     | Auto-attached when a stored token exists; issued to recipient on send |
| `cstoken` (native, derived)           | ⚠️     | Attached for cold contacts **if** NCTSalt + recipient LID are present |
| `SubscribePresence` pre-warm          | ❌     | Server ignores without prior context                                  |
| Signal state reset                    | ❌     | Unrelated to token mechanism                                          |
| LID retry                             | ❌     | Same token requirement applies to LID                                 |
| Wait for inbound message              | ✅     | Token issued on first inbound; next send succeeds                     |
| WhatsApp Business API                 | ✅     | Officially supported, tokens guaranteed by Meta                       |
| GraphQL MEX reachout-timelock query   | ❌     | query_id hash not public; changes per WA version                      |

For a truly cold **first** send with no NCTSalt/LID context, there is still **no guaranteed client-side fix** — it is an intentional Meta anti-spam mechanism.

## Open TODOs

- **Simplify our 463 retry ladder** — remove the ineffective `subscribePresencePreWarm` + `resetSignalStateForTargets` stage and reconsider the tctoken-only cold-contact short-circuit (see warning above).
- **Add an NCTSalt readiness gate** before driving cold outbound on a freshly paired session (`Store.NCTSalt.GetNCTSalt`).
- **Consider bumping whatsmeow** — newer pins exist (e.g. `v0.0.0-20260529...`); track upstream for further token changes.
- **Token expiration handling** is now native: ~28-day rolling window with automatic pruning (`deleteExpiredPrivacyTokens`).
- **Handle `events.NotifyAccountReachoutTimelock`** — currently unhandled, only logged generically. Consider tracking `is_active`/`time_enforcement_ends` per account to gate or warn on cold outbound while enforcement is active (see dedicated section above).

## How sibling libraries handle it

- **whatsmeow** ([#1074](https://github.com/tulir/whatsmeow/issues/1074)): the practical resolution was the tctoken/cstoken lifecycle (we already ship it).
- **WAHA** ([#2050](https://github.com/devlikeapro/waha/issues/2050), [#1992](https://github.com/devlikeapro/waha/issues/1992)): their GOWS engine lacked the lifecycle; the fix is to adopt the updated whatsmeow.
- **Baileys** ([#2441](https://github.com/WhiskeySockets/Baileys/issues/2441)): implemented tctoken (PRs #2257/#2339), cstoken (#2438), and stale-token recycling — same direction.

## References

- [whatsmeow issue #1074 — error 463](https://github.com/tulir/whatsmeow/issues/1074)
- [Baileys issue #2441 — 463 investigation](https://github.com/WhiskeySockets/Baileys/issues/2441)
- [go-whatsapp-web-multidevice PR #695 — pre-warm + retry](https://github.com/aldinokemal/go-whatsapp-web-multidevice/pull/695)
- [WAHA issue #1992 — cold contact 463](https://github.com/devlikeapro/waha/issues/1992)
- [WAHA issue #2050 — tctoken lifecycle not implemented in GOWS](https://github.com/devlikeapro/waha/issues/2050)
- whatsmeow source (our pinned version): `tctoken.go`, `cstoken.go`, `send.go`, `appstate.go`, `message.go`
- whatsmeow source for the account-level notification: `notification.go`, `types/events/events.go` (`NotifyAccountReachoutTimelock`)
- [Our LID vs phone investigation — ISSUE-lid-vs-phone.md](./ISSUE-lid-vs-phone.md)
- Implementation: `src/whatsmeow/whatsmeow_connection.go` — region `error 463`
