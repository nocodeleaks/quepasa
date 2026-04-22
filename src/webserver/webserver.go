package webserver

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	environment "github.com/nocodeleaks/quepasa/environment"
	log "github.com/sirupsen/logrus"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// RouterConfigurator lets feature packages attach routes to the main router
// without coupling the webserver package to specific modules.
type RouterConfigurator func(r chi.Router)

// configurators stores all route registration hooks collected during package init.
var configurators []RouterConfigurator

// RegisterRouterConfigurator appends a route configuration hook to the main router.
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

	// Recoverer should wrap all previous middlewares.
	// chi applies middlewares in reverse when building the final handler chain,
	// so adding Recoverer last makes it the outermost wrapper.
	r.Use(middleware.Recoverer)

	// Install SPA fallback behavior before feature routes. This only affects
	// unresolved paths because chi executes NotFound after normal route matching.
	ServeSPA(r)

	// Execute feature-level route registration hooks (API, Form, Swagger, etc.).
	for _, configurator := range configurators {
		configurator(r)
	}

	return r
}

// ServeSPA configures SPA fallback behavior without interfering with existing matched routes.
// In development it can proxy to a local Vite server; in production it serves assets/frontend
// when a built SPA is present on disk.
func ServeSPA(r chi.Router) {
	workDir, _ := os.Getwd()

	// Mount the imported PR frontend as an explicit alternate app so it does not
	// interfere with the classic UI or the primary SPA fallback.
	ServeBuiltSPAAtPrefix(r, "/spa-app", filepath.Join(workDir, "assets", "frontend-alt"))

	if useFrontendDevProxy() {
		proxy := NewReverseProxy(frontendDevTarget())
		r.NotFound(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// Keep explicit API-like namespaces out of the SPA proxy so transport
			// failures still surface as ordinary 404 responses.
			if !shouldServeSPAPath(req.URL.Path) {
				http.NotFound(w, req)
				return
			}
			proxy.ServeHTTP(w, req)
		}))
		return
	}

	frontendDir := filepath.Join(workDir, "assets", "frontend")
	indexPath := filepath.Join(frontendDir, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		return
	}

	fs := http.FileServer(http.Dir(frontendDir))
	r.NotFound(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet || !shouldServeSPAPath(req.URL.Path) {
			http.NotFound(w, req)
			return
		}

		// Serve concrete built files directly before falling back to index.html.
		relativePath := strings.TrimPrefix(path.Clean(req.URL.Path), "/")
		staticPath := filepath.Join(frontendDir, filepath.FromSlash(relativePath))
		if info, err := os.Stat(staticPath); err == nil && !info.IsDir() {
			fs.ServeHTTP(w, req)
			return
		}

		http.ServeFile(w, req, indexPath)
	}))
}

// ServeBuiltSPAAtPrefix mounts a prebuilt SPA bundle under an explicit URL
// prefix and falls back to its index.html for client-side navigation.
func ServeBuiltSPAAtPrefix(r chi.Router, prefix string, spaDir string) {
	indexPath := filepath.Join(spaDir, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		return
	}

	serve := func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.NotFound(w, req)
			return
		}

		requestPath := path.Clean(req.URL.Path)
		if requestPath == "." {
			requestPath = "/"
		}

		if requestPath == prefix || requestPath == prefix+"/" {
			http.ServeFile(w, req, indexPath)
			return
		}

		relativePath := strings.TrimPrefix(requestPath, prefix+"/")
		staticPath := filepath.Join(spaDir, filepath.FromSlash(relativePath))
		if info, err := os.Stat(staticPath); err == nil && !info.IsDir() {
			http.ServeFile(w, req, staticPath)
			return
		}

		http.ServeFile(w, req, indexPath)
	}

	r.Get(prefix, serve)
	r.Get(prefix+"/", serve)
	r.Get(prefix+"/*", serve)
}

// useFrontendDevProxy enables the Vite reverse proxy explicitly for local SPA work.
func useFrontendDevProxy() bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv("QUEPASA_DEV_FRONTEND")))
	return value == "1" || value == "true"
}

// frontendDevTarget resolves the Vite dev-server origin used by the reverse proxy.
func frontendDevTarget() string {
	host := strings.TrimSpace(os.Getenv("QUEPASA_FRONTEND_HOST"))
	if host == "" {
		host = "http://127.0.0.1"
	}

	port := strings.TrimSpace(os.Getenv("QUEPASA_FRONTEND_DEV_PORT"))
	if port == "" {
		port = "5173"
	}

	return host + ":" + port
}

// shouldServeSPAPath filters requests that may be safely handled by the SPA
// fallback without shadowing API, static, form, or other transport routes.
func shouldServeSPAPath(requestPath string) bool {
	if requestPath == "" {
		return true
	}

	if requestPath == "/" {
		return true
	}

	if strings.HasPrefix(requestPath, "/api") ||
		strings.HasPrefix(requestPath, "/cable") ||
		strings.HasPrefix(requestPath, "/spa") ||
		strings.HasPrefix(requestPath, "/assets") ||
		strings.HasPrefix(requestPath, "/form") ||
		strings.HasPrefix(requestPath, "/mcp") ||
		strings.HasPrefix(requestPath, "/swagger") ||
		isLegacyAPIPath(requestPath) {
		return false
	}

	return true
}

// isLegacyAPIPath marks historical root-level API prefixes that must never be
// interpreted as SPA navigation routes.
func isLegacyAPIPath(requestPath string) bool {
	legacyPrefixes := []string{
		"/health",
		"/healthapi",
		"/current/",
		"/v3/",
		"/v4/",
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
		"/contact",
		"/contacts",
		"/invite",
		"/isonwhatsapp",
		"/useridentifier",
		"/getphone",
		"/userinfo",
		"/spam",
	}

	for _, prefix := range legacyPrefixes {
		if strings.HasPrefix(requestPath, prefix) {
			return true
		}
	}

	return false
}

// NewReverseProxy creates a reverse proxy to a target (http://host:port).
// Forwarded host headers are preserved so the dev frontend can inspect origin data.
func NewReverseProxy(target string) *httputil.ReverseProxy {
	targetURL, err := url.Parse(target)
	if err != nil {
		targetURL, _ = url.Parse("http://127.0.0.1:5173")
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Header.Set("X-Forwarded-Host", req.Host)
		req.Header.Set("X-Origin-Host", targetURL.Host)
	}

	return proxy
}

func ServeStaticContent(r chi.Router) {
	r.Group(func(r chi.Router) {
		// Serve shared static assets from the project assets directory.
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
