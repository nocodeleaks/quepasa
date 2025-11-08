package mcp

import (
	"encoding/json"
)

// MCPTool represents a generic MCP tool interface
type MCPTool interface {
	Execute(params json.RawMessage) (interface{}, error)
	Name() string
	Description() string
	InputSchema() map[string]interface{}
}

// MCPToolRegistry manages available MCP tools
type MCPToolRegistry struct {
	tools map[string]MCPTool
}

// NewMCPToolRegistry creates a new tool registry
func NewMCPToolRegistry() *MCPToolRegistry {
	return &MCPToolRegistry{
		tools: make(map[string]MCPTool),
	}
}

// Register adds a tool to the registry
func (r *MCPToolRegistry) Register(tool MCPTool) {
	r.tools[tool.Name()] = tool
}

// Get retrieves a tool by name
func (r *MCPToolRegistry) Get(name string) (MCPTool, bool) {
	tool, exists := r.tools[name]
	return tool, exists
}

// List returns all available tools
func (r *MCPToolRegistry) List() []MCPTool {
	tools := make([]MCPTool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}
