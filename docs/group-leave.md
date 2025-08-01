# Leave Group Functionality Documentation

## Overview
This document describes the implementation of the **Leave Group** feature for the Quepasa WhatsApp API, which allows the bot to leave WhatsApp groups.

## Feature Implementation

### API Endpoint
- **Method**: `POST`
- **URL**: `/v3/groups/leave`
- **Content-Type**: `application/json`

### Request Structure
```go
type LeaveGroupRequest struct {
    ChatId string `json:"chatId"` // Required: Group Chat ID to leave
}
```

### Example Request
```json
{
    "chatId": "120363027885165765@g.us"
}
```

### Implementation Details
- **Controller**: `LeaveGroupController` in `api_handlers+GroupsController.go`
- **Group Manager Method**: `LeaveGroup` in `whatsmeow_group_manager.go`
- **Validation**: Validates group JID format using `whatsapp.IsValidGroupId`
- **Error Handling**: Returns appropriate error messages for invalid group IDs

### Response
```json
{
    "success": true,
    "message": "successfully left the group"
}
```

## Technical Implementation

### Interface Updates
- Added `LeaveGroup(groupID string) error` to `WhatsappGroupManagerInterface`

### File Structure
```
src/
├── api/
│   ├── leave_group_request.go         # Request struct for leave group
│   ├── api_handlers+GroupsController.go   # Leave group controller
│   └── api_handlers.go                # Route definitions
├── whatsmeow/
│   └── whatsmeow_group_manager.go     # Leave group method implementation
├── whatsapp/
│   └── whatsapp_group_manager_interface.go # Interface updates
└── models/
    └── qp_group_manager.go            # QP layer group manager updates
```

### Implementation Flow
1. **Request Validation**: Controller validates JSON structure and required fields
2. **Group ID Validation**: Validates group JID format and server type
3. **Leave Operation**: Calls WhatsApp leave group API
4. **Response**: Returns success or error message

### Error Handling
- Input validation (required field: chatId)
- Group JID format validation
- Group server type verification (@g.us)
- Connection availability checks
- WhatsApp API error propagation
- Consistent error response format

## Usage Examples

### Leave Group via cURL
```bash
curl -X POST "http://localhost:31000/v3/groups/leave" \
  -H "Content-Type: application/json" \
  -H "X-QUEPASA-TOKEN: your-token-here" \
  -d '{
    "chatId": "120363027885165765@g.us"
  }'
```

### JavaScript Example
```javascript
const response = await fetch('http://localhost:31000/v3/groups/leave', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-QUEPASA-TOKEN': 'your-token-here'
  },
  body: JSON.stringify({
    chatId: '120363027885165765@g.us'
  })
});

const result = await response.json();
console.log(result);
```

### Python Example
```python
import requests

url = "http://localhost:31000/v3/groups/leave"
headers = {
    "Content-Type": "application/json",
    "X-QUEPASA-TOKEN": "your-token-here"
}
data = {
    "chatId": "120363027885165765@g.us"
}

response = requests.post(url, json=data, headers=headers)
result = response.json()
print(result)
```

## Error Responses

### Missing Chat ID
```json
{
    "success": false,
    "message": "chatId is required"
}
```

### Invalid Group JID Format
```json
{
    "success": false,
    "message": "invalid group JID format: [provided_id]"
}
```

### Leave Operation Failed
```json
{
    "success": false,
    "message": "failed to leave group: [error details]"
}
```

### Connection Error
```json
{
    "success": false,
    "message": "client not defined"
}
```

## Group ID Format
WhatsApp group IDs follow this pattern:
- Format: `{unique_id}@g.us`
- Example: `120363027885165765@g.us`
- The `@g.us` suffix indicates it's a group chat
- The numeric prefix is the unique group identifier

## Notes and Limitations
- Requires an active WhatsApp connection
- The bot must be a participant in the group to leave it
- Once left, the bot cannot rejoin without being re-invited
- Group leaving is immediate and cannot be undone
- The action is visible to all group participants
- All operations are logged for debugging purposes
- Admin privileges are not required to leave a group

## Related Operations
For other group management operations, see:
- Group creation: `POST /v3/groups/create`
- Group info: `GET /v3/groups/get`
- Group participants: `PUT /v3/groups/participants`
- Group settings: `PUT /v3/groups/name`, `PUT /v3/groups/description`, `PUT /v3/groups/photo`
