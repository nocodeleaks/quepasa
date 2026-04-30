package api

import "net/http"

// SPASessionsController returns the user's sessions, including disconnected records.
func SPASessionsController(w http.ResponseWriter, r *http.Request) {
	SPAServersController(w, r)
}

// SPASessionsSearchController performs lightweight session-side filtering for SPA clients.
func SPASessionsSearchController(w http.ResponseWriter, r *http.Request) {
	SPAServersSearchController(w, r)
}

// SPASessionCreateController creates a new pre-configured session owned by the SPA user.
func SPASessionCreateController(w http.ResponseWriter, r *http.Request) {
	SPAServerCreateController(w, r)
}

// SPASessionGetController returns session information for a token owned by the user.
func SPASessionGetController(w http.ResponseWriter, r *http.Request) {
	SPAServerInfoController(w, r)
}

// SPASessionUpdateController patches persisted session configuration for the SPA user.
func SPASessionUpdateController(w http.ResponseWriter, r *http.Request) {
	SPAServerUpdateController(w, r)
}

// SPASessionDeleteController deletes a session owned by the SPA user.
func SPASessionDeleteController(w http.ResponseWriter, r *http.Request) {
	SPAServerDeleteController(w, r)
}

// SPASessionQRCodeController returns a QR code payload for a session that is not yet ready.
func SPASessionQRCodeController(w http.ResponseWriter, r *http.Request) {
	SPAServerQRCodeController(w, r)
}

// SPASessionPairCodeController returns a phone pairing code for a session token owned by the user.
func SPASessionPairCodeController(w http.ResponseWriter, r *http.Request) {
	SPAServerPairCodeController(w, r)
}

// SPASessionEnableController starts a session through the SPA HTTP surface.
func SPASessionEnableController(w http.ResponseWriter, r *http.Request) {
	SPAServerEnableController(w, r)
}

// SPASessionDisableController stops a session through the SPA HTTP surface.
func SPASessionDisableController(w http.ResponseWriter, r *http.Request) {
	SPAServerDisableController(w, r)
}

// SPASessionDebugToggleController toggles session debug mode through the SPA auth surface.
func SPASessionDebugToggleController(w http.ResponseWriter, r *http.Request) {
	SPAServerDebugToggleController(w, r)
}

// SPASessionOptionToggleController toggles a persisted session option explicitly by name.
func SPASessionOptionToggleController(w http.ResponseWriter, r *http.Request) {
	SPAServerOptionToggleController(w, r)
}
