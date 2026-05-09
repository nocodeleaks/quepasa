package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// N8nQuepasaIntegrationTests covers all HTTP requests made from n8n workflows to QuePasa API v4
// Reference: docs/N8N_QUEPASA_API_REQUESTS_MAPPING.md

// ============================================================================
// Test Fixtures
// ============================================================================

type MockServer struct {
	LastRequest *http.Request
	Response    interface{}
	StatusCode  int
}

func (m *MockServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.LastRequest = r
	w.WriteHeader(m.StatusCode)
	if m.Response != nil {
		json.NewEncoder(w).Encode(m.Response)
	}
}

// ============================================================================
// 1. QuepasaChatControl Tests - Send Text Message
// ============================================================================

func TestN8n_SendTextMessage(t *testing.T) {
	// Test: QuepasaChatControl.json -> Send text message via custom node
	// Endpoint: POST /messages/sendtext (implicit v4)
	// Auth: token parameter

	mockServer := &MockServer{
		Response: map[string]interface{}{
			"success": true,
			"message": "Text message sent successfully",
			"id":      "msg123",
		},
		StatusCode: http.StatusOK,
	}

	server := httptest.NewServer(mockServer)
	defer server.Close()

	// Simulate n8n custom node request
	requestBody := map[string]interface{}{
		"chatid": "5511988887777@s.whatsapp.net",
		"text":   "Convite do grupo: https://chat.whatsapp.com/...",
	}

	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", server.URL+"/messages/sendtext", bytes.NewBuffer(body))
	req.Header.Set("X-QUEPASA-TOKEN", "test-token-123")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "POST", mockServer.LastRequest.Method)
	assert.Equal(t, "test-token-123", mockServer.LastRequest.Header.Get("X-QUEPASA-TOKEN"))
}

func TestN8n_SendTextMessage_ValidationChatID(t *testing.T) {
	// Validation: chatid must be valid WhatsApp ID format
	tests := []struct {
		chatid      string
		expectValid bool
		description string
	}{
		{"5511988887777@s.whatsapp.net", true, "Valid individual chat ID"},
		{"5511988887777@g.us", true, "Valid group chat ID"},
		{"121281638842371@lid", true, "Valid LID format"},
		{"invalid-chatid", false, "Invalid format - missing domain"},
		{"", false, "Empty chatid"},
		{"5511988887777", false, "Missing WhatsApp domain"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			requestBody := map[string]interface{}{
				"chatid": tt.chatid,
				"text":   "Test message",
			}

			body, _ := json.Marshal(requestBody)
			req, _ := http.NewRequest("POST", "http://localhost/messages/sendtext", bytes.NewBuffer(body))

			// Validate chatid format
			isValid := false
			if tt.chatid != "" && (bytes.Contains([]byte(tt.chatid), []byte("@s.whatsapp.net")) ||
				bytes.Contains([]byte(tt.chatid), []byte("@g.us")) ||
				bytes.Contains([]byte(tt.chatid), []byte("@lid"))) {
				isValid = true
			}

			assert.Equal(t, tt.expectValid, isValid, "Failed for: "+tt.description)
			_ = req // use req to avoid unused error
		})
	}
}

// ============================================================================
// 2. QuepasaChatControl Tests - Get Group Invite Link
// ============================================================================

func TestN8n_GetGroupInviteLink(t *testing.T) {
	// Test: QuepasaChatControl.json -> Get invite link for group
	// Endpoint: GET /control/invite (implicit v4)
	// Auth: token parameter

	mockServer := &MockServer{
		Response: map[string]interface{}{
			"success":    true,
			"url":        "https://chat.whatsapp.com/ABC123DEF456",
			"inviteCode": "ABC123DEF456",
		},
		StatusCode: http.StatusOK,
	}

	server := httptest.NewServer(mockServer)
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL+"/control/invite?chatid=5511988887777@g.us", nil)
	req.Header.Set("X-QUEPASA-TOKEN", "test-token-123")

	client := &http.Client{}
	resp, err := client.Do(req)

	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "GET", mockServer.LastRequest.Method)

	var responseData map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&responseData)
	assert.Equal(t, true, responseData["success"])
	assert.Contains(t, responseData["url"], "https://chat.whatsapp.com")
}

func TestN8n_GetGroupInviteLink_OnlyGroups(t *testing.T) {
	// Validation: Can only get invite link for groups (@g.us), not individual chats
	tests := []struct {
		chatid      string
		shouldFail  bool
		description string
	}{
		{"5511988887777@g.us", false, "Valid group"},
		{"5511988887777@s.whatsapp.net", true, "Individual chat - should fail"},
		{"121281638842371@lid", true, "LID - not a group"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			isGroup := bytes.Contains([]byte(tt.chatid), []byte("@g.us"))
			assert.Equal(t, !tt.shouldFail, isGroup)
		})
	}
}

// ============================================================================
// 3. PostToChatwoot Tests - Download Media
// ============================================================================

func TestN8n_DownloadMediaFromQuepasa(t *testing.T) {
	// Test: PostToChatwoot.json -> Download media attachment
	// Endpoint: GET /download/:messageid (implicit v4)
	// Auth: token parameter

	mockServer := &MockServer{
		Response:   []byte("PNG binary data..."),
		StatusCode: http.StatusOK,
	}

	server := httptest.NewServer(mockServer)
	defer server.Close()

	req, _ := http.NewRequest("GET",
		server.URL+"/download/msg123?filename=photo.jpg",
		nil)
	req.Header.Set("X-QUEPASA-TOKEN", "test-token-123")

	client := &http.Client{}
	resp, err := client.Do(req)

	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "GET", mockServer.LastRequest.Method)
}

func TestN8n_DownloadMediaValidation(t *testing.T) {
	// Validation: messageid must not be empty
	tests := []struct {
		messageid   string
		expectValid bool
		description string
	}{
		{"msg123", true, "Valid message ID"},
		{"echo_12345_67890", true, "Echo message ID format"},
		{"", false, "Empty message ID"},
		{"   ", false, "Whitespace only"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			isValid := tt.messageid != "" && bytes.TrimSpace([]byte(tt.messageid)) != nil
			assert.Equal(t, tt.expectValid, isValid)
		})
	}
}

// ============================================================================
// 4. ChatwootProfileUpdate Tests - Picture Info
// ============================================================================

func TestN8n_GetContactPictureInfo(t *testing.T) {
	// Test: ChatwootProfileUpdate.json -> Get profile picture info
	// Endpoint: GET /picinfo/:chatid (implicit v4)
	// Auth: X-QUEPASA-TOKEN header

	mockServer := &MockServer{
		Response: map[string]interface{}{
			"success":   true,
			"picture":   "https://quepasa.server/downloads/pic_abc123",
			"source_id": "contact123",
		},
		StatusCode: http.StatusOK,
	}

	server := httptest.NewServer(mockServer)
	defer server.Close()

	req, _ := http.NewRequest("GET",
		server.URL+"/picinfo/5511988887777@s.whatsapp.net",
		nil)
	req.Header.Set("X-QUEPASA-TOKEN", "test-token-123")

	client := &http.Client{}
	resp, err := client.Do(req)

	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "GET", mockServer.LastRequest.Method)
	assert.Equal(t, "test-token-123", mockServer.LastRequest.Header.Get("X-QUEPASA-TOKEN"))

	var responseData map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&responseData)
	assert.True(t, responseData["success"].(bool))
}

// ============================================================================
// 5. Webhook Registration Tests
// ============================================================================

func TestN8n_RegisterWebhook(t *testing.T) {
	// Test: QuepasaInboxControl_typebot.json -> Register webhook
	// Endpoint: POST /webhooks (implicit v4)
	// Auth: token parameter

	mockServer := &MockServer{
		Response: map[string]interface{}{
			"success":    true,
			"webhook_id": "wh_123",
			"url":        "https://n8n.server/webhook/quepasa",
		},
		StatusCode: http.StatusOK,
	}

	server := httptest.NewServer(mockServer)
	defer server.Close()

	requestBody := map[string]interface{}{
		"url":    "https://n8n.server/webhook/quepasa",
		"events": []string{"messages", "status"},
	}

	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", server.URL+"/webhooks", bytes.NewBuffer(body))
	req.Header.Set("X-QUEPASA-TOKEN", "test-token-123")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "POST", mockServer.LastRequest.Method)
}

// ============================================================================
// 6. Authentication Tests
// ============================================================================

func TestN8n_AuthenticationTokenHeader(t *testing.T) {
	// Test: All requests should include X-QUEPASA-TOKEN header
	// or use token in custom node parameters

	mockServer := &MockServer{
		Response:   map[string]interface{}{"success": true},
		StatusCode: http.StatusOK,
	}

	server := httptest.NewServer(mockServer)
	defer server.Close()

	tests := []struct {
		description string
		setupFunc   func(*http.Request)
		expectAuth  bool
	}{
		{
			"Request with token header",
			func(r *http.Request) {
				r.Header.Set("X-QUEPASA-TOKEN", "valid-token")
			},
			true,
		},
		{
			"Request without token header",
			func(r *http.Request) {
				// No auth header
			},
			false,
		},
		{
			"Request with empty token",
			func(r *http.Request) {
				r.Header.Set("X-QUEPASA-TOKEN", "")
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			req, _ := http.NewRequest("GET", server.URL+"/messages", nil)
			tt.setupFunc(req)

			client := &http.Client{}
			resp, err := client.Do(req)
			defer resp.Body.Close()

			require.NoError(t, err)

			hasAuth := mockServer.LastRequest.Header.Get("X-QUEPASA-TOKEN") != ""
			assert.Equal(t, tt.expectAuth, hasAuth)
		})
	}
}

func TestN8n_MasterKeyAuthentication(t *testing.T) {
	// Test: Bootstrap endpoints might accept X-QUEPASA-MASTERKEY
	// Reference: docs/USAGE-authentication-modes.md

	mockServer := &MockServer{
		Response:   map[string]interface{}{"success": true},
		StatusCode: http.StatusOK,
	}

	server := httptest.NewServer(mockServer)
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL+"/system/health", nil)
	req.Header.Set("X-QUEPASA-MASTERKEY", "master-secret-key")

	client := &http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()

	require.NoError(t, err)
	assert.NotEmpty(t, mockServer.LastRequest.Header.Get("X-QUEPASA-MASTERKEY"))
}

// ============================================================================
// 7. Request/Response Format Tests
// ============================================================================

func TestN8n_RequestBodyFormat_SendText(t *testing.T) {
	// Test: Verify request body structure for send text
	tests := []struct {
		name        string
		requestBody map[string]interface{}
		expectValid bool
	}{
		{
			"Valid send text request",
			map[string]interface{}{
				"chatid": "5511988887777@s.whatsapp.net",
				"text":   "Hello World",
			},
			true,
		},
		{
			"Missing chatid",
			map[string]interface{}{
				"text": "Hello World",
			},
			false,
		},
		{
			"Missing text",
			map[string]interface{}{
				"chatid": "5511988887777@s.whatsapp.net",
			},
			false,
		},
		{
			"Empty text",
			map[string]interface{}{
				"chatid": "5511988887777@s.whatsapp.net",
				"text":   "",
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasRequired := false
			if val, ok := tt.requestBody["chatid"]; ok && val != "" {
				if val2, ok2 := tt.requestBody["text"]; ok2 && val2 != "" {
					hasRequired = true
				}
			}
			assert.Equal(t, tt.expectValid, hasRequired)
		})
	}
}

func TestN8n_ResponseFormat_StandardSuccess(t *testing.T) {
	// Test: All successful responses should follow standard format
	// { "success": true, "data": {...} } or { "success": true, ... }

	successResponse := map[string]interface{}{
		"success": true,
		"message": "Operation completed",
		"id":      "resource-123",
	}

	data, _ := json.Marshal(successResponse)
	assert.NotEmpty(t, data)

	var parsed map[string]interface{}
	json.Unmarshal(data, &parsed)
	assert.True(t, parsed["success"].(bool))
}

func TestN8n_ResponseFormat_ErrorResponse(t *testing.T) {
	// Test: Error responses should include error details

	errorResponse := map[string]interface{}{
		"success": false,
		"error":   "Invalid chatid format",
		"code":    "INVALID_REQUEST",
	}

	data, _ := json.Marshal(errorResponse)
	assert.NotEmpty(t, data)

	var parsed map[string]interface{}
	json.Unmarshal(data, &parsed)
	assert.False(t, parsed["success"].(bool))
	assert.NotEmpty(t, parsed["error"])
}

// ============================================================================
// 8. Integration Scenario Tests
// ============================================================================

func TestN8n_ScenarioQuepasaChatControl_FullFlow(t *testing.T) {
	// Test: Complete flow - Get invite link and send message
	// Simulates: QuepasaChatControl.json workflow

	mockServer := &MockServer{
		StatusCode: http.StatusOK,
	}

	server := httptest.NewServer(mockServer)
	defer server.Close()

	// Step 1: Get invite link
	mockServer.Response = map[string]interface{}{
		"success": true,
		"url":     "https://chat.whatsapp.com/ABC123",
	}

	req1, _ := http.NewRequest("GET",
		server.URL+"/control/invite?chatid=5511988887777@g.us",
		nil)
	req1.Header.Set("X-QUEPASA-TOKEN", "test-token")

	client := &http.Client{}
	resp1, err1 := client.Do(req1)
	require.NoError(t, err1)
	defer resp1.Body.Close()
	assert.Equal(t, http.StatusOK, resp1.StatusCode)

	// Step 2: Send message with invite link
	mockServer.Response = map[string]interface{}{
		"success": true,
		"id":      "msg123",
	}

	requestBody := map[string]interface{}{
		"chatid": "5511988887777@g.us",
		"text":   "Para convidar: https://chat.whatsapp.com/ABC123",
	}
	body, _ := json.Marshal(requestBody)

	req2, _ := http.NewRequest("POST", server.URL+"/messages/sendtext", bytes.NewBuffer(body))
	req2.Header.Set("X-QUEPASA-TOKEN", "test-token")
	req2.Header.Set("Content-Type", "application/json")

	resp2, err2 := client.Do(req2)
	require.NoError(t, err2)
	defer resp2.Body.Close()
	assert.Equal(t, http.StatusOK, resp2.StatusCode)
}

func TestN8n_ScenarioPostToChatwoot_FullFlow(t *testing.T) {
	// Test: Download media and post to Chatwoot
	// Simulates: PostToChatwoot.json workflow

	mockServer := &MockServer{
		StatusCode: http.StatusOK,
	}

	server := httptest.NewServer(mockServer)
	defer server.Close()

	// Step 1: Download from QuePasa
	mockServer.Response = []byte("image data")

	req1, _ := http.NewRequest("GET",
		server.URL+"/download/msg123?filename=photo.jpg",
		nil)
	req1.Header.Set("X-QUEPASA-TOKEN", "qp-token")

	client := &http.Client{}
	resp1, err1 := client.Do(req1)
	require.NoError(t, err1)
	defer resp1.Body.Close()
	assert.Equal(t, http.StatusOK, resp1.StatusCode)

	// Step 2: Get picture info
	mockServer.Response = map[string]interface{}{
		"success": true,
		"picture": "https://quepasa.server/pic.jpg",
	}

	req2, _ := http.NewRequest("GET",
		server.URL+"/picinfo/5511988887777@s.whatsapp.net",
		nil)
	req2.Header.Set("X-QUEPASA-TOKEN", "qp-token")

	resp2, err2 := client.Do(req2)
	require.NoError(t, err2)
	defer resp2.Body.Close()
	assert.Equal(t, http.StatusOK, resp2.StatusCode)
}

// ============================================================================
// 9. Error Handling Tests
// ============================================================================

func TestN8n_ErrorHandling_InvalidToken(t *testing.T) {
	// Test: Invalid or expired token should return 401
	mockServer := &MockServer{
		Response: map[string]interface{}{
			"success": false,
			"error":   "Invalid token",
		},
		StatusCode: http.StatusUnauthorized,
	}

	server := httptest.NewServer(mockServer)
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL+"/messages", nil)
	req.Header.Set("X-QUEPASA-TOKEN", "invalid-token")

	client := &http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()

	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestN8n_ErrorHandling_NotFound(t *testing.T) {
	// Test: Non-existent resources should return 404
	mockServer := &MockServer{
		Response: map[string]interface{}{
			"success": false,
			"error":   "Chat not found",
		},
		StatusCode: http.StatusNotFound,
	}

	server := httptest.NewServer(mockServer)
	defer server.Close()

	req, _ := http.NewRequest("GET",
		server.URL+"/control/invite?chatid=nonexistent@g.us",
		nil)
	req.Header.Set("X-QUEPASA-TOKEN", "test-token")

	client := &http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()

	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestN8n_ErrorHandling_ServerError(t *testing.T) {
	// Test: Server errors should return 500
	mockServer := &MockServer{
		Response: map[string]interface{}{
			"success": false,
			"error":   "Internal server error",
		},
		StatusCode: http.StatusInternalServerError,
	}

	server := httptest.NewServer(mockServer)
	defer server.Close()

	req, _ := http.NewRequest("POST", server.URL+"/messages/sendtext", nil)
	req.Header.Set("X-QUEPASA-TOKEN", "test-token")

	client := &http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()

	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

// ============================================================================
// 10. Edge Cases and Special Scenarios
// ============================================================================

func TestN8n_SpecialCharactersInText(t *testing.T) {
	// Test: Text messages with special characters, emojis, line breaks
	tests := []struct {
		text        string
		description string
	}{
		{"Hello 👋 World", "With emoji"},
		{"Line 1\nLine 2\nLine 3", "With line breaks"},
		{"Special chars: !@#$%^&*()", "With special characters"},
		{"Unicode: 你好世界", "With Unicode characters"},
		{"Very long message " + fmt.Sprintf("%0100d", 0), "Long message"},
		{"", "Empty text should fail"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			requestBody := map[string]interface{}{
				"chatid": "5511988887777@s.whatsapp.net",
				"text":   tt.text,
			}

			body, err := json.Marshal(requestBody)
			require.NoError(t, err)

			isValid := len(tt.text) > 0 && len(body) > 0
			if tt.description == "Empty text should fail" {
				assert.False(t, isValid)
			} else {
				assert.True(t, isValid)
			}
		})
	}
}

func TestN8n_MultipleAuthenticationMethods(t *testing.T) {
	// Test: Verify correct auth method precedence
	// X-QUEPASA-TOKEN > X-QUEPASA-MASTERKEY > query param > body

	tests := []struct {
		name           string
		headerToken    string
		headerMaster   string
		expectPriority string
	}{
		{
			"Both header token and master key - token wins",
			"session-token",
			"master-key",
			"session-token",
		},
		{
			"Only master key header",
			"",
			"master-key",
			"master-key",
		},
		{
			"Only session token header",
			"session-token",
			"",
			"session-token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "http://localhost/messages", nil)

			if tt.headerToken != "" {
				req.Header.Set("X-QUEPASA-TOKEN", tt.headerToken)
			}
			if tt.headerMaster != "" {
				req.Header.Set("X-QUEPASA-MASTERKEY", tt.headerMaster)
			}

			// In real scenario, API would select token over master key
			priority := ""
			if token := req.Header.Get("X-QUEPASA-TOKEN"); token != "" {
				priority = token
			} else if master := req.Header.Get("X-QUEPASA-MASTERKEY"); master != "" {
				priority = master
			}

			assert.Equal(t, tt.expectPriority, priority)
		})
	}
}

func TestN8n_RateLimiting(t *testing.T) {
	// Test: Rapid requests should respect rate limits
	// This is a conceptual test - actual rate limiting is server-side

	mockServer := &MockServer{
		StatusCode: http.StatusOK,
	}

	server := httptest.NewServer(mockServer)
	defer server.Close()

	client := &http.Client{}

	successCount := 0
	for i := 0; i < 5; i++ {
		req, _ := http.NewRequest("GET", server.URL+"/messages", nil)
		req.Header.Set("X-QUEPASA-TOKEN", "test-token")

		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			successCount++
		}
		if resp != nil {
			resp.Body.Close()
		}
	}

	// All requests should succeed in test environment
	assert.GreaterOrEqual(t, successCount, 1)
}
