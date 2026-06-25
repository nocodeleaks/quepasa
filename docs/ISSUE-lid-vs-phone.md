# LID vs Phone Number: WhatsApp Account Migration

## Background

WhatsApp is migrating accounts in phases to a privacy-first identifier system based on **LID** (Local Identifier).
LIDs are opaque identifiers in the format `<opaque_number>@lid` that hide the user's phone number.

The migration is server-initiated — WhatsApp decides when to migrate each account.
The client learns about its own migration status through session sync events.

---

## The `LIDMigrationTimestamp` Field

Stored in `whatsmeow/store.Device`:

```go
LIDMigrationTimestamp int64
```

This is a Unix timestamp set by the whatsmeow client when the WhatsApp server signals that the account has been migrated to LID-based chat database. Two sources populate it:

### Source 1: LID sync message (`storeLIDSyncMessage`)

Triggered by a `waLidMigrationSyncPayload.LIDMigrationMappingSyncPayload` received during app state sync.
Contains:
- `ChatDbMigrationTimestamp` — the migration timestamp
- `PnToLidMappings` — batch of `PN → LID` pairs to populate the local store

File: `message.go` → `storeLIDSyncMessage()`

### Source 2: Global settings history sync (`storeGlobalSettings`)

Triggered by `waHistorySync.GlobalSettings` received during history sync.
Contains:
- `ChatDbLidMigrationTimestamp`

File: `message.go` → `storeGlobalSettings()`

In both cases, if `LIDMigrationTimestamp == 0` and the received value `> 0`, it is persisted to the store with `Store.Save()`.

---

## How `LIDMigrationTimestamp` Affects Sending

Source: `send.go`, function `Client.SendMessage()`, lines 323–345.

```go
} else if to.Server == types.HiddenUserServer {
    // Destination is already @lid → just set own identity to LID, send as-is
    ownID = cli.getOwnLID()

} else if to.Server == types.DefaultUserServer && cli.Store.LIDMigrationTimestamp > 0 && !req.Peer {
    // Destination is @s.whatsapp.net AND account was migrated → force convert PN → LID
    toLID, err = cli.Store.LIDs.GetLIDForPN(ctx, to)
    if toLID.IsEmpty() {
        // Fallback: fetch from server via GetUserInfo
    }
    to = toLID           // replace destination with @lid
    ownID = cli.getOwnLID()
}
```

### Behavior matrix

| Account state | Send to `@s.whatsapp.net` | Send to `@lid` |
|---|---|---|
| Not migrated (`= 0`) | Sent as PN directly | Sent as LID directly |
| Migrated (`> 0`) | **Converted to LID before send** | Sent as LID directly |

### Key observations

- Sending to `@lid` **never converts** to phone number — it always sends as-is.
- Sending to `@s.whatsapp.net` on a migrated account **always converts** to LID automatically.
- On a migrated account, if the LID mapping is missing from the local store and `GetUserInfo` also fails to return one, the send **fails with an error** (no silent fallback to PN).
- `HiddenUserServer` (`@lid`) and `DefaultUserServer` (`@s.whatsapp.net`) both route to `sendDM()` — the same underlying send function.

---

## LID Store and Mapping

The LID-to-phone bidirectional map is maintained in two places:

### In-memory cache

Thread-safe singleton maps in QuePasa:
- `lidToPhone map[string]string`
- `phoneToLID map[string]string`

File: `src/whatsmeow/whatsmeow_contact_maps.go`

### Persistent store (whatsmeow SQLite)

Table: `whatsmeow_lid_map` (or equivalent in the active store backend)

Methods:
- `Store.LIDs.GetLIDForPN(ctx, pnJID)` → returns `@lid` for a phone JID
- `Store.LIDs.GetPNForLID(ctx, lidJID)` → returns `@s.whatsapp.net` for a LID
- `Store.LIDs.PutManyLIDMappings(ctx, pairs)` → bulk insert during sync

**Important:** Not all contacts have a mapping. A contact only has a LID entry in the store if:
1. A LID sync message was received that included that contact, OR
2. The contact appeared in a history sync, OR
3. The contact sent a message that carried LID metadata

---

## Implications for QuePasa Send Flow

### Standard send (`POST /messages/send`)

File: `src/api/api_handlers+SendController.go`

- If `chatid` ends with `@lid`, QuePasa tries to resolve it to a phone number via `server.GetPhoneFromLID(lid)`.
- If resolved, converts to `@s.whatsapp.net` and sends normally.
- If not resolved, may attempt direct send to `@lid`.
- On migrated accounts, the whatsmeow layer will convert back to LID anyway — making the PN→LID→PN→LID round-trip redundant.

### Direct LID send (`POST /messages/lid/direct`)

File: `src/api/api_handlers+LIDSendController.go`

- Bypasses the API-layer LID→phone conversion.
- Calls `wmConn.Client.SendMessage` directly with the `@lid` JID.
- Logs `chatid`, parsed JID, and returned `msgid` on success.
- On migrated accounts: behaves identically to normal send (whatsmeow sends `@lid` as-is).
- On non-migrated accounts: valid test of whether the server accepts `@lid` routing directly.

---

## Wire-Level Evidence

From `.dist/wa_cdp_callflow.json` (captured locally):

```
"to": "62556332941345@lid"
```

The XMPP/Noise frame shows `to="62556332941345@lid"` intact — confirming the LID reaches WhatsApp servers without conversion by the client.

---

## Recipient Normalization Flow (QuePasa)

### 1) Generic endpoint formatting

`FormatEndpoint(...)` accepts `@lid` as a valid suffix and does not convert it to phone automatically.

File: `src/whatsapp/whatsapp_extensions.go`

Implication: a message can reach lower layers still targeting `something@lid`.

### 2) API send path (best-effort conversion)

In API send flow, if `chat.id` ends with `@lid`, QuePasa tries:

1. Resolve phone using `server.GetPhoneFromLID(lid)`
2. Convert to `@s.whatsapp.net` via `PhoneToWid(phone)`
3. If mapping lookup fails, it may try direct send to `@lid`

File: `src/api/api_handlers+SendController.go`

### 3) Connection send path

`WhatsmeowConnection.Send(...)` validates and sends the provided JID.
It does not force `@lid` to `@s.whatsapp.net` before first attempt.

File: `src/whatsmeow/whatsmeow_connection.go`

Special behavior:

- If send fails with error `463` to a phone JID, code retries once via resolved LID (`resolveLIDRetryJID`).
- Retry direction is phone -> LID (not LID -> phone).

---

## Why `@lid` Delivery Fails in Practice

Most common causes:

1. Missing LID->phone mapping in active store/cache for that recipient.
2. Attempt to send directly to `@lid` when mapping is absent.
3. Session/privacy token state mismatch (observed around send error `463`).
4. Store-only mode cannot resolve mapping (`GetPhoneFromLID` requires active connection manager).

---

## Verified Related Paths

- API conversion and fallback logic:
    - `src/api/api_handlers+SendController.go`
- Connection send and 463 retry:
    - `src/whatsmeow/whatsmeow_connection.go`
- Mapping resolver methods:
    - `src/whatsmeow/whatsmeow_contact_manager.go`
    - `src/whatsmeow/whatsmeow_contact_maps.go`
- Server-level wrappers:
    - `src/models/server_messaging.go`
- Direct `@lid` testing endpoint:
    - `src/api/api_handlers+LIDSendController.go`
    - `src/api/api_routes_messages.go`

---

## Recommended Stable Policy

For outbound messaging reliability:

1. If input is `@lid`, resolve to phone first whenever mapping is available.
2. Send to `@s.whatsapp.net` when phone resolution succeeds.
3. If resolution fails, return explicit resolution error instead of silently relying on direct `@lid` send.
4. Keep direct `@lid` send as controlled fallback (for diagnostics and controlled scenarios).

This matches operational behavior where converting `@lid` to phone before dispatch tends to be more reliable.

---

## Debug Checklist

When a `@lid` send fails:

1. Confirm active connected session (`Ready`).
2. Call mapping endpoint or internal resolver for that LID (`GetPhoneFromLID`).
3. If empty or error, treat as unresolved mapping (do not derive phone from LID text).
4. If resolved, resend using `phone@s.whatsapp.net`.
5. If send returns `463`, inspect retry logs and signal/privacy state reset logs in connection send path.

---

## Brazilian Mobile Phone: The Extra Digit 9 Issue

### Background

Brazil migrated mobile phone numbers from 8 digits to 9 digits (with the extra `9` prefix) starting in 2012, region by region. The migration was phased by DDD (area code):

- **DDDs ≤ 30** (São Paulo capital and region): migrated first
- **DDDs > 30** (rest of the country): migrated later, with many older numbers still registered on WhatsApp **without** the extra `9`

This creates an ambiguity: a number like `+5547976090095` (9-digit mobile) may exist on WhatsApp as `+554767609095` (8-digit legacy), depending on when the user registered their account.

### The Problem for Outbound Sends

When an external system sends a message with the full 9-digit number and the WhatsApp account was registered with the legacy 8-digit number, the send fails silently — WhatsApp reports the number as not found.

### QuePasa Solution: `REMOVEDIGIT9`

QuePasa implements a pre-send resolution step controlled by the environment variable:

```
REMOVEDIGIT9=true   # default: false
```

File: `src/models/qp_env.go` → `ShouldRemoveDigit9()`

### Eligibility Check for Phone-Only Send: `RemoveDigit9IfElegible`

File: `src/library/phone.go`

```go
// Whatsapp issue on understanding mobile phones with ddd bigger than 30, only mobile
func RemoveDigit9IfElegible(source string) (response string, err error) {
    if len(source) == 14 {
        // mobile phones with 9 digit (subscriber must start with 5-9)
        r, _ := regexp.Compile(`\+55([4-9][1-9]|[3-9][1-9])9[5-9]\d{7}$`)
        if r.MatchString(source) {
            prefix := source[0:5]          // "+55" + DDD (2 digits)
            response = prefix + source[6:14] // skip the extra "9" at position 5
        }
    }
    return
}
```

**Regex breakdown:** `\+55([4-9][1-9]|[3-9][1-9])9[5-9]\d{7}$`

| Part | Meaning |
|---|---|
| `\+55` | Brazil country code |
| `([4-9][1-9]\|[3-9][1-9])` | DDD with first digit 3–9 and second digit 1–9 (covers DDDs > 30 like 31, 41, 47, 51, 61, 71, etc.) |
| `9` | The extra mobile digit |
| `[5-9]` | First subscriber digit for mobile numbers |
| `\d{7}` | Remaining subscriber digits |
| Total length | 14 chars including `+` |

**Examples:**

| Input (9-digit) | Output (8-digit) | DDD | Eligible? |
|---|---|---|---|
| `+5547976090095` | `+554767609095` | 47 (SC) | ✅ |
| `+5521967609095` | not eligible in this function | 21 (RJ) | ❌ |
| `+552140627711` | fixed line | 21 (RJ) | ❌ |
| `+554767022587` | already 8-digit form | 47 (SC) | n/a |

> **Note:** DDD `11` (São Paulo): first digit `1`, second digit `1` → `([4-9][1-9]|[3-9][1-9])` does NOT match `11` (first digit must be 3–9). This means SP capital is excluded by design, as it was migrated early and accounts there are expected to already be 9-digit.

> **Mobile rule:** digit-9 transformation applies only when subscriber starts with `5-9`. Fixed-line numbers (subscriber `2-4`) are never transformed.

### LID Store Augmentation (Implemented Workaround)

Files:
- `src/library/phone.go`
- `src/whatsmeow/whatsmeow_contact_manager.go`

To reduce `PN -> LID` misses on migrated accounts, QuePasa now augments whatsmeow's persistent LID store with both BR mobile variants:

- `AddDigit9BRAllDDDs` and `RemoveDigit9BRAllDDDs` are used for **LID store only**.
- They accept **all Brazilian DDDs** (`11-99`) but still enforce **mobile-only** subscriber prefix (`5-9`).
- This is safe because extra unused mappings in `Store.LIDs` are harmless.

Implemented behavior in `GetLIDFromPhone` / `GetPhoneFromLID`:

1. On a successful mapping, persist the alternate BR mobile variant in `Store.LIDs` and cache both forms in memory.
2. If direct `GetLIDForPN` lookup fails, try the BR variant (`add 9` or `remove 9`) before returning empty.
3. If variant lookup succeeds, backfill the original form into `Store.LIDs` to self-heal future lookups.

This solves cases like:
- `+5521967609095` <-> `+552167609095`
- `+5547967022587` <-> `+554767022587`

while still leaving fixed lines untouched, e.g. `+552140627711`.

### Send Flow with `REMOVEDIGIT9=true`

File: `src/models/server_messaging.go` → `QpWhatsappServer.SendMessage()`

```go
if ENV.ShouldRemoveDigit9() {
    phone, _ := whatsapp.GetPhoneIfValid(msg.Chat.Id)
    if len(phone) > 0 {
        phoneWithout9, _ := library.RemoveDigit9IfElegible(phone)
        if len(phoneWithout9) > 0 {
            valids, err := contactManager.IsOnWhatsApp(phone, phoneWithout9)
            // IsOnWhatsApp queries WhatsApp for both versions simultaneously
            for _, valid := range valids {
                msg.Chat.Id = valid  // use whichever version WhatsApp confirms
                break
            }
        }
    }
}
```

**Steps:**
1. Extract and validate phone from `msg.Chat.Id`
2. Check if it is eligible (14-char Brazilian mobile with extra 9)
3. If eligible, generate the 8-digit variant
4. Query `IsOnWhatsApp` with **both** the 9-digit and 8-digit versions
5. Use whichever version WhatsApp confirms as registered — first result wins

### Interaction with LID

This digit-9 normalization runs **before** the whatsmeow send call, at the QuePasa application layer.

On accounts with `LIDMigrationTimestamp > 0`, after the digit-9 resolution resolves the correct PN (`@s.whatsapp.net`), whatsmeow will then convert that PN to `@lid` transparently before hitting the wire — so the two mechanisms compose cleanly:

```
Input chatid (9-digit)
    → REMOVEDIGIT9 check (QuePasa layer)
        → IsOnWhatsApp confirms correct PN variant
            → send to correct @s.whatsapp.net
                → whatsmeow LIDMigration converts to @lid (if migrated)
                    → wire: to="opaque@lid"
```

### Summary

| Scenario | Behavior |
|---|---|
| `REMOVEDIGIT9=false` (default) | No normalization; send to exact chatid as received |
| `REMOVEDIGIT9=true`, number not eligible (DDD ≤ 30, fixed-line, or already 8-digit) | No normalization; send as-is |
| `REMOVEDIGIT9=true`, number eligible, 9-digit confirmed on WhatsApp | Send to 9-digit |
| `REMOVEDIGIT9=true`, number eligible, 8-digit confirmed on WhatsApp | Send to 8-digit (legacy) |
| `REMOVEDIGIT9=true`, neither version found | `IsOnWhatsApp` returns empty; original chatid used |

---

## Summary

| Question | Answer |
|---|---|
| Does whatsmeow convert `@lid` → `@s.whatsapp.net` on send? | **No.** Never. |
| Does whatsmeow convert `@s.whatsapp.net` → `@lid` on send? | **Yes**, but only when `LIDMigrationTimestamp > 0`. |
| What sets `LIDMigrationTimestamp`? | WhatsApp server via LID sync or history sync. |
| What happens if mapping is missing on migrated account? | Send fails (no silent fallback). |
| Is `Wid` in the send response the destination? | **No.** It is the sender's WID (`server.GetWId()`). |
| Are `@lid` and `@s.whatsapp.net` handled by the same send function? | **Yes** — both route to `sendDM()`. |
