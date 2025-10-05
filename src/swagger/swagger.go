package swagger

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/nocodeleaks/quepasa/environment"
	"github.com/nocodeleaks/quepasa/webserver"
	httpSwagger "github.com/swaggo/http-swagger"
)

func init() {
	// Automatically registers the Swagger configuration in the webserver
	// This allows Swagger to be configured without the webserver module
	// needing to know specifically about Swagger
	webserver.RegisterRouterConfigurator(Configure)
}

// Configure automatically configures Swagger UI in the router
// if enabled in settings. This function should be called from main.go
// to avoid the webserver module needing to know specifically about Swagger.
func Configure(r chi.Router) {
	if environment.Settings.Swagger.Enabled {
		ServeSwaggerUI(r)
	}
}

// ServeSwaggerJSON serves the swagger.json with dynamic basePath based on API_PREFIX environment variable
func ServeSwaggerJSON(w http.ResponseWriter, r *http.Request, apiPrefix string) {
	w.Header().Set("Content-Type", "application/json")

	// Read the swagger.json file
	swaggerPath := filepath.Join("swagger", "swagger.json")
	spec, err := ioutil.ReadFile(swaggerPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Parse the JSON
	var swaggerDoc map[string]interface{}
	if err := json.Unmarshal(spec, &swaggerDoc); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update the basePath dynamically
	swaggerDoc["basePath"] = "/" + apiPrefix + "/"

	// Convert back to JSON
	modifiedSpec, err := json.Marshal(swaggerDoc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(modifiedSpec)
}

// ServeSwaggerUI configures Swagger UI routes in the provided router
// Should only be called if IsSwaggerEnabled() returns true
func ServeSwaggerUI(r chi.Router) {
	prefix := environment.Settings.Swagger.Prefix
	apiPrefix := environment.Settings.API.Prefix

	// Serve swagger.json with dynamic basePath
	r.Get("/swagger.json", func(w http.ResponseWriter, req *http.Request) {
		ServeSwaggerJSON(w, req, apiPrefix)
	})

	// Handle both /{prefix} and /{prefix}/ routes
	r.Get("/"+prefix, func(w http.ResponseWriter, req *http.Request) {
		// Serve the Swagger UI directly for /{prefix}
		httpSwagger.Handler(
			httpSwagger.URL("/swagger.json"), // Use our dynamic swagger.json
			httpSwagger.UIConfig(map[string]string{
				"defaultModelsExpandDepth": "-1", // Hide models section
			}),
		).ServeHTTP(w, req)
	})

	// Configure Swagger UI to hide models/schemas section
	r.Mount("/"+prefix+"/", httpSwagger.Handler(
		httpSwagger.URL("/swagger.json"), // Use our dynamic swagger.json
		httpSwagger.UIConfig(map[string]string{
			"defaultModelsExpandDepth": "-1", // Hide models section
		}),
	))
}
