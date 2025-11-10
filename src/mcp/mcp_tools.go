package mcp

import (
	"encoding/json"

	models "github.com/nocodeleaks/quepasa/models"
)

// MCPToolContext holds the authentication context for tool execution
type MCPToolContext struct {
	Server   *models.QpWhatsappServer // nil for master key, specific server for bot token
	IsMaster bool                     // true if authenticated with master key
}

// MCPTool represents a tool that requires authentication context
type MCPTool interface {
	ExecuteWithContext(ctx *MCPToolContext, params json.RawMessage) (interface{}, error)
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
