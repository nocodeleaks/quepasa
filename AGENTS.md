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

### health
- **Description**: Check server health and status
- **Authentication**: Master key or bot token
- **Parameters**: None
- **Returns**: 
  - Master key: Global system health (total_servers, connected_servers, disconnected_servers)
  - Bot token: Specific server health (status, connected, timestamp, server_info)

## Implementation Status

### ✅ Completed
- [x] SSE endpoint (GET /mcp)
- [x] JSON-RPC endpoint (POST /mcp)
- [x] Bearer token authentication
- [x] Master key vs Bot token differentiation
- [x] Initialize method with protocol version negotiation
- [x] Tools/list method
- [x] Tools/call method
- [x] Health tool with dual behavior (master/bot)
- [x] JSON-RPC 2.0 compliant responses
- [x] Environment variables configuration

### 🚧 In Progress
- [ ] Additional MCP tools (send, receive, contacts, groups)
- [ ] Error handling improvements
- [ ] Logging optimization

### 📋 Pending
- [ ] Resource support (files, attachments)
- [ ] Prompt templates
- [ ] Session management
- [ ] Rate limiting
- [ ] Metrics and monitoring

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

### Test Health Tool
```bash
curl -X POST http://localhost:31000/mcp \
  -H "Authorization: Bearer quepasa-master-key-dev" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
      "name": "health",
      "arguments": {}
    }
  }'
```

## Next Steps

### Priority 1: Core Tools
- [ ] **send_message** - Send text/media messages
- [ ] **receive_messages** - Get messages from webhook cache
- [ ] **get_contacts** - List contacts
- [ ] **get_groups** - List groups

### Priority 2: Advanced Tools
- [ ] **send_location** - Send location messages
- [ ] **send_contact** - Send contact cards
- [ ] **create_group** - Create new groups
- [ ] **manage_group** - Add/remove participants

### Priority 3: System Tools
- [ ] **get_qrcode** - Get QR code for pairing
- [ ] **disconnect** - Disconnect server
- [ ] **restart** - Restart connection

## Notes / Known Issues

### 2025-11-10: MCP Server Functional ✅
- SSE connection working correctly
- Bearer token authentication implemented
- Initialize method returns proper capabilities
- Health tool returns JSON (master: global stats, bot: server-specific)
- JSON-RPC 2.0 format compliant
- Multiple protocol versions supported (2024-11-05, 2025-03-26, 2025-06-18)

### Authentication Flow
1. Client sends Bearer token via Authorization header
2. Server checks if token matches MASTERKEY (super user)
3. If not, checks if token is a valid bot token (server-specific)
4. Registers tools with appropriate context (nil for master, server for bot)
5. Tools adapt behavior based on context

## Development Guidelines
- Follow MCP protocol specification
- Implement proper authentication middleware
- Return structured JSON responses
- Log all MCP operations with appropriate levels
- Handle errors gracefully with JSON-RPC error format
- Use Bearer token authentication (standard HTTP)
- Support multiple protocol versions
- Tools should handle both master and bot contexts
