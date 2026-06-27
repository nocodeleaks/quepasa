package mcp

import (
	"encoding/json"
	"testing"
)

type MockTool struct {
	name        string
	description string
	schema      map[string]interface{}
	executeFunc func(ctx *MCPToolContext, params json.RawMessage) (interface{}, error)
}

func (m *MockTool) Name() string {
	return m.name
}

func (m *MockTool) Description() string {
	return m.description
}

func (m *MockTool) InputSchema() map[string]interface{} {
	return m.schema
}

func (m *MockTool) ExecuteWithContext(ctx *MCPToolContext, params json.RawMessage) (interface{}, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, params)
	}
	return map[string]interface{}{"status": "ok"}, nil
}

func TestNewMCPToolRegistry(t *testing.T) {
	registry := NewMCPToolRegistry()
	if registry == nil {
		t.Fatal("Expected registry to be created")
	}
	if len(registry.tools) != 0 {
		t.Errorf("Expected empty registry, got %d tools", len(registry.tools))
	}
}

func TestRegisterTool(t *testing.T) {
	registry := NewMCPToolRegistry()
	tool := &MockTool{
		name:        "test_tool",
		description: "Test tool",
		schema:      map[string]interface{}{"type": "object"},
	}
	registry.Register(tool)
	if len(registry.tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(registry.tools))
	}
	retrievedTool, exists := registry.Get("test_tool")
	if !exists {
		t.Error("Expected tool to exist")
	}
	if retrievedTool.Name() != "test_tool" {
		t.Errorf("Expected name 'test_tool', got '%s'", retrievedTool.Name())
	}
}

func TestListTools(t *testing.T) {
	registry := NewMCPToolRegistry()
	tool1 := &MockTool{name: "tool1", description: "Tool 1"}
	tool2 := &MockTool{name: "tool2", description: "Tool 2"}
	registry.Register(tool1)
	registry.Register(tool2)
	tools := registry.List()
	if len(tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(tools))
	}
}

func BenchmarkRegisterTool(b *testing.B) {
	registry := NewMCPToolRegistry()
	tool := &MockTool{name: "bench_tool", description: "Bench"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registry.Register(tool)
	}
}

func BenchmarkGetTool(b *testing.B) {
	registry := NewMCPToolRegistry()
	tool := &MockTool{name: "bench_tool"}
	registry.Register(tool)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = registry.Get("bench_tool")
	}
}
