package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	environment "github.com/nocodeleaks/quepasa/environment"
	log "github.com/sirupsen/logrus"
)

// APIHandlerTool wraps a standard API handler as an MCP tool
type APIHandlerTool struct {
	name        string
	description string
	method      string // GET, POST, PUT, DELETE, PATCH
	path        string
	handler     http.HandlerFunc
	inputSchema map[string]interface{}
}

// NewAPIHandlerTool creates a new API handler tool
func NewAPIHandlerTool(name, description, method, path string, handler http.HandlerFunc, schema map[string]interface{}) *APIHandlerTool {
	return &APIHandlerTool{
		name:        name,
		description: description,
		method:      strings.ToUpper(method),
		path:        path,
		handler:     handler,
		inputSchema: schema,
	}
}

// ExecuteWithContext runs the API handler with authentication context
func (t *APIHandlerTool) ExecuteWithContext(ctx *MCPToolContext, params json.RawMessage) (interface{}, error) {
	// Parse parameters
	var paramsMap map[string]interface{}
	if len(params) > 0 {
		if err := json.Unmarshal(params, &paramsMap); err != nil {
			return nil, fmt.Errorf("invalid parameters: %v", err)
		}
	} else {
		paramsMap = make(map[string]interface{})
	}

	// Determine authentication token to use
	var authToken string
	var useMasterKey bool

	if ctx.Server != nil {
		// Bot token from authentication context
		authToken = ctx.Server.Token
		useMasterKey = false
	} else if ctx.IsMaster {
		// Master key authentication - check if token parameter was provided
		if tokenParam, ok := paramsMap["token"].(string); ok && tokenParam != "" {
			// User provided specific server token as parameter
			authToken = tokenParam
			useMasterKey = true
			log.Debugf("MCP API Tool: Master key with target server token: %s", authToken)
		} else {
			// Master key without server token - only valid for global endpoints
			authToken = ""
			useMasterKey = true
			log.Debug("MCP API Tool: Master key without specific server token")
		}
	}

	// Build request path with path parameters
	requestPath := t.path
	bodyParams := make(map[string]interface{})

	// Process path parameters (e.g., {messageId}, {chatId}, {phone})
	for key, value := range paramsMap {
		// Skip 'token' parameter - it's used for authentication, not as path/body param
		if key == "token" {
			continue
		}

		placeholder := fmt.Sprintf("{%s}", key)
		if strings.Contains(requestPath, placeholder) {
			requestPath = strings.ReplaceAll(requestPath, placeholder, fmt.Sprintf("%v", value))
		} else {
			bodyParams[key] = value
		}
	}

	log.Debugf("MCP API Tool: %s %s", t.method, requestPath)

	// Create request body
	var bodyReader io.Reader
	if len(bodyParams) > 0 && (t.method == "POST" || t.method == "PUT" || t.method == "PATCH") {
		bodyJSON, err := json.Marshal(bodyParams)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %v", err)
		}
		bodyReader = bytes.NewReader(bodyJSON)
		log.Debugf("MCP API Tool: Request body: %s", string(bodyJSON))
	}

	// Create HTTP request
	req := httptest.NewRequest(t.method, requestPath, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Add authentication headers
	if useMasterKey {
		// Master key authentication
		masterKey := environment.Settings.API.MasterKey
		req.Header.Set("X-QUEPASA-MASTERKEY", masterKey)
		log.Debug("MCP API Tool: Using MASTER KEY authentication")

		// If master key provided a specific server token, add it too
		if authToken != "" {
			req.Header.Set("X-QUEPASA-TOKEN", authToken)
			log.Debugf("MCP API Tool: With target SERVER TOKEN: %s", authToken)
		}
	} else if authToken != "" {
		// Direct bot token authentication
		req.Header.Set("X-QUEPASA-TOKEN", authToken)
		log.Debugf("MCP API Tool: Using SERVER TOKEN authentication: %s", authToken)
	} else {
		return nil, fmt.Errorf("no authentication context available")
	}

	// Create response recorder
	rec := httptest.NewRecorder()

	// Execute handler
	t.handler(rec, req)

	// Get response
	result := rec.Result()
	defer result.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	log.Debugf("MCP API Tool: Response status: %d, body length: %d", result.StatusCode, len(responseBody))

	// Check for errors
	if result.StatusCode >= 400 {
		// Try to parse error message
		var errorData map[string]interface{}
		if json.Unmarshal(responseBody, &errorData) == nil {
			return nil, fmt.Errorf("API error (%d): %v", result.StatusCode, errorData)
		}
		return nil, fmt.Errorf("API error (%d): %s", result.StatusCode, string(responseBody))
	}

	// Parse successful response
	if len(responseBody) == 0 {
		return map[string]interface{}{
			"status": "success",
			"code":   result.StatusCode,
		}, nil
	}

	var responseData interface{}
	if err := json.Unmarshal(responseBody, &responseData); err != nil {
		// If not JSON, return as text
		return map[string]interface{}{
			"status": "success",
			"code":   result.StatusCode,
			"data":   string(responseBody),
		}, nil
	}

	return responseData, nil
}

// Name returns the tool name
func (t *APIHandlerTool) Name() string {
	return t.name
}

// Description returns the tool description
func (t *APIHandlerTool) Description() string {
	return t.description
}

// InputSchema returns the JSON schema for the tool input
func (t *APIHandlerTool) InputSchema() map[string]interface{} {
	if t.inputSchema != nil {
		return t.inputSchema
	}

	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
		"required":   []string{},
	}
}
