package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	api "github.com/nocodeleaks/quepasa/api/models"
)

// TestInfoEndpoint_NoAuthentication tests /info endpoint without authentication
func TestInfoEndpoint_NoAuthentication(t *testing.T) {
	// Setup test environment
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	// Create request without any authentication headers or parameters
	req := httptest.NewRequest(http.MethodGet, "/info", nil)
	rec := httptest.NewRecorder()

	// Call the controller
	GetInformationController(rec, req)

	// Check response
	resp := rec.Result()
	defer resp.Body.Close()

	// Should return 200 OK with basic application information
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 without authentication, got %d", resp.StatusCode)
	}

	// Parse response
	var response api.InformationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Should be successful
	if !response.Success {
		t.Errorf("Expected success=true, got: %+v", response)
	}

	// Should have version
	if response.Version == "" {
		t.Error("Expected version field")
	}

	// Should NOT have server info
	if response.Server != nil {
		t.Error("Expected no server field without authentication")
	}

	t.Logf("No auth response: Version=%s", response.Version)
}

// TestInfoEndpoint_WithBotToken tests /info endpoint with X-QUEPASA-TOKEN
func TestInfoEndpoint_WithBotToken(t *testing.T) {
	// Setup test environment
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	// Create test user and server
	testToken := "test-bot-token-12345"
	testUser := "testuser"
	testPassword := "testpass123"

	CreateTestUser(t, testUser, testPassword)
	server := CreateTestServer(t, testToken, testUser)

	// Create request with bot token in header
	req := httptest.NewRequest(http.MethodGet, "/info", nil)
	req.Header.Set("X-QUEPASA-TOKEN", testToken)
	rec := httptest.NewRecorder()

	// Call the controller
	GetInformationController(rec, req)

	// Check response
	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 with bot token, got %d", resp.StatusCode)
	}

	// Parse response
	var response api.InformationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Should be successful
	if !response.Success {
		t.Errorf("Expected success=true with bot token, got: %+v", response)
	}

	// Should have version
	if response.Version == "" {
		t.Error("Expected version field")
	}

	// Should have server info
	if response.Server == nil {
		t.Error("Expected server info in response")
	} else {
		if response.Server.Token != testToken {
			t.Errorf("Expected token %s, got %s", testToken, response.Server.Token)
		}
		t.Logf("Server info: Token=%s, User=%s, Connected=%v",
			response.Server.Token, response.Server.User, response.Server.GetStatus())
	}

	t.Log("Bot token correctly returns base information")

	// Verify server object is returned
	if server == nil {
		t.Error("Expected server object to be created")
	}
}

// TestInfoEndpoint_WithMasterKey tests /info endpoint with MASTERKEY
func TestInfoEndpoint_WithMasterKey(t *testing.T) {
	// Setup test environment
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	// Setup test master key
	testMasterKey := "test-master-key-456"
	cleanup := SetupTestMasterKey(t, testMasterKey)
	defer cleanup()

	// Create test user and server
	testToken := "test-master-token-67890"
	testUser := "masteruser"
	testPassword := "testpass456"

	CreateTestUser(t, testUser, testPassword)
	server := CreateTestServer(t, testToken, testUser)

	// Test 1: Master key in header with bot token
	t.Run("MasterKeyInHeaderWithToken", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/info", nil)
		req.Header.Set("X-QUEPASA-MASTERKEY", testMasterKey)
		req.Header.Set("X-QUEPASA-TOKEN", testToken)
		rec := httptest.NewRecorder()

		GetInformationController(rec, req)

		resp := rec.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 with master key, got %d", resp.StatusCode)
		}

		var response api.InformationResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if !response.Success {
			t.Errorf("Expected success=true with master key, got: %+v", response)
		}

		// Should have version
		if response.Version == "" {
			t.Error("Expected version field")
		}

		if response.Server == nil {
			t.Error("Expected server info in response")
		} else {
			t.Logf("Master key access - Server info: Token=%s, User=%s",
				response.Server.Token, response.Server.User)
		}
	})

	// Test 2: Master key in query parameter
	t.Run("MasterKeyInQueryParameter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/info?masterkey="+testMasterKey+"&token="+testToken, nil)
		rec := httptest.NewRecorder()

		GetInformationController(rec, req)

		resp := rec.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 with master key in query, got %d", resp.StatusCode)
		}

		var response api.InformationResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if !response.Success {
			t.Errorf("Expected success=true with master key in query, got: %+v", response)
		}

		t.Logf("Master key query access successful")
	})

	// Verify server exists
	if server == nil {
		t.Error("Expected server object to be created")
	}
}

// TestInfoEndpoint_InvalidToken tests /info endpoint with invalid bot token
func TestInfoEndpoint_InvalidToken(t *testing.T) {
	// Setup test environment
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	req := httptest.NewRequest(http.MethodGet, "/info", nil)
	req.Header.Set("X-QUEPASA-TOKEN", "invalid-token-does-not-exist")
	rec := httptest.NewRecorder()

	GetInformationController(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	// Should return 204 No Content (server not found)
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected status 204 with invalid token, got %d", resp.StatusCode)
	}

	t.Log("Invalid token correctly returned 204 No Content")
}

func TestInfoPost_AllowsMultiplePlaceholderServersWithEmptyWid(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	testUser := "testuser"
	testPassword := "testpass123"
	CreateTestUser(t, testUser, testPassword)

	create := func(t *testing.T, token string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/info?user="+testUser, nil)
		req.Header.Set("X-QUEPASA-TOKEN", token)
		rec := httptest.NewRecorder()
		CreateInformationController(rec, req)
		return rec
	}

	// First placeholder (wid empty)
	resp1 := create(t, "test-token-001")
	if resp1.Result().StatusCode != http.StatusCreated {
		t.Fatalf("Expected status 201 for first placeholder server, got %d", resp1.Result().StatusCode)
	}

	// Second placeholder (wid empty again) must also be accepted
	resp2 := create(t, "test-token-002")
	if resp2.Result().StatusCode != http.StatusCreated {
		body := resp2.Body.String()
		t.Fatalf("Expected status 201 for second placeholder server, got %d. Body: %s", resp2.Result().StatusCode, body)
	}

	// Basic response validation
	var response api.InformationResponse
	if err := json.NewDecoder(resp2.Result().Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if !response.Success {
		t.Fatalf("Expected success=true for second placeholder creation, got: %+v", response)
	}
}

// TestInfoEndpoint_AuthenticationPriority tests authentication header priority
func TestInfoEndpoint_AuthenticationPriority(t *testing.T) {
	// Setup test environment
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	// Setup test master key
	testMasterKey := "test-master-key-789"
	cleanup := SetupTestMasterKey(t, testMasterKey)
	defer cleanup()

	// Create test user and server
	testToken := "priority-test-token"
	testUser := "priorityuser"
	testPassword := "testpass789"

	CreateTestUser(t, testUser, testPassword)
	server := CreateTestServer(t, testToken, testUser)

	// Send request with both token and masterkey
	req := httptest.NewRequest(http.MethodGet, "/info", nil)
	req.Header.Set("X-QUEPASA-TOKEN", testToken)
	req.Header.Set("X-QUEPASA-MASTERKEY", testMasterKey)
	rec := httptest.NewRecorder()

	GetInformationController(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 with both credentials, got %d", resp.StatusCode)
	}

	var response api.InformationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected success with both credentials")
	}

	t.Log("Authentication with both token and masterkey handled correctly")

	// Verify server exists
	if server == nil {
		t.Error("Expected server object to be created")
	}
}
