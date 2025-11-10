# MCP Module - AI Agent Instructions

## Module Scope
Model Context Protocol (MCP) server implementation for QuePasa API integration.

## Overview
This module implements the MCP (Model Context Protocol) server for QuePasa, allowing AI assistants and other tools to interact with the WhatsApp API through a standardized protocol.

## Architecture
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
