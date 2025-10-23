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
