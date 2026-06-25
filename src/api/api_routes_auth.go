package api

import "github.com/go-chi/chi/v5"

func registerCanonicalPublicAuthRoutes(r chi.Router) {
	r.Get("/auth/config", LoginConfigController)
	r.Post("/auth/login", CanonicalLoginPostController)
}

func registerCanonicalProtectedAuthRoutes(r chi.Router) {
	r.Get("/auth/session", AuthenticatedSessionController)
	r.Get("/auth/account", AuthenticatedAccountController)
	r.Get("/auth/masterkey/status", AuthenticatedMasterKeyStatusController)
}
