package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	api "github.com/nocodeleaks/quepasa/api/models"
)

// TestHealthEndpoint_NoAuthentication tests /health endpoint without authentication
func TestHealthEndpoint_NoAuthentication(t *testing.T) {
	// Setup test environment
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	// Create request without any authentication headers or parameters
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	// Call the controller
	HealthController(rec, req)

	// Check response
	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 without authentication, got %d", resp.StatusCode)
	}

	// Parse response
	var response api.HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Should return basic health info
	if !response.Success {
		t.Errorf("Expected success=true without authentication, got: %+v", response)
	}

	// Should NOT have items array (no master key)
	if response.Items != nil {
		t.Errorf("Expected no items array without authentication, got: %+v", response.Items)
	}

	// Should NOT have State/StateCode/Wid (no token)
	if response.State != nil || response.Wid != "" {
		t.Errorf("Expected no State/Wid without authentication, got State=%v, Wid=%s",
			response.State, response.Wid)
	}

	t.Logf("No auth response: Success=%v, Timestamp=%v", response.Success, response.Timestamp)
}

// TestHealthEndpoint_WithBotToken tests /health endpoint with X-QUEPASA-TOKEN
func TestHealthEndpoint_WithBotToken(t *testing.T) {
	// Setup test environment
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	// Create test user and server
	testToken := "health-bot-token-12345"
	testUser := "healthuser"
	testPassword := "healthpass123"

	CreateTestUser(t, testUser, testPassword)
	server := CreateTestServer(t, testToken, testUser)

	// Create request with bot token in header
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("X-QUEPASA-TOKEN", testToken)
	rec := httptest.NewRecorder()

	// Call the controller
	HealthController(rec, req)

	// Check response
	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 with bot token, got %d", resp.StatusCode)
	}

	// Parse response
	var response api.HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Note: Success may be false if server is UnVerified/disconnected
	// But HTTP status is always 200 OK for bot token authentication

	// Should have State, StateCode IN BODY (not in items array)
	if response.State == nil || response.State.String() == "" {
		t.Error("Expected State field in response body with bot token")
	}
	// Note: StateCode is computed automatically from State during JSON serialization
	// Note: Wid may be empty for UnVerified servers

	// Should NOT have items array (only with master key)
	if response.Items != nil {
		t.Errorf("Expected no items array with bot token, got: %+v", response.Items)
	}

	t.Logf("Bot token response: State=%s, Wid=%s",
		response.State.String(), response.Wid)

	// Verify server object is returned
	if server == nil {
		t.Error("Expected server object to be created")
	}
}

// TestHealthEndpoint_WithMasterKey tests /health endpoint with MASTERKEY
func TestHealthEndpoint_WithMasterKey(t *testing.T) {
	// Setup test environment
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	// Setup test master key
	testMasterKey := "health-master-key-456"
	cleanup := SetupTestMasterKey(t, testMasterKey)
	defer cleanup()

	// Create test servers
	testToken1 := "health-master-token-001"
	testToken2 := "health-master-token-002"
	testUser1 := "healthmaster1"
	testUser2 := "healthmaster2"
	testPassword := "healthpass456"

	CreateTestUser(t, testUser1, testPassword)
	CreateTestUser(t, testUser2, testPassword)
	server1 := CreateTestServer(t, testToken1, testUser1)
	server2 := CreateTestServer(t, testToken2, testUser2)

	// Test 1: Master key in header
	t.Run("MasterKeyInHeader", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		req.Header.Set("X-QUEPASA-MASTERKEY", testMasterKey)
		rec := httptest.NewRecorder()

		HealthController(rec, req)

		resp := rec.Result()
		defer resp.Body.Close()

		// Master key: returns 400 if servers are unhealthy (based on stats)
		// This is the only case where HTTP status reflects health
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 200 or 400 with master key, got %d", resp.StatusCode)
		}

		var response api.HealthResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Note: Success may be false if servers are not in healthy state
		// This is expected for test servers that are UnVerified

		// Should have items array with all servers
		if response.Items == nil || len(response.Items) == 0 {
			t.Error("Expected items array with servers using master key")
		} else {
			t.Logf("Master key returned %d servers in items array", len(response.Items))
			for _, item := range response.Items {
				t.Logf("  Server: Token=%s, State=%s, StateCode=%d, Wid=%s",
					item.Token, item.State, item.StateCode, item.Wid)
			}
		}

		// Should NOT have State/StateCode/Wid in main body (only in items array)
		if response.State != nil || response.Wid != "" {
			t.Errorf("Expected no State/Wid in main body with master key, got State=%v, Wid=%s",
				response.State, response.Wid)
		}
	})

	// Test 2: Master key in query parameter
	t.Run("MasterKeyInQueryParameter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health?masterkey="+testMasterKey, nil)
		rec := httptest.NewRecorder()

		HealthController(rec, req)

		resp := rec.Result()
		defer resp.Body.Close()

		// Master key: can return 400 if servers unhealthy
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 200 or 400 with master key, got %d", resp.StatusCode)
		}

		var response api.HealthResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Note: Success may be false if servers are not healthy

		// Should have items array
		if response.Items == nil || len(response.Items) == 0 {
			t.Error("Expected items array with servers using master key in query")
		} else {
			t.Logf("Master key query returned %d servers", len(response.Items))
		}
	})

	// Verify servers exist
	if server1 == nil || server2 == nil {
		t.Error("Expected server objects to be created")
	}
}

// TestHealthEndpoint_InvalidToken tests /health endpoint with invalid bot token
func TestHealthEndpoint_InvalidToken(t *testing.T) {
	// Setup test environment
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("X-QUEPASA-TOKEN", "invalid-health-token-does-not-exist")
	rec := httptest.NewRecorder()

	HealthController(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	// Invalid token returns 200 OK with Success=false
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 with invalid token, got %d", resp.StatusCode)
	}

	var response api.HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Should return Success=false for invalid token
	if response.Success {
		t.Errorf("Expected success=false with invalid token, got: %+v", response)
	}

	// Should NOT have State/StateCode/Wid (invalid token)
	if response.State != nil || response.Wid != "" {
		t.Errorf("Expected no State/Wid with invalid token")
	}

	t.Log("Invalid token correctly returned basic health info")
}

// TestHealthEndpoint_AuthenticationPriority tests authentication header priority
func TestHealthEndpoint_AuthenticationPriority(t *testing.T) {
	// Setup test environment
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	// Setup test master key
	testMasterKey := "health-master-key-789"
	cleanup := SetupTestMasterKey(t, testMasterKey)
	defer cleanup()

	// Create test user and server
	testToken := "priority-health-token"
	testUser := "priorityhealthuser"
	testPassword := "healthpass789"

	CreateTestUser(t, testUser, testPassword)
	server := CreateTestServer(t, testToken, testUser)

	// Send request with both token and masterkey
	// Master key should take precedence
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("X-QUEPASA-TOKEN", testToken)
	req.Header.Set("X-QUEPASA-MASTERKEY", testMasterKey)
	rec := httptest.NewRecorder()

	HealthController(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	// Master key takes precedence, can return 400 if unhealthy
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 200 or 400 with master key, got %d", resp.StatusCode)
	}

	var response api.HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Note: Success may be false if servers are not healthy

	// Should return items array (master key takes precedence)
	if response.Items == nil || len(response.Items) == 0 {
		t.Error("Expected items array when master key is present")
	}

	// Should NOT have State/StateCode/Wid in main body
	if response.State != nil || response.Wid != "" {
		t.Errorf("Expected no State/Wid in main body when master key present")
	}

	t.Log("Master key correctly took precedence over bot token")

	// Verify server exists
	if server == nil {
		t.Error("Expected server object to be created")
	}
}

// TestHealthEndpoint_WithUserPassword tests /health endpoint with X-QUEPASA-USER and X-QUEPASA-PASSWORD
func TestHealthEndpoint_WithUserPassword(t *testing.T) {
	// Setup test environment
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	// Create test users
	testUser1 := "healthuser1"
	testPassword1 := "healthpass1"
	testUser2 := "healthuser2"
	testPassword2 := "healthpass2"

	CreateTestUser(t, testUser1, testPassword1)
	CreateTestUser(t, testUser2, testPassword2)

	// Create servers for user1
	server1 := CreateTestServer(t, "user1-token-001", testUser1)
	server2 := CreateTestServer(t, "user1-token-002", testUser1)

	// Create servers for user2
	server3 := CreateTestServer(t, "user2-token-001", testUser2)

	// Test 1: Valid user credentials - should return only user's servers
	t.Run("ValidUserCredentials", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		req.Header.Set("X-QUEPASA-USER", testUser1)
		req.Header.Set("X-QUEPASA-PASSWORD", testPassword1)
		rec := httptest.NewRecorder()

		HealthController(rec, req)

		resp := rec.Result()
		defer resp.Body.Close()

		// Can return 400 if servers are unhealthy
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 200 or 400 with valid credentials, got %d", resp.StatusCode)
		}

		var response api.HealthResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Should have items array with user's servers only
		if response.Items == nil {
			t.Fatal("Expected items array with user credentials")
		}

		if len(response.Items) != 2 {
			t.Errorf("Expected 2 servers for user1, got %d", len(response.Items))
		}

		// Verify returned servers belong to user1
		for _, item := range response.Items {
			found := false
			if item.Token == "user1-token-001" || item.Token == "user1-token-002" {
				found = true
			}
			if !found {
				t.Errorf("Unexpected server token in response: %s", item.Token)
			}
		}

		// Should have stats
		if response.Stats == nil {
			t.Error("Expected stats field with user credentials")
		} else {
			if response.Stats.Total != 2 {
				t.Errorf("Expected total=2 in stats, got %d", response.Stats.Total)
			}
		}

		t.Logf("User %s has %d servers with %d healthy, %d unhealthy",
			testUser1, response.Stats.Total, response.Stats.Healthy, response.Stats.Unhealthy)
	})

	// Test 2: Invalid username
	t.Run("InvalidUsername", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		req.Header.Set("X-QUEPASA-USER", "nonexistent")
		req.Header.Set("X-QUEPASA-PASSWORD", "wrongpass")
		rec := httptest.NewRecorder()

		HealthController(rec, req)

		resp := rec.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 with invalid credentials, got %d", resp.StatusCode)
		}

		var response api.HealthResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Should return error
		if response.Success {
			t.Error("Expected success=false with invalid username")
		}

		t.Log("Invalid username correctly rejected")
	})

	// Test 3: Valid username but wrong password
	t.Run("WrongPassword", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		req.Header.Set("X-QUEPASA-USER", testUser1)
		req.Header.Set("X-QUEPASA-PASSWORD", "wrongpassword")
		rec := httptest.NewRecorder()

		HealthController(rec, req)

		resp := rec.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 with wrong password, got %d", resp.StatusCode)
		}

		var response api.HealthResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Should return error
		if response.Success {
			t.Error("Expected success=false with wrong password")
		}

		t.Log("Wrong password correctly rejected")
	})

	// Test 4: User with no servers
	t.Run("UserWithNoServers", func(t *testing.T) {
		testUser3 := "healthuser3"
		testPassword3 := "healthpass3"
		CreateTestUser(t, testUser3, testPassword3)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		req.Header.Set("X-QUEPASA-USER", testUser3)
		req.Header.Set("X-QUEPASA-PASSWORD", testPassword3)
		rec := httptest.NewRecorder()

		HealthController(rec, req)

		resp := rec.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 for user with no servers, got %d", resp.StatusCode)
		}

		var response api.HealthResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Should be successful but with empty items
		if !response.Success {
			t.Errorf("Expected success=true for user with no servers")
		}

		if response.Items != nil && len(response.Items) > 0 {
			t.Errorf("Expected no items for user with no servers, got %d", len(response.Items))
		}

		t.Log("User with no servers handled correctly")
	})

	// Verify servers exist
	if server1 == nil || server2 == nil || server3 == nil {
		t.Error("Expected all server objects to be created")
	}
}
