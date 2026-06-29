package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/jwtauth"
	library "github.com/nocodeleaks/quepasa/library"
	models "github.com/nocodeleaks/quepasa/models"
)

// userKeyChain mirrors the production middleware order (JWT verifier, then the
// authenticated handler) so the personal API key path is exercised exactly as
// wired in RegisterAuthenticatedControllers.
func userKeyChain(next http.Handler) http.Handler {
	return jwtauth.Verifier(GetAuthenticatedTokenAuth())(AuthenticatedAPIHandler(next))
}

// TestPersonalAPIKeyAuthenticatesUser verifies the full path: a stored key hash
// authenticates the owning user via X-QUEPASA-USERKEY, a wrong key is rejected,
// and a revoked key stops working.
func TestPersonalAPIKeyAuthenticatesUser(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	CreateTestUser(t, "owner@example.com", "Password123!")

	plain, hash, err := models.GenerateAPIKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	if err := models.WhatsappService.DB.Users.SetAPIKey("owner@example.com", hash); err != nil {
		t.Fatalf("set key: %v", err)
	}

	// Inner handler records the user the middleware resolved.
	var resolved string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if u, uErr := GetAuthenticatedUser(r); uErr == nil && u != nil {
			resolved = u.Username
		}
		w.WriteHeader(http.StatusOK)
	})

	t.Run("valid key authenticates owner", func(t *testing.T) {
		resolved = ""
		req := httptest.NewRequest(http.MethodGet, "/api/account", nil)
		req.Header.Set(library.HeaderUserKey, plain)
		rec := httptest.NewRecorder()
		userKeyChain(inner).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("valid key got status %d", rec.Code)
		}
		if resolved != "owner@example.com" {
			t.Fatalf("middleware resolved %q, want owner@example.com", resolved)
		}
	})

	t.Run("wrong key is rejected", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/account", nil)
		req.Header.Set(library.HeaderUserKey, "qp_not_a_real_key")
		rec := httptest.NewRecorder()
		userKeyChain(inner).ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("wrong key got status %d, want 401", rec.Code)
		}
	})

	t.Run("revoked key stops working", func(t *testing.T) {
		if err := models.WhatsappService.DB.Users.ClearAPIKey("owner@example.com"); err != nil {
			t.Fatalf("clear key: %v", err)
		}
		req := httptest.NewRequest(http.MethodGet, "/api/account", nil)
		req.Header.Set(library.HeaderUserKey, plain)
		rec := httptest.NewRecorder()
		userKeyChain(inner).ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("revoked key got status %d, want 401", rec.Code)
		}
	})
}
