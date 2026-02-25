package webserver

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	environment "github.com/nocodeleaks/quepasa/environment"
	log "github.com/sirupsen/logrus"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// RouterConfigurator é uma função que pode configurar rotas adicionais no router
type RouterConfigurator func(r chi.Router)

// configurators armazena as funções de configuração adicionais que serão executadas
// durante a criação do router. Módulos externos podem registrar suas configurações aqui.
var configurators []RouterConfigurator

// RegisterRouterConfigurator permite que módulos externos registrem funções
// para configurar rotas adicionais no router principal
func RegisterRouterConfigurator(configurator RouterConfigurator) {
	configurators = append(configurators, configurator)
}

func WebServerStart(logentry *log.Entry) error {
	r := newRouter()
	webAPIPort := environment.Settings.WebServer.Port
	webAPIHost := environment.Settings.WebServer.Host

	var timeout = 30 * time.Second
	server := http.Server{
		Addr:         fmt.Sprintf("%s:%d", webAPIHost, webAPIPort),
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		Handler:      r,
	}

	logentry.Infof("starting web server on port: %d", webAPIPort)
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

	return err
}

func newRouter() chi.Router {
	r := chi.NewRouter()

	r.Use(MiddlewareForNormalizePaths)
	r.Use(middleware.StripSlashes)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	if environment.Settings.WebServer.Logs {
		r.Use(middleware.Logger)
	}

	r.Use(middleware.Recoverer)

	// Serve SPA root (dev proxy or static build) before other configurators
	ServeSPA(r)

	// Form and SignalR routes are now configured automatically via configurators

	// Execute registered configurators (e.g., Swagger, custom modules)
	for _, configurator := range configurators {
		configurator(r)
	}

	return r
}

// ServeSPA registers handlers to serve the frontend single-page application.
// In development (QUEPASA_DEV_FRONTEND=1) it proxies requests to the local Vite dev server.
// In production it serves files from assets/frontend and falls back to index.html for SPA routes.
func ServeSPA(r chi.Router) {
	// Check if dev proxy is enabled
	if v := os.Getenv("QUEPASA_DEV_FRONTEND"); v == "1" || v == "true" {
		viteHost := os.Getenv("QUEPASA_FRONTEND_HOST")
		if viteHost == "" {
			viteHost = "http://127.0.0.1"
		}
		vitePort := os.Getenv("QUEPASA_FRONTEND_DEV_PORT")
		if vitePort == "" {
			vitePort = "5173"
		}
		viteBasePath := os.Getenv("QUEPASA_FRONTEND_BASE_PATH")
		if viteBasePath == "" {
			viteBasePath = "/assets/frontend"
		}
		target := viteHost + ":" + vitePort
		// create reverse proxy WITHOUT base path rewrite
		proxy := NewReverseProxy(target, "")
		log.Infof("SPA dev proxy enabled, proxying to %s", target)

		// In dev mode with base: '/', we just proxy everything that isn't API or Form
		r.NotFound(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// Let API, Form and legacy API paths return 404 if not matched by registered routes
			// Legacy API paths (/info, /health, /v3, /v4, /current, /swagger, /mcp) should return 404 if not matched
			if strings.HasPrefix(req.URL.Path, "/api") || strings.HasPrefix(req.URL.Path, "/form") || isLegacyAPIPath(req.URL.Path) {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			// Proxy everything else to Vite (including all GET/POST/etc)
			log.Infof("Proxying request: %s %s -> 127.0.0.1:5173%s", req.Method, req.URL.Path, req.URL.Path)
			proxy.ServeHTTP(w, req)
		}))
		return
	}

	// Production: serve files from assets/frontend
	workDir, _ := os.Getwd()
	frontendDir := filepath.Join(workDir, "assets", "frontend")
	fs := http.FileServer(http.Dir(frontendDir))

	// Use NotFound handler to serve SPA index or static file when nothing else matched
	r.NotFound(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Only serve SPA for GET requests and not for API, assets, form or legacy API paths
		// Legacy API paths (/info, /health, /v3, /v4, /current, /swagger, /mcp) should return 404 if not matched
		if req.Method != http.MethodGet ||
			strings.HasPrefix(req.URL.Path, "/api") ||
			strings.HasPrefix(req.URL.Path, "/assets") ||
			strings.HasPrefix(req.URL.Path, "/form") ||
			isLegacyAPIPath(req.URL.Path) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Try to serve a static file first
		path := filepath.Join(frontendDir, filepath.Clean(req.URL.Path))
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			fs.ServeHTTP(w, req)
			return
		}

		// Fallback to index.html
		indexPath := filepath.Join(frontendDir, "index.html")
		http.ServeFile(w, req, indexPath)
	}))
}

// isLegacyAPIPath checks if the path is a legacy API path that should NOT be served by the SPA.
// These paths are API endpoints that existed before the /api prefix was standardized.
// If API_PREFIX is empty, these paths are used for API routes.
func isLegacyAPIPath(path string) bool {
	// List of legacy API path prefixes that should not be served by SPA
	legacyPrefixes := []string{
		"/info",
		"/health",
		"/healthapi",
		"/v3/",
		"/v4/",
		"/current/",
		"/swagger",
		"/mcp",
		"/scan",
		"/paircode",
		"/command",
		"/message",
		"/send",
		"/receive",
		"/download",
		"/webhook",
		"/picinfo",
		"/picdata",
		"/contacts",
		"/groups",
		"/invite",
		"/account",
		"/rabbitmq",
		"/read",
		"/edit",
		"/isonwhatsapp",
		"/useridentifier",
		"/getphone",
		"/environment",
		"/login",
	}

	for _, prefix := range legacyPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

// NewReverseProxy creates a reverse proxy to a target (http://host:port)
// basePath is the base path to prepend to requests (e.g., "/assets/frontend")
func NewReverseProxy(target string, basePath string) *httputil.ReverseProxy {
	url, err := url.Parse(target)
	if err != nil {
		// fallback to a proxy to localhost:5173
		url, _ = url.Parse("http://127.0.0.1:5173")
	}
	proxy := httputil.NewSingleHostReverseProxy(url)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		// Remove the extra log here since we log in NotFound handler
		// Rewrite path to include base path if provided
		if basePath != "" {
			originalPath := req.URL.Path
			if originalPath == "/" {
				req.URL.Path = basePath + "/"
			} else if !strings.HasPrefix(originalPath, basePath) {
				req.URL.Path = basePath + originalPath
			}
		}
		// allow websocket upgrades to pass through
		req.Header.Set("X-Forwarded-Host", req.Host)
		req.Header.Set("X-Origin-Host", url.Host)
	}
	return proxy
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
