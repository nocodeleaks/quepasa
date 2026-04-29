package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestRegisterAPIV5ControllersMountsCanonicalAliases(t *testing.T) {
	router := chi.NewRouter()
	RegisterAPIV5Controllers(router)

	for _, path := range []string{"/system/version", "/v5/system/version"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		res := httptest.NewRecorder()

		router.ServeHTTP(res, req)

		if res.Code != http.StatusOK {
			t.Fatalf("expected status 200 for %s, got %d", path, res.Code)
		}
	}
}

func TestRegisterAPIControllersKeepsLegacyAliasesMounted(t *testing.T) {
	router := chi.NewRouter()
	RegisterAPIControllers(router)

	for _, path := range []string{"/healthapi", "/current/health", "/v4/health"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		res := httptest.NewRecorder()

		router.ServeHTTP(res, req)

		if res.Code == http.StatusNotFound {
			t.Fatalf("expected legacy route %s to remain mounted", path)
		}
	}
}

func TestRegisterAPIV3ControllersKeepsVersionedAliasMounted(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	CreateTestUser(t, "owner@example.com", "Password123!")
	CreateTestServer(t, "test-token", "owner@example.com")

	router := chi.NewRouter()
	RegisterAPIV3Controllers(router)

	req := httptest.NewRequest(http.MethodGet, "/v3/bot/test-token", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code == http.StatusNotFound {
		t.Fatal("expected v3 bot route to remain mounted")
	}
}

func TestConfigureNoLongerMountsSPAAlias(t *testing.T) {
	router := chi.NewRouter()
	Configure(router)

	req := httptest.NewRequest(http.MethodGet, "/spa/login/config", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected /spa/login/config to be retired, got %d", res.Code)
	}
}
