package environment

// MCP environment variable names
const (
	ENV_MCP_ENABLED = "MCP_ENABLED" // enable/disable MCP server
	ENV_MCP_PATH    = "MCP_PATH"    // MCP endpoint path
)

// MCPSettings holds all MCP configuration loaded from environment
type MCPSettings struct {
	Enabled bool   `json:"enabled"`
	Path    string `json:"path"`
}

// NewMCPSettings creates a new MCP settings by loading all values from environment
func NewMCPSettings() MCPSettings {
	return MCPSettings{
		Enabled: getEnvOrDefaultBool(ENV_MCP_ENABLED, false),
		Path:    getEnvOrDefaultString(ENV_MCP_PATH, "/mcp"),
	}
}
