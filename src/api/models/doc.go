// Package api contains HTTP transport contracts used by the API module.
//
// These types intentionally live outside the shared `models` package so request
// and response payloads can evolve without polluting domain/runtime code with
// delivery-specific shapes.
package api
