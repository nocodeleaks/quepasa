package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	environment "github.com/nocodeleaks/quepasa/environment"
)

func withAllowedOrigins(t *testing.T, origins []string) func() {
	t.Helper()
	old := environment.Settings.API.AllowedOrigins
	environment.Settings.API.AllowedOrigins = origins
	return func() { environment.Settings.API.AllowedOrigins = old }
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
}

// TestCORSDisabledByDefault: with no allow-list, the middleware must be a no-op —
// no CORS headers and OPTIONS is NOT short-circuited (left to the router).
func TestCORSDisabledByDefault(t *testing.T) {
	defer withAllowedOrigins(t, nil)()

	req := httptest.NewRequest(http.MethodOptions, "/api/anything", nil)
	req.Header.Set("Origin", "https://evil.example")
	rec := httptest.NewRecorder()

	reached := false
	APICORSMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)

	if !reached {
		t.Fatal("middleware short-circuited a request while CORS was disabled")
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected no CORS header when disabled, got %q", got)
	}
}

// TestCORSReflectsAllowedOrigin: an exact-match origin is reflected with
// credentials and Vary; a non-matching origin gets nothing.
func TestCORSReflectsAllowedOrigin(t *testing.T) {
	defer withAllowedOrigins(t, []string{"https://app.example", "https://admin.example"})()

	t.Run("match", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/x", nil)
		req.Header.Set("Origin", "https://admin.example")
		rec := httptest.NewRecorder()
		APICORSMiddleware(okHandler()).ServeHTTP(rec, req)

		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://admin.example" {
			t.Fatalf("allow-origin: got %q, want https://admin.example", got)
		}
		if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
			t.Fatalf("allow-credentials: got %q, want true", got)
		}
		if got := rec.Header().Get("Vary"); got != "Origin" {
			t.Fatalf("vary: got %q, want Origin", got)
		}
	})

	t.Run("non-match", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/x", nil)
		req.Header.Set("Origin", "https://evil.example")
		rec := httptest.NewRecorder()
		APICORSMiddleware(okHandler()).ServeHTTP(rec, req)

		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Fatalf("non-allowed origin must not be reflected, got %q", got)
		}
	})
}

// TestCORSWildcardWithoutCredentials: "*" allows any origin but never with
// credentials (browsers reject credentialed wildcard responses).
func TestCORSWildcardWithoutCredentials(t *testing.T) {
	defer withAllowedOrigins(t, []string{"*"})()

	req := httptest.NewRequest(http.MethodGet, "/api/x", nil)
	req.Header.Set("Origin", "https://whatever.example")
	rec := httptest.NewRecorder()
	APICORSMiddleware(okHandler()).ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("allow-origin: got %q, want *", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Fatalf("wildcard must not set allow-credentials, got %q", got)
	}
}

// TestCORSPreflightShortCircuits: a configured allow-list answers OPTIONS with
// 204 and does not reach the handler.
func TestCORSPreflightShortCircuits(t *testing.T) {
	defer withAllowedOrigins(t, []string{"https://app.example"})()

	req := httptest.NewRequest(http.MethodOptions, "/api/x", nil)
	req.Header.Set("Origin", "https://app.example")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rec := httptest.NewRecorder()

	reached := false
	APICORSMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
	})).ServeHTTP(rec, req)

	if reached {
		t.Fatal("preflight reached the underlying handler")
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("preflight status: got %d, want 204", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://app.example" {
		t.Fatalf("preflight allow-origin: got %q", got)
	}
}
