# Presence API Documentation

## Overview
The Presence API allows you to control the global presence status of your WhatsApp bot, making it appear as "available" (online) or "unavailable" (offline) to other users.

## Endpoint

**POST** `/presence`

## Authentication
Requires `X-QUEPASA-TOKEN` header with a valid bot token.

## Request Body

```json
{
  "presence": "available"
}
```

### Parameters

| Field | Type | Required | Description | Valid Values |
|-------|------|----------|-------------|--------------|
| `presence` | string | Yes | The presence status to set | `"available"` or `"unavailable"` |

## Response

### Success Response (200 OK)

```json
{
  "success": true,
  "message": "presence updated to available"
}
```

### Error Responses

#### Invalid Presence Value (400 Bad Request)
```json
{
  "success": false,
  "message": "invalid presence value: online (must be 'available' or 'unavailable')"
}
```

#### Server Not Ready (503 Service Unavailable)
```json
{
  "success": false,
  "message": "server not ready, current status: disconnected"
}
```

#### Missing Token (401 Unauthorized)
```json
{
  "success": false,
  "message": "token not found"
}
```

## Examples

### Set Bot as Available (Online)

```bash
curl -X POST http://localhost:31000/presence \
  -H "Content-Type: application/json" \
  -H "X-QUEPASA-TOKEN: your-bot-token-here" \
  -d '{
    "presence": "available"
  }'
```

### Set Bot as Unavailable (Offline)

```bash
curl -X POST http://localhost:31000/presence \
  -H "Content-Type: application/json" \
  -H "X-QUEPASA-TOKEN: your-bot-token-here" \
  -d '{
    "presence": "unavailable"
  }'
```

### Python Example

```python
import requests

url = "http://localhost:31000/presence"
headers = {
    "Content-Type": "application/json",
    "X-QUEPASA-TOKEN": "your-bot-token-here"
}
payload = {
    "presence": "available"
}

response = requests.post(url, json=payload, headers=headers)
print(response.json())
```

### JavaScript/Node.js Example

```javascript
const axios = require('axios');

const url = 'http://localhost:31000/presence';
const headers = {
  'Content-Type': 'application/json',
  'X-QUEPASA-TOKEN': 'your-bot-token-here'
};
const data = {
  presence: 'available'
};

axios.post(url, data, { headers })
  .then(response => console.log(response.data))
  .catch(error => console.error(error));
```

## Difference from Chat Presence

This endpoint controls the **global presence** of the bot (online/offline status), while the `/chat/presence` endpoint controls **typing indicators** in specific chats (typing text, recording audio, or paused).

| Feature | `/presence` | `/chat/presence` |
|---------|-------------|------------------|
| Scope | Global (entire account) | Per-chat |
| Purpose | Show online/offline status | Show typing/recording indicators |
| Values | `available`, `unavailable` | `text`, `audio`, `paused` |
| Requires ChatId | No | Yes |

## Notes

1. **Presence Updates**: WhatsApp may override your presence status based on actual activity
2. **Rate Limiting**: Avoid sending too many presence updates in quick succession
3. **Server Status**: The bot must be connected (`Ready` status) to send presence updates
4. **Automatic Behavior**: WhatsApp automatically sets you as unavailable after a period of inactivity

## Related Endpoints

- **POST** `/chat/presence` - Control typing indicators in specific chats
- **GET** `/bot` - Check bot connection status

## Implementation Details

The endpoint uses the WhatsApp `SendPresence` method from the whatsmeow library:
- Documentation: https://pkg.go.dev/go.mau.fi/whatsmeow#Client.SendPresence
- This is a global presence update affecting the entire account
- The presence is sent to WhatsApp's servers and broadcast to contacts

## Troubleshooting

### "Server not ready" error
- Check bot connection status with `GET /bot`
- Ensure the bot is paired and connected to WhatsApp

### Presence not updating
- WhatsApp may delay presence updates
- Check that you're using valid values: `"available"` or `"unavailable"`
- Ensure proper authentication with valid token

### Rate limiting
- If you receive errors about rate limiting, reduce the frequency of presence updates
- Recommended: Update presence only when necessary (bot startup, shutdown, or status changes)
