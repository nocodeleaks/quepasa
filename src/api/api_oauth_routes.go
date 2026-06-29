package api

import (
	"github.com/go-chi/chi/v5"
	oauth "github.com/nocodeleaks/quepasa/oauth"
)

// RegisterOAuthRoutes mounts the OAuth login/callback handlers. These routes are
// public (no auth middleware) and mounted outside the /api prefix so the callback
// URL is stable regardless of API_PREFIX configuration.
func RegisterOAuthRoutes(r chi.Router) {
	if !oauth.IsEnabled() {
		return
	}

	r.Get("/oauth/login", oauth.OAuthLoginHandler)
	r.Get("/oauth/callback", oauth.OAuthCallbackHandler)
}
