# API Module

## Overview

REST API endpoints for QuePasa WhatsApp bot platform.

## Structure

### API Versions

- **v1**: Legacy endpoints (deprecated)
- **v3**: Latest stable version
- **latest**: Non-versioned routes (current development)

### Controller Pattern

Controllers follow the naming convention: `api_handlers+*Controller.go`

Examples:
- `api_handlers+AccountController.go`
- `api_handlers+MessageController.go`
- `api_handlers+GroupsController.go`

## Swagger Documentation

### Automatic Generation

The API documentation is automatically generated from code comments using `swaggo/swag`.

### Comment Format

```go
// @Summary      Brief description of the endpoint
// @Description  Detailed description
// @Tags         category
// @Accept       json
// @Produce      json
// @Param        paramName paramType dataType required "Description"
// @Success      200 {object} ResponseModel
// @Failure      400 {object} ErrorResponse
// @Router       /api/v3/endpoint [method]
func HandlerFunction(w http.ResponseWriter, r *http.Request) {
    // Implementation
}
```

### Access Swagger UI

- URL: `http://localhost:31000/swagger`
- Alternative: `http://localhost:31000/swagger/`
- OpenAPI Spec: `http://localhost:31000/swagger/doc.json`

### Generate Documentation

```bash
# Using swag CLI
swag init -g main.go --output ./swagger

# Using VS Code task
# Run task: "Generate Swagger Docs"
```

## Development Guidelines

For AI agent-specific instructions, check `/.github/instructions/` and use the file whose first tag matches the module context.

### Key Patterns

- Latest routes in files without version suffix (e.g., `api_handlers.go`)
- Use centralized `environment.Settings` for configuration
- Follow controller naming conventions
- Document all endpoints with Swagger comments

## Integration

- Uses Chi router from webserver module
- Integrates with metrics for Prometheus monitoring
- Supports webhook and RabbitMQ dispatching
- Handles multiple WhatsApp accounts via `QpWhatsappServer`
