package webserver

import (
	"os"
	"path/filepath"
	"testing"
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
	originalDevFrontend := os.Getenv("QUEPASA_DEV_FRONTEND")
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
		_ = os.Setenv("QUEPASA_DEV_FRONTEND", originalDevFrontend)
	})

	mustMkdirAll(t, filepath.Join(tempDir, "apps", "vuejs", "client"))
	mustWriteFile(t, filepath.Join(tempDir, "apps", "vuejs", "package.json"), "{}")
	mustWriteFile(t, filepath.Join(tempDir, "apps", "vuejs", "client", "index.html"), "<html>source</html>")
	if err := os.Setenv("QUEPASA_DEV_FRONTEND", "1"); err != nil {
		t.Fatalf("setenv: %v", err)
	}

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
