# MCP Module - AI Agent Instructions

## Module Scope
Model Context Protocol (MCP) server implementation for QuePasa API integration.

## Overview
This module implements the MCP (Model Context Protocol) server for QuePasa, allowing AI assistants and other tools to interact with the WhatsApp API through a standardized protocol.

## Architecture
- MCP server endpoint: `/mcp`
- Authentication: Master key (full access) or Bot token (server-specific access)
- Protocol: SSE (Server-Sent Events) based communication
- Tools: Exposed as MCP tools for AI assistants

## Authentication Levels
1. **Master Key** - Full access to all servers and operations
   - Uses `MASTERKEY` environment variable
   - Can access any bot/server
   - Administrative privileges

2. **Bot Token** - Server-specific access
   - Uses individual bot tokens
   - Limited to specific server operations
   - Standard user privileges

## Environment Variables
- **`MCP_ENABLED`** - Enable/disable MCP server (default: `false`)
- **`MCP_PATH`** - MCP endpoint path (default: `/mcp`)

## Available Tools

### 1. health
- **Description**: Check server health and status
- **Authentication**: Master key or bot token
- **Parameters**: None (uses authentication context)
- **Returns**: Server health information

## Implementation Guidelines
- Follow MCP protocol specification
- Implement proper authentication middleware
- Return structured responses
- Log all MCP operations
- Handle errors gracefully

## Testing
- Test with both authentication methods
- Verify tool responses
- Check error handling
- Validate MCP protocol compliance

## Notes / Known issues

- 2025-11-08: Couldn't run MCP request due to SSE error: Non-200 status code (405). Reproduction: invoke the MCP tools endpoint (/mcp) using SSE/long-poll and observe 405 response when attempting to open the SSE stream. Likely causes: wrong HTTP method for SSE endpoint or route mismatch (server expects POST on /mcp/tools/call while SSE client used GET). Next step: investigate server route handling for SSE and add dedicated SSE handler or correct client method. Continue debugging tomorrow.

