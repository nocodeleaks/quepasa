package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	environment "github.com/nocodeleaks/quepasa/environment"
	webserver "github.com/nocodeleaks/quepasa/webserver"
	log "github.com/sirupsen/logrus"
)

func init() {
	// Register API routes as a webserver configurator so the webserver package does
	// not need direct knowledge about API concerns.
	webserver.RegisterRouterConfigurator(Configure)

	// Log prefix resolution once at startup to make route shape explicit in logs.
	apiPrefix := environment.Settings.API.Prefix
	if apiPrefix == "" {
		log.Info("API routes initialized: prefix=/ (root)")
	} else {
		log.Infof("API routes initialized: prefix=/%s", apiPrefix)
	}
}

// Configure automatically configures API routes in the router
// if enabled in settings. This function should be called from main.go
// to avoid the webserver module needing to know specifically about API.
func Configure(r chi.Router) {
	apiPrefix := environment.Settings.API.Prefix

	r.Group(func(r chi.Router) {
		timeout := environment.Settings.API.GetAPITimeout()

		// Apply one timeout policy to all HTTP API routes, including SPA endpoints.
		r.Use(middleware.Timeout(timeout))

		/* CORS TESTING
		r.Use(cors.Handler(cors.Options{
			//AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
			//AllowedOrigins: []string{"https://*", "http://*"},
			AllowOriginFunc: func(r *http.Request, origin string) bool { return true },
			//AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			//AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			//ExposedHeaders:   []string{"Link"},
			//AllowCredentials: false,
			// MaxAge: 300, // Maximum value not ignored by any of major browsers
		}))
		*/

		// Mount API routes under the configured prefix
		r.Route("/"+apiPrefix, func(r chi.Router) {
			r.Group(RegisterAPIControllers)
			r.Group(RegisterAPIV3Controllers)
		})

		// Preserve legacy root-level routes when API_PREFIX is configured so older
		// clients keep working while newer clients can migrate to the prefixed API.
		if apiPrefix != "" {
			r.Group(RegisterAPIControllers)
			r.Group(RegisterAPIV3Controllers)
		}

		// SPA-only endpoints live under /spa so they stay clearly separated from the
		// shared/public API surface and can evolve with frontend-specific contracts.
		r.Route("/spa", func(r chi.Router) {
			r.Group(RegisterSPAControllers)
		})
	})
}
