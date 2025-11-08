# API Module - AI Agent Instructions

## Module Scope
REST API endpoints, GraphQL, gRPC, HTTP handlers, controllers.

## Structure
- Latest routes: files without version suffix (e.g., `api_handlers.go`)
- Controllers: `api_handlers+*Controller.go` pattern
- Versions: v1 (deprecated), v2, v3, latest

## Swagger Comments
- Required: `@Summary`, `@Description`, `@Tags`, `@Accept`, `@Produce`
- Parameters: `@Param name location type required "description"`
- Responses: `@Success`, `@Failure` with model types
- Router: `@Router /path [method]`

## Patterns
- Use centralized `environment.Settings.*` for configuration
- Integrate metrics for monitoring
- Support webhook and RabbitMQ dispatching
- Handle multiple WhatsApp accounts

## Error Handling
- Return structured JSON errors
- Include error codes and messages
- Log errors with context
- Use appropriate HTTP status codes

## Development
- Document all endpoints with Swagger comments
- Follow controller naming conventions
- Keep versioned and latest endpoints separate
- Test with multiple API versions

## Toggle System

### Current Implementation
The API provides toggle functionality to control WhatsApp event processing through boolean options.

### Available Toggles
- **Groups**: Controls processing of group messages
- **Broadcasts**: Controls processing of broadcast messages
- **ReadReceipts**: Controls read receipt confirmations
- **Calls**: Controls call event processing
- **Devel**: Controls debug mode (devel) for enhanced logging

### Toggle States
- `unset` (0): Not configured (default)
- `true` (1): Enabled
- `false` (-1): Disabled
- `forcedtrue` (2): Force enabled
- `forcedfalse` (-2): Force disabled

### API Endpoints
- **GET** `/bot/{token}/command?action={toggle_name}` - Toggle via command
  - Actions: `groups`, `broadcasts`, `readreceipts`, `calls`, `debug`
  - Response: `{"success": true, "status": "groups toggled: true"}`

### Web Interface
- **POST** `/form/toggle?key=server-{toggle_name}&token={token}` - Web form toggle
- Used in account management interface for server-level configuration

### Usage Locations
- **Server Level**: Applied to entire WhatsApp server instance
- **Webhook Level**: Can be configured per webhook endpoint
- **RabbitMQ Level**: Can be configured per RabbitMQ connection

### Implementation Files
- `models/qp_whatsapp_extensions.go`: Toggle functions (`ToggleGroups`, `ToggleBroadcasts`, etc.)
- `api/api_handlers.go`: Command controller integration
- `form/form_extensions.go`: Web form toggle controller
- `whatsapp/whatsapp_options_extended.go`: Option evaluation logic

## Planned Improvements: PATCH API for Server Configuration

### New Endpoint
- **PATCH** `/bot/{token}/config` - Update server configuration options
- Body: JSON with toggle options to update

### Request Format
```json
{
  "groups": "true",
  "broadcasts": "false",
  "readReceipts": "unset",
  "calls": "forcedtrue",
  "devel": true
}
```

### Response Format
```json
{
  "success": true,
  "updated": {
    "groups": "true",
    "broadcasts": "false",
    "devel": true
  },
  "message": "Server configuration updated successfully"
}
```

### Implementation Plan
1. Create new PATCH endpoint in `api_handlers.go`
2. Add request/response models for configuration updates
3. Implement validation for toggle values
4. Add proper error handling and logging
5. Update Swagger documentation
6. Test with all toggle combinations
