package models

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nocodeleaks/quepasa/environment"
	"github.com/nocodeleaks/quepasa/whatsapp"
)

// TestWebhookRetrySuccess tests successful webhook delivery on first attempt
func TestWebhookRetrySuccess(t *testing.T) {
	// Setup test server that responds with 200 OK
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	}))
	defer server.Close()

	// Create webhook with test server URL
	webhook := &QpWebhook{
		Url: server.URL,
		Wid: "test@whatsapp",
	}

	// Create test message
	message := &whatsapp.WhatsappMessage{
		Id:   "test-message-id",
		Text: "Test message",
	}

	// Execute webhook post
	err := webhook.Post(message)

	// Verify success
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if webhook.Success == nil {
		t.Error("Expected Success timestamp to be set")
	}

	if webhook.Failure != nil {
		t.Error("Expected Failure timestamp to be nil")
	}
}

// TestWebhookRetryFailure tests webhook retry logic with failures
func TestWebhookRetryFailure(t *testing.T) {
	// Setup test server that always responds with 500 error
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error")
	}))
	defer server.Close()

	// Set test environment values for faster testing
	originalRetryCount := environment.Settings.API.WebhookRetryCount
	originalRetryDelay := environment.Settings.API.WebhookRetryDelay
	originalTimeout := environment.Settings.API.WebhookTimeout

	environment.Settings.API.WebhookRetryCount = 2 // 3 total attempts (0, 1, 2)
	environment.Settings.API.WebhookRetryDelay = 0 // No delay for faster testing
	environment.Settings.API.WebhookTimeout = 1    // 1 second timeout

	defer func() {
		// Restore original values
		environment.Settings.API.WebhookRetryCount = originalRetryCount
		environment.Settings.API.WebhookRetryDelay = originalRetryDelay
		environment.Settings.API.WebhookTimeout = originalTimeout
	}()

	// Create webhook with test server URL
	webhook := &QpWebhook{
		Url: server.URL,
		Wid: "test@whatsapp",
	}

	// Create test message
	message := &whatsapp.WhatsappMessage{
		Id:   "test-message-id",
		Text: "Test message",
	}

	// Execute webhook post
	err := webhook.Post(message)

	// Verify failure
	if err == nil {
		t.Error("Expected error, got nil")
	}

	if err != ErrInvalidResponse {
		t.Errorf("Expected ErrInvalidResponse, got: %v", err)
	}

	// Verify retry attempts (should be 3: initial + 2 retries)
	expectedAttempts := 3
	if attemptCount != expectedAttempts {
		t.Errorf("Expected %d attempts, got %d", expectedAttempts, attemptCount)
	}

	if webhook.Failure == nil {
		t.Error("Expected Failure timestamp to be set")
	}

	if webhook.Success != nil {
		t.Error("Expected Success timestamp to be nil")
	}
}

// TestWebhookRetrySuccessAfterFailure tests webhook success after initial failure
func TestWebhookRetrySuccessAfterFailure(t *testing.T) {
	// Setup test server that fails first two attempts, then succeeds
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "Internal Server Error")
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "OK")
		}
	}))
	defer server.Close()

	// Set test environment values
	originalRetryCount := environment.Settings.API.WebhookRetryCount
	originalRetryDelay := environment.Settings.API.WebhookRetryDelay
	originalTimeout := environment.Settings.API.WebhookTimeout

	environment.Settings.API.WebhookRetryCount = 3 // 4 total attempts (0, 1, 2, 3)
	environment.Settings.API.WebhookRetryDelay = 0 // No delay for faster testing
	environment.Settings.API.WebhookTimeout = 1    // 1 second timeout

	defer func() {
		// Restore original values
		environment.Settings.API.WebhookRetryCount = originalRetryCount
		environment.Settings.API.WebhookRetryDelay = originalRetryDelay
		environment.Settings.API.WebhookTimeout = originalTimeout
	}()

	// Create webhook with test server URL
	webhook := &QpWebhook{
		Url: server.URL,
		Wid: "test@whatsapp",
	}

	// Create test message
	message := &whatsapp.WhatsappMessage{
		Id:   "test-message-id",
		Text: "Test message",
	}

	// Execute webhook post
	err := webhook.Post(message)

	// Verify success
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify retry attempts (should be 3: 2 failures + 1 success)
	expectedAttempts := 3
	if attemptCount != expectedAttempts {
		t.Errorf("Expected %d attempts, got %d", expectedAttempts, attemptCount)
	}

	if webhook.Success == nil {
		t.Error("Expected Success timestamp to be set")
	}

	if webhook.Failure != nil {
		t.Error("Expected Failure timestamp to be nil after success")
	}
}

// TestWebhookPayloadFormat tests that the webhook payload is correctly formatted
func TestWebhookPayloadFormat(t *testing.T) {
	var receivedPayload *QpWebhookPayload

	// Setup test server to capture the payload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check Content-Type header
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
		}

		// Check User-Agent header
		userAgent := r.Header.Get("User-Agent")
		if userAgent != "Quepasa" {
			t.Errorf("Expected User-Agent 'Quepasa', got '%s'", userAgent)
		}

		// Check X-QUEPASA-WID header
		wid := r.Header.Get("X-QUEPASA-WID")
		if wid != "test@whatsapp" {
			t.Errorf("Expected X-QUEPASA-WID 'test@whatsapp', got '%s'", wid)
		}

		// Decode payload
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&receivedPayload)
		if err != nil {
			t.Errorf("Failed to decode payload: %v", err)
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	}))
	defer server.Close()

	// Create webhook with extra data
	extraData := map[string]interface{}{
		"source": "test",
		"version": "1.0",
	}

	webhook := &QpWebhook{
		Url:   server.URL,
		Wid:   "test@whatsapp",
		Extra: extraData,
	}

	// Create test message
	message := &whatsapp.WhatsappMessage{
		Id:   "test-message-id",
		Text: "Test message",
		Chat: whatsapp.WhatsappChat{
			Id: "test-chat-id",
		},
	}

	// Execute webhook post
	err := webhook.Post(message)

	// Verify success
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify payload structure
	if receivedPayload == nil {
		t.Fatal("No payload received")
	}

	if receivedPayload.WhatsappMessage == nil {
		t.Error("Expected WhatsappMessage in payload")
	}

	if receivedPayload.WhatsappMessage.Id != "test-message-id" {
		t.Errorf("Expected message ID 'test-message-id', got '%s'", receivedPayload.WhatsappMessage.Id)
	}

	if receivedPayload.Extra == nil {
		t.Error("Expected Extra data in payload")
	}
}

// TestWebhookTimeoutHandling tests webhook timeout handling
func TestWebhookTimeoutHandling(t *testing.T) {
	// Setup test server that delays response longer than timeout
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Delay longer than timeout
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	}))
	defer server.Close()

	// Set short timeout for testing
	originalTimeout := environment.Settings.API.WebhookTimeout
	originalRetryCount := environment.Settings.API.WebhookRetryCount

	environment.Settings.API.WebhookTimeout = 1    // 1 second timeout
	environment.Settings.API.WebhookRetryCount = 1 // 2 total attempts

	defer func() {
		// Restore original values
		environment.Settings.API.WebhookTimeout = originalTimeout
		environment.Settings.API.WebhookRetryCount = originalRetryCount
	}()

	// Create webhook with test server URL
	webhook := &QpWebhook{
		Url: server.URL,
		Wid: "test@whatsapp",
	}

	// Create test message
	message := &whatsapp.WhatsappMessage{
		Id:   "test-message-id",
		Text: "Test message",
	}

	// Execute webhook post
	start := time.Now()
	err := webhook.Post(message)
	duration := time.Since(start)

	// Verify timeout error
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}

	// Verify that it didn't take too long (should timeout quickly)
	if duration > 5*time.Second {
		t.Errorf("Webhook took too long: %v", duration)
	}

	if webhook.Failure == nil {
		t.Error("Expected Failure timestamp to be set")
	}
}
