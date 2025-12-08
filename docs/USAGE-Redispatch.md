# Re-dispatch Message Endpoint

## Overview

This endpoint allows you to force re-dispatch of a message that is stored in cache. This is useful when a webhook or RabbitMQ dispatch failed and you want to retry sending the message without waiting for it to be naturally re-processed.

**Important:** The re-dispatch applies **ALL original dispatching validations**, ensuring the same behavior as the original dispatch system.

## Endpoint

**POST** `/redispatch/{messageid}`

## Authentication

Requires authentication via bot token:

- Header: `X-QUEPASA-TOKEN` or `Authorization: Bearer <token>`

## Parameters

### Path Parameters

- `messageid` (string, required): The ID of the message to re-dispatch

## Dispatching Validations Applied

The redispatch endpoint uses the same validation logic as the original dispatching system (`PostToDispatchingFromServer`):

### 1. TrackId Validation (Loop Prevention)
- **Purpose**: Prevents infinite loops when `ForwardInternal` is enabled
- **Logic**: Message is dispatched only if:
  - Message is NOT from internal system (`!message.FromInternal`), OR
  - `ForwardInternal` is true AND (no `TrackId` configured OR `TrackId` doesn't match message's `TrackId`)
- **Use Case**: Avoids re-sending messages that originated from the same webhook/integration

### 2. Message Type Filters

Each dispatching endpoint can configure filters:

- **Read Receipts**: Skipped if `ReadReceipts = false` (message.Id == "readreceipt")
- **Group Messages**: Skipped if `Groups = false` (message.FromGroup())
- **Broadcast Messages**: Skipped if `Broadcasts = false` (message.FromBroadcast())
- **Call Messages**: Skipped if `Calls = false` (message.Type == CallMessageType)

### 3. Per-Dispatching Configuration

Each webhook/RabbitMQ configuration has its own filters. The message will only be sent to dispatching endpoints that match ALL filter criteria.

## Response

### Success (200 OK)
```json
{
  "success": true,
  "message": "message <messageid> re-dispatched successfully"
}
```

### Error Responses

#### 400 Bad Request
```json
{
  "success": false,
  "message": "message ID is required"
}
```

#### 404 Not Found
```json
{
  "success": false,
  "message": "message not present on cache, id: <messageid>"
}
```

#### 503 Service Unavailable
```json
{
  "success": false,
  "message": "server not ready"
}
```

## Usage Examples

### cURL
```bash
curl -X POST "http://localhost:31000/redispatch/3EB0XXXXXXXXXXXXXX" \
  -H "X-QUEPASA-TOKEN: your-bot-token-here"
```

### PowerShell
```powershell
$headers = @{
    "X-QUEPASA-TOKEN" = "your-bot-token-here"
}
Invoke-RestMethod -Uri "http://localhost:31000/redispatch/3EB0XXXXXXXXXXXXXX" `
    -Method POST -Headers $headers
```

### Python
```python
import requests

url = "http://localhost:31000/redispatch/3EB0XXXXXXXXXXXXXX"
headers = {
    "X-QUEPASA-TOKEN": "your-bot-token-here"
}

response = requests.post(url, headers=headers)
print(response.json())
```

## How It Works

1. The endpoint retrieves the message from the server's message cache using the provided message ID
2. Calls `PostToDispatchingFromServer()` which applies **all original dispatching validations**
3. For each configured dispatching endpoint (webhook/RabbitMQ):
   - Validates TrackId (prevents loops)
   - Applies message type filters (groups, broadcasts, calls, read receipts)
   - Checks ForwardInternal configuration
4. Sends the message only to dispatching endpoints that pass all validation checks

## Validation Examples

### Example 1: TrackId Loop Prevention

```json
// Webhook configuration
{
  "url": "https://myapp.com/webhook",
  "forwardInternal": true,
  "trackId": "myapp-integration"
}

// Message from internal system (originated from myapp)
{
  "id": "3EB0ABC123",
  "trackId": "myapp-integration",
  "fromInternal": true
}
```

**Result**: Message will NOT be re-dispatched (avoids infinite loop)

### Example 2: Group Message Filter

```json
// Webhook configuration
{
  "url": "https://myapp.com/webhook",
  "groups": false  // Don't send group messages
}

// Group message
{
  "id": "3EB0XYZ789",
  "chat": {"id": "123456@g.us"}  // Group chat
}
```

**Result**: Message will NOT be re-dispatched to this webhook

### Example 3: Multiple Webhooks with Different Filters

```json
// Webhook 1: All messages
{
  "url": "https://app1.com/webhook",
  "groups": true,
  "broadcasts": true
}

// Webhook 2: Only private chats
{
  "url": "https://app2.com/webhook",
  "groups": false,
  "broadcasts": false
}

// Group message
{
  "id": "3EB0ABC456",
  "chat": {"id": "123456@g.us"}
}
```

**Result**: Message dispatched ONLY to Webhook 1 (Webhook 2 filters out groups)

## Use Cases

- **Retry Failed Webhooks**: When a webhook endpoint was temporarily down and you want to resend the message
- **Manual Dispatch**: Force dispatch of a specific message to newly configured webhooks (respecting filters)
- **Testing**: Verify webhook configurations by resending test messages
- **Recovery**: Recover from temporary network issues that prevented initial dispatch
- **Selective Re-dispatch**: Messages are only sent to webhooks that match their filters (groups, calls, etc.)

## Notes

- The message must exist in the cache (default retention: 124 hours / ~5 days)
- Message IDs are case-insensitive (automatically converted to uppercase)
- The re-dispatch applies **ALL dispatching validations** (same as original dispatch)
- **TrackId validation** prevents infinite loops when ForwardInternal is enabled
- **Message type filters** are respected (groups, broadcasts, calls, read receipts)
- Each dispatching endpoint (webhook/RabbitMQ) has its own configuration and filters
- Messages may be dispatched to some endpoints and skipped on others based on filters
- Original message metadata (timestamp, sender, TrackId, etc.) is preserved

## Related Endpoints

- `GET /receive` - List cached messages
- `GET /message/{messageid}` - Get specific message details
- `GET /webhook` - View webhook configurations
- `GET /rabbitmq` - View RabbitMQ configurations
