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

		// Apply one timeout policy to all HTTP API routes.
		r.Use(middleware.Timeout(timeout))
		r.Use(APIEventMiddleware)

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

		// Mount API routes under the configured prefix.
		// The prefix is controlled exclusively by the API_PREFIX environment variable
		// (default: "api", see environment/api_settings.go). The official web client reads
		// the effective prefix from window.__QUEPASA_CONFIG__.apiBase injected at
		// serve time, so it adapts automatically.
		r.Route("/"+apiPrefix, func(r chi.Router) {
			defaultVersion := environment.Settings.API.DefaultVersion
			r.Group(func(router chi.Router) {
				RegisterAPIV5Controllers(router, defaultVersion == CurrentCanonicalAPIVersion)
			})
			r.Group(func(router chi.Router) {
				RegisterAPIControllers(router, defaultVersion == CurrentAPIVersion)
			})
			r.Group(RegisterAPIV3Controllers)
			r.Group(RegisterAuthenticatedPublicControllers)
			r.Group(RegisterAuthenticatedControllers)
		})

		// Preserve legacy root-level routes when API_PREFIX is configured so older
		// clients keep working while newer clients can migrate to the prefixed API.
		if apiPrefix != "" {
			r.Group(func(router chi.Router) {
				RegisterAPIControllers(router, true)
			})
			r.Group(RegisterAPIV3Controllers)
		}
	})
}
