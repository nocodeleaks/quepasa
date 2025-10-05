# WebServer Module

## Overview
HTTP server module for QuePasa, handli## Integration
- Uses centralized `environment.Settings.WebServer` for configuration
- HTTP logging controlled by `environment.Settings.WebServer.Logs`
- Integrates with metrics for Prometheus monitoring
- Supports multiple API versions (v1, v2, v3, latest)eb interface, API endpoints, and WebSocket connections.

## Configuration

### Environment Variables

#### New Variables (Recommended)
- `WEBSERVER_PORT` (uint32): Port for the web server (default: 31000)
- `WEBSERVER_HOST` (string): Host address to bind (default: "0.0.0.0")
- `WEBSERVER_LOGS` (bool): Enable HTTP request logging (default: false)

#### Legacy Variables (Backward Compatibility)
- `WEBAPIPORT`: Falls back to this if `WEBSERVER_PORT` not set
- `WEBAPIHOST`: Falls back to this if `WEBSERVER_HOST` not set

### Type Safety
- Port uses `uint32` to prevent negative values
- Automatic parsing with error logging via `getEnvOrDefaultUint32()`

### Configuration Access
```go
import "github.com/nocodeleaks/quepasa/environment"

// Access centralized settings
port := environment.Settings.WebServer.Port
host := environment.Settings.WebServer.Host
logs := environment.Settings.WebServer.Logs
```

## Features
- Chi HTTP router with middleware support
- Swagger/OpenAPI documentation
- WebSocket connections for real-time updates
- Form handling and validation
- Static asset serving
- CORS configuration
- Request logging and metrics

## Swagger Documentation

### Access Points
- `/swagger` - Swagger UI
- `/swagger/` - Alternative access (no redirect loop)
- `/swagger/doc.json` - OpenAPI spec

### Configuration
- Models section hidden by default (`defaultModelsExpandDepth: "-1"`)
- Auto-generated from code comments using `swaggo/swag`

### Generating Documentation
```bash
# Using VS Code task
# Run: "Generate Swagger Docs"

# Or manually
swag init -g main.go --output ./swagger
```

### Comment Standards
```go
// @Summary      Brief description
// @Description  Detailed description
// @Tags         tag1,tag2
// @Accept       json
// @Produce      json
// @Param        id path string true "ID"
// @Success      200 {object} ResponseType
// @Router       /endpoint/{id} [get]
```

## Integration
- Uses centralized `environment.Settings.WebServer` for configuration
- Integrates with metrics module for Prometheus monitoring
- Supports multiple API versions (v1, v2, v3, latest)

## Development
- Check `AGENTS.md` for AI agent-specific guidelines
- Controller pattern: `api_handlers+*Controller.go` for API endpoints
- Latest routes in non-versioned files (e.g., `api_handlers.go`)
