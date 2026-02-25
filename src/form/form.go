package form

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	environment "github.com/nocodeleaks/quepasa/environment"
	webserver "github.com/nocodeleaks/quepasa/webserver"
)

var viewsBasePath string

func init() {
	// Automatically registers the Form configuration in the webserver
	// This allows Form to be configured without the webserver module
	// needing to know specifically about Form
	webserver.RegisterRouterConfigurator(Configure)
}

// GetViewPath returns the full path to a view file
func GetViewPath(viewPath string) string {
	return filepath.Join(viewsBasePath, viewPath)
}

// Configure automatically configures Form routes in the router
// if enabled in settings. This function should be called from main.go
// to avoid the webserver module needing to know specifically about Form.
func Configure(r chi.Router) {
	if environment.Settings.Form.Enabled {

		// Set views base path relative to the current working directory
		// This ensures templates are found regardless of where the executable is run from
		workDir, _ := os.Getwd()
		viewsBasePath = filepath.Join(workDir, "views")

		// Form routes, extra content
		ServeForms(r)
	}
}

func ServeForms(r chi.Router) {

	// Static Forms content (skip in dev frontend mode - SPA proxy handles assets)
	if v := os.Getenv("QUEPASA_DEV_FRONTEND"); v != "1" && v != "true" {
		ServeStaticContent(r)
	}

	// setting group
	r.Group(func(r chi.Router) {

		// setting timeout for the group
		r.Use(middleware.Timeout(30 * time.Second))

		// web routes
		// authenticated web routes
		r.Group(RegisterFormAuthenticatedControllers)

		// unauthenticated web routes
		r.Group(RegisterFormControllers)
	})
}

func ServeStaticContent(r chi.Router) {

	// setting group
	r.Group(func(r chi.Router) {

		// static files
		workDir, _ := os.Getwd()
		assetsDir := filepath.Join(workDir, "assets")
		root := http.Dir(assetsDir)

		path := "/assets"

		if strings.ContainsAny(path, "{}*") {
			panic("FileServer does not permit URL parameters.")
		}

		fs := http.StripPrefix(path, http.FileServer(root))
		if path != "/" && path[len(path)-1] != '/' {
			r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
			path += "/"
		}
		path += "*"
		r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fs.ServeHTTP(w, r)
		}))
	})
}
