package mcp

import (
	log "github.com/sirupsen/logrus"
)

// RegisterAPITools registers all API endpoints as MCP tools using auto-discovery
func (s *MCPServer) RegisterAPITools() {
	log.Info("MCP: Starting auto-discovery of API endpoints...")

	// Get path to API controllers directory - try multiple possible paths
	apiDir := "api"
	log.Infof("MCP: Scanning directory: %s", apiDir)

	// Parse Swagger annotations from all API controllers
	endpoints, err := ParseSwaggerAnnotations(apiDir)
	if err != nil {
		log.Errorf("MCP: Failed to parse Swagger annotations: %v", err)

		// Try alternative path
		apiDir = "./api"
		log.Infof("MCP: Retrying with path: %s", apiDir)
		endpoints, err = ParseSwaggerAnnotations(apiDir)
		if err != nil {
			log.Errorf("MCP: Failed again with alternative path: %v", err)
			return
		}
	}

	log.Infof("MCP: Found %d endpoints with Swagger annotations", len(endpoints))

	// Register each discovered endpoint as an MCP tool
	registered := 0
	skipped := 0

	for _, endpoint := range endpoints {
		// Skip if marked with @MCPHidden
		if endpoint.MCPHidden {
			log.Debugf("MCP: Skipping hidden endpoint: %s (%s)", endpoint.FuncName, endpoint.Path)
			skipped++
			continue
		}

		// Skip if no valid route
		if endpoint.Path == "" || endpoint.Method == "" {
			log.Debugf("MCP: Skipping endpoint without route: %s", endpoint.FuncName)
			skipped++
			continue
		}

		// Get handler function by name
		handler, err := GetHandlerByName(endpoint.FuncName)
		if err != nil {
			log.Warnf("MCP: Handler not found for %s: %v", endpoint.FuncName, err)
			skipped++
			continue
		}

		// Generate tool name
		toolName := GenerateToolName(endpoint)
		if toolName == "" {
			log.Warnf("MCP: Could not generate tool name for %s", endpoint.FuncName)
			skipped++
			continue
		}

		// Generate input schema from parameters
		inputSchema := GenerateInputSchema(endpoint)

		// Create and register the tool
		tool := &APIHandlerTool{
			name:        toolName,
			description: endpoint.Description,
			method:      endpoint.Method,
			path:        endpoint.Path,
			handler:     handler,
			inputSchema: inputSchema,
		}

		s.registry.Register(tool)
		registered++

		log.Debugf("MCP: Registered tool '%s' -> %s %s (%s)",
			toolName, endpoint.Method, endpoint.Path, endpoint.FuncName)
	}

	log.Infof("MCP: Auto-discovery complete. Registered: %d tools, Skipped: %d endpoints",
		registered, skipped)
}
