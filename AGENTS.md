# Contacts Endpoint Improvements - AI Agent Instructions

## Branch: feature/improve-contacts-endpoint

## Module Scope
Improve the `/contacts` endpoint to provide better contact information and add search functionality.

## Branch: feature/improve-contacts-endpoint

## Module Scope
Improve the `/contacts` endpoint to provide better contact information and add search functionality.

## Current Implementation Analysis

### Existing Endpoint: GET /contacts
- **File:** `src/api/api_handlers+ContactsController.go`
- **Method:** GET
- **Returns:** All contacts with Id, LId, Title (name), Phone
- **Data Source:** `whatsmeow.Client.Store.Contacts.GetAllContacts()` (uses local database cache, NOT direct WhatsApp queries)
- **Issue:** Some contacts are returned without names (Title is empty)

### Contact Data Flow
1. `ContactsController` → `server.GetContacts()`
2. `WhatsmeowContactManager.GetContacts()` → `Client.Store.Contacts.GetAllContacts()`
3. Merges @lid and @s.whatsapp.net contacts by phone number
4. Returns `[]WhatsappChat` with Id, LId, Title, Phone

## Improvements to Implement

### ✅ Priority 1: Fix Missing Contact Names

**Problem:** Some contacts return without Title (name) even though they might have information available.

**Investigation Tasks:**
- [ ] Analyze why some contacts have empty Title field
- [ ] Check if FullName, BusinessName, or PushName are being properly extracted from whatsmeow store
- [ ] Verify contact merging logic (lines 80-150 in whatsmeow_contact_manager.go)
- [ ] Test with real data to identify patterns of missing names

**Possible Causes:**
1. Contact info not synced from WhatsApp
2. Priority order of name fields (FullName → BusinessName → PushName) missing fallbacks
3. Contact merging losing name information
4. Store database not properly populated

**Solution Strategy:**
1. Add better fallback logic for name extraction
2. Consider using phone number as fallback if no name available
3. Add logging to track why names are missing
4. Potentially trigger contact sync if contact info is incomplete

### ✅ Priority 2: Add Search Endpoint

**Endpoint:** POST /search (or POST /contacts/search)

**Why POST instead of GET:**
- More appropriate for search operations with multiple criteria
- Allows complex filter objects in request body
- Better for future extensibility

**Request Body:**
```json
{
  "query": "string",        // Search in name and phone (optional)
  "has_name": true|false,   // Filter contacts with/without name (optional)
  "has_lid": true|false,    // Filter contacts with/without LID (optional)
  "phone": "string"         // Search by specific phone (optional)
}
```

**Response:**
```json
{
  "result": "success",
  "total": 10,
  "contacts": [
    {
      "id": "5511999999999@s.whatsapp.net",
      "lid": "ABC123@lid",
      "title": "Contact Name",
      "phone": "+5511999999999"
    }
  ]
}
```

**Implementation Requirements:**
- Create new controller: `api_handlers+ContactsSearchController.go`
- Add route: `POST /search` or `POST /contacts/search`
- Implement search logic in `WhatsmeowContactManager` or controller
- Add Swagger annotations (@Summary, @Description, @Param, @Router)
- Support case-insensitive search
- Support partial matching in query field
- Return empty array if no matches

### ❌ NOT Implementing (Out of Scope)

**Pagination:** Not needed - administrators expect full results and will handle large datasets.

**Sorting:** Not needed - clients will sort results after receiving them based on their needs.

**Profile Pictures:** No public API available for bulk retrieval. Individual picture endpoint already exists at `/picinfo/{chatId}`.

**Status/Last Message:** Not relevant for contact listing purposes.

**Cache Implementation:** Already uses whatsmeow database cache (`Client.Store.Contacts`), no additional caching needed.

**Export Formats:** API returns JSON only - each application handles its own export format needs.

**Type Filtering (groups vs personal):** Not needed - API consumers can identify type by ID format.

## Technical Guidelines

### File Naming Convention
- Controllers: `api_handlers+<Name>Controller.go`
- Search controller: `api_handlers+ContactsSearchController.go`

### Swagger Documentation
**CRITICAL:** Always regenerate Swagger after API changes:
```bash
cd src
swag init --output ./swagger
```

### Testing Requirements
- Test with contacts that have names
- Test with contacts without names  
- Test search with various criteria
- Test search with no results
- Test search with special characters
- Test with LID and non-LID contacts

### Error Handling
- Return appropriate HTTP status codes
- Use existing `QpResponse` error handling pattern
- Log errors for debugging

## Implementation Checklist

### Phase 1: Contact Name Improvements
- [ ] Analyze GetContacts() implementation thoroughly
- [ ] Add logging to track name extraction process
- [ ] Implement better fallback logic for missing names
- [ ] Test with real WhatsApp data
- [ ] Document findings

### Phase 2: Search Endpoint
- [ ] Create ContactsSearchController.go
- [ ] Define search request/response models
- [ ] Implement search logic
- [ ] Add Swagger annotations
- [ ] Register route in api_handlers.go
- [ ] Regenerate Swagger documentation
- [ ] Test all search scenarios

### Phase 3: Testing & Documentation
- [ ] Test with various contact scenarios
- [ ] Update README if needed
- [ ] Verify Swagger docs are correct
- [ ] Test with Postman/curl
- [ ] Get user approval before committing

## Notes

**Cache Verification (COMPLETED):**
- ✅ Confirmed: `GetContacts()` uses `Client.Store.Contacts.GetAllContacts()`
- ✅ This is the whatsmeow local database cache (SQLite/Postgres)
- ✅ No direct WhatsApp queries for contact listing
- ✅ No additional cache layer needed

**Version Update:**
- Remember to update `QpVersion` in `models/qp_defaults.go` before merging to main
- Follow project versioning guidelines

## Questions for User

1. Preferred search endpoint path: `/search` or `/contacts/search`?
2. Should search be case-sensitive or case-insensitive?
3. Any specific test cases or contact scenarios to prioritize?
4. Should we add search to MCP tools as well?

## Related Files

- `src/api/api_handlers+ContactsController.go` - Current contacts endpoint
- `src/whatsmeow/whatsmeow_contact_manager.go` - Contact retrieval logic
- `src/models/qp_contacts_response.go` - Response model
- `src/whatsapp/whatsapp_chat.go` - Contact data structure
- `src/api/api_handlers.go` - Route registration

---

# Previous Module Documentation (MCP)

*Note: The content below is from the previous MCP module implementation. Keep for reference but focus on the Contacts Improvement task above.*

---

## MCP Module - AI Agent Instructions (ARCHIVED)

## Overview
- MCP server endpoint: `/mcp`
- Authentication: Bearer token (Master key or Bot token)
- Protocol: JSON-RPC 2.0 with SSE support
- Tools: Exposed as MCP tools for AI assistants

## Authentication Levels

### 1. Master Key - Full Access (Super User)
- Uses `MASTERKEY` environment variable
- Header: `Authorization: Bearer <MASTERKEY>`
- Can access all servers and operations
- Administrative privileges
- Tools return global system information

### 2. Bot Token - Server-Specific Access
- Uses individual bot tokens
- Header: `Authorization: Bearer <bot-token>`
- Limited to specific server operations
- Standard user privileges
- Tools return server-specific information

## Environment Variables
- **`MCP_ENABLED`** - Enable/disable MCP server (default: `false`)
- **`MCP_PATH`** - MCP endpoint path (default: `/mcp`)
- **`MASTERKEY`** - Master key for super admin access

## Protocol Flow

### SSE Connection (GET /mcp)
1. Client opens SSE connection with Bearer token
2. Server authenticates and keeps connection alive
3. Connection remains open for real-time updates

### JSON-RPC Messages (POST /mcp)
All messages follow JSON-RPC 2.0 format:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "method_name",
  "params": {}
}
```

### Supported Methods
1. **initialize** - Returns server capabilities and protocol version
2. **notifications/initialized** - Client confirms initialization (no response)
3. **tools/list** - Returns available tools
4. **tools/call** - Executes a specific tool

## Available Tools

### System Tools

#### health
- **Description**: Check server health and status
- **Authentication**: Master key or bot token
- **Parameters**: None
- **Returns**: 
  - Master key: Global system health (total_servers, connected_servers, disconnected_servers)
  - Bot token: Specific server health (status, connected, timestamp, server_info)

#### list_servers
- **Description**: List available WhatsApp bot servers
- **Authentication**: Master key (all servers) or bot token (current server only)
- **Parameters**: None
- **Returns**: Array of servers with token, wid, connected status, etc.

### API Tools (WhatsApp Operations)

All API tools support dual authentication mode:
- **Master Key**: Requires `token` parameter to specify target server
- **Bot Token**: Uses authenticated server automatically (token parameter ignored)

#### send_message
- **Description**: Send WhatsApp message (text, image, document, audio, video)
- **Method**: POST /send
- **Parameters**:
  - `token` (string, optional): Bot token (required with master key)
  - `chatId` (string, required): Phone number with country code
  - `text` (string, optional): Message text
  - `attachment` (object, optional): Media attachment with url, mimetype, filename

#### receive_messages
- **Description**: Get received messages from webhook cache
- **Method**: GET /receive
- **Parameters**:
  - `token` (string, optional): Bot token (required with master key)
  - `timestamp` (string, optional): Filter messages after timestamp

#### get_qrcode
- **Description**: Get QR code for WhatsApp pairing
- **Method**: GET /scan
- **Parameters**:
  - `token` (string, optional): Bot token (required with master key)

#### download_media
- **Description**: Download media from message
- **Method**: GET /download/{messageId}
- **Parameters**:
  - `token` (string, optional): Bot token (required with master key)
  - `messageId` (string, required): Message ID to download

#### get_contacts
- **Description**: Get WhatsApp contacts list
- **Method**: GET /contacts
- **Parameters**:
  - `token` (string, optional): Bot token (required with master key)

#### get_groups
- **Description**: Get WhatsApp groups list
- **Method**: GET /groups/getall
- **Parameters**:
  - `token` (string, optional): Bot token (required with master key)

#### is_on_whatsapp
- **Description**: Check if phone number is registered on WhatsApp
- **Method**: GET /isonwhatsapp/{phone}
- **Parameters**:
  - `token` (string, optional): Bot token (required with master key)
  - `phone` (string, required): Phone number with country code

#### get_picture
- **Description**: Get profile picture URL for contact or group
- **Method**: GET /picinfo/{chatId}
- **Parameters**:
  - `token` (string, optional): Bot token (required with master key)
  - `chatId` (string, required): Chat ID (phone number or group ID)

#### mark_as_read
- **Description**: Mark chat messages as read
- **Method**: POST /chat/markread
- **Parameters**:
  - `token` (string, optional): Bot token (required with master key)
  - `chatId` (string, required): Chat ID to mark as read

## Implementation Status

### ✅ Completed (2025-11-10)
- [x] SSE endpoint (GET /mcp)
- [x] JSON-RPC endpoint (POST /mcp)
- [x] Bearer token authentication (master key + bot token)
- [x] Context-based tool execution (MCPToolContext)
- [x] Initialize method with protocol version negotiation
- [x] Tools/list method
- [x] Tools/call method
- [x] JSON-RPC 2.0 compliant responses
- [x] Environment variables configuration
- [x] Health tool (dual behavior: master/bot)
- [x] List servers tool
- [x] **Auto-discovery system** - 38 API tools auto-registered from Swagger annotations
- [x] **Swagger annotation parser** - Extracts @Summary, @Router, @Param, @MCPHidden, @MCPTool
- [x] **Handler registry** - Maps 38+ controller names to functions
- [x] 40 total tools (2 system + 38 API)
- [x] Header-only authentication (no tokens in URL paths)
- [x] Master key + token parameter support
- [x] Current API version endpoints (no /v3/ prefix)
- [x] JSON schema generation from Swagger params

### 📋 Future Enhancements
- [ ] Resource support (files, attachments via MCP resources)
- [ ] Prompt templates
- [ ] Session management improvements
- [ ] Rate limiting per client
- [ ] Metrics and monitoring (Prometheus integration)
- [ ] Additional tools: send_location, send_contact, create_group, manage_group
- [ ] Bulk operations support

## Testing

### Test SSE Connection

```bash
curl -N -H "Authorization: Bearer quepasa-master-key-dev" \
  http://localhost:31000/mcp
```

### Test Initialize

```bash
curl -X POST http://localhost:31000/mcp \
  -H "Authorization: Bearer quepasa-master-key-dev" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2024-11-05",
      "clientInfo": {"name": "test", "version": "1.0"}
    }
  }'
```

### Test Tools List

```bash
curl -X POST http://localhost:31000/mcp \
  -H "Authorization: Bearer quepasa-master-key-dev" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/list"
  }'
```

### Test API Tool (with Master Key)

```bash
curl -X POST http://localhost:31000/mcp \
  -H "Authorization: Bearer quepasa-master-key-dev" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
      "name": "get_contacts",
      "arguments": {
        "token": "your-bot-token-here"
      }
    }
  }'
```

### Test API Tool (with Bot Token)

```bash
curl -X POST http://localhost:31000/mcp \
  -H "Authorization: Bearer your-bot-token-here" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 4,
    "method": "tools/call",
    "params": {
      "name": "get_contacts",
      "arguments": {}
    }
  }'
```

## Next Steps

### Priority 1: Additional Tools

- [ ] **send_location** - Send location messages
- [ ] **send_contact** - Send contact cards
- [ ] **create_group** - Create new groups
- [ ] **manage_group** - Add/remove participants
- [ ] **get_group_invite** - Get group invite link
- [ ] **send_reaction** - React to messages

### Priority 2: Advanced Features

- [ ] **Resource support** - Files, attachments via MCP resources
- [ ] **Prompt templates** - Pre-defined conversation templates
- [ ] **Session management** - Persistent MCP sessions
- [ ] **Rate limiting** - Per-client request throttling
- [ ] **Metrics integration** - Prometheus monitoring
- [ ] **Bulk operations** - Send to multiple contacts at once

### Priority 3: Developer Experience

- [ ] **OpenAPI/Swagger** - MCP tools documentation
- [ ] **SDK examples** - Python, Node.js, Go clients
- [ ] **Error codes** - Standardized error responses
- [ ] **Webhooks** - Real-time event notifications via MCP

## Notes / Known Issues

### 2025-11-10: MCP Server Fully Functional ✅

**Implemented Features:**

- SSE connection working correctly
- Bearer token authentication implemented
- Initialize method returns proper capabilities
- Health tool returns JSON (master: global stats, bot: server-specific)
- JSON-RPC 2.0 format compliant
- Multiple protocol versions supported (2024-11-05, 2025-03-26, 2025-06-18)
- 11 tools available (2 system + 9 API)
- Header-only authentication (tokens never in URL)
- Master key + token parameter for cross-server operations
- Context-based tool execution

**Authentication Architecture:**

1. Client sends Bearer token via Authorization header
2. Server checks if token matches MASTERKEY (super user)
3. If not, checks if token is a valid bot token (server-specific)
4. Creates MCPToolContext{Server, IsMaster} for tool execution
5. Tools adapt behavior based on context

**Security Design:**

- ✅ Tokens only in headers (X-QUEPASA-TOKEN, X-QUEPASA-MASTERKEY)
- ✅ No tokens in URL paths (clean URLs: /send, /contacts, etc.)
- ✅ Master key can specify target server via token parameter
- ✅ Bot token automatically uses authenticated server
- ✅ Dual authentication: Master key + server token for admin operations

### Known Limitations

- No resource support yet (files must be sent as URLs in attachments)
- No prompt templates (tools only, no conversational patterns)
- No rate limiting (consider implementing for production use)
- No bulk operations (each message requires separate tool call)

## Development Guidelines

- Follow MCP protocol specification
- Implement proper authentication middleware
- Return structured JSON responses
- Log all MCP operations with appropriate levels
- Handle errors gracefully with JSON-RPC error format
- Use Bearer token authentication (standard HTTP)
- Support multiple protocol versions
- Tools should handle both master and bot contexts
- Never expose tokens in URL paths (security requirement)
- Use current API version (no /v3/ prefix)
