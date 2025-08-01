# Edit Message Functionality Documentation

## Overview
This document describes the implementation of the **Edit Message** feature for the Quepasa WhatsApp API, which allows editing of previously sent messages.

## Feature Implementation

### API Endpoint
- **Method**: `PUT`
- **URL**: `/v3/message/edit`
- **Content-Type**: `application/json`

### Request Structure
```go
type EditMessageRequest struct {
    MessageId string `json:"messageId"` // Required: Message ID to edit
    Content   string `json:"content"`   // Required: New content for the message
}
```

### Example Request
```json
{
    "messageId": "3EB0C7F0A5D0C7F0A5D0C7F0A5D0C7F0",
    "content": "This is the edited message content"
}
```

### Implementation Details
- **Controller**: `EditMessageController` in `api_handlers+MessageController.go`
- **Connection Method**: `Edit` in `whatsmeow_connection.go`
- **Validation**: Validates message ID and content presence
- **Error Handling**: Returns appropriate error messages for invalid requests

### Response
```json
{
    "success": true,
    "message": "message edited successfully"
}
```

## Technical Implementation

### Interface Updates
- Added `Edit(IWhatsappMessage, string) error` to `IWhatsappConnection` interface

### File Structure
```
src/
├── api/
│   ├── edit_message_request.go        # Request struct for edit message
│   ├── api_handlers+MessageController.go  # Edit message controller
│   └── api_handlers.go                # Route definitions
├── whatsmeow/
│   └── whatsmeow_connection.go        # Edit method implementation
└── whatsapp/
    └── whatsapp_connection_interface.go    # Interface updates
```

### Implementation Flow
1. **Request Validation**: Controller validates JSON structure and required fields
2. **Message Retrieval**: Fetches original message using message ID
3. **Edit Operation**: Calls WhatsApp edit API with new content
4. **Response**: Returns success or error message

### Error Handling
- Input validation (required fields: messageId, content)
- Message existence verification
- Connection availability checks
- WhatsApp API error propagation
- Consistent error response format

## Usage Examples

### Edit Message via cURL
```bash
curl -X PUT "http://localhost:31000/v3/message/edit" \
  -H "Content-Type: application/json" \
  -H "X-QUEPASA-TOKEN: your-token-here" \
  -d '{
    "messageId": "3EB0C7F0A5D0C7F0A5D0C7F0A5D0C7F0",
    "content": "Updated message content"
  }'
```

### JavaScript Example
```javascript
const response = await fetch('http://localhost:31000/v3/message/edit', {
  method: 'PUT',
  headers: {
    'Content-Type': 'application/json',
    'X-QUEPASA-TOKEN': 'your-token-here'
  },
  body: JSON.stringify({
    messageId: '3EB0C7F0A5D0C7F0A5D0C7F0A5D0C7F0',
    content: 'Updated message content'
  })
});

const result = await response.json();
console.log(result);
```

## Error Responses

### Missing Message ID
```json
{
    "success": false,
    "message": "messageId is required"
}
```

### Missing Content
```json
{
    "success": false,
    "message": "content is required"
}
```

### Message Not Found
```json
{
    "success": false,
    "message": "message not found: [error details]"
}
```

### Edit Failed
```json
{
    "success": false,
    "message": "failed to edit message: [error details]"
}
```

## Notes and Limitations
- Requires an active WhatsApp connection
- Message editing is subject to WhatsApp's editing limitations:
  - Time limits (typically 15 minutes after sending)
  - Message type restrictions (text messages only)
  - Edit history is maintained by WhatsApp
- Only messages sent by the bot can be edited
- All operations are logged for debugging purposes
- Edited messages will show "edited" indicator in WhatsApp clients
