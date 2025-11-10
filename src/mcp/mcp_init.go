package mcp

import (
	"github.com/go-chi/chi/v5"
	webserver "github.com/nocodeleaks/quepasa/webserver"
	log "github.com/sirupsen/logrus"
)

var mcpServerInstance *MCPServer

func init() {
	log.Info(">>> MCP PACKAGE LOADED <<<")

	// Initialize MCP server
	mcpServerInstance = NewMCPServer()

	log.Infof("MCP server created, enabled=%v", mcpServerInstance.IsEnabled())

	// Register all tools once at startup
	if mcpServerInstance.IsEnabled() {
		log.Info("=== MCP SERVER INITIALIZATION ===")
		mcpServerInstance.RegisterTools()
		log.Info("=== MCP SERVER READY ===")
	} else {
		log.Warn("MCP server is DISABLED (set MCP_ENABLED=true to enable)")
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
