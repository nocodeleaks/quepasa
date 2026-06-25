package cable

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	dispatchservice "github.com/nocodeleaks/quepasa/dispatch/service"
	webserver "github.com/nocodeleaks/quepasa/webserver"
)

// CableHub is the singleton websocket cable hub used by the application.
//
// The hub owns every active client connection, subscription index, and command
// registry. Keeping a single shared instance allows the model layer to publish
// events without needing to know anything about HTTP routing details.
var CableHub = NewHub()

func init() {
	// Register the websocket transport in the main router without forcing the
	// webserver package to know about this feature explicitly.
	webserver.RegisterRouterConfigurator(Configure)

	// Register the hub as a realtime publisher so model-layer events can reach
	// websocket clients through a transport-neutral callback interface.
	dispatchservice.RegisterRealtimePublisher(CableHub)
}

// Configure mounts the authenticated websocket cable endpoint.
//
// `/cable` is the single transport route because the websocket bus is not a
// SPA-only concern and we are defining this contract now instead of preserving
// an older websocket surface.
func Configure(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(CableVerifier())
		r.Use(CableAuthenticator)
		r.Get("/cable", ServeCable)
	})
}

// ServeCable upgrades an authenticated HTTP request to the cable websocket.
func ServeCable(w http.ResponseWriter, r *http.Request) {
	user, err := GetCableUser(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	CableHub.ServeWS(w, r, user)
}
