# API Module - AI Agent Instructions

## Module Scope
REST API endpoints for QuePasa WhatsApp bot management and operations.

## Overview
This module implements the HTTP REST API for QuePasa, providing endpoints for bot management, message operations, and WhatsApp integration.

## Key Patterns

### Server Pre-Configuration Flow
**Problem**: When scanning QR code, messages arrive immediately after connection. If webhooks aren't configured yet, initial messages are lost.

**Solution**: POST /info endpoint allows creating/updating server configuration **before** QR code scanning.

#### Flow Example:
1. **Create/Update Server Configuration** (before QR scan)
   ```bash
   POST /info
   Headers: X-QUEPASA-TOKEN: your-custom-token
   Body: {
     "groups": true,
     "broadcasts": false,
     "readreceipts": true,
     "calls": true,
     "devel": false
   }
   ```

2. **Add Webhook** (before QR scan)
   ```bash
   POST /webhook
   Headers: X-QUEPASA-TOKEN: your-custom-token
   Body: {
     "url": "https://your-webhook-url.com/webhook",
     "forwardinternal": false
   }
   ```

3. **Scan QR Code** (connection established)
   ```bash
   GET /scan?token=your-custom-token
   ```

4. **Messages automatically dispatched** to pre-configured webhook

### Information Endpoint (/info)

#### POST /info - Create or Update Server Configuration
- **Purpose**: Pre-configure server before QR scan (AddOrUpdate pattern)
- **Authentication**: Token via `X-QUEPASA-TOKEN` header
- **Method**: POST
- **Body** (all fields optional):
  ```json
  {
    "groups": true|false,
    "broadcasts": true|false,
    "readreceipts": true|false,
    "calls": true|false,
    "devel": true|false
  }
  ```
- **Behavior**:
  - If server with token exists: updates configuration
  - If server doesn't exist: creates new server with configuration
  - No WhatsApp connection established (only configuration)
- **Response**: Server information with configured settings

#### GET /info - Get Server Information
- **Purpose**: Retrieve current server configuration and status
- **Returns**: Server details, connection status, settings

#### PATCH /info - Update Server Settings
- **Purpose**: Update existing server settings (requires active server)
- **Body**: Same as POST, but requires server to exist

#### DELETE /info - Delete Server
- **Purpose**: Remove server and all configurations
- **Effect**: Disconnects WhatsApp, removes from database

## Implementation Guidelines

### Pre-Configuration Pattern
1. Always allow POST /info without active connection
2. Token from header identifies the server (enables AddOrUpdate)
3. Configuration persists in database
4. QR scan (/scan) reuses existing token's configuration
5. First messages after QR scan use pre-configured webhooks

### Error Handling
- Missing token: 400 Bad Request
- Invalid configuration: 400 Bad Request
- Server operations use transactions where possible

## Testing Pre-Configuration

```bash
# 1. Create server with config (no connection yet)
curl -X POST http://localhost:31000/info \
  -H "X-QUEPASA-TOKEN: test-token-123" \
  -H "Content-Type: application/json" \
  -d '{"groups": true, "readreceipts": true}'

# 2. Add webhook (before QR scan)
curl -X POST http://localhost:31000/webhook \
  -H "X-QUEPASA-TOKEN: test-token-123" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://webhook.site/your-uuid"}'

# 3. Scan QR code (establishes connection)
curl http://localhost:31000/scan?token=test-token-123

# 4. Messages received immediately dispatched to webhook
```

## Benefits
- ✅ No lost messages during initial connection
- ✅ Clean separation: config first, connect second
- ✅ RESTful pattern (POST creates, GET reads, PATCH updates, DELETE removes)
- ✅ Token-based identification (no auto-generation)
- ✅ AddOrUpdate pattern (idempotent POST)
