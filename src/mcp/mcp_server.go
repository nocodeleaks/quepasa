package mcp

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

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

// GetRegistry returns the tool registry
func (s *MCPServer) GetRegistry() *MCPToolRegistry {
	return s.registry
}

// RegisterTools registers all available MCP tools (called once at startup)
func (s *MCPServer) RegisterTools() {
	log.Info("MCP: Registering tools...")

	// Register health tool
	s.registry.Register(&HealthTool{})
	log.Debug("MCP: Registered HealthTool")

	// Register list servers tool
	s.registry.Register(&ListServersTool{})
	log.Debug("MCP: Registered ListServersTool")

	// Register API endpoint tools
	s.RegisterAPITools()

	totalTools := len(s.registry.List())
	log.Infof("MCP: === TOTAL TOOLS REGISTERED: %d ===", totalTools)
}

// MCPRequest represents an incoming MCP request
type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int            `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// MCPResponse represents an MCP response
type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      *int        `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

// MCPError represents an MCP error response
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// HandleRequest handles MCP protocol requests (JSON-RPC over POST)
func (s *MCPServer) HandleRequest(w http.ResponseWriter, r *http.Request) {
	if !s.enabled {
		http.Error(w, "MCP server is disabled", http.StatusServiceUnavailable)
		return
	}

	// Authenticate request
	server, authLevel, err := s.authenticate(r)
	if err != nil {
		s.sendError(w, nil, 401, fmt.Sprintf("Authentication failed: %v", err))
		return
	}

	// Parse request
	var req MCPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, nil, 400, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	log.Infof("MCP: %s (auth=%s)", req.Method, authLevel)

	// Create authentication context for tool execution
	ctx := &MCPToolContext{
		Server:   server,
		IsMaster: (authLevel == "master"),
	}

	// Handle method
	switch req.Method {
	case "initialize":
		s.handleInitialize(w, req.ID, req.Params)
	case "tools/list":
		s.handleToolsList(w, req.ID)
	case "tools/call":
		s.handleToolCall(w, req.ID, req.Params, ctx)
	case "notifications/initialized":
		// Client confirms initialization - no response needed
		return
	default:
		s.sendError(w, req.ID, 404, fmt.Sprintf("Unknown method: %s", req.Method))
	}
}

// HandleSSE handles MCP protocol over Server-Sent Events (SSE)
func (s *MCPServer) HandleSSE(w http.ResponseWriter, r *http.Request) {
	if !s.enabled {
		log.Warn("MCP SSE rejected: server is disabled")
		http.Error(w, "MCP server is disabled", http.StatusServiceUnavailable)
		return
	}

	// Generate unique connection ID
	rand.Seed(time.Now().UnixNano())
	connID := fmt.Sprintf("%04x", rand.Intn(0xFFFF))

	// Authenticate request
	_, authLevel, err := s.authenticate(r)
	if err != nil {
		log.Warnf("MCP SSE [%s] authentication failed: %v", connID, err)
		http.Error(w, fmt.Sprintf("Authentication failed: %v", err), http.StatusUnauthorized)
		return
	}

	log.Infof("MCP SSE [%s] connection established: level=%s", connID, authLevel)

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}
	flusher.Flush()

	log.Debugf("MCP SSE [%s] ready, waiting for messages...", connID)

	// Keep connection alive - messages will come via POST
	<-r.Context().Done()
	log.Infof("MCP SSE [%s] connection closed", connID)
}

// authenticate authenticates the request and returns the server and auth level
func (s *MCPServer) authenticate(r *http.Request) (*models.QpWhatsappServer, string, error) {
	// Check Authorization header (Bearer token)
	authHeader := r.Header.Get("Authorization")
	log.Debugf("MCP auth: Authorization header present=%v", authHeader != "")

	if authHeader == "" {
		log.Warn("MCP auth: no Authorization header provided")
		return nil, "", fmt.Errorf("no authentication provided, expected: Authorization: Bearer <token>")
	}

	// Extract Bearer token
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		log.Warn("MCP auth: Authorization header without Bearer prefix")
		return nil, "", fmt.Errorf("invalid authorization format, expected: Bearer <token>")
	}

	token := strings.TrimPrefix(authHeader, bearerPrefix)
	token = strings.TrimSpace(token)

	// PRIORITY 1: Check if token matches MASTERKEY (super user - full access)
	expectedMasterKey := environment.Settings.API.MasterKey

	if token == expectedMasterKey && expectedMasterKey != "" {
		log.Debug("MCP auth: MASTER KEY authenticated - SUPER USER with full access to all servers")
		return nil, "master", nil
	}

	// PRIORITY 2: Check if token matches a specific server token (limited access)
	log.Debugf("MCP auth: checking if token is a server identifier (total_servers=%d)", len(models.WhatsappService.Servers))
	server, ok := models.WhatsappService.Servers[token]
	if ok {
		log.Debugf("MCP auth: SERVER TOKEN authenticated - limited access to server: %s", server.Token)
		return server, "bot", nil
	}

	// No match found
	log.Warnf("MCP auth: token not recognized as master key or server identifier (token_prefix=%s...)", token[:min(len(token), 10)])
	return nil, "", fmt.Errorf("invalid token - must be either MASTERKEY or a valid server token")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// handleToolsList returns the list of available tools
func (s *MCPServer) handleToolsList(w http.ResponseWriter, id *int) {
	tools := s.registry.List()
	log.Infof("MCP: tools/list returning %d tools", len(tools))

	toolInfos := make([]map[string]interface{}, 0, len(tools))

	for _, tool := range tools {
		toolInfo := map[string]interface{}{
			"name":        tool.Name(),
			"description": tool.Description(),
			"inputSchema": tool.InputSchema(),
		}
		toolInfos = append(toolInfos, toolInfo)
		log.Debugf("MCP: tool=%s, desc=%s", tool.Name(), tool.Description())
	}

	log.Infof("MCP: Sending response with %d tools", len(toolInfos))

	response := map[string]interface{}{
		"tools": toolInfos,
	}

	// Log first 2 tools for debugging
	if len(toolInfos) > 0 {
		log.Debugf("MCP: First tool example: %+v", toolInfos[0])
	}

	s.sendResponse(w, id, response)
}

// handleInitialize handles the initialize method
func (s *MCPServer) handleInitialize(w http.ResponseWriter, id *int, params json.RawMessage) {
	// Parse client protocol version if provided
	var initParams struct {
		ProtocolVersion string `json:"protocolVersion"`
		ClientInfo      struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"clientInfo"`
	}

	_ = json.Unmarshal(params, &initParams)

	// Support multiple protocol versions
	supportedVersions := []string{"2024-11-05", "2025-03-26", "2025-06-18"}
	negotiatedVersion := "2024-11-05" // Default

	// If client specified version, use it if supported
	if initParams.ProtocolVersion != "" {
		for _, v := range supportedVersions {
			if v == initParams.ProtocolVersion {
				negotiatedVersion = v
				break
			}
		}
	}

	log.Infof("MCP initialize: client=%s/%s, protocol=%s",
		initParams.ClientInfo.Name, initParams.ClientInfo.Version, negotiatedVersion)

	s.sendResponse(w, id, map[string]interface{}{
		"protocolVersion": negotiatedVersion,
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{
				"listChanged": false,
			},
			"resources": map[string]interface{}{
				"subscribe":   false,
				"listChanged": false,
			},
			"prompts": map[string]interface{}{
				"listChanged": false,
			},
			"logging": map[string]interface{}{},
		},
		"serverInfo": map[string]interface{}{
			"name":    "QuePasa MCP Server",
			"version": "1.0.0",
		},
	})
}

// handleToolsList returns the list of available tools
func (s *MCPServer) handleToolsListOld(w http.ResponseWriter) {
	tools := s.registry.List()
	toolInfos := make([]map[string]interface{}, 0, len(tools))

	for _, tool := range tools {
		toolInfos = append(toolInfos, map[string]interface{}{
			"name":        tool.Name(),
			"description": tool.Description(),
			"inputSchema": tool.InputSchema(),
		})
	}

	s.sendResponse(w, nil, map[string]interface{}{
		"tools": toolInfos,
	})
}

// handleToolCall executes a tool call
func (s *MCPServer) handleToolCall(w http.ResponseWriter, id *int, params json.RawMessage, ctx *MCPToolContext) {
	var callParams struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}

	if err := json.Unmarshal(params, &callParams); err != nil {
		s.sendError(w, id, 400, fmt.Sprintf("Invalid parameters: %v", err))
		return
	}

	// Get tool
	tool, exists := s.registry.Get(callParams.Name)
	if !exists {
		s.sendError(w, id, 404, fmt.Sprintf("Tool not found: %s", callParams.Name))
		return
	}

	// Execute tool with context
	result, err := tool.ExecuteWithContext(ctx, callParams.Arguments)
	if err != nil {
		s.sendError(w, id, 500, fmt.Sprintf("Tool execution failed: %v", err))
		return
	}

	// Serialize result to JSON string
	resultJSON, err := json.Marshal(result)
	if err != nil {
		s.sendError(w, id, 500, fmt.Sprintf("Failed to serialize result: %v", err))
		return
	}

	s.sendResponse(w, id, map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": string(resultJSON),
			},
		},
	})
}

// sendResponse sends a successful MCP response
func (s *MCPServer) sendResponse(w http.ResponseWriter, id *int, result interface{}) {
	w.Header().Set("Content-Type", "application/json")

	response := &MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}

	// Log response for debugging
	if jsonBytes, err := json.Marshal(response); err == nil {
		if len(jsonBytes) > 500 {
			log.Debugf("MCP: Response JSON (truncated): %s...", string(jsonBytes[:500]))
		} else {
			log.Debugf("MCP: Response JSON: %s", string(jsonBytes))
		}
	}

	json.NewEncoder(w).Encode(response)
}

// sendError sends an MCP error response
func (s *MCPServer) sendError(w http.ResponseWriter, id *int, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // MCP uses 200 with error object
	json.NewEncoder(w).Encode(&MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
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
