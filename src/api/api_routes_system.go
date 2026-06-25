package api

import "github.com/go-chi/chi/v5"

func registerCanonicalSystemRoutes(r chi.Router) {
	r.Get("/system/health", HealthController)
	r.Head("/system/health", HealthController)
	r.Get("/system/version", VersionController)
	r.Get("/system/environment", EnvironmentController)

	// Alias routes for convenience
	r.Get("/health", HealthController)
	r.Head("/health", HealthController)
}
