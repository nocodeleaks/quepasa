package mcp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	environment "github.com/nocodeleaks/quepasa/environment"
	models "github.com/nocodeleaks/quepasa/models"
	log "github.com/sirupsen/logrus"
)

// MCPServer represents the Model Context Protocol server
type MCPServer struct {
	enabled  bool
	path     string
	registry *MCPToolRegistry
}

// NewMCPServer creates a new MCP server instance
func NewMCPServer() *MCPServer {
	enabled := environment.Settings.MCP.Enabled
	path := environment.Settings.MCP.Path
	if path == "" {
		path = "/mcp"
	}

	server := &MCPServer{
		enabled:  enabled,
		path:     path,
		registry: NewMCPToolRegistry(),
	}

	log.Infof("MCP Server initialized: enabled=%v, path=%s", enabled, path)
	return server
}

// IsEnabled returns whether MCP server is enabled
func (s *MCPServer) IsEnabled() bool {
	return s.enabled
}

// GetPath returns the MCP endpoint path
func (s *MCPServer) GetPath() string {
	return s.path
}

// RegisterTools registers all available MCP tools
func (s *MCPServer) RegisterTools(server *models.QpWhatsappServer) {
	// Register health tool
	healthTool := &HealthTool{server: server}
	s.registry.Register(healthTool)

	log.Debugf("Registered MCP tools: %d", len(s.registry.List()))
}

// MCPRequest represents an incoming MCP request
type MCPRequest struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
}

// MCPResponse represents an MCP response
type MCPResponse struct {
	Result interface{} `json:"result,omitempty"`
	Error  *MCPError   `json:"error,omitempty"`
}

// MCPError represents an MCP error response
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// HandleRequest handles MCP protocol requests
func (s *MCPServer) HandleRequest(w http.ResponseWriter, r *http.Request) {
	if !s.enabled {
		http.Error(w, "MCP server is disabled", http.StatusServiceUnavailable)
		return
	}

	// Authenticate request
	server, authLevel, err := s.authenticate(r)
	if err != nil {
		s.sendError(w, 401, fmt.Sprintf("Authentication failed: %v", err))
		return
	}

	log.Debugf("MCP request authenticated: level=%s", authLevel)

	// Register tools for the authenticated server
	if server != nil {
		s.RegisterTools(server)
	}

	// Parse request
	var req MCPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, 400, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	// Handle method
	switch req.Method {
	case "tools/list":
		s.handleToolsList(w)
	case "tools/call":
		s.handleToolCall(w, req.Params)
	default:
		s.sendError(w, 404, fmt.Sprintf("Unknown method: %s", req.Method))
	}
}

// authenticate authenticates the request and returns the server and auth level
func (s *MCPServer) authenticate(r *http.Request) (*models.QpWhatsappServer, string, error) {
	// Check for master key
	masterKey := r.Header.Get("X-QUEPASA-MASTERKEY")
	if masterKey != "" && masterKey == environment.Settings.API.MasterKey {
		return nil, "master", nil
	}

	// Check for bot token
	token := r.Header.Get("X-QUEPASA-TOKEN")
	if token == "" {
		return nil, "", fmt.Errorf("no authentication provided")
	}

	// Get server by token
	server, ok := models.WhatsappService.Servers[token]
	if !ok {
		return nil, "", fmt.Errorf("invalid token")
	}

	return server, "bot", nil
}

// handleToolsList returns the list of available tools
func (s *MCPServer) handleToolsList(w http.ResponseWriter) {
	tools := s.registry.List()
	toolInfos := make([]map[string]interface{}, 0, len(tools))

	for _, tool := range tools {
		toolInfos = append(toolInfos, map[string]interface{}{
			"name":        tool.Name(),
			"description": tool.Description(),
			"inputSchema": tool.InputSchema(),
		})
	}

	s.sendResponse(w, map[string]interface{}{
		"tools": toolInfos,
	})
}

// handleToolCall executes a tool call
func (s *MCPServer) handleToolCall(w http.ResponseWriter, params json.RawMessage) {
	var callParams struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}

	if err := json.Unmarshal(params, &callParams); err != nil {
		s.sendError(w, 400, fmt.Sprintf("Invalid parameters: %v", err))
		return
	}

	// Get tool
	tool, exists := s.registry.Get(callParams.Name)
	if !exists {
		s.sendError(w, 404, fmt.Sprintf("Tool not found: %s", callParams.Name))
		return
	}

	// Execute tool
	result, err := tool.Execute(callParams.Arguments)
	if err != nil {
		s.sendError(w, 500, fmt.Sprintf("Tool execution failed: %v", err))
		return
	}

	s.sendResponse(w, map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": fmt.Sprintf("%v", result),
			},
		},
	})
}

// sendResponse sends a successful MCP response
func (s *MCPServer) sendResponse(w http.ResponseWriter, result interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&MCPResponse{
		Result: result,
	})
}

// sendError sends an MCP error response
func (s *MCPServer) sendError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // MCP uses 200 with error object
	json.NewEncoder(w).Encode(&MCPResponse{
		Error: &MCPError{
			Code:    code,
			Message: message,
		},
	})
}

// RegisterRoutes registers MCP routes with the HTTP router
func RegisterMCPRoutes(mux *http.ServeMux, mcpServer *MCPServer) {
	if !mcpServer.IsEnabled() {
		log.Info("MCP server is disabled, skipping route registration")
		return
	}

	path := mcpServer.GetPath()
	
	// Remove trailing slash if present
	path = strings.TrimSuffix(path, "/")
	
	mux.HandleFunc(path, mcpServer.HandleRequest)
	mux.HandleFunc(path+"/", mcpServer.HandleRequest)

	log.Infof("MCP routes registered at: %s", path)
}
