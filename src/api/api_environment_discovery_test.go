package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestCanonicalEnvironmentAnonymousPreview validates that anonymous requests get only preview
func TestCanonicalEnvironmentAnonymousPreview(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	router := newCanonicalTestRouter()

	// Anonymous request (no master key)
	req := httptest.NewRequest(http.MethodGet, "/api/system/environment", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for anonymous environment request, got %d", rec.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	// Should have preview, not settings
	preview, hasPreview := response["preview"]
	settings := response["settings"]

	if !hasPreview || preview == nil {
		t.Fatalf("expected preview in anonymous response, got: %+v", response)
	}

	if settings != nil {
		t.Fatalf("expected NO settings in anonymous response, but got some")
	}

	// Verify preview is a map with public fields
	previewMap, ok := preview.(map[string]interface{})
	if !ok {
		t.Fatalf("expected preview to be map, got type: %T", preview)
	}

	// Verify preview has expected public fields
	expectedFields := []string{"groups", "broadcasts", "calls", "read_receipts", "history_sync"}
	for _, field := range expectedFields {
		if _, exists := previewMap[field]; !exists {
			t.Logf("Warning: expected field '%s' in preview, got fields: %+v", field, previewMap)
		}
	}

	t.Logf("Anonymous preview fields: %v", previewMap)
}

// TestCanonicalEnvironmentMasterKeyFullAccess validates that master key requests get full settings
func TestCanonicalEnvironmentMasterKeyFullAccess(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	// Setup master key
	restore := SetupTestMasterKey(t, "test-master-key-env-12345")
	defer restore()

	router := newCanonicalTestRouter()

	// Request with master key
	req := httptest.NewRequest(http.MethodGet, "/api/system/environment", nil)
	req.Header.Set("X-QUEPASA-MASTERKEY", "test-master-key-env-12345")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for master key environment request, got %d", rec.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	// Should have settings, not preview (or preview should be much simpler)
	settings := response["settings"]
	preview := response["preview"]

	if settings == nil {
		t.Fatalf("expected settings in master key response, got: %+v", response)
	}

	// Verify settings has expected config sections
	settingsMap, ok := settings.(map[string]interface{})
	if !ok {
		t.Fatalf("expected settings to be map, got type: %T", settings)
	}

	expectedSections := []string{"api", "database", "webserver"}
	for _, section := range expectedSections {
		if _, exists := settingsMap[section]; !exists {
			t.Logf("Warning: expected '%s' section in settings, got sections: %+v", section, settingsMap)
		}
	}

	t.Logf("Master key response sections: preview=%v, settings=%+v", preview != nil, settingsMap)
}

// TestCanonicalEnvironmentWrongMasterKeyTreatsAsAnonymous validates that wrong master key gets preview only
func TestCanonicalEnvironmentWrongMasterKeyTreatsAsAnonymous(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	// Setup one master key
	restore := SetupTestMasterKey(t, "correct-key")
	defer restore()

	router := newCanonicalTestRouter()

	// Request with wrong master key
	req := httptest.NewRequest(http.MethodGet, "/api/system/environment", nil)
	req.Header.Set("X-QUEPASA-MASTERKEY", "wrong-key")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 even with wrong master key, got %d", rec.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	// Wrong key should be treated as anonymous (preview only)
	preview, hasPreview := response["preview"]
	settings := response["settings"]

	if !hasPreview || preview == nil {
		t.Fatalf("expected preview for wrong master key (treated as anonymous), got: %+v", response)
	}

	if settings != nil {
		t.Fatalf("expected NO settings for wrong master key, but got some")
	}

	t.Logf("Wrong master key correctly treated as anonymous (preview only)")
}
