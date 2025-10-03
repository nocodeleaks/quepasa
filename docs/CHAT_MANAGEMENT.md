# Chat Management - Complete Documentation

## Table of Contents
1. [API Documentation](#api-documentation)
2. [Technical Implementation](#technical-implementation)
3. [Known Issues and Workarounds](#known-issues-and-workarounds)
4. [Troubleshooting](#troubleshooting)

---

## API Documentation

### Endpoints Overview

#### 1. Mark Chat as Read
**Endpoint:** `POST /chat/markread`

Marks a WhatsApp chat as read, removing the unread badge.

**Request:**
```json
{
  "chatid": "5511999999999"
}
```

**Success Response (200):**
```json
{
  "success": true,
  "message": "chat 5511999999999@s.whatsapp.net marked as read"
}
```

---

#### 2. Mark Chat as Unread
**Endpoint:** `POST /chat/markunread`

Marks a WhatsApp chat as unread, showing the unread badge.

**Request:**
```json
{
  "chatid": "5511999999999"
}
```

**Success Response (200):**
```json
{
  "success": true,
  "message": "chat 5511999999999@s.whatsapp.net marked as unread"
}
```

---

#### 3. Archive/Unarchive Chat
**Endpoint:** `POST /chat/archive`

Archives or unarchives a WhatsApp chat.

**Request:**
```json
{
  "chatid": "5511999999999",
  "archive": true
}
```

**Parameters:**
- `chatid` (string, required): The chat ID
- `archive` (boolean, required): `true` to archive, `false` to unarchive

**Success Response (200):**
```json
{
  "success": true,
  "message": "chat 5511999999999@s.whatsapp.net archived successfully"
}
```

---

### Chat ID Format

Accepts multiple formats:
- Phone number: `5511999999999`
- With country code: `+5511999999999`
- Full JID: `5511999999999@s.whatsapp.net`
- Group JID: `120363XXXXXXXXXX@g.us`

### Authentication

All endpoints require authentication via API key:
```
X-QUEPASA-TOKEN: your-api-key-here
```

---

### Usage Examples

#### cURL Examples

**Mark as read:**
```bash
curl -X POST https://your-server/chat/markread \
  -H "Content-Type: application/json" \
  -H "X-QUEPASA-TOKEN: your-api-key" \
  -d '{"chatid": "5511999999999"}'
```

**Mark as unread:**
```bash
curl -X POST https://your-server/chat/markunread \
  -H "Content-Type: application/json" \
  -H "X-QUEPASA-TOKEN: your-api-key" \
  -d '{"chatid": "5511999999999"}'
```

**Archive:**
```bash
curl -X POST https://your-server/chat/archive \
  -H "Content-Type: application/json" \
  -H "X-QUEPASA-TOKEN: your-api-key" \
  -d '{"chatid": "5511999999999", "archive": true}'
```

#### JavaScript/Node.js Example

```javascript
const axios = require('axios');

const baseURL = 'https://your-server';
const apiKey = 'your-api-key';

async function markChatAsRead(chatId) {
  const response = await axios.post(
    `${baseURL}/chat/markread`,
    { chatid: chatId },
    { headers: { 'X-QUEPASA-TOKEN': apiKey } }
  );
  return response.data;
}

async function markChatAsUnread(chatId) {
  const response = await axios.post(
    `${baseURL}/chat/markunread`,
    { chatid: chatId },
    { headers: { 'X-QUEPASA-TOKEN': apiKey } }
  );
  return response.data;
}

async function archiveChat(chatId, archive = true) {
  const response = await axios.post(
    `${baseURL}/chat/archive`,
    { chatid: chatId, archive },
    { headers: { 'X-QUEPASA-TOKEN': apiKey } }
  );
  return response.data;
}
```

#### Python Example

```python
import requests

base_url = 'https://your-server'
api_key = 'your-api-key'
headers = {'X-QUEPASA-TOKEN': api_key, 'Content-Type': 'application/json'}

def mark_chat_as_read(chat_id):
    response = requests.post(
        f'{base_url}/chat/markread',
        json={'chatid': chat_id},
        headers=headers
    )
    return response.json()

def mark_chat_as_unread(chat_id):
    response = requests.post(
        f'{base_url}/chat/markunread',
        json={'chatid': chat_id},
        headers=headers
    )
    return response.json()

def archive_chat(chat_id, archive=True):
    response = requests.post(
        f'{base_url}/chat/archive',
        json={'chatid': chat_id, 'archive': archive},
        headers=headers
    )
    return response.json()
```

---

## Technical Implementation

### App State Protocol

All chat management operations use WhatsApp's **App State Protocol**:

- **Mark as Read/Unread**: `appstate.BuildMarkChatAsRead(jid, read, timestamp, messageKey)`
- **Archive/Unarchive**: `appstate.BuildArchive(jid, archive, timestamp, messageKey)`

### App State Types

| Operation | App State Type | Priority |
|-----------|---------------|----------|
| Mark as Read/Unread | `regular_low` | Low |
| Archive/Unarchive | `regular_low` | Low |
| Mute/Unmute | `regular_high` | High |
| Pin/Unpin | `regular_low` | Low |

### Implementation Flow

```go
// 1. Parse chat JID
jid, err := types.ParseJID(chatId)

// 2. Build app state patch
patch := appstate.BuildMarkChatAsRead(jid, true, time.Time{}, nil)

// 3. Send to WhatsApp
ctx := context.Background()
err = client.SendAppState(ctx, patch)
```

### Multi-Device Sync

Changes are automatically synchronized across all connected devices using WhatsApp's App State system.

---

## Known Issues and Workarounds

### WhatsApp App State Bug (Issue #858)

#### Overview

There is a **known bug in the whatsmeow library** that affects app state mutations. This is **NOT a bug in QuePasa** - it's an upstream library issue.

#### GitHub Issues
- **Primary**: [tulir/whatsmeow#858](https://github.com/tulir/whatsmeow/issues/858) - "mismatching LTHash"
- **Related**: [tulir/whatsmeow#382](https://github.com/tulir/whatsmeow/issues/382) - 409 conflicts
- **Status**: **OPEN** (as of 2024-07-22)

#### Duplicate Issues
The same bug reported in: #686, #813, #508, #518, #651

#### Root Cause

According to **Tulir** (whatsmeow author):
- `ErrMismatchingLTHash` - Failed to verify patch hash
- LTHash (Long Term Hash) verification fails when decoding app state patches
- This **probably causes** 409 conflict errors

**Technical Details:**
```go
// From appstate/decode.go in whatsmeow
func validateSnapshotMAC(currentState HashState, keyID, expectedSnapshotMAC []byte) error {
    snapshotMAC := currentState.generateSnapshotMAC(name, keys.SnapshotMAC)
    if !bytes.Equal(snapshotMAC, expectedSnapshotMAC) {
        return ErrMismatchingLTHash // ERROR HERE
    }
}
```

**Error sequence:**
1. Client sends app state patch (e.g., archive chat)
2. Server responds with 409 conflict (version mismatch)
3. Client attempts to resync by fetching patches
4. Patch verification fails: calculated LTHash ≠ expected LTHash
5. Operation fails

---

### Error Types

#### 1. Conflict Error (409)

**What it is:**
- Local app state version is outdated
- Another device made changes that haven't synced
- Version mismatch between client and server

**Error example:**
```xml
<error code="409" text="conflict"/>
```

**API response:**
```json
{
  "success": false,
  "status": "server returned error updating app state: conflict"
}
```

#### 2. LTHash Mismatch Error

**Error message:**
```
failed to verify patch v121: mismatching LTHash
```

**What it means:**
- Local app state hash doesn't match server's hash
- App state got out of sync
- Multiple devices making rapid changes

---

### Current Implementation Strategy

**QuePasa uses a SIMPLE approach**: Return errors directly to the user without automatic retry.

```go
func sendAppState(conn *WhatsmeowConnection, patch appstate.PatchInfo) error {
    ctx := context.Background()
    return conn.Client.SendAppState(ctx, patch) // Direct pass-through
}
```

**Why this approach:**
- ✅ Simple and predictable
- ✅ User knows immediately if operation failed
- ✅ No hidden retry delays
- ✅ User can decide when to retry

**Error handling:**
```json
{
  "success": false,
  "status": "server returned error: conflict"
}
```

---

### Alternative Approaches (Not Implemented)

#### Option 1: Automatic Retry with Full Resync

#### tested and don't work reliably

```go
func sendAppStateWithRetry(conn *WhatsmeowConnection, patch appstate.PatchInfo) error {
    err := conn.Client.SendAppState(ctx, patch)
    
    if strings.Contains(err.Error(), "conflict") || strings.Contains(err.Error(), "409") {
        // Full resync with snapshot
        conn.Client.FetchAppState(ctx, patch.Type, true, false)
        time.Sleep(1000 * time.Millisecond)
        
        // Retry
        return conn.Client.SendAppState(ctx, patch)
    }
    return err
}
```

**Pros:**
- Automatic recovery
- User doesn't see temporary failures

**Cons:**
- Adds 1+ second latency on conflicts
- Hidden complexity
- May mask underlying issues

#### Option 2: Message-Level Marking (Actually Implemented for Mark Read)

Instead of marking the entire chat state, mark individual messages:

```go
// Requires message IDs
func (source *QpWhatsappServer) MarkRead(id string) (err error) {
	msg, err := source.Handler.GetById(id)
	if err != nil {
		return
	}
	source.GetLogger().Infof("marking msg %s as read", id)
	return source.connection.MarkRead(msg)
}
```

**Pros:**
- More stable (no app state conflicts)

**Cons:**
- Requires knowing message IDs first
- Can't mark entire chat at once

---

## Troubleshooting

### Issue: 409 Conflict Error

**Symptoms:**
```json
{
  "success": false,
  "status": "server returned error: conflict"
}
```
---

### Issue: "mismatching LTHash" Error

**Symptoms:**
```
failed to decode app state patches: mismatching LTHash
```

**Causes:**
- App state hash out of sync
- Multiple devices making changes simultaneously

**Solutions:**
1. **Reconnect WhatsApp**: Disconnect and reconnect the bot
2. **Clear app state**: May require re-pairing device
3. **Wait for sync**: Sometimes resolves itself

---


---

## Monitoring and Logs

### Log Patterns

**Successful operation:**
```
[INFO] marked chat 5511999999999@s.whatsapp.net as read
```

**Conflict error:**
```
[ERROR] failed to mark chat as read: server returned error: conflict
```

**LTHash error:**
```
[ERROR] failed to decode app state patches: mismatching LTHash
```

### Metrics to Track

- Number of 409 conflicts per hour
- Operation success rate
- Average response time
- Failed operations by error type

---

## Future Improvements

### When Whatsmeow Fixes Bug #858

1. ✅ Monitor issue #858 for resolution
2. ✅ Update dependency: `go get -u go.mau.fi/whatsmeow`
3. ✅ Test if errors still occur
4. ⚠️ Consider implementing automatic retry if upstream fix is robust
5. ✅ Update this documentation

### Potential Enhancements

1. **Automatic retry** with exponential backoff
2. **Queue-based retry** for failed operations
3. **Client-side deduplication** to prevent duplicate requests
4. **Message-level marking** as alternative to app state

---

## References

- [WhatsApp Multi-Device App State Protocol](https://github.com/tulir/whatsmeow/blob/main/appstate)
- [Issue #858 - LTHash Mismatch](https://github.com/tulir/whatsmeow/issues/858)
- [Issue #382 - 409 Conflicts](https://github.com/tulir/whatsmeow/issues/382)

---

## Version History

- **v3.25.2207.0128**: Initial implementation with direct error pass-through
- **Future**: May add automatic retry when upstream bug is fixed

---

## FAQ

**Q: Why do I get 409 errors?**  
A: This is a known bug in whatsmeow library (#858). WhatsApp synchronizes state across devices, and conflicts occur when multiple devices update simultaneously.

**Q: What's the difference between app state and message-level marking?**  
A: App state marks the entire chat state, message-level marks specific messages. Both valid approaches.

**Q: Is this safe to use in production?**  
A: Yes, but implement proper error handling and retry logic in your client code.

**Q: Will this be fixed?**  
A: Waiting for whatsmeow library to fix issue #858. Monitor upstream for updates.

---

**Last Updated:** October 2025  
**QuePasa Version:** 3.25.1003.1248
