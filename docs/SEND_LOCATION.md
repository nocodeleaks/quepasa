# Location Messages Documentation

## Overview

QuePasa supports sending location messages through WhatsApp using the whatsmeow library. This feature allows you to send geographic coordinates with optional metadata like name, address, and URL.

## API Endpoint

**POST** `/v3/bot/{token}/send`

## Request Format

### Headers
```
Content-Type: application/json
X-QUEPASA-CHATID: <phone_number>@s.whatsapp.net
```

### JSON Body
```json
{
  "location": {
    "latitude": -23.550520,
    "longitude": -46.633308,
    "name": "São Paulo",
    "address": "Av. Paulista, 1578 - Bela Vista, São Paulo - SP",
    "url": "https://maps.google.com/?q=-23.550520,-46.633308"
  }
}
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `latitude` | float64 | Yes | Geographic latitude in degrees (-90 to 90) |
| `longitude` | float64 | Yes | Geographic longitude in degrees (-180 to 180) |
| `name` | string | No | Location name or title (displayed in chat) |
| `address` | string | No | Full address of the location |
| `url` | string | No | URL to the location (usually Google Maps link) |

## Examples

### Minimal Example
```json
{
  "location": {
    "latitude": -23.550520,
    "longitude": -46.633308
  }
}
```

### Complete Example
```json
{
  "location": {
    "latitude": -23.550520,
    "longitude": -46.633308,
    "name": "Avenida Paulista",
    "address": "Av. Paulista, 1578 - Bela Vista, São Paulo - SP, 01310-200",
    "url": "https://maps.google.com/?q=-23.550520,-46.633308"
  }
}
```

### Using PowerShell
```powershell
$headers = @{
    "Content-Type" = "application/json"
    "X-QUEPASA-CHATID" = "5535900000000@s.whatsapp.net"
}

$body = @{
    location = @{
        latitude = -23.550520
        longitude = -46.633308
        name = "São Paulo"
        address = "Av. Paulista, 1578 - Bela Vista, São Paulo - SP"
        url = "https://maps.google.com/?q=-23.550520,-46.633308"
    }
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:31000/v3/bot/YOUR_TOKEN/send" -Method Post -Headers $headers -Body $body
```

### Using cURL
```bash
curl -X POST "http://localhost:31000/v3/bot/YOUR_TOKEN/send" \
  -H "Content-Type: application/json" \
  -H "X-QUEPASA-CHATID: 5535900000000@s.whatsapp.net" \
  -d '{
    "location": {
      "latitude": -23.550520,
      "longitude": -46.633308,
      "name": "São Paulo",
      "address": "Av. Paulista, 1578",
      "url": "https://maps.google.com/?q=-23.550520,-46.633308"
    }
  }'
```

## Response

### Success Response (200 OK)
```json
{
  "result": {
    "id": "3EB0ABC123DEF456",
    "chat": {
      "id": "5535900000000@s.whatsapp.net",
      "phone": "5535900000000"
    },
    "text": "São Paulo",
    "type": "location",
    "fromMe": true,
    "timestamp": 1696615906
  }
}
```

### Error Response (400 Bad Request)
```json
{
  "error": "latitude and longitude are required for location messages"
}
```

## Implementation Details

### Data Flow
1. **API Request** → `QpSendRequest` with `Location` field
2. **Model Conversion** → `ToWhatsappMessage()` creates `WhatsappMessage` with `LocationMessageType`
3. **Type Preservation** → Type is maintained through validation and processing
4. **WhatsApp Send** → `WhatsmeowConnection.Send()` creates protobuf `LocationMessage`
5. **Delivery** → whatsmeow sends to WhatsApp servers

### Key Components

#### 1. WhatsappLocation Struct
```go
type WhatsappLocation struct {
    Latitude  float64 `json:"latitude"`
    Longitude float64 `json:"longitude"`
    Name      string  `json:"name,omitempty"`
    Address   string  `json:"address,omitempty"`
    URL       string  `json:"url,omitempty"`
}
```

#### 2. QpSendRequest
- Contains `Location *WhatsappLocation` field
- `ToWhatsappMessage()` converts to internal format
- Sets `Type = LocationMessageType`
- Returns early to prevent type override

#### 3. WhatsmeowConnection
- Detects `LocationMessageType` before media processing
- Creates `waE2E.LocationMessage` protobuf
- Maps fields: `DegreesLatitude`, `DegreesLongitude`, `Name`
- Sends directly without upload

### Type Safety
The implementation ensures location messages are never treated as file attachments:
- Early type detection in `SendWithMessageType()`
- Priority check in `WhatsmeowConnection.Send()`
- Direct protobuf creation without upload

## Validation Rules

1. **Required Fields**: `latitude` and `longitude` must be present
2. **Valid Ranges**: 
   - Latitude: -90 to 90 degrees
   - Longitude: -180 to 180 degrees
3. **Optional Fields**: All other fields (name, address, url) are optional

## Common Use Cases

### 1. Share Business Location
```json
{
  "location": {
    "latitude": -23.550520,
    "longitude": -46.633308,
    "name": "QuePasa Office",
    "address": "Av. Paulista, 1578 - São Paulo, SP"
  }
}
```

### 2. Share Meeting Point
```json
{
  "location": {
    "latitude": -23.561414,
    "longitude": -46.655882,
    "name": "Meeting Point - 3 PM"
  }
}
```

### 3. Share Point of Interest
```json
{
  "location": {
    "latitude": -22.970722,
    "longitude": -43.182365,
    "name": "Christ the Redeemer",
    "url": "https://maps.google.com/?q=-22.970722,-43.182365"
  }
}
```

## Troubleshooting

### Issue: "text not found, do not send empty messages"
**Solution**: Ensure you're sending the `location` object, not trying to send empty text.

### Issue: Message sent as text instead of location
**Solution**: Verify that both `latitude` and `longitude` are numbers, not strings.

### Issue: Location not showing on map
**Solution**: Check that coordinates are valid and within proper ranges.

## Technical Notes

### WhatsApp Protobuf
Location messages use the `LocationMessage` type from WhatsApp's protobuf schema:
```protobuf
message LocationMessage {
    optional double degreesLatitude = 1;
    optional double degreesLongitude = 2;
    optional string name = 3;
    optional string address = 4;
    optional string url = 5;
    optional ContextInfo contextInfo = 17;
}
```

### Message Type Constant
```go
const LocationMessageType WhatsappMessageType = "location"
```

### MIME Type
Location attachments use: `text/x-uri; location`

## Version History

- **v3.25.XXXX.XXXX**: Initial implementation of location messages
- Support for optional fields: name, address, url
- Integration with existing message type system
- Full compatibility with LID (Limited Identifier) chat IDs

## Related Documentation

- [Webhook System](WEBHOOK_SYSTEM_DOCUMENTATION.md)
- [Chat Management](CHAT_MANAGEMENT.md)
- [API Documentation](../src/swagger/)

## Support

For issues or questions about location messages:
1. Check this documentation first
2. Review example code in `tests/test-location-send.ps1`
3. Enable debug logging to trace message flow
4. Open an issue on GitHub with logs and request details
