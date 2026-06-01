# Error 463 — NackCallerReachoutTimelocked

## What is it?

WhatsApp error 463 (`NackCallerReachoutTimelocked`) is a server-side rejection applied when a message is sent to a **cold contact** — a recipient who has never initiated a conversation with the sending number via the WhatsApp Web/multidevice protocol.

The server requires a **privacy token (`tctoken`)** to be included in the outgoing message node. Without it, the server counts the attempt as an unsolicited outreach and rejects with 463.

## The tctoken (privacy token)

The `tctoken` is a per-pair cryptographic token stored in the whatsmeow database table:

```sql
whatsmeow_privacy_tokens (our_jid, their_jid, token, timestamp)
```

**Key properties:**

- Strictly per `our_jid ↔ their_jid` pair — a token held by number A for contact X cannot be used by number B
- Populated only when the recipient sends a message to us (server pushes a `privacy_token` notification)
- Can also be populated via `SubscribePresence` — but **only if the server already has relationship context** for the pair
- Has a ~28-day sliding expiration window (refreshed on each use)

The whatsmeow send path (`send.go`) automatically includes the token in the message node if it exists in the store:

```go
if tcToken, err := cli.Store.PrivacyTokens.GetPrivacyToken(ctx, to); err != nil {
    // log warning
} else if tcToken != nil {
    node.Content = append(node.GetChildren(), waBinary.Node{
        Tag:     "tctoken",
        Content: tcToken.Token,
    })
}
```

If the token is absent, the message is sent without it — and 463 results for cold contacts.

## Production investigation (2026-06-01)

### What was confirmed

**Scenario A — truly cold contact (`55xxxxxxxxxx`, LID `25xxxxxxxxxxxxx@lid`):**

```
11:37:34 → SendMessage to 55xxxxxxxxxx@s.whatsapp.net → 463
11:37:34 → no privacy token for 55xxxxxxxxxx@s.whatsapp.net
11:37:34 → retry via LID 25xxxxxxxxxxxxx@lid → 463
11:37:34 → no privacy token for 25xxxxxxxxxxxxx@lid
11:37:34 → SubscribePresence → "without privacy token" (server ignores)
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

- `55xxxxxxxxxx` was discovered as a group participant of `55yyyyyyyyyy`, not `55zzzzzzzzzz`
- The token for that LID in `55zzzzzzzzzz`'s store was absent — tokens from other numbers are not shared
- Pre-warming via `SubscribePresence` from `55zzzzzzzzzz` was ignored by the server

### Key findings

1. **SubscribePresence does not create tokens for cold contacts.** The server logs `"without privacy token"` and ignores the subscribe. This was confirmed in multiple cases.

2. **Tokens are per sender, not global.** A token discovered via a shared group on another connected number is useless for a different sender — the DB key is `(our_jid, their_jid)`.

3. **Pre-warming proactively before send is ineffective** for the same reason: no prior relationship context → server does not issue token.

4. **Once warm, sends succeed without retries.** The WhatsApp server tracks the relationship after the first inbound message and accepts subsequent sends without a token check.

5. **The 3-second SubscribePresence wait was wasting time.** For cold contacts, the token never arrives, so the wait is unconditionally wasted.

## Our implementation (whatsmeow_connection.go)

### Retry flow

```
SendMessage(phone@s.whatsapp.net)
    └── 463
        └── resolveLIDRetryJID(phone) → LID found?
                ├── NO → return error
                └── YES
                    └── hasPrivacyToken(LID)?
                            ├── NO (cold contact) → return error immediately ✅
                            └── YES (token exists but may be stale)
                                └── SendMessage(LID)
                                    └── 463
                                        └── hasPrivacyToken(LID)?
                                                ├── NO → return error
                                                └── YES
                                                    └── SubscribePresence (3s wait)
                                                        └── resetSignalState
                                                            └── SendMessage(LID) → final attempt
```

### hasPrivacyToken dual-lookup

The token may be stored under either the LID JID or the phone JID depending on how the contact first interacted with us:

```go
func (source *WhatsmeowConnection) hasPrivacyToken(target types.JID) bool {
    // 1. Check by LID directly
    token, err := store.GetPrivacyToken(ctx, target)
    if err == nil && token != nil { return true }

    // 2. If LID, also check the phone equivalent
    if target.Server == types.HiddenUserServer {
        phone := contactManager.GetPhoneFromContactId(target)
        phoneJID := phone + "@s.whatsapp.net"
        token, err = store.GetPrivacyToken(ctx, phoneJID)
        if err == nil && token != nil { return true }
    }
    return false
}
```

## What cannot be fixed

For truly cold contacts, there is **no client-side workaround**. This is an intentional Meta anti-spam mechanism. The options are:

| Approach                            | Works? | Notes                                             |
| ----------------------------------- | ------ | ------------------------------------------------- |
| `SubscribePresence` pre-warm        | ❌     | Server ignores without prior context              |
| Signal state reset                  | ❌     | Unrelated to token mechanism                      |
| LID retry                           | ❌     | Same token requirement applies to LID             |
| Wait for inbound message            | ✅     | Token issued on first inbound; next send succeeds |
| WhatsApp Business API               | ✅     | Officially supported, tokens guaranteed by Meta   |
| GraphQL MEX reachout-timelock query | ❌     | query_id hash not public; changes per WA version  |

## Open TODOs

- **Monitor whatsmeow upstream** for a proactive token acquisition path — no such mechanism exists as of 2026-06-01.
- **GraphQL MEX query** (`WAWebMexFetchReachoutTimelockJobQuery`): whatsmeow already has `sendMexIQ()` and the `w:mex` namespace wired (used for newsletters). However, WhatsApp uses **persisted queries** — only a numeric hash (`query_id`) is sent over the wire, never the query text. The hash for the timelock query is not publicly documented, is extracted from WhatsApp APK decompilation, and changes per client version. Not viable to implement reliably.
- **Token expiration handling**: tokens have a ~28-day TTL. If a previously-warm contact's token expires and send fails with 463, the current code will attempt the pre-warm + signal reset path, which may succeed if the server reissues the token via SubscribePresence.

## References

- [whatsmeow issue #1074 — error 463](https://github.com/tulir/whatsmeow/issues/1074)
- [Baileys issue #2441 — 463 investigation](https://github.com/WhiskeySockets/Baileys/issues/2441)
- [go-whatsapp-web-multidevice PR #695 — pre-warm + retry](https://github.com/aldinokemal/go-whatsapp-web-multidevice/pull/695)
- [WAHA issue #1992 — cold contact 463](https://github.com/devlikeapro/waha/issues/1992)
- [whatsmeow_privacy_tokens schema — store/sqlstore/store.go](https://github.com/tulir/whatsmeow/blob/main/store/sqlstore/store.go)
- [Our LID vs phone investigation — ISSUE-lid-vs-phone.md](./ISSUE-lid-vs-phone.md)
- Implementation: `src/whatsmeow/whatsmeow_connection.go` — region `error 463`
