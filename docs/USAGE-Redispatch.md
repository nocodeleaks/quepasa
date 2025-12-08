# Re-dispatch Message Endpoint

## Overview
This endpoint allows you to force re-dispatch of a message that is stored in cache. This is useful when a webhook or RabbitMQ dispatch failed and you want to retry sending the message without waiting for it to be naturally re-processed.

## Endpoint
**POST** `/redispatch/{messageid}`

## Authentication
Requires authentication via bot token:
- Header: `X-QUEPASA-TOKEN` or `Authorization: Bearer <token>`

## Parameters

### Path Parameters
- `messageid` (string, required): The ID of the message to re-dispatch

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
2. If the message exists, it calls the `Trigger()` method on the handler
3. The `Trigger()` method dispatches the message to all configured webhooks and RabbitMQ queues
4. The message will be sent with its original content and metadata

## Use Cases

- **Retry Failed Webhooks**: When a webhook endpoint was temporarily down and you want to resend the message
- **Manual Dispatch**: Force dispatch of a specific message to newly configured webhooks
- **Testing**: Verify webhook configurations by resending test messages
- **Recovery**: Recover from temporary network issues that prevented initial dispatch

## Notes

- The message must exist in the cache (default retention: 124 hours / ~5 days)
- Message IDs are case-insensitive (automatically converted to uppercase)
- The re-dispatch will trigger ALL configured dispatching methods (webhooks + RabbitMQ)
- Original message metadata (timestamp, sender, etc.) is preserved
- The server must be in "Ready" state for re-dispatch to work

## Related Endpoints

- `GET /receive` - List cached messages
- `GET /message/{messageid}` - Get specific message details
- `GET /webhook` - View webhook configurations
- `GET /rabbitmq` - View RabbitMQ configurations
