# TODO - Whatsmeow Features Implementation Roadmap

## 📋 Task Objective
Implement missing WhatsApp features from the whatsmeow library that are not yet exposed in the QuePasa API. Currently at **~62% feature coverage** (32 implemented vs 20+ not implemented).

---

## �️ Architecture Overview

### Pipeline (Inbound messages)

```
WebSocket (whatsmeow) 
  → EventsHandler()          [whatsmeow_handlers.go]
  → Message() / Receipt()    [per event type dispatcher]
  → HandleKnowingMessages()  [whatsmeow_handlers_message_extensions.go]
  → Follow()                 [whatsmeow_handlers.go]
  → WAHandlers.Message()     [dispatch to webhooks/RabbitMQ]
```

**Critical rule**: if `message.Type == UnhandledMessageType` the message is **discarded before Follow()** — it never reaches webhooks.

### Extension Pattern (established by poll / interactive)

Every new send capability follows this file template:
```
src/whatsmeow/whatsmeow_extensions+<feature>.go
```
Each builds a `waE2E.Message` struct and is called from `WhatsmeowConnection.Send()` via type-switch.

### Reusable Infrastructure

| Helper | File | Purpose |
|--------|------|---------|
| `GetMediaTypeFromAttachment()` | whatsmeow_extensions.go | Infer media type |
| `NewWhatsmeowMessageAttachment()` | whatsmeow_extensions.go | Build message with upload |
| `GetInReplyContextInfo()` | whatsmeow_connection.go | Add reply context |
| `GetContextInfo()` | whatsmeow_connection.go | Process mentions |
| `PhoneToJID()` | whatsmeow_extensions.go | Convert phone → JID |
| `ImproveTimestamp()` | whatsmeow_extensions.go | Normalize timestamp |

### ⚠️ Universal Critical Rule — MessageSecret

Every outgoing `waE2E.Message` **must** include:
```go
MessageContextInfo: &waE2E.MessageContextInfo{
    MessageSecret: random.Bytes(32), // MANDATORY — silent failure without it
}
```

---

## �🎯 Priority Categories

### 🚨 **PRIORITY HIGH** - Quick wins with high user value

#### 1. **📤 Message Reactions (Send)**
- **Status**: Not started
- **Current state**: Receiving reactions works; sending emoji reactions is missing
- **Complexity**: Low
- **Impact**: Users can automate emoji reactions to messages

**Architecture analysis**:
- Receive already works: `HandleReactionMessage()` sets `out.InReaction=true`, `out.InReply=msgID`, `out.Text=emoji`
- Fields already exist in `WhatsappMessage`: `InReaction bool`, `InReply string`, `Text string`
- Send: build `waE2E.ReactionMessage{Text: emoji, Key: {ID: msgID}}` and call `Client.SendMessage()`

- **Files to create/modify**:
  - [ ] `src/whatsmeow/whatsmeow_extensions+reactions.go` — `SendReaction(chatID, msgID, emoji)` + `RemoveReaction(chatID, msgID)`
  - [ ] `src/api/api_handlers+ReactionsController.go` — API endpoint
  - [ ] `src/api/` routes file — register route
- **Endpoint**: `POST /api/messages/{messageId}/react` — body: `{token, chatid, emoji}`
- **Remove reaction**: send with `emoji=""` (WhatsApp clears it)

**Checklist**:
- [ ] `whatsmeow_extensions+reactions.go` with `SendReaction` and `RemoveReaction`
- [ ] `api_handlers+ReactionsController.go`
- [ ] Route registered in routes file
- [ ] Regenerate Swagger after

**⚠️ Risks**:
- Emoji can be multi-codepoint unicode (e.g., family emoji = 7 codepoints) — validate as rune, not byte length
- Removing reaction with `emoji=""` may silently fail on older WhatsApp versions

**Test**: Send reaction → webhook receives `InReaction=true`, `InReply=<msgID>`, `Text=<emoji>`

#### 2. ~~**🔗 Message Forwarding**~~ — **OUT OF SCOPE**
- **Status**: Will NOT be implemented (by design decision)
- **Reason**: No intention to implement forwarding functionality in QuePasa
- **Current state**: whatsmeow has `Client.ForwardMessage()` available but will not be used

#### 3. **🌐 Broadcast Lists Support**
- **Status**: Not started
- **Current state**: Broadcast messages are filtered/ignored in handlers
- **Complexity**: Medium
- **Impact**: Send to multiple contacts with individual privacy

**Architecture analysis**:
- Broadcasts are **explicitly discarded** in `whatsmeow_handlers.go` line ~865:
  ```go
  if strings.Contains(chatID, "@broadcast") { /* DISCARDED */ }
  ```
- To receive broadcasts in webhooks: remove that filter (design decision needed)
- To send: target JID is `xxx@broadcast` — works like a normal `SendMessage` call
- `Client.CreateBroadcastList()` exists in whatsmeow for creating lists

**⚠️ Pending decision**: Do we want to receive broadcast messages in webhooks, or only send?

- **Files to create/modify**:
  - [ ] `src/whatsmeow/whatsmeow_handlers.go` — remove/conditionalize `@broadcast` filter
  - [ ] `src/whatsmeow/whatsmeow_extensions+broadcast.go` — broadcast operations
  - [ ] `src/api/api_handlers+BroadcastController.go` — API endpoints
- **Endpoints**:
  - [ ] `POST /api/broadcasts/create` — Create broadcast list
  - [ ] `POST /api/broadcasts/{id}/send` — Send to broadcast
  - [ ] `GET /api/broadcasts` — List all broadcasts

**⚠️ Risks**:
- Broadcast IDs are not persistent across reconnections — do not persist IDs without re-validation
- Recipients of a broadcast list are not synced automatically; list must be recreated if changed

**Test**: Create broadcast, verify in WhatsApp, send message; confirm webhook receives (if reception enabled)

#### 4. **🔐 Block/Unblock Contacts**
- **Status**: Not started
- **Current state**: Methods exist in whatsmeow (`BlockJID`, `UnblockJID`) but no API
- **Complexity**: Low
- **Impact**: Control access and privacy

**Architecture analysis**:
- `WhatsmeowContactManager` already exists — add 2 methods: `BlockContact(wid)` and `UnblockContact(wid)`
- Parse JID → call `Client.BlockJID()` / `Client.UnblockJID()` → return error
- Add methods to contact manager interface in `src/whatsapp/`

- **Files to create/modify**:
  - [ ] `src/whatsmeow/whatsmeow_contact_manager.go` — add `BlockContact` and `UnblockContact`
  - [ ] `src/whatsapp/` contact interface — add method signatures
  - [ ] `src/api/api_handlers+BlockController.go` — block/unblock endpoints
- **Endpoints**:
  - [ ] `POST /api/contacts/{contactId}/block`
  - [ ] `DELETE /api/contacts/{contactId}/block` (unblock)

**⚠️ Risks**:
- `BlockJID` is NOT idempotent — may return error if contact is already blocked; handle gracefully
- `UnblockJID` same — may error if not blocked

**Test**: Block contact, verify status in WhatsApp; unblock, verify restored

---

### ⚠️ **PRIORITY MEDIUM** - Enhancements & partial features

#### 5. **📱 Ephemeral Messages (Disappearing)**
- **Status**: Detected but not processed
- **Current state**: Handler has comment: `"handling ephemeral message not implemented"`
- **Location**: [whatsmeow_handlers_message_extensions.go](../src/whatsmeow/whatsmeow_handlers_message_extensions.go#L203)
- **Complexity**: Low (but needs investigation first)
- **Impact**: Properly flag and document disappearing messages in webhooks

**Architecture analysis**:
- `HandleEphemeralMessage()` receives `waE2E.FutureProofMessage` — this is a **generic future-proof type**, NOT exclusive to ephemeral/disappearing messages
- May be View Once, Disappearing, or another future type
- Expiration can be in `ContextInfo.Expiration` (seconds) OR `ContextInfo.EphemeralSettingTimestamp` (milliseconds) — inconsistent units

- **Files to modify**:
  - [ ] Investigate what `FutureProofMessage` contains at runtime (add debug logging first)
  - [ ] Add `ExpiresAt int64` field to `WhatsappMessage` model (unix seconds, 0 = never)
  - [ ] `whatsmeow_handlers_message_extensions.go` — implement real handling
  - [ ] Include `ExpiresAt` in webhook payload

**⚠️ Risks**:
- `FutureProofMessage` is a catch-all wrapper — content type must be detected dynamically
- Unit mismatch: `Expiration` field is seconds, `EphemeralSettingTimestamp` is milliseconds

**Test**: Send disappearing message, verify `ExpiresAt > 0` in webhook payload

#### 6. **🔔 WhatsApp Status/Stories Support**
- **Status**: Not started
- **Current state**: No publish or view functionality
- **Complexity**: Medium
- **Impact**: Publish automated status updates

**Architecture analysis**:
- Status is sent to the special JID `status@broadcast`
- Media upload uses the same logic as normal messages but media type must be `whatsmeow.MediaStatusRoomImage` (special constant)
- Receiving status updates requires adding `*events.StatusUpdate` handler in `EventsHandler()` — currently not present

- **Files to create/modify**:
  - [ ] `src/whatsmeow/whatsmeow_extensions+status.go` — `PublishStatus(attachment, caption, expiration)`
  - [ ] `src/whatsmeow/whatsmeow_handlers.go` — add `*events.StatusUpdate` case in `EventsHandler`
  - [ ] `src/api/api_handlers+StatusController.go` — API endpoints
- **Endpoints**:
  - [ ] `POST /api/status/publish` — Publish status with media
  - [ ] `GET /api/status/viewed` — Get viewing notifications
  - [ ] `GET /api/status/list` — List contact statuses

**⚠️ Risks**:
- Privacy settings of the account control who sees the status — publishing succeeds but may be invisible to contacts if privacy settings are restrictive
- Status expiry (24h) is handled by WhatsApp server, no need to manage locally

**Test**: Publish status, verify visibility on another device

#### 7. **🎫 Group Invitation Links**
- **Status**: Almost ready ✅ (backend exists, API endpoint missing)
- **Current state**: `GetInvite()` already exists in `whatsmeow_group_manager.go` — just needs API exposure
- **Complexity**: Very low
- **Impact**: Better group invite workflows — **recommended as first implementation**

**Architecture analysis**:
- `WhatsmeowGroupManager.GetInvite(groupId)` calls `Client.GetGroupInviteLink(ctx, jid, false)` — already works
- Revoke: same method with `revokeCurrentLink=true` parameter
- `GroupsController` already has the server/group resolution pattern to follow

- **Files to modify**:
  - [ ] `src/whatsmeow/whatsmeow_group_manager.go` — add `RevokeInviteLink(groupId)` method
  - [ ] `src/api/api_handlers+GroupsController.go` — add `GetGroupInviteLinkController` and `RevokeGroupInviteLinkController`
  - [ ] Routes file — register endpoints
- **New endpoints**:
  - [ ] `GET /api/groups/{groupId}/invite-link` — Get current invite link
  - [ ] `POST /api/groups/{groupId}/revoke-link` — Revoke link (generates new one)

**Test**: Generate link → open in WhatsApp → join group; revoke → previous link becomes invalid

#### 8. **🔐 Privacy Settings (Granular)**
- **Status**: Not started
- **Current state**: No support for who sees photo/presence/read status
- **Complexity**: Medium
- **Impact**: Better privacy control
- **Files to create/modify**:
  - [ ] `src/whatsmeow/whatsmeow_extensions+privacy.go` - Privacy settings
  - [ ] `src/api/api_handlers+PrivacyController.go` - API endpoint
- **Endpoint**: `PUT /api/account/privacy` with `{whoCanSeePhoto, whoCanSeeStatus, etc}`
- **Test**: Set privacy options, verify in WhatsApp

---

### 📦 **PRIORITY LOW** - Specialized/Future features

#### 9. **💳 Payment Messages Support**
- **Status**: Not started
- **Current state**: WhatsApp supports PaymentInviteMessage, RequestPaymentMessage
- **Complexity**: High (requires payment API integration)
- **Impact**: Limited to WhatsApp Business accounts
- **Files needed**: TBD (payment system integration)
- **Note**: Requires separate payment processing system

#### 10. **👥 Communities Support**
- **Status**: Not started
- **Current state**: WhatsApp Communities API exists but complex implementation
- **Complexity**: High
- **Impact**: Manage communities and subcategories
- **Note**: Significant effort, defer to later release

#### 11. **🌟 Newsletter Support**
- **Status**: Not started
- **Current state**: Ignored in handlers (line 865-866)
- **Complexity**: Medium
- **Impact**: Newsletter distribution features
- **Note**: Similar to broadcast lists but with subscriber management

#### 12. **📞 Call Events (Metadata)**
- **Status**: Events received; no API exposure
- **Current state**: Intentionally omitted per design (`CallManager` excluded)
- **Note**: Handlers exist but not exposed; document why in API docs
- **Complexity**: Low (if just exposing events)

---

## 📊 Implementation Checklist

### Phase 1: High Priority — Quick Wins

| Feature | Effort | Complexity | Recommended Order |
|---------|--------|------------|-------------------|
| Group Invite Link (expose existing) | 1-2h | Very Low | **#1 — already built** |
| Block/Unblock Contacts | 3-4h | Low | **#2** |
| Message Reactions (Send) | 4-5h | Low | **#3** |
| Ephemeral Messages Flag | 3-4h | Low (after investigation) | **#4** |

- [ ] Group Invitation Links — expose `GetInvite()` to API
- [ ] Contact Block/Unblock — add to ContactManager + controller
- [ ] Message Reactions (Send) — `whatsmeow_extensions+reactions.go` + controller
- [ ] Ephemeral Messages — investigate `FutureProofMessage`, add `ExpiresAt` field

### Phase 2: Medium Priority

| Feature | Effort | Notes |
|---------|--------|-------|
| Broadcast Lists | 8-10h | Requires design decision on reception |
| Status/Stories Support | 8-10h | New event handler needed |
| Privacy Settings | 4-6h | — |

- [ ] Broadcast Lists — decision on reception first, then create/send
- [ ] Status/Stories Support — `whatsmeow_extensions+status.go` + event handler
- [ ] Privacy Settings — `whatsmeow_extensions+privacy.go`

### Phase 3: Low Priority / Specialized

- [ ] Payment Messages — 10+ hours (external payment system dependency)
- [ ] Communities — 15+ hours (complex)
- [ ] Newsletters — 8-10 hours

---

## 🔄 Current Status

**Overall Progress**: 32 / 52 features = 62%

### By Category
- ✅ Messaging & Communication: 9/13 (69%)
- ✅ Chat Management: 5/7 (71%)
- ✅ Group Management: 7/10 (70%)
- ✅ Contacts & Presence: 5/7 (71%)
- ✅ Auth & Connection: 4/5 (80%)
- ⚠️ Advanced/Optional: 2/10 (20%)

---

## 📝 Next Steps

1. **Group Invite Links** — expose existing `GetInvite()` to API (1-2h, already built)
2. **Block/Unblock Contacts** — add to ContactManager + controller (3-4h)
3. **Message Reactions (send)** — `whatsmeow_extensions+reactions.go` + controller (4-5h)
4. **Investigate Ephemeral** — add debug logging to `HandleEphemeralMessage`, inspect `FutureProofMessage` at runtime
5. **Decide on Broadcast reception** — define whether received broadcasts should reach webhooks
6. **Create test suite** for new endpoints (TDD approach)
7. **Update API documentation** with new features
8. **Consider v6 API release** to bundle these changes

---

## 🚫 Immutable Constraints Discovered

1. **Call Managers intentionally omitted** — Do not include `CallManager`/`SIPCallManager` in whatsmeow_connection.go
   - Call events ARE received but not exposed; handlers exist in `whatsmeow_handlers+call.go`

2. **Phone number privacy (LID)** — LIDs never contain phone numbers; they are opaque identifiers
   - Mapping must come from whatsmeow DB (`whatsmeow_lid_map` table)
   - Not all LIDs have mappings — expected behavior

3. **Broadcast filtering** — `@broadcast` and `@newsletter` messages are currently discarded in handlers (~line 865)
   - Intentional; changing this requires design decision

4. **Message Forwarding** — Intentionally out of scope; will NOT be implemented
   - `Client.ForwardMessage()` exists in whatsmeow but won't be used

5. **MessageSecret mandatory** — Every outgoing `waE2E.Message` must include `MessageContextInfo.MessageSecret = random.Bytes(32)` or it fails silently

6. **ContextInfo is type-specific** — Not all `waE2E` message types accept `ContextInfo`; adding it to the wrong sub-type is silently ignored

7. **Version management** — `QpVersion` in `models/qp_defaults.go` must be updated on each merge to main
   - Format: `3.YY.MMDD.HHMM`

---

## 📚 Related Files & References

### Core Files
- Connection wrapper: [src/whatsmeow/whatsmeow_connection.go](../src/whatsmeow/whatsmeow_connection.go)
- Event handlers: [src/whatsmeow/whatsmeow_handlers.go](../src/whatsmeow/whatsmeow_handlers.go)
- Message type handlers: [src/whatsmeow/whatsmeow_handlers_message_extensions.go](../src/whatsmeow/whatsmeow_handlers_message_extensions.go)
- Group manager: [src/whatsmeow/whatsmeow_group_manager.go](../src/whatsmeow/whatsmeow_group_manager.go)
- Contact manager: [src/whatsmeow/whatsmeow_contact_manager.go](../src/whatsmeow/whatsmeow_contact_manager.go)
- Extensions (send patterns): [src/whatsmeow/whatsmeow_extensions+poll.go](../src/whatsmeow/whatsmeow_extensions+poll.go), [whatsmeow_extensions+interactive.go](../src/whatsmeow/whatsmeow_extensions+interactive.go)

### API Patterns
- API handlers: [src/api/api_handlers.go](../src/api/api_handlers.go)
- Controller example: [src/api/api_handlers+GroupsController.go](../src/api/api_handlers+GroupsController.go)

### Instructions
- Main architecture: [.github/copilot-instructions.md](../.github/copilot-instructions.md)

---

**Last Updated**: 29 de abril de 2026  
**Status**: Planning Phase — Phase 1 recommended start: Group Invite Links
