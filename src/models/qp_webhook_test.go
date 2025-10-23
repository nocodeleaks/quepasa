package models

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

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
		"source":  "test",
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
