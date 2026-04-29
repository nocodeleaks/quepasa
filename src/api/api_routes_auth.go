package api

import "github.com/go-chi/chi/v5"

func registerCanonicalPublicAuthRoutes(r chi.Router) {
	r.Get("/auth/config", LoginConfigController)
}

func registerCanonicalProtectedAuthRoutes(r chi.Router) {
	r.Get("/auth/session", SPASessionController)
	r.Get("/auth/account", SPAAccountController)
	r.Get("/auth/masterkey/status", SPAMasterKeyStatusController)
}
