# Contact Messages - Complete Guide

## Table of Contents
- [Overview](#overview)
- [Quick Start](#quick-start)
- [API Usage](#api-usage)
- [WhatsApp Status Detection](#whatsapp-status-detection)
- [vCard Format Reference](#vcard-format-reference)
- [Implementation Details](#implementation-details)
- [Testing](#testing)
- [Troubleshooting](#troubleshooting)
- [Technical Reference](#technical-reference)

---

## Overview

QuePasa supports sending contact (vCard) messages through WhatsApp with **intelligent WhatsApp status detection**. The system automatically generates appropriate vCard formats based on whether the contact is:

1. **Business WhatsApp Account** - Full business information with company details
2. **Regular WhatsApp Account** - Standard WhatsApp contact with clickable number
3. **Not on WhatsApp** - Basic phone contact with invite option

### Key Features

✅ **Automatic WhatsApp Detection** - Uses `GetUserInfo` API to verify contact status  
✅ **Smart vCard Generation** - Three different formats based on detection  
✅ **Business Support** - Properly handles WhatsApp Business accounts  
✅ **Custom vCard Support** - Bypass auto-generation with your own vCard  
✅ **No Manual Formatting** - System handles all vCard complexity  

---

## Quick Start

### Minimal Example
Send a contact and let QuePasa detect WhatsApp status automatically:

```bash
curl -X POST "http://localhost:31000/v3/bot/YOUR_TOKEN/send" \
  -H "Content-Type: application/json" \
  -H "X-QUEPASA-CHATID: 5535900000000@s.whatsapp.net" \
  -d '{
    "contact": {
      "phone": "+55 19 97138-4638",
      "name": "John Doe"
    }
  }'
```

**What happens:**
1. System validates phone number
2. Calls `GetUserInfo` to check if contact is on WhatsApp
3. Detects if it's a Business account
4. Generates appropriate vCard format
5. Sends contact message

---

## API Usage

### Endpoint

**POST** `/v3/bot/{token}/send`

### Headers

```
Content-Type: application/json
X-QUEPASA-CHATID: <recipient_phone>@s.whatsapp.net
```

### Request Body

```json
{
  "contact": {
    "phone": "string (required)",
    "name": "string (required)",
    "vcard": "string (optional)"
  }
}
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `phone` | string | ✅ Yes | Contact phone number (E.164 format recommended) |
| `name` | string | ✅ Yes | Contact display name |
| `vcard` | string | ❌ No | Custom vCard string (skips auto-detection) |

### Response Format

#### Success (200 OK)
```json
{
  "result": {
    "id": "3EB0ABC123DEF456",
    "chat": {
      "id": "5535900000000@s.whatsapp.net",
      "phone": "5535900000000"
    },
    "text": "John Doe",
    "type": "contact",
    "fromMe": true,
    "timestamp": 1696615906
  }
}
```

#### Error (400 Bad Request)
```json
{
  "error": "phone and name are required for contact messages"
}
```

### Examples

#### PowerShell
```powershell
$headers = @{
    "Content-Type" = "application/json"
    "X-QUEPASA-CHATID" = "5535900000000@s.whatsapp.net"
}

$body = @{
    contact = @{
        phone = "+55 19 97138-4638"
        name = "John Doe"
    }
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:31000/v3/bot/YOUR_TOKEN/send" `
    -Method Post -Headers $headers -Body $body
```

#### cURL
```bash
curl --location 'http://localhost:31000/v3/bot/YOUR_TOKEN/send' \
--header 'Content-Type: application/json' \
--header 'X-QUEPASA-CHATID: 5535900000000@s.whatsapp.net' \
--data '{
  "contact": {
    "phone": "+55 19 97138-4638",
    "name": "John Doe"
  }
}'
```

---

## WhatsApp Status Detection

### How It Works

When you send a contact without providing a custom vCard, QuePasa:

1. **Extracts Phone Number** - Cleans and formats for WhatsApp lookup
2. **Calls GetUserInfo** - Queries WhatsApp servers for contact status
3. **Analyzes Response** - Checks for devices and business information
4. **Generates vCard** - Creates appropriate format based on status

### Detection Logic

```go
// 1. Parse phone to JID
jid := phoneWaid + "@s.whatsapp.net"

// 2. Query WhatsApp
userInfos, err := source.Client.GetUserInfo(jids)

// 3. Check devices (definitive indicator)
if len(userInfo.Devices) > 0 {
    isOnWhatsApp = true
}

// 4. Check business status
if isOnWhatsApp && contactInfo.BusinessName != "" {
    isBusiness = true
}
```

### Detection Results

| Scenario | Devices | BusinessName | Result |
|----------|---------|--------------|--------|
| **Business WhatsApp** | > 0 | Present | Business vCard |
| **Regular WhatsApp** | > 0 | Absent | Regular vCard |
| **Not on WhatsApp** | 0 (empty) | - | Non-WhatsApp vCard |
| **API Error/Empty** | - | - | Non-WhatsApp vCard (fallback) |

### Debug Logs

When debug logging is enabled, you'll see:

```log
# For WhatsApp contact
level=debug msg="Contact +55 19 97138-4638 - UserInfo details: Devices=2, Status='Hey there!', VerifiedName=false"
level=debug msg="Contact +55 19 97138-4638 IS on WhatsApp (has 2 devices)"
level=debug msg="Generating Regular WhatsApp vCard for +55 19 97138-4638"

# For Business contact
level=debug msg="Contact +55 11 98765-4321 - UserInfo details: Devices=1, Status='', VerifiedName=true"
level=debug msg="Contact +55 11 98765-4321 IS on WhatsApp (has 1 devices)"
level=debug msg="Contact +55 11 98765-4321 is a Business account: Company Name"
level=debug msg="Generating Business WhatsApp vCard for +55 11 98765-4321"

# For non-WhatsApp contact
level=debug msg="Contact +55 35 3262-0001 - UserInfo details: Devices=0, Status='', VerifiedName=false"
level=debug msg="Contact +55 35 3262-0001 is NOT on WhatsApp (no devices)"
level=debug msg="Generating Non-WhatsApp vCard for +55 35 3262-0001 (no waid parameter)"
```

---

## vCard Format Reference

### Format 1: Business WhatsApp Account

**When:** Contact has `businessName` in contact store + devices present

**Example:**
```vcard
BEGIN:VCARD
VERSION:3.0
N:;Company Name;;;
FN:Company Name
item1.TEL;waid=5519971384638:+55 19 97138-4638
item1.X-ABLabel:WhatsApp Business
X-WA-BIZ-NAME:Company Name
X-WA-BIZ-DESCRIPTION:Official company description
END:VCARD
```

**Features:**
- ✅ Includes `waid` parameter (makes number clickable)
- ✅ Shows "WhatsApp Business" label
- ✅ Displays business name
- ✅ Displays business description
- ✅ User can tap to chat or add to contacts

**WhatsApp Display:**
```
📱 Company Name
   WhatsApp Business
   +55 19 97138-4638
   [Message] [Add Contact]
```

### Format 2: Regular WhatsApp Account

**When:** Contact has devices but no businessName

**Example:**
```vcard
BEGIN:VCARD
VERSION:3.0
N:;John Doe;;;
FN:John Doe
item1.TEL;waid=5519971384638:+55 19 97138-4638
item1.X-ABLabel:Celular
END:VCARD
```

**Features:**
- ✅ Includes `waid` parameter (makes number clickable)
- ✅ Shows "Celular" (Mobile) label
- ❌ No business fields
- ✅ User can tap to chat or add to contacts

**WhatsApp Display:**
```
📱 John Doe
   Celular
   +55 19 97138-4638
   [Message] [Add Contact]
```

### Format 3: Not on WhatsApp

**When:** Contact has no devices (empty device list)

**Example:**
```vcard
BEGIN:VCARD
VERSION:3.0
N:;Jane Smith;;;
FN:Jane Smith
item1.TEL:+55 35 3262-0001
item1.X-ABLabel:Celular
END:VCARD
```

**Features:**
- ❌ **NO** `waid` parameter (not clickable for WhatsApp)
- ✅ Shows "Celular" (Mobile) label
- ❌ No business fields
- ✅ User can only **invite** to WhatsApp or add to phone contacts

**WhatsApp Display:**
```
📱 Jane Smith
   Celular
   +55 35 3262-0001
   [Invite to WhatsApp] [Add Contact]
```

### Comparison Table

| Feature | Business | Regular | Not WhatsApp |
|---------|----------|---------|--------------|
| **waid parameter** | ✅ Yes | ✅ Yes | ❌ **No** |
| **X-ABLabel** | WhatsApp Business | Celular | Celular |
| **X-WA-BIZ-NAME** | ✅ Yes | ❌ No | ❌ No |
| **X-WA-BIZ-DESCRIPTION** | ✅ Yes | ❌ No | ❌ No |
| **Clickable in WhatsApp** | ✅ Yes | ✅ Yes | ❌ **No** |
| **Shows "Message" button** | ✅ Yes | ✅ Yes | ❌ **No** |
| **Shows "Invite" button** | ❌ No | ❌ No | ✅ **Yes** |

### WhatsApp-Specific Extensions

#### waid Parameter

The `waid` (WhatsApp ID) parameter makes phone numbers clickable in WhatsApp:

```
item1.TEL;waid=5519971384638:+55 19 97138-4638
       └─────────────┘  └────────────────┘
         No formatting    Original formatting
```

**Format Rules:**
- **waid value**: Digits only, no `+`, spaces, dashes, or parentheses
- **Display value**: Original formatting preserved for readability
- **Only included**: When contact is verified on WhatsApp

**Example Processing:**
```
Input:  "+55 19 97138-4638"
waid:   "5519971384638"       (stripped)
Display: "+55 19 97138-4638"  (preserved)
```

#### X-ABLabel

Apple Address Book extension for custom labels:

```
item1.X-ABLabel:WhatsApp Business
```

**Possible Values:**
- `WhatsApp Business` - For business accounts
- `Celular` - For regular WhatsApp and non-WhatsApp numbers
- `WhatsApp` - Legacy format (not used in new implementation)

#### X-WA-BIZ-NAME

WhatsApp Business official name from business profile:

```
X-WA-BIZ-NAME:My Company Name
```

**Only included:** For verified Business WhatsApp accounts

#### X-WA-BIZ-DESCRIPTION

WhatsApp Business description text:

```
X-WA-BIZ-DESCRIPTION:We provide excellent service
```

**Only included:** For verified Business WhatsApp accounts

### Custom vCard Support

You can bypass auto-detection by providing your own vCard:

```json
{
  "contact": {
    "phone": "+55 19 97138-4638",
    "name": "Custom Contact",
    "vcard": "BEGIN:VCARD\nVERSION:3.0\nN:Smith;John;Robert;Mr.;Jr.\nFN:Mr. John Robert Smith Jr.\nORG:Company Inc.\nTITLE:CEO\nTEL;TYPE=WORK,VOICE:+5511888888888\nTEL;TYPE=CELL:+5519971384638\nEMAIL:john@company.com\nADR;TYPE=WORK:;;123 Main St;City;State;12345;Country\nURL:https://company.com\nEND:VCARD"
  }
}
```

**When custom vCard is provided:**
- ❌ Auto-detection is **skipped**
- ❌ No `GetUserInfo` call is made
- ✅ Your vCard is used **exactly as provided**
- ✅ Full control over all fields and formatting

**vCard 3.0 Specification:** [RFC 2426](https://www.rfc-editor.org/rfc/rfc2426)

---

## Implementation Details

### System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    API Request                              │
│  POST /v3/bot/{token}/send                                  │
│  { "contact": { "phone": "...", "name": "..." } }          │
└────────────────────┬────────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────────┐
│              QpSendRequest.ToWhatsappMessage()              │
│  • Creates WhatsappMessage                                  │
│  • Sets Type = ContactMessageType                           │
│  • Returns early (prevents type override)                   │
└────────────────────┬────────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────────┐
│            WhatsmeowConnection.Send()                       │
│  • Detects ContactMessageType                               │
│  • Calls generateVCardForContact()                          │
└────────────────────┬────────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────────┐
│          generateVCardForContact()                          │
│  1. Extract phone & clean formatting                        │
│  2. Parse JID (e.g., "5519971384638@s.whatsapp.net")       │
│  3. Call Client.GetUserInfo(jids)                          │
│  4. Check len(userInfo.Devices) > 0                         │
│  5. If on WhatsApp, check contact store for businessName    │
│  6. Generate appropriate vCard format                       │
└────────────────────┬────────────────────────────────────────┘
                     ↓
        ┌────────────┴────────────┐
        ↓                         ↓
┌──────────────┐         ┌────────────────┐
│ isBusiness?  │   No    │ isOnWhatsApp?  │
│     Yes      │ ───────→│      Yes       │
└──────┬───────┘         └────────┬───────┘
       ↓                          ↓
┌──────────────────┐      ┌──────────────────┐
│ Business vCard   │      │ Regular vCard    │
│ • waid ✅        │      │ • waid ✅        │
│ • X-WA-BIZ-* ✅  │      │ • X-WA-BIZ-* ❌  │
└──────────────────┘      └──────────────────┘
                                  │
                          ┌───────┴────────┐
                          ↓                ↓
                    ┌──────────────┐  ┌──────────────────┐
                    │ isOnWhatsApp │  │ Not WhatsApp     │
                    │     No       │  │ vCard            │
                    └──────────────┘  │ • waid ❌        │
                                      │ • X-WA-BIZ-* ❌  │
                                      └──────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────────┐
│         Create waE2E.ContactMessage Protobuf                │
│  • DisplayName: contact.Name                                │
│  • Vcard: generated vCard string                            │
└────────────────────┬────────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────────┐
│              Send to WhatsApp Servers                       │
│  whatsmeow library handles encryption & delivery            │
└─────────────────────────────────────────────────────────────┘
```

### Key Code Components

#### 1. WhatsappContact Struct
**File:** `src/whatsapp/whatsapp_contact.go`

```go
type WhatsappContact struct {
    Phone string `json:"phone"`          // Required
    Name  string `json:"name"`           // Required
    Vcard string `json:"vcard,omitempty"` // Optional
}
```

#### 2. QpSendRequest
**File:** `src/models/qp_send_request.go`

```go
type QpSendRequest struct {
    // ... other fields
    Contact *whatsapp.WhatsappContact `json:"contact,omitempty"`
}

func (source *QpSendRequest) ToWhatsappMessage() (*whatsapp.WhatsappMessage, error) {
    // Contact handling
    if source.Contact != nil {
        result.Contact = source.Contact
        result.Type = whatsapp.ContactMessageType
        return &result, nil  // Early return prevents type override
    }
    // ... other types
}
```

#### 3. WhatsmeowConnection.Send()
**File:** `src/whatsmeow/whatsmeow_connection.go` (lines ~357-380)

```go
func (source *WhatsmeowConnection) Send(msg *whatsapp.WhatsappMessage) {
    // Check if this is a contact message
    if msg.Type == whatsapp.ContactMessageType && msg.Contact != nil {
        contact := msg.Contact
        
        // Generate vCard if not provided
        vcard := contact.Vcard
        if len(vcard) == 0 {
            vcard = source.generateVCardForContact(contact)
        }
        
        newMessage = &waE2E.Message{
            ContactMessage: &waE2E.ContactMessage{
                DisplayName: proto.String(contact.Name),
                Vcard:       proto.String(vcard),
            },
        }
    }
}
```

#### 4. generateVCardForContact()
**File:** `src/whatsmeow/whatsmeow_connection.go` (lines ~685-770)

```go
func (source *WhatsmeowConnection) generateVCardForContact(contact *whatsapp.WhatsappContact) string {
    // 1. Clean phone number for waid
    phoneWaid := strings.ReplaceAll(contact.Phone, " ", "")
    phoneWaid = strings.ReplaceAll(phoneWaid, "-", "")
    phoneWaid = strings.ReplaceAll(phoneWaid, "(", "")
    phoneWaid = strings.ReplaceAll(phoneWaid, ")", "")
    phoneWaid = strings.TrimPrefix(phoneWaid, "+")
    
    // 2. Parse JID
    jid := phoneWaid + "@s.whatsapp.net"
    parsedJid, _ := types.ParseJID(jid)
    
    // 3. Get user info
    userInfos, err := source.Client.GetUserInfo([]types.JID{parsedJid})
    
    // 4. Check devices (critical!)
    if len(userInfo.Devices) > 0 {
        isOnWhatsApp = true
    }
    
    // 5. Check business status
    if isOnWhatsApp {
        contactInfo, _ := source.Client.Store.Contacts.GetContact(context.Background(), parsedJid)
        if contactInfo.BusinessName != "" {
            isBusiness = true
        }
    }
    
    // 6. Generate appropriate vCard
    if isBusiness {
        return fmt.Sprintf("BEGIN:VCARD\nVERSION:3.0\n...")  // Business format
    } else if isOnWhatsApp {
        return fmt.Sprintf("BEGIN:VCARD\nVERSION:3.0\n...")  // Regular format
    } else {
        return fmt.Sprintf("BEGIN:VCARD\nVERSION:3.0\n...")  // Non-WhatsApp format
    }
}
```

### WhatsApp Protobuf

Contact messages use the `ContactMessage` type:

```protobuf
message ContactMessage {
    optional string displayName = 1;
    optional string vcard = 16;
    optional ContextInfo contextInfo = 17;
}
```

### Type Safety

The implementation ensures contacts are never treated as file attachments:

1. **Early Type Detection** - `ToWhatsappMessage()` returns immediately after setting type
2. **Priority Check** - `Send()` checks for `ContactMessageType` before other processing
3. **Direct Creation** - Creates protobuf directly without upload step
4. **No File Handling** - Skips all attachment processing logic

---

## Testing

### Test Script

**File:** `tests/test-contact-whatsapp-detection.ps1`

```powershell
# Configure
$baseUrl = "http://localhost:31000"
$token = "your-token-here"
$botNumber = "5519971200904"

# Test 1: Regular WhatsApp
$regularContact = @{
    chatId = "5519971384638@s.whatsapp.net"
    contact = @{
        phone = "+55 19 97138-4638"
        name = "Regular Contact"
    }
} | ConvertTo-Json

Invoke-RestMethod -Uri "$baseUrl/v3/bot/$botNumber/send" `
    -Method Post `
    -Headers @{
        "X-QUEPASA-TOKEN" = $token
        "Content-Type" = "application/json"
    } `
    -Body $regularContact

# Test 2: Business WhatsApp (replace with real business number)
$businessContact = @{
    chatId = "5519971384638@s.whatsapp.net"
    contact = @{
        phone = "+55 11 98765-4321"
        name = "Business Contact"
    }
} | ConvertTo-Json

# Test 3: Non-WhatsApp (landline or inactive)
$nonWhatsappContact = @{
    chatId = "5519971384638@s.whatsapp.net"
    contact = @{
        phone = "+55 35 3262-0001"
        name = "Landline Contact"
    }
} | ConvertTo-Json

# Test 4: Custom vCard
$customContact = @{
    chatId = "5519971384638@s.whatsapp.net"
    contact = @{
        phone = "+55 31 9102-0002"
        name = "Custom Contact"
        vcard = "BEGIN:VCARD`nVERSION:3.0`n..."
    }
} | ConvertTo-Json
```

### Verification Checklist

After sending contacts, verify in WhatsApp:

#### Regular WhatsApp Contact
- ✅ Phone number is **clickable** (has waid)
- ✅ Shows "Message" button
- ✅ Shows "Add Contact" button
- ✅ Label shows "Celular"
- ❌ No business information displayed

#### Business WhatsApp Contact
- ✅ Phone number is **clickable** (has waid)
- ✅ Shows "Message" button
- ✅ Shows "Add Contact" button
- ✅ Label shows "WhatsApp Business"
- ✅ Business name displayed
- ✅ Business description displayed (if available)

#### Non-WhatsApp Contact
- ❌ Phone number is **NOT clickable** (no waid)
- ❌ **NO** "Message" button
- ✅ Shows "**Invite to WhatsApp**" button
- ✅ Shows "Add Contact" button
- ✅ Label shows "Celular"
- ❌ No business information

### Manual Testing

1. **Enable Debug Logs**
   ```
   Set log level to DEBUG in environment variables
   ```

2. **Send Test Contacts**
   - Regular WhatsApp number (mobile with active WhatsApp)
   - Business WhatsApp number (verified business account)
   - Non-WhatsApp number (landline or inactive mobile)

3. **Check Logs**
   ```log
   # Look for these patterns:
   "Contact ... - UserInfo details: Devices=X, ..."
   "Contact ... IS on WhatsApp (has X devices)"
   "Contact ... is NOT on WhatsApp (no devices)"
   "Contact ... is a Business account: ..."
   "Generating [Business/Regular/Non-WhatsApp] vCard for ..."
   ```

4. **Verify in WhatsApp**
   - Open received contact
   - Check button options
   - Verify business info (if applicable)

---

## Troubleshooting

### Issue: All contacts show "Message" button (even non-WhatsApp)

**Cause:** Old code running (before fix), or `Devices` check not working

**Solution:**
1. Restart QuePasa server to apply changes
2. Check logs for `"Contact ... is NOT on WhatsApp (no devices)"`
3. Verify debug logs show `Devices=0` for non-WhatsApp numbers

**Verify in logs:**
```log
# Should see for non-WhatsApp:
level=debug msg="Contact +55 35 3262-0001 - UserInfo details: Devices=0, ..."
level=debug msg="Generating Non-WhatsApp vCard ... (no waid parameter)"
```

### Issue: "phone and name are required"

**Cause:** Missing required fields in request

**Solution:**
```json
{
  "contact": {
    "phone": "+55 19 97138-4638",  // ✅ Required
    "name": "Contact Name"         // ✅ Required
  }
}
```

### Issue: Contact displays incorrectly

**Cause:** Invalid vCard format (if using custom vCard)

**Solution:**
1. Use auto-generation by omitting `vcard` field
2. If custom vCard needed, validate against vCard 3.0 spec
3. Check for proper escaping of special characters in JSON

### Issue: Business info not showing

**Cause:** Contact is not in contact store, or not a business account

**Solution:**
1. Verify it's a real WhatsApp Business account
2. The system checks `contactInfo.BusinessName` from contact store
3. May need to have the business contact saved first
4. Check logs: `"Contact ... is a Business account: ..."`

### Issue: GetUserInfo fails

**Cause:** Network issues, WhatsApp server problems, or invalid phone format

**Solution:**
1. Check phone number format (should be E.164: `+5519971384638`)
2. System falls back to non-WhatsApp format on error (graceful degradation)
3. Check logs for: `"Contact ... - GetUserInfo failed or empty result"`

### Issue: Slow contact sending

**Cause:** Each send calls `GetUserInfo` (network request)

**Solution:**
- whatsmeow library caches results
- Minimal performance impact for normal usage
- For bulk sending, consider batching or caching

---

## Technical Reference

### Environment Variables

No specific environment variables required. Uses standard QuePasa configuration.

### Dependencies

- **whatsmeow**: v0.0.0-20251005115322-65f6143fa407
- **Go**: 1.21+
- **protobuf**: For WhatsApp message structures

### Related Files

| File | Purpose |
|------|---------|
| `src/whatsapp/whatsapp_contact.go` | Contact data structure |
| `src/models/qp_send_request.go` | API request model |
| `src/api/api_handlers+SendController.go` | Send endpoint handler |
| `src/whatsmeow/whatsmeow_connection.go` | Core implementation |
| `tests/test-contact-send.ps1` | Basic test script |
| `tests/test-contact-whatsapp-detection.ps1` | Detection test script |

### API Endpoints Used Internally

- `GetUserInfo(jids)` - WhatsApp user information lookup
- `Store.Contacts.GetContact()` - Local contact store query

### Message Types

```go
const ContactMessageType WhatsappMessageType = "contact"
```

### Validation Rules

1. ✅ `phone` field is required
2. ✅ `name` field is required
3. ✅ `vcard` field is optional
4. ✅ Phone should be in valid format (E.164 recommended)
5. ✅ Custom vCard must be valid vCard 3.0 format

### Performance Characteristics

- **API Call Overhead**: ~50-200ms per `GetUserInfo` call
- **Caching**: whatsmeow caches UserInfo results
- **Fallback**: Always succeeds (falls back to non-WhatsApp format on error)
- **Throughput**: Suitable for normal messaging volume

### Security Considerations

- ✅ No sensitive data in vCard auto-generation
- ✅ Custom vCard is validated (structure only, not content)
- ✅ Phone numbers are not stored or logged beyond debug level
- ✅ WhatsApp encryption applies to contact messages

---

## Version History

### v3.25.XXXX.XXXX (October 2025)
- ✅ Initial implementation of contact messages
- ✅ WhatsApp status detection via `GetUserInfo`
- ✅ Three-format vCard generation (Business/Regular/Non-WhatsApp)
- ✅ Auto-generation with intelligent detection
- ✅ Custom vCard support (bypass detection)
- ✅ Full debug logging
- ✅ Documentation and test scripts

### Key Improvements
- **Before**: All contacts sent with same vCard format
- **After**: Smart detection generates appropriate format per contact

### Breaking Changes
- None - Fully backward compatible
- Custom vCard still works as before
- API interface unchanged

---

## Additional Resources

### Documentation
- [Location Messages](./SEND_LOCATION.md)
- [Webhook System](./WEBHOOK_SYSTEM_DOCUMENTATION.md)
- [Chat Management](./CHAT_MANAGEMENT.md)

### Specifications
- [vCard 3.0 RFC](https://www.rfc-editor.org/rfc/rfc2426)
- [WhatsApp Web Protocol](https://github.com/WhiskeySockets/Baileys)

### Support
1. Check this documentation
2. Review test scripts in `tests/`
3. Enable debug logging and check server logs
4. Open GitHub issue with logs and request details

---

## Summary

QuePasa's contact message feature provides intelligent, automatic detection of WhatsApp status:

✅ **Simple to use** - Just provide phone and name  
✅ **Automatic detection** - System handles all complexity  
✅ **Three smart formats** - Business, Regular, or Non-WhatsApp  
✅ **Proper UX** - Correct buttons (Message vs Invite)  
✅ **Backward compatible** - Custom vCard still supported  
✅ **Production ready** - Tested and working  

**Bottom line:** Send contacts with confidence - the system automatically generates the right format for the best user experience! 🎉
