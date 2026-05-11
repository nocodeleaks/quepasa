package api

import "net/http"

// AuthenticatedSessionsController returns the user's sessions, including disconnected records.
func AuthenticatedSessionsController(w http.ResponseWriter, r *http.Request) {
	AuthenticatedServersController(w, r)
}

// AuthenticatedSessionsSearchController performs lightweight session-side filtering.
func AuthenticatedSessionsSearchController(w http.ResponseWriter, r *http.Request) {
	AuthenticatedServersSearchController(w, r)
}

// AuthenticatedSessionCreateController creates a new pre-configured session owned by the authenticated user.
func AuthenticatedSessionCreateController(w http.ResponseWriter, r *http.Request) {
	AuthenticatedServerCreateController(w, r)
}

// AuthenticatedSessionGetController returns session information for a token owned by the user.
func AuthenticatedSessionGetController(w http.ResponseWriter, r *http.Request) {
	AuthenticatedServerInfoController(w, r)
}

// AuthenticatedSessionUpdateController patches persisted session configuration for the authenticated user.
func AuthenticatedSessionUpdateController(w http.ResponseWriter, r *http.Request) {
	AuthenticatedServerUpdateController(w, r)
}

// AuthenticatedSessionDeleteController deletes a session owned by the authenticated user.
func AuthenticatedSessionDeleteController(w http.ResponseWriter, r *http.Request) {
	AuthenticatedServerDeleteController(w, r)
}

// AuthenticatedSessionQRCodeController returns a QR code payload for a session that is not yet ready.
func AuthenticatedSessionQRCodeController(w http.ResponseWriter, r *http.Request) {
	AuthenticatedServerQRCodeController(w, r)
}

// AuthenticatedSessionPairCodeController returns a phone pairing code for a session token owned by the user.
func AuthenticatedSessionPairCodeController(w http.ResponseWriter, r *http.Request) {
	AuthenticatedServerPairCodeController(w, r)
}

// AuthenticatedSessionEnableController starts a session through the authenticated HTTP surface.
func AuthenticatedSessionEnableController(w http.ResponseWriter, r *http.Request) {
	AuthenticatedServerEnableController(w, r)
}

// AuthenticatedSessionDisableController stops a session through the authenticated HTTP surface.
func AuthenticatedSessionDisableController(w http.ResponseWriter, r *http.Request) {
	AuthenticatedServerDisableController(w, r)
}

// AuthenticatedSessionDebugToggleController toggles session debug mode through the authenticated API surface.
func AuthenticatedSessionDebugToggleController(w http.ResponseWriter, r *http.Request) {
	AuthenticatedServerDebugToggleController(w, r)
}

// AuthenticatedSessionOptionToggleController toggles a persisted session option explicitly by name.
func AuthenticatedSessionOptionToggleController(w http.ResponseWriter, r *http.Request) {
	AuthenticatedServerOptionToggleController(w, r)
}
