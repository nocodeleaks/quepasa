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
- **Status**: ✅ IMPLEMENTED
- **Endpoint**: `POST /messages/react`
- **Request**: `{"chatid": "...", "messageid": "...", "fromme": true, "emoji": "👍"}`
- **Remove reaction**: send with `emoji=""` (WhatsApp clears it)

**Files created/modified**:
  - [x] `src/whatsmeow/whatsmeow_extensions+reactions.go` — `SendReaction(chatID, msgID, fromMe, emoji)` on `WhatsmeowConnection`
  - [x] `src/whatsapp/whatsapp_connection_interface.go` — `SendReaction` added to `IWhatsappConnection`
  - [x] `src/api/api_handlers+ReactionsController.go` — `SendReactionController`, `ReactionRequest` DTO
  - [x] `src/api/api_routes_messages.go` — `POST /messages/react` route registered
  - [x] Swagger regenerated

**Checklist**:
  - [x] `whatsmeow_extensions+reactions.go` with `SendReaction`
  - [x] `api_handlers+ReactionsController.go`
  - [x] Route registered in routes file
  - [x] Swagger regenerated

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

#### 4. **🔐 Block/Unblock Contacts** ✅ IMPLEMENTED
- **Status**: Implemented
- **Current state**: `BlockContact`/`UnblockContact` in interface, `WhatsmeowContactManager`, `QpContactManager`, and HTTP controller
- **Complexity**: Low
- **Impact**: Control access and privacy

**Files created/modified**:
  - [x] `src/whatsmeow/whatsmeow_contact_manager.go` — `BlockContact` and `UnblockContact` using `Client.UpdateBlocklist`
  - [x] `src/whatsmeow/whatsmeow_contact_manager_store.go` — stub methods (store-only access returns error)
  - [x] `src/whatsapp/whatsapp_contact_manager_interface.go` — interface method signatures
  - [x] `src/models/qp_contact_manager.go` — delegation to underlying contact manager
  - [x] `src/api/api_handlers+BlockController.go` — `BlockContactController` and `UnblockContactController`
  - [x] `src/api/api_routes_contacts.go` — routes registered
- **Endpoints**:
  - [x] `POST /contacts/block` — body `{wid: "...@s.whatsapp.net"}`
  - [x] `DELETE /contacts/block` — body `{wid: "...@s.whatsapp.net"}`

**Test**: Block contact, verify status in WhatsApp; unblock, verify restored

---

### ⚠️ **PRIORITY MEDIUM** - Enhancements & partial features

#### 5. **📱 Ephemeral Messages (Disappearing)**

**Architecture analysis**:


**⚠️ Risks**:

**Test**: Send disappearing message, verify `ExpiresAt > 0` in webhook payload
#### 5. **📱 Ephemeral Messages (Disappearing)**
- **Status**: ✅ IMPLEMENTED
- **Current state**: `ExpiresAt` added to `WhatsappMessage`; normal flow uses `evt.IsEphemeral`; fallback `HandleEphemeralMessage` recursively processes inner message
- **Complexity**: Low
- **Impact**: Properly flag disappearing messages in webhooks with expiry timestamp

**Architecture notes**:
- whatsmeow auto-unwraps `EphemeralMessage` and sets `evt.IsEphemeral = true` in the normal flow
- `extractExpirationFromMessage()` checks `ContextInfo.Expiration` (seconds) across all common message types
- `ExpiresAt = message.Timestamp.Unix() + int64(expiration)` — absolute unix timestamp
- Fallback `HandleEphemeralMessage` handles edge cases (history sync, re-requested messages) by recursively calling `HandleKnowingMessages` then setting `ExpiresAt`
#### 6. **🔔 WhatsApp Status/Stories Support**
- **Status**: ✅ IMPLEMENTED
- **Current state**: `PublishStatus(text, attachment)` added to `IWhatsappConnection` and `WhatsmeowConnection`; sends to `types.StatusBroadcastJID`; text-only and media (image/video) stories supported; `UserAbout` and `UserStatusMute` events registered in router

**Files created/modified**:
  - [x] `src/whatsmeow/whatsmeow_extensions+status.go` — `PublishStatus` on `WhatsmeowConnection`
  - [x] `src/whatsapp/whatsapp_connection_interface.go` — `PublishStatus` added to `IWhatsappConnection`
  - [x] `src/api/api_handlers+StatusController.go` — `PublishStatusController`, `StatusPublishRequest` DTO
  - [x] `src/api/api_routes_status.go` — `POST /status/publish` route registered
  - [x] `src/whatsmeow/whatsmeow_event_router.go` — `UserAbout` and `UserStatusMute` events registered
  - [x] Swagger regenerated
- **Endpoints**:
  - [x] `POST /status/publish` — body `{"text": "...", "attachment": {...}}`

**⚠️ Notes**:
- Privacy settings of the account control who sees the status — publishing succeeds but visibility depends on account privacy settings
- Status expiry (24h) is handled by WhatsApp server
- Media types: image and video supported; audio/document not typical for status

#### 7. **🎫 Group Invitation Links**
- **Status**: ✅ IMPLEMENTED
- **Current state**: `GetInvite()` and `RevokeInvite()` both implemented and exposed via API
- **Complexity**: Very low
- **Impact**: Better group invite workflows

**Files created/modified**:
  - [x] `src/whatsapp/whatsapp_group_manager_interface.go` — `RevokeInvite` added to interface
  - [x] `src/whatsmeow/whatsmeow_group_manager.go` — `RevokeInvite` calls `GetGroupInviteLink(ctx, jid, true)`
  - [x] `src/models/qp_group_manager.go` — `RevokeInvite` delegates to underlying group manager
  - [x] `src/api/api_handlers+SPAGroupController.go` — `SPAGroupRevokeInviteController`
  - [x] `src/api/api_handlers+GroupsController.go` — `GetGroupInviteLinkController` + `RevokeGroupInviteLinkController` with Swagger
  - [x] `src/api/api_spa_routes.go` — `DELETE /server/{token}/group/{groupid}/invite`
  - [x] `src/api/api_routes_groups.go` — `DELETE /groups/invite` canonical alias
  - [x] Swagger regenerated
- **Endpoints**:
  - [x] `GET /groups/invite?groupId=xxx` — Get current invite link
  - [x] `DELETE /groups/invite?groupId=xxx` — Revoke link (generates new one)

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
- [x] Contact Block/Unblock — add to ContactManager + controller
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
