# WebServer Instruction

## Scope
- Module: `src/webserver`
- Responsibility: HTTP server initialization, routing, middleware, websocket handling, static/web assets integration.

## Configuration Rules
- Use `environment.Settings.WebServer.Port` for port.
- Use `environment.Settings.WebServer.Host` for host.
- Use `environment.Settings.WebServer.Logs` for HTTP logging behavior.
- Do not introduce parallel config sources for webserver runtime.

## Implementation Rules
- Keep Chi router as the routing layer.
- Keep middleware composition centralized and deterministic.
- Preserve centralized error handling behavior.
- Keep CORS behavior compatible with API routes.
- Keep graceful shutdown behavior intact.

## Swagger Rules
- Keep both `/swagger` and `/swagger/` routes working.
- Avoid redirect loops for swagger routes.
- Keep swagger UI behavior stable with hidden models default.
- Keep integration aligned with generated swagger artifacts in `src/swagger`.

## Key Files
- `src/webserver/webserver.go`
- `src/webserver/websocket.go`
- `src/webserver/middleware.go`
