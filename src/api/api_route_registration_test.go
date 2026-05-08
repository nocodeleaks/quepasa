package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	environment "github.com/nocodeleaks/quepasa/environment"
)

func TestRegisterAPIV5ControllersMountsCanonicalAliases(t *testing.T) {
	router := chi.NewRouter()
	RegisterAPIV5Controllers(router, true)

	for _, path := range []string{"/system/version", "/v5/system/version"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		res := httptest.NewRecorder()

		router.ServeHTTP(res, req)

		if res.Code != http.StatusOK {
			t.Fatalf("expected status 200 for %s, got %d", path, res.Code)
		}
	}
}

func TestRegisterAPIV5ControllersCanSkipUnversionedAlias(t *testing.T) {
	router := chi.NewRouter()
	RegisterAPIV5Controllers(router, false)

	versionedReq := httptest.NewRequest(http.MethodGet, "/v5/system/version", nil)
	versionedRes := httptest.NewRecorder()
	router.ServeHTTP(versionedRes, versionedReq)
	if versionedRes.Code != http.StatusOK {
		t.Fatalf("expected /v5/system/version to remain mounted, got %d", versionedRes.Code)
	}

	unversionedReq := httptest.NewRequest(http.MethodGet, "/system/version", nil)
	unversionedRes := httptest.NewRecorder()
	router.ServeHTTP(unversionedRes, unversionedReq)
	if unversionedRes.Code != http.StatusNotFound {
		t.Fatalf("expected /system/version to be skipped when unversioned alias is disabled, got %d", unversionedRes.Code)
	}
}

func TestRegisterAPIControllersKeepsLegacyAliasesMounted(t *testing.T) {
	router := chi.NewRouter()
	RegisterAPIControllers(router, true)

	for _, path := range []string{"/healthapi", "/current/health", "/v4/health", "/health"} {
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

func TestConfigureUsesDefaultVersionForUnversionedAlias(t *testing.T) {
	previous := environment.Settings.API.DefaultVersion
	defer func() { environment.Settings.API.DefaultVersion = previous }()

	environment.Settings.API.DefaultVersion = CurrentAPIVersion

	router := chi.NewRouter()
	Configure(router)

	legacyReq := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	legacyRes := httptest.NewRecorder()
	router.ServeHTTP(legacyRes, legacyReq)
	if legacyRes.Code == http.StatusNotFound {
		t.Fatalf("expected /api/health to be mounted when default version is %s", CurrentAPIVersion)
	}

	canonicalReq := httptest.NewRequest(http.MethodGet, "/api/system/version", nil)
	canonicalRes := httptest.NewRecorder()
	router.ServeHTTP(canonicalRes, canonicalReq)
	if canonicalRes.Code != http.StatusNotFound {
		t.Fatalf("expected /api/system/version to be absent when default version is %s, got %d", CurrentAPIVersion, canonicalRes.Code)
	}
}
