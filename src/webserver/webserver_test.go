package webserver

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-chi/chi/v5"
	environment "github.com/nocodeleaks/quepasa/environment"
)

func TestResolveFrontendAppRequestKeepsConsoleWhenVueJSAlsoExists(t *testing.T) {
	tempDir := t.TempDir()
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})

	mustMkdirAll(t, filepath.Join(tempDir, "apps", "vuejs", "dist"))
	mustMkdirAll(t, filepath.Join(tempDir, "apps", "console"))
	mustWriteFile(t, filepath.Join(tempDir, "apps", "vuejs", "package.json"), "{}")
	mustWriteFile(t, filepath.Join(tempDir, "apps", "vuejs", "dist", "index.html"), "<html>vuejs</html>")
	mustWriteFile(t, filepath.Join(tempDir, "apps", "console", "index.html"), "<html>console</html>")

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	app, relativePath, ok := resolveFrontendAppRequest("/apps/console/server/abc")
	if !ok {
		t.Fatal("expected console app to resolve")
	}

	if app.Slug != "console" {
		t.Fatalf("expected %q slug, got %q", "console", app.Slug)
	}

	if relativePath != "/server/abc" {
		t.Fatalf("expected relative path /server/abc, got %q", relativePath)
	}
}

func TestDiscoverFrontendAppsKeepsConsoleWhenVueJSExists(t *testing.T) {
	tempDir := t.TempDir()
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})

	mustMkdirAll(t, filepath.Join(tempDir, "apps", "vuejs", "dist"))
	mustMkdirAll(t, filepath.Join(tempDir, "apps", "console"))
	mustWriteFile(t, filepath.Join(tempDir, "apps", "vuejs", "package.json"), "{}")
	mustWriteFile(t, filepath.Join(tempDir, "apps", "vuejs", "dist", "index.html"), "<html>vuejs</html>")
	mustWriteFile(t, filepath.Join(tempDir, "apps", "console", "index.html"), "<html>console</html>")

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	apps := DiscoverFrontendApps()
	if len(apps) != 2 {
		t.Fatalf("expected 2 apps, got %d", len(apps))
	}

	if apps[0].Slug != "console" {
		t.Fatalf("expected first slug %q, got %q", "console", apps[0].Slug)
	}

	if apps[1].Slug != "vuejs" {
		t.Fatalf("expected second slug %q, got %q", "vuejs", apps[1].Slug)
	}
}

func TestResolveFrontendAppRequestKeepsConsoleWhenVueJSIsMissing(t *testing.T) {
	tempDir := t.TempDir()
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})

	mustMkdirAll(t, filepath.Join(tempDir, "apps", "console"))
	mustWriteFile(t, filepath.Join(tempDir, "apps", "console", "index.html"), "<html>console</html>")

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	app, relativePath, ok := resolveFrontendAppRequest("/apps/console/login")
	if !ok {
		t.Fatal("expected console app to resolve")
	}

	if app.Slug != "console" {
		t.Fatalf("expected %q slug, got %q", "console", app.Slug)
	}

	if relativePath != "/login" {
		t.Fatalf("expected relative path /login, got %q", relativePath)
	}
}

func TestDiscoverFrontendAppsUsesClientIndexDuringDevProxy(t *testing.T) {
	tempDir := t.TempDir()
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})

	mustMkdirAll(t, filepath.Join(tempDir, "apps", "vuejs", "client"))
	mustWriteFile(t, filepath.Join(tempDir, "apps", "vuejs", "package.json"), "{}")
	mustWriteFile(t, filepath.Join(tempDir, "apps", "vuejs", "client", "index.html"), "<html>source</html>")
	t.Setenv("QUEPASA_DEV_FRONTEND", "1")
	previousDevFrontend := environment.Settings.WebServer.DevFrontend
	environment.Settings.WebServer.DevFrontend = true
	t.Cleanup(func() { environment.Settings.WebServer.DevFrontend = previousDevFrontend })

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	apps := DiscoverFrontendApps()
	if len(apps) != 1 {
		t.Fatalf("expected 1 app, got %d", len(apps))
	}

	wantIndex := filepath.Join(tempDir, "apps", "vuejs", "client", "index.html")
	if apps[0].IndexFile != wantIndex {
		t.Fatalf("expected client index %q, got %q", wantIndex, apps[0].IndexFile)
	}
}

func TestNormalizeDefaultFrontendAppPath(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "empty", value: "", want: ""},
		{name: "slug", value: "vuejs", want: "/apps/vuejs/"},
		{name: "hidden slug", value: ".custom", want: "/apps/.custom/"},
		{name: "apps path", value: "/apps/.custom/", want: "/apps/.custom/"},
		{name: "apps child path", value: "apps/form/account", want: "/apps/form/account"},
		{name: "external url rejected", value: "https://example.com/apps/vuejs", want: ""},
		{name: "protocol relative rejected", value: "//example.com/apps/vuejs", want: ""},
		{name: "path traversal rejected", value: "../api", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeDefaultFrontendAppPath(tt.value); got != tt.want {
				t.Fatalf("normalizeDefaultFrontendAppPath(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestServeAppsRedirectsRootWhenDefaultAppConfigured(t *testing.T) {
	previous := environment.Settings.WebServer.DefaultApp
	environment.Settings.WebServer.DefaultApp = "console"
	t.Cleanup(func() { environment.Settings.WebServer.DefaultApp = previous })

	r := chi.NewRouter()
	ServeApps(r)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("expected status %d, got %d", http.StatusFound, rec.Code)
	}
	if got := rec.Header().Get("Location"); got != "/apps/console/" {
		t.Fatalf("expected redirect to /apps/console/, got %q", got)
	}
}

func TestDiscoverFrontendAppsFindsSrcAppsFromRepositoryRoot(t *testing.T) {
	tempDir := t.TempDir()
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})

	mustMkdirAll(t, filepath.Join(tempDir, "src", "apps", "vuejs", "dist"))
	mustMkdirAll(t, filepath.Join(tempDir, "src", "apps", "console"))
	mustMkdirAll(t, filepath.Join(tempDir, "src", "assets"))
	mustWriteFile(t, filepath.Join(tempDir, "src", "apps", "vuejs", "package.json"), "{}")
	mustWriteFile(t, filepath.Join(tempDir, "src", "apps", "vuejs", "dist", "index.html"), "<html>vuejs</html>")
	mustWriteFile(t, filepath.Join(tempDir, "src", "apps", "console", "index.html"), "<html>console</html>")

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	app, relativePath, ok := resolveFrontendAppRequest("/apps/console/server/abc")
	if !ok {
		t.Fatal("expected console app to resolve from repository root")
	}

	if app.Slug != "console" {
		t.Fatalf("expected %q slug, got %q", "console", app.Slug)
	}

	if relativePath != "/server/abc" {
		t.Fatalf("expected relative path /server/abc, got %q", relativePath)
	}

	apps := DiscoverFrontendApps()
	if len(apps) != 2 {
		t.Fatalf("expected 2 apps, got %d", len(apps))
	}

	if apps[0].Slug != "console" {
		t.Fatalf("expected first slug %q, got %q", "console", apps[0].Slug)
	}

	if apps[1].Slug != "vuejs" {
		t.Fatalf("expected second slug %q, got %q", "vuejs", apps[1].Slug)
	}
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func mustWriteFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
