package mcp

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewMCPServer(t *testing.T) {
	server := &MCPServer{
		enabled:  true,
		path:     "/mcp",
		registry: NewMCPToolRegistry(),
	}
	if !server.IsEnabled() {
		t.Error("Expected server to be enabled")
	}
	if server.GetPath() != "/mcp" {
		t.Errorf("Expected path '/mcp', got '%s'", server.GetPath())
	}
}

func TestHandleRequestDisabled(t *testing.T) {
	server := &MCPServer{
		enabled:  false,
		path:     "/mcp",
		registry: NewMCPToolRegistry(),
	}
	req := httptest.NewRequest("POST", "/mcp", bytes.NewBufferString("{}"))
	w := httptest.NewRecorder()
	server.HandleRequest(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", w.Code)
	}
}

func TestHandleRequestNoAuth(t *testing.T) {
	server := &MCPServer{
		enabled:  true,
		path:     "/mcp",
		registry: NewMCPToolRegistry(),
	}
	reqBody := `{"jsonrpc":"2.0","method":"tools/list","id":1}`
	req := httptest.NewRequest("POST", "/mcp", bytes.NewBufferString(reqBody))
	w := httptest.NewRecorder()
	server.HandleRequest(w, req)
	var response MCPResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if response.Error == nil {
		t.Error("Expected error in response")
	}
}

func TestSendError(t *testing.T) {
	server := &MCPServer{
		enabled:  true,
		path:     "/mcp",
		registry: NewMCPToolRegistry(),
	}
	id := 42
	w := httptest.NewRecorder()
	server.sendError(w, &id, 404, "Test error message")
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	var response MCPResponse
	json.NewDecoder(w.Body).Decode(&response)
	if response.Error == nil {
		t.Fatal("Expected error in response")
	}
	if response.Error.Code != 404 {
		t.Errorf("Expected error code 404, got %d", response.Error.Code)
	}
}

func TestAuthenticateNoHeader(t *testing.T) {
	server := &MCPServer{
		enabled:  true,
		path:     "/mcp",
		registry: NewMCPToolRegistry(),
	}
	req := httptest.NewRequest("POST", "/mcp", nil)
	_, _, err := server.authenticate(req)
	if err == nil {
		t.Error("Expected error when no Authorization header provided")
	}
	if !strings.Contains(err.Error(), "no authentication provided") {
		t.Errorf("Expected 'no authentication provided' error, got: %v", err)
	}
}

func BenchmarkHandleRequest(b *testing.B) {
	server := &MCPServer{
		enabled:  true,
		path:     "/mcp",
		registry: NewMCPToolRegistry(),
	}
	server.registry.Register(&HealthTool{})
	reqBody := `{"jsonrpc":"2.0","method":"tools/list","id":1}`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/mcp", bytes.NewBufferString(reqBody))
		req.Header.Set("Authorization", "Bearer test-token")
		w := httptest.NewRecorder()
		server.HandleRequest(w, req)
	}
}
