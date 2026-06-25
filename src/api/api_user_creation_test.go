package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestCanonicalUserCreationFirstUserNoMasterKey validates that first user can be created without master key
func TestCanonicalUserCreationFirstUserNoMasterKey(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	router := newCanonicalTestRouter()

	body := map[string]string{
		"email":    "firstuser@example.com",
		"password": "MySecurePassword123!@#Password",
	}
	bodyBytes, _ := json.Marshal(body)

	// First user creation without master key should succeed
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected first user creation to succeed (200), got %d: %s", rec.Code, rec.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response["result"] != "success" {
		t.Fatalf("expected success result, got %v", response["result"])
	}

	if response["username"] != "firstuser@example.com" {
		t.Fatalf("expected username to be firstuser@example.com, got %v", response["username"])
	}
}

// TestCanonicalUserCreationSecondUserRequiresMasterKey validates that second user requires master key
func TestCanonicalUserCreationSecondUserRequiresMasterKey(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	router := newCanonicalTestRouter()

	// Create first user
	firstBody := map[string]string{
		"email":    "firstuser@example.com",
		"password": "MySecurePassword123!@#Password",
	}
	firstBodyBytes, _ := json.Marshal(firstBody)

	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(firstBodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("first user creation failed: got %d", rec.Code)
	}

	// Attempt to create second user without master key should fail
	secondBody := map[string]string{
		"email":    "seconduser@example.com",
		"password": "MySecurePassword123!@#Password",
	}
	secondBodyBytes, _ := json.Marshal(secondBody)

	req = httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(secondBodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected second user creation without master key to fail (403), got %d: %s", rec.Code, rec.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		// If not JSON, just check the status code was correct
		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected 403 Forbidden, got %d", rec.Code)
		}
		return
	}
}

// TestCanonicalUserCreationSecondUserWithMasterKey validates that second user can be created with master key
func TestCanonicalUserCreationSecondUserWithMasterKey(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	// Set up master key for this test
	restore := SetupTestMasterKey(t, "test-master-key-12345")
	defer restore()

	router := newCanonicalTestRouter()

	// Create first user without master key
	firstBody := map[string]string{
		"email":    "firstuser@example.com",
		"password": "MySecurePassword123!@#Password",
	}
	firstBodyBytes, _ := json.Marshal(firstBody)

	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(firstBodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("first user creation failed: got %d", rec.Code)
	}

	// Create second user WITH master key should succeed
	secondBody := map[string]string{
		"email":    "seconduser@example.com",
		"password": "MySecurePassword123!@#Password",
	}
	secondBodyBytes, _ := json.Marshal(secondBody)

	req = httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(secondBodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-QUEPASA-MASTERKEY", "test-master-key-12345")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected second user creation with master key to succeed (200), got %d: %s", rec.Code, rec.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response["result"] != "success" {
		t.Fatalf("expected success result, got %v", response["result"])
	}

	if response["username"] != "seconduser@example.com" {
		t.Fatalf("expected username to be seconduser@example.com, got %v", response["username"])
	}
}

// TestCanonicalUserCreationInvalidPassword validates password strength requirement
func TestCanonicalUserCreationInvalidPassword(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	router := newCanonicalTestRouter()

	body := map[string]string{
		"email":    "user@example.com",
		"password": "weak",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected weak password to fail (400), got %d", rec.Code)
	}
}

// TestCanonicalUserCreationDuplicateEmail validates that duplicate email is rejected
func TestCanonicalUserCreationDuplicateEmail(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	router := newCanonicalTestRouter()

	body := map[string]string{
		"email":    "user@example.com",
		"password": "MySecurePassword123!@#Password",
	}
	bodyBytes, _ := json.Marshal(body)

	// First creation should succeed
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("first user creation failed: got %d", rec.Code)
	}

	// Second creation with same email should fail (need master key for second user anyway)
	restore := SetupTestMasterKey(t, "test-master-key-12345")
	defer restore()

	req = httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-QUEPASA-MASTERKEY", "test-master-key-12345")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected duplicate email to fail (400), got %d: %s", rec.Code, rec.Body.String())
	}
}
