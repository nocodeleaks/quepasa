package mcp

import (
	"github.com/go-chi/chi/v5"
	webserver "github.com/nocodeleaks/quepasa/webserver"
	log "github.com/sirupsen/logrus"
)

var mcpServerInstance *MCPServer

func init() {
	// Initialize MCP server
	mcpServerInstance = NewMCPServer()

	// Register all tools once at startup
	if mcpServerInstance.IsEnabled() {
		mcpServerInstance.RegisterTools()
	}

	// Register MCP routes with the web server
	webserver.RegisterRouterConfigurator(func(r chi.Router) {
		if mcpServerInstance.IsEnabled() {
			path := mcpServerInstance.GetPath()

			log.Infof("Registering MCP routes at: %s", path)

			// Support both POST (JSON-RPC) and GET (SSE) methods
			r.Post(path, mcpServerInstance.HandleRequest)
			r.Post(path+"/", mcpServerInstance.HandleRequest)
			r.Get(path, mcpServerInstance.HandleSSE)
			r.Get(path+"/", mcpServerInstance.HandleSSE)
		} else {
			log.Info("MCP server is disabled, skipping route registration")
		}
	})
}

// GetMCPServer returns the global MCP server instance
func GetMCPServer() *MCPServer {
	return mcpServerInstance
}
