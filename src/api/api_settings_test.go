package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	models "github.com/nocodeleaks/quepasa/models"
)

// TestCanonicalSettingsRequiresMasterKey: missing/invalid master key -> 401 (unlike env, /settings rejects).
func TestCanonicalSettingsRequiresMasterKey(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	restore := SetupTestMasterKey(t, "settings-master-key")
	defer restore()

	router := newCanonicalTestRouter()

	// No master key
	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without master key, got %d", rec.Code)
	}

	// Wrong master key
	req = httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	req.Header.Set("X-QUEPASA-MASTERKEY", "wrong-key")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 with wrong master key, got %d", rec.Code)
	}
}

// TestCanonicalSettingsGetReturnsEnvAndGlobal: valid master key GET returns env + global tiers,
// and the global override seeded in the cache surfaces (mirrors the PUT-then-GET intent without DB).
func TestCanonicalSettingsGetReturnsEnvAndGlobal(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	restore := SetupTestMasterKey(t, "settings-master-key")
	defer restore()

	// Seed a global override (runtime cache) the way SetGlobalMessageConfig would leave it.
	five := 5
	models.SetGlobalMessageConfigForTest(models.GlobalMessageConfig{StoreRetentionDays: &five})
	defer models.SetGlobalMessageConfigForTest(models.GlobalMessageConfig{})

	router := newCanonicalTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	req.Header.Set("X-QUEPASA-MASTERKEY", "settings-master-key")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 with master key, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	var resp struct {
		Env    map[string]interface{}     `json:"env"`
		Global models.GlobalMessageConfig `json:"global"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.Env == nil {
		t.Fatalf("expected env tier in response, got: %s", rec.Body.String())
	}
	if _, ok := resp.Env["store_retention_days"]; !ok {
		t.Fatalf("expected env.store_retention_days, got: %+v", resp.Env)
	}
	if _, ok := resp.Env["dispatch_types"]; !ok {
		t.Fatalf("expected env.dispatch_types, got: %+v", resp.Env)
	}
	if resp.Global.StoreRetentionDays == nil || *resp.Global.StoreRetentionDays != 5 {
		t.Fatalf("expected global.store_retention_days=5, got: %+v", resp.Global)
	}
}
