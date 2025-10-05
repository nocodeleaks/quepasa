# WebServer Module - AI Agent Instructions

## Module Scope
HTTP server, routing, middleware, forms, websockets, static assets.

## Configuration
- Use `environment.Settings.WebServer.Port` (uint32)
- Use `environment.Settings.WebServer.Host` (string)
- Use `environment.Settings.WebServer.Logs` (bool) - HTTP request logging

## Key Files
- `webserver.go`: Main server initialization and routing
- `websocket.go`: WebSocket connection handling
- `middleware.go`: HTTP middleware stack

## Patterns
- Chi router for HTTP routing
- Middleware composition for request processing
- Centralized error handling
- CORS configuration for API access

## Swagger Integration
- Both `/swagger` and `/swagger/` routes supported
- No redirect loops (direct handler registration)
- Models hidden by default in UI
- Auto-generated from code comments

## Code Style
- Use `environment.Settings.WebServer.*` for all configuration
- Log configuration changes at startup
- Handle graceful shutdown
- Metrics integration via Prometheus
