package webserver

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
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

// FrontendApp describes one auto-discovered frontend mounted under /apps/<slug>.
type FrontendApp struct {
	Slug      string
	Path      string
	URLPath   string
	IndexFile string
	// BackendManaged marks apps whose routes are handled by Go code.
	// They appear in the apps listing but are excluded from static/SPA file serving.
	BackendManaged bool
}

// configurators stores all route registration hooks collected during package init.
var configurators []RouterConfigurator

// RegisterRouterConfigurator appends a route configuration hook to the main router.
func RegisterRouterConfigurator(configurator RouterConfigurator) {
	configurators = append(configurators, configurator)
}

// backendApps holds apps registered programmatically as backend-managed
// (i.e., routes handled by Go, not by static file serving).
var backendApps []FrontendApp

// RegisterBackendApp registers an app that should appear in the /apps/ listing
// but whose paths are fully handled by Go route handlers.
// No static file serving or SPA fallback will be applied to this app.
func RegisterBackendApp(slug, urlPath string) {
	backendApps = append(backendApps, FrontendApp{
		Slug:           slug,
		URLPath:        urlPath,
		BackendManaged: true,
	})
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

	// Execute feature-level route registration hooks (API, Form, Swagger, etc.).
	for _, configurator := range configurators {
		configurator(r)
	}

	// Install apps fallback behavior after feature routes so explicit handlers
	// like /apps/form/login are matched first and the wildcard only handles
	// unresolved app bundle paths.
	ServeApps(r)

	return r
}

// ServeApps configures apps fallback behavior without interfering with existing matched routes.
// App bundles are served from ./apps/<slug>.
func ServeApps(r chi.Router) {
	ServeDiscoveredApps(r)

	if useFrontendDevProxy() {
		proxy := NewReverseProxy(frontendDevTarget())
		r.NotFound(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// Keep explicit API-like namespaces out of the SPA proxy so transport
			// failures still surface as ordinary 404 responses.
			if !shouldServeAppsPath(req.URL.Path) {
				http.NotFound(w, req)
				return
			}
			proxy.ServeHTTP(w, req)
		}))
		return
	}
}

// DiscoverFrontendApps scans ./apps and returns every subdirectory that exposes a
// browser-loadable index file directly at its root (index.html, index.htm, etc.).
func DiscoverFrontendApps() []FrontendApp {
	appsDir := filepath.Join(resolveFrontendContentDir(), "apps")

	entries, err := os.ReadDir(appsDir)
	if err != nil {
		return nil
	}

	apps := make([]FrontendApp, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		slug := strings.TrimSpace(entry.Name())
		if slug == "" {
			continue
		}

		appDir := filepath.Join(appsDir, slug)
		publicDir, indexPath, ok := findFrontendAppPublicDir(appDir)
		if !ok {
			continue
		}

		urlPath := "/apps/" + slug
		apps = append(apps, FrontendApp{
			Slug:      slug,
			Path:      publicDir,
			URLPath:   urlPath,
			IndexFile: indexPath,
		})
	}

	sort.Slice(apps, func(i, j int) bool {
		return strings.ToLower(apps[i].Slug) < strings.ToLower(apps[j].Slug)
	})

	// Merge backend-managed apps, keeping sort order.
	for _, ba := range backendApps {
		if _, found := findExactFrontendAppBySlug(apps, ba.Slug); !found {
			apps = append(apps, ba)
		}
	}
	sort.Slice(apps, func(i, j int) bool {
		return strings.ToLower(apps[i].Slug) < strings.ToLower(apps[j].Slug)
	})

	return apps
}

// ServeDiscoveredApps mounts /apps/<slug> for every auto-discovered app folder.
func ServeDiscoveredApps(r chi.Router) {
	r.Get("/apps", ServeFrontendAppsIndex)
	r.Get("/apps/", ServeFrontendAppsIndex)
	r.Get("/apps/*", ServeDiscoveredAppRequest)
}

// ServeFrontendAppsIndex lists discovered frontend apps in a minimal HTML page.
func ServeFrontendAppsIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	apps := DiscoverFrontendApps()
	if len(apps) == 0 {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte("<!doctype html><html><head><meta charset=\"utf-8\"><title>QuePasa Apps</title></head><body><h1>QuePasa Apps</h1><ul>"))
	for _, app := range apps {
		_, _ = w.Write([]byte(fmt.Sprintf("<li><a href=\"%s/\">%s</a></li>", app.URLPath, app.Slug)))
	}
	_, _ = w.Write([]byte("</ul></body></html>"))
}

// ServeDiscoveredAppRequest serves one discovered app under /apps/<slug>.
func ServeDiscoveredAppRequest(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.NotFound(w, req)
		return
	}

	app, relativePath, ok := resolveFrontendAppRequest(req.URL.Path)
	if !ok {
		http.NotFound(w, req)
		return
	}

	// Backend-managed apps have their routes handled entirely by Go handlers.
	// Passing the request through here would bypass them, so return 404 and let
	// chi continue to the explicit routes registered by those packages.
	if app.BackendManaged {
		http.NotFound(w, req)
		return
	}

	if useFrontendDevProxy() {
		proxyFrontendAppRequest(w, req, app, relativePath)
		return
	}

	if relativePath == "" || relativePath == "/" {
		serveIndexWithConfig(w, req, app.IndexFile)
		return
	}

	staticPath := filepath.Join(app.Path, filepath.FromSlash(strings.TrimPrefix(relativePath, "/")))
	if info, err := os.Stat(staticPath); err == nil && !info.IsDir() {
		http.ServeFile(w, req, staticPath)
		return
	}
	if info, err := os.Stat(staticPath); err == nil && info.IsDir() {
		if indexPath, ok := findIndexFileInDir(staticPath); ok {
			http.ServeFile(w, req, indexPath)
			return
		}
	}

	serveIndexWithConfig(w, req, app.IndexFile)
}

// serveIndexWithConfig reads an SPA index.html and injects a runtime config
// block so browser clients can discover the server's API prefix without a
// separate HTTP round-trip.
func serveIndexWithConfig(w http.ResponseWriter, req *http.Request, indexFile string) {
	content, err := os.ReadFile(indexFile)
	if err != nil {
		http.ServeFile(w, req, indexFile)
		return
	}

	// Resolve the effective API base path from the environment setting.
	// This is injected into the SPA so it can call the correct routes without
	// any hardcoded prefix. The default is "api" (see environment/api_settings.go).
	prefix := strings.Trim(environment.Settings.API.Prefix, "/")
	var apiBase string
	if prefix != "" {
		apiBase = "/" + prefix
	} else {
		apiBase = ""
	}

	script := fmt.Sprintf(
		`<script>window.quepasa={"apiBase":%q};</script>`,
		apiBase,
	)
	html := strings.Replace(string(content), "</head>", script+"</head>", 1)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write([]byte(html))
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

// shouldServeAppsPath filters requests that may be safely handled by the apps
// fallback without shadowing API, static, form, or other transport routes.
func shouldServeAppsPath(requestPath string) bool {
	if requestPath == "" {
		return true
	}

	if requestPath == "/" {
		return true
	}

	if strings.HasPrefix(requestPath, "/api") ||
		strings.HasPrefix(requestPath, "/cable") ||
		strings.HasPrefix(requestPath, "/apps") ||
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

func resolveFrontendAppRequest(requestPath string) (FrontendApp, string, bool) {
	cleaned := path.Clean(requestPath)
	if cleaned == "." {
		cleaned = "/"
	}

	if !strings.HasPrefix(cleaned, "/apps/") {
		return FrontendApp{}, "", false
	}

	trimmed := strings.TrimPrefix(cleaned, "/apps/")
	parts := strings.SplitN(trimmed, "/", 2)
	slug := strings.TrimSpace(parts[0])
	if slug == "" {
		return FrontendApp{}, "", false
	}

	apps := DiscoverFrontendApps()
	app, found := findExactFrontendAppBySlug(apps, slug)
	if found {
		relativePath := "/"
		if len(parts) == 2 {
			relativePath = "/" + parts[1]
		}
		return app, relativePath, true
	}

	return FrontendApp{}, "", false
}

func findExactFrontendAppBySlug(apps []FrontendApp, slug string) (FrontendApp, bool) {
	for _, app := range apps {
		if app.Slug == slug {
			return app, true
		}
	}

	return FrontendApp{}, false
}

func resolveFrontendContentDir() string {
	candidates := make([]string, 0, 5)
	if workDir, err := os.Getwd(); err == nil {
		candidates = append(candidates, workDir, filepath.Join(workDir, "src"))
	}

	if executablePath, err := os.Executable(); err == nil {
		executableDir := filepath.Dir(executablePath)
		candidates = append(candidates,
			executableDir,
			filepath.Join(executableDir, "src"),
			filepath.Join(executableDir, "..", "src"),
		)
	}

	for _, candidate := range candidates {
		if isFrontendContentDir(candidate) {
			return filepath.Clean(candidate)
		}
	}

	if len(candidates) > 0 {
		return filepath.Clean(candidates[0])
	}

	return "."
}

func isFrontendContentDir(dir string) bool {
	if dir == "" {
		return false
	}

	if info, err := os.Stat(filepath.Join(dir, "apps")); err != nil || !info.IsDir() {
		return false
	}

	if info, err := os.Stat(filepath.Join(dir, "assets")); err != nil || !info.IsDir() {
		return false
	}

	return true
}

func findFrontendAppPublicDir(appDir string) (string, string, bool) {
	distDir := filepath.Join(appDir, "dist")
	if indexPath, ok := findIndexFileInDir(distDir); ok {
		return distDir, indexPath, true
	}

	if isLegacyFrontendBundleDir(appDir) {
		if indexPath, ok := findIndexFileInDir(appDir); ok {
			return appDir, indexPath, true
		}
	}

	if useFrontendDevProxy() {
		clientDir := filepath.Join(appDir, "client")
		if indexPath, ok := findIndexFileInDir(clientDir); ok {
			return appDir, indexPath, true
		}
	}

	return "", "", false
}

func findIndexFileInDir(dir string) (string, bool) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", false
	}

	indexCandidates := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := strings.ToLower(strings.TrimSpace(entry.Name()))
		if !strings.HasPrefix(name, "index.") {
			continue
		}

		indexCandidates = append(indexCandidates, filepath.Join(dir, entry.Name()))
	}

	if len(indexCandidates) == 0 {
		return "", false
	}

	sort.Strings(indexCandidates)
	return indexCandidates[0], true
}

func isLegacyFrontendBundleDir(appDir string) bool {
	if _, err := os.Stat(filepath.Join(appDir, "package.json")); err == nil {
		return false
	}

	return true
}

func proxyFrontendAppRequest(w http.ResponseWriter, req *http.Request, app FrontendApp, relativePath string) {
	proxy := NewReverseProxy(frontendDevTarget())
	proxyReq := req.Clone(req.Context())
	proxyReq.URL.Path = app.URLPath
	if relativePath != "" && relativePath != "/" {
		proxyReq.URL.Path += relativePath
	} else {
		proxyReq.URL.Path += "/"
	}
	proxyReq.URL.RawPath = proxyReq.URL.Path
	proxy.ServeHTTP(w, proxyReq)
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
		contentDir := resolveFrontendContentDir()
		faviconPath := filepath.Join(contentDir, "assets", "favicon-32x32.png")
		r.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, faviconPath)
		})

		// Serve shared static assets from the project assets directory.
		assetsDir := filepath.Join(contentDir, "assets")
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
