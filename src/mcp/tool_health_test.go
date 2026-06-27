package mcp

import (
	"encoding/json"
	"testing"

	models "github.com/nocodeleaks/quepasa/models"
)

func TestHealthToolName(t *testing.T) {
	tool := &HealthTool{}
	if tool.Name() != "health" {
		t.Errorf("Expected name 'health', got '%s'", tool.Name())
	}
}

func TestHealthToolDescription(t *testing.T) {
	tool := &HealthTool{}
	desc := tool.Description()
	if desc == "" {
		t.Error("Expected non-empty description")
	}
}

func TestHealthToolInputSchema(t *testing.T) {
	tool := &HealthTool{}
	schema := tool.InputSchema()
	if schema == nil {
		t.Fatal("Expected schema, got nil")
	}
	if schema["type"] != "object" {
		t.Errorf("Expected type 'object', got '%v'", schema["type"])
	}
}

func TestHealthToolExecuteWithMasterKey(t *testing.T) {
	tool := &HealthTool{}
	ctx := &MCPToolContext{
		Server:   nil,
		IsMaster: true,
	}
	params := json.RawMessage(`{}`)
	result, err := tool.ExecuteWithContext(ctx, params)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	resultMap, ok := result.(map[string]interface{})
	if ok {
		if resultMap["access_level"] != "master" {
			t.Errorf("Expected access_level 'master', got '%v'", resultMap["access_level"])
		}
	}
}

func TestHealthToolExecuteWithBotToken(t *testing.T) {
	tool := &HealthTool{}
	mockServer := &models.QpWhatsappServer{
		Token: "test-bot-token",
	}
	ctx := &MCPToolContext{
		Server:   mockServer,
		IsMaster: false,
	}
	params := json.RawMessage(`{}`)
	result, err := tool.ExecuteWithContext(ctx, params)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
}

func BenchmarkHealthToolExecute(b *testing.B) {
	tool := &HealthTool{}
	ctx := &MCPToolContext{
		Server:   nil,
		IsMaster: true,
	}
	params := json.RawMessage(`{}`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tool.ExecuteWithContext(ctx, params)
	}
}
