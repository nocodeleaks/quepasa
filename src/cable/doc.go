// Package cable implements the realtime command/event transport used by the
// application websocket bus.
//
// The package owns websocket authentication, connection fan-out, subscription
// management, command handling, and transport-specific payloads. It should not
// depend on HTTP request DTOs from the REST API layer.
package cable
