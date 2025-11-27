package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	api "github.com/nocodeleaks/quepasa/api/models"
	environment "github.com/nocodeleaks/quepasa/environment"
)

// Helper function to mask sensitive strings
func maskString(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return s[:2] + "****" + s[len(s)-2:]
}

// TestEnvironmentEndpoint_NoAuthentication tests /environment endpoint without authentication
func TestEnvironmentEndpoint_NoAuthentication(t *testing.T) {
	// Setup test environment
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	// Create request without any authentication headers or parameters
	req := httptest.NewRequest(http.MethodGet, "/environment", nil)
	rec := httptest.NewRecorder()

	// Call the controller
	EnvironmentController(rec, req)

	// Check response
	resp := rec.Result()
	defer resp.Body.Close()

	// Should return 200 OK with preview (no master key required)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 without authentication (preview mode), got %d", resp.StatusCode)
	}

	// Parse as EnvironmentSettingsPreview
	var preview environment.EnvironmentSettingsPreview
	if err := json.NewDecoder(resp.Body).Decode(&preview); err != nil {
		t.Fatalf("Failed to decode preview response: %v", err)
	}

	// Should have preview fields
	if preview.Groups == "" {
		t.Error("Expected groups field in preview")
	}
	if preview.Broadcasts == "" {
		t.Error("Expected broadcasts field in preview")
	}
	if preview.ReadReceipts == "" {
		t.Error("Expected read_receipts field in preview")
	}
	if preview.Calls == "" {
		t.Error("Expected calls field in preview")
	}

	t.Logf("No authentication returned preview: Groups=%s, Broadcasts=%s, ReadReceipts=%s, Calls=%s",
		preview.Groups, preview.Broadcasts, preview.ReadReceipts, preview.Calls)
}

// TestEnvironmentEndpoint_WithBotToken tests /environment endpoint with bot token
func TestEnvironmentEndpoint_WithBotToken(t *testing.T) {
	// Setup test environment
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	// Create test user and server
	testToken := "env-bot-token-12345"
	testUser := "envuser"
	testPassword := "envpass123"

	CreateTestUser(t, testUser, testPassword)
	server := CreateTestServer(t, testToken, testUser)

	// Create request with bot token in header
	req := httptest.NewRequest(http.MethodGet, "/environment", nil)
	req.Header.Set("X-QUEPASA-TOKEN", testToken)
	rec := httptest.NewRecorder()

	// Call the controller
	EnvironmentController(rec, req)

	// Check response
	resp := rec.Result()
	defer resp.Body.Close()

	// Should return 200 OK with preview (bot token is not master key, so gets preview)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 with bot token (preview mode), got %d", resp.StatusCode)
	}

	// Parse as EnvironmentSettingsPreview
	var preview environment.EnvironmentSettingsPreview
	if err := json.NewDecoder(resp.Body).Decode(&preview); err != nil {
		t.Fatalf("Failed to decode preview response: %v", err)
	}

	// Should have preview fields (not full settings)
	if preview.Groups == "" {
		t.Error("Expected groups field in preview")
	}

	t.Log("Bot token correctly returned preview (not full settings)")

	// Verify server exists
	if server == nil {
		t.Error("Expected server object to be created")
	}
}

// TestEnvironmentEndpoint_WithMasterKey tests /environment endpoint with master key
func TestEnvironmentEndpoint_WithMasterKey(t *testing.T) {
	// Setup test environment
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	// Setup test master key
	testMasterKey := "env-master-key-456"
	cleanup := SetupTestMasterKey(t, testMasterKey)
	defer cleanup()

	// Test 1: Master key in header
	t.Run("MasterKeyInHeader", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/environment", nil)
		req.Header.Set("X-QUEPASA-MASTERKEY", testMasterKey)
		rec := httptest.NewRecorder()

		EnvironmentController(rec, req)

		resp := rec.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 with master key, got %d", resp.StatusCode)
		}

		var response api.EnvironmentResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if !response.Success {
			t.Errorf("Expected success=true with master key, got: %+v", response)
		}

		// Should have environment settings
		if response.Settings == nil {
			t.Error("Expected environment settings with master key")
		} else {
			t.Logf("Environment settings present:")
			t.Logf("  Database: Driver=%s, Host=%s, Port=%s, Database=%s",
				response.Settings.Database.Driver, response.Settings.Database.Host,
				response.Settings.Database.Port, response.Settings.Database.Database)
			t.Logf("  API: MasterKey=%s, Prefix=%s, Timeout=%d",
				maskString(response.Settings.API.MasterKey),
				response.Settings.API.Prefix, response.Settings.API.Timeout)
			t.Logf("  WhatsApp: Groups=%v, Broadcasts=%v, Calls=%v, ReadReceipts=%v",
				response.Settings.WhatsApp.Groups, response.Settings.WhatsApp.Broadcasts,
				response.Settings.WhatsApp.Calls, response.Settings.WhatsApp.ReadReceipts)
			t.Logf("  Whatsmeow: LogLevel=%s, DBLogLevel=%s",
				response.Settings.Whatsmeow.LogLevel, response.Settings.Whatsmeow.DBLogLevel)

			// Verify main settings structure is present
			if response.Settings.Database.Driver == "" {
				t.Error("Expected Database settings to be populated")
			}
			if response.Settings.API.MasterKey == "" {
				t.Error("Expected API settings to be populated")
			}
		}
	})

	// Test 2: Master key in query parameter
	t.Run("MasterKeyInQueryParameter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/environment?masterkey="+testMasterKey, nil)
		rec := httptest.NewRecorder()

		EnvironmentController(rec, req)

		resp := rec.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 with master key in query, got %d", resp.StatusCode)
		}

		var response api.EnvironmentResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if !response.Success {
			t.Errorf("Expected success=true with master key in query, got: %+v", response)
		}

		// Should have environment settings
		if response.Settings == nil {
			t.Error("Expected environment settings with master key in query")
		} else {
			t.Log("Environment settings correctly present with master key query parameter")
		}

		t.Logf("Master key query access successful")
	})
}

// TestEnvironmentEndpoint_InvalidMasterKey tests /environment endpoint with invalid master key
func TestEnvironmentEndpoint_InvalidMasterKey(t *testing.T) {
	// Setup test environment
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	// Setup test master key
	testMasterKey := "env-master-key-789"
	cleanup := SetupTestMasterKey(t, testMasterKey)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/environment", nil)
	req.Header.Set("X-QUEPASA-MASTERKEY", "invalid-master-key-wrong")
	rec := httptest.NewRecorder()

	EnvironmentController(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	// Should return 200 OK with preview (invalid master key, so gets preview)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 with invalid master key (preview mode), got %d", resp.StatusCode)
	}

	// Parse as EnvironmentSettingsPreview
	var preview environment.EnvironmentSettingsPreview
	if err := json.NewDecoder(resp.Body).Decode(&preview); err != nil {
		t.Fatalf("Failed to decode preview response: %v", err)
	}

	// Should have preview fields (not full settings)
	if preview.Groups == "" {
		t.Error("Expected groups field in preview")
	}

	t.Log("Invalid master key correctly returned preview")
}

// TestEnvironmentEndpoint_AuthenticationPriority tests master key vs bot token
func TestEnvironmentEndpoint_AuthenticationPriority(t *testing.T) {
	// Setup test environment
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	// Setup test master key
	testMasterKey := "env-master-key-priority"
	cleanup := SetupTestMasterKey(t, testMasterKey)
	defer cleanup()

	// Create test user and server
	testToken := "priority-env-token"
	testUser := "priorityenvuser"
	testPassword := "envpass123"

	CreateTestUser(t, testUser, testPassword)
	server := CreateTestServer(t, testToken, testUser)

	// Send request with both token and masterkey
	// Master key should grant access
	req := httptest.NewRequest(http.MethodGet, "/environment", nil)
	req.Header.Set("X-QUEPASA-TOKEN", testToken)
	req.Header.Set("X-QUEPASA-MASTERKEY", testMasterKey)
	rec := httptest.NewRecorder()

	EnvironmentController(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 with master key, got %d", resp.StatusCode)
	}

	var response api.EnvironmentResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected success with master key")
	}

	// Should have environment settings (master key grants access)
	if response.Settings == nil {
		t.Error("Expected environment settings when master key is present")
	} else {
		t.Log("Environment settings correctly shown when master key present")
	}

	t.Log("Master key correctly granted access to environment settings")

	// Verify server exists
	if server == nil {
		t.Error("Expected server object to be created")
	}
}
