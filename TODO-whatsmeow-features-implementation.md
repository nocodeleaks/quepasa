# TODO - Whatsmeow Features Implementation Roadmap

## 📋 Task Objective
Implement missing WhatsApp features from the whatsmeow library that are not yet exposed in the QuePasa API. Currently at **~62% feature coverage** (32 implemented vs 20+ not implemented).

---

## 🎯 Priority Categories

### 🚨 **PRIORITY HIGH** - Quick wins with high user value

#### 1. **📤 Message Reactions (Send)**
- **Status**: Not started
- **Current state**: Receiving reactions works; sending emoji reactions is missing
- **Complexity**: Low
- **Impact**: Users can automate emoji reactions to messages
- **Files to create/modify**:
  - [ ] `src/whatsmeow/whatsmeow_extensions+reactions.go` - SendReaction method
  - [ ] `src/api/api_handlers+ReactionsController.go` - API endpoint
- **Endpoint**: `POST /api/messages/{messageId}/react` with `{token, emoji}`
- **Test**: Manual verification via API endpoint

#### 2. ~~**🔗 Message Forwarding**~~ — **OUT OF SCOPE**
- **Status**: Will NOT be implemented (by design decision)
- **Reason**: No intention to implement forwarding functionality in QuePasa
- **Current state**: whatsmeow has `Client.ForwardMessage()` available but will not be used

#### 3. **🌐 Broadcast Lists Support**
- **Status**: Not started
- **Current state**: Broadcast messages are filtered/ignored in handlers
- **Complexity**: Medium
- **Impact**: Send to multiple contacts with individual privacy
- **Files to create/modify**:
  - [ ] `src/whatsmeow/whatsmeow_extensions+broadcast.go` - Broadcast operations
  - [ ] `src/api/api_handlers+BroadcastController.go` - API endpoints
- **Endpoints**:
  - [ ] `POST /api/broadcasts/create` - Create broadcast list
  - [ ] `POST /api/broadcasts/{id}/send` - Send to broadcast
  - [ ] `GET /api/broadcasts` - List all broadcasts
- **Test**: Create broadcast, verify in WhatsApp, send message

#### 4. **🔐 Block/Unblock Contacts**
- **Status**: Not started
- **Current state**: Methods exist in whatsmeow (`BlockJID`, `UnblockJID`) but no API
- **Complexity**: Low
- **Impact**: Control access and privacy
- **Files to create/modify**:
  - [ ] `src/api/api_handlers+BlockController.go` - Block/unblock endpoints
- **Endpoints**:
  - [ ] `POST /api/contacts/{contactId}/block`
  - [ ] `DELETE /api/contacts/{contactId}/block` (unblock)
- **Test**: Block contact, verify status in WhatsApp

---

### ⚠️ **PRIORITY MEDIUM** - Enhancements & partial features

#### 5. **📱 Ephemeral Messages (Disappearing)**
- **Status**: Detected but not processed
- **Current state**: Handler has comment: `"handling ephemeral message not implemented"`
- **Location**: [whatsmeow_handlers_message_extensions.go](z:\Desenvolvimento\nocodeleaks-quepasa\src\whatsmeow\whatsmeow_handlers_message_extensions.go#L203)
- **Complexity**: Low
- **Impact**: Properly flag and document disappearing messages
- **Files to modify**:
  - [ ] Add `IsEphemeral` flag to QpWhatsappMessage model
  - [ ] Implement message type detection
  - [ ] Include in webhook payload with flag
- **Test**: Send disappearing message, verify flag in webhook

#### 6. **🔔 WhatsApp Status/Stories Support**
- **Status**: Not started
- **Current state**: No publish or view functionality
- **Complexity**: Medium
- **Impact**: Publish automated status updates
- **Files to create/modify**:
  - [ ] `src/whatsmeow/whatsmeow_extensions+status.go` - Status operations
  - [ ] `src/api/api_handlers+StatusController.go` - API endpoints
- **Endpoints**:
  - [ ] `POST /api/status/publish` - Publish status with media
  - [ ] `GET /api/status/viewed` - Get viewing notifications
  - [ ] `GET /api/status/list` - List contact statuses
- **Test**: Publish status, verify visibility

#### 7. **🎫 Group Invitation Links**
- **Status**: Partially implemented
- **Current state**: Basic group management works but link generation is incomplete
- **Complexity**: Medium
- **Impact**: Better group invite workflows
- **Files to modify**:
  - [ ] `src/whatsmeow/whatsmeow_group_manager.go` - Add link generation
  - [ ] `src/api/api_handlers+GroupsController.go` - Enhance endpoints
- **New endpoints**:
  - [ ] `POST /api/groups/{groupId}/invite-link` - Generate invite link
  - [ ] `POST /api/groups/{groupId}/revoke-link` - Revoke link
- **Test**: Generate link, verify usability in WhatsApp

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

### Phase 1: High Priority (Weeks 1-2)
- [ ] Message Reactions (Send) - Lines ~1, Estimate: 4-6 hours
- [ ] Contact Block/Unblock - Estimate: 3-4 hours
- [ ] Ephemeral Messages Flag - Estimate: 2-3 hours

### Phase 2: Medium Priority (Weeks 3-4)
- [ ] Broadcast Lists - Estimate: 8-10 hours
- [ ] Status/Stories Support - Estimate: 6-8 hours
- [ ] Group Invitation Links - Estimate: 4-6 hours

### Phase 3: Low Priority (Weeks 5+)
- [ ] Privacy Settings - Estimate: 4-6 hours
- [ ] Payment Messages - Estimate: 10+ hours (depends on payment system)
- [ ] Communities - Estimate: 15+ hours (complex feature)
- [ ] Newsletters - Estimate: 8-10 hours

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

1. **Create reaction endpoint** first (Phase 1, quick win)
2. **Add ephemeral flag** to message model (Phase 1, minimal effort)
3. **Block/unblock contacts** API (Phase 1, high value)
4. **Plan Phase 2** once Phase 1 is complete
5. **Create test suite** for new endpoints (TDD approach)
6. **Update API documentation** with new features
7. **Consider v6 API release** to bundle these changes

---

## 🚫 Immutable Constraints Discovered

1. **Call Managers intentionally omitted** - Per design decision, do not include CallManager/SIPCallManager in whatsmeow_connection.go
   - Reason: Documented in copilot-instructions.md
   - Current state: Call events ARE received and can be exposed if needed, but not implemented

2. **Phone number privacy (LID)** - LIDs never contain phone numbers; they are opaque identifiers
   - Mapping must come from whatsmeow database (whatsmeow_lid_map table)
   - This constraint applies to all contact-related features

3. **Broadcast vs Direct message** - Need to handle broadcast list metadata separately from regular messages

4. **Message Forwarding** - Intentionally out of scope; will NOT be implemented
   - `Client.ForwardMessage()` is available in whatsmeow but won't be used

5. **Version management** - QpVersion in models/qp_defaults.go must be updated on each merge to main
   - Format: 3.YY.MMDD.HHMM (e.g., 3.25.1114.2100)

---

## 📚 Related Files & References

- Main instruction: `/github/copilot-instructions.md`
- Current handlers: `src/whatsmeow/whatsmeow_handlers*.go`
- Connection wrapper: `src/whatsmeow/whatsmeow_connection.go`
- API patterns: `src/api/api_handlers+*.go`
- Group management: `src/whatsmeow/whatsmeow_group_manager.go`
- Contact manager: `src/whatsmeow/whatsmeow_contact_manager.go`
- Message extensions: `src/whatsmeow/whatsmeow_extensions+*.go`

---

## 📞 Communication Notes

- **Status**: Responses in pt-BR, code/comments in English
- **Branch**: develop
- **API versioning**: Latest (non-versioned) routes preferred for new features
- **Build command**: `go build -o .dist/quepasa.exe`
- **Swagger regeneration**: Required after any API endpoint changes (`swag init --output ./swagger`)

---

**Last Updated**: 29 de abril de 2026  
**Created by**: Code Assistant Analysis  
**Status**: Planning Phase
