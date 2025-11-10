package mcp

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	log "github.com/sirupsen/logrus"
)

// MCPEndpoint represents an API endpoint that can be exposed as MCP tool
type MCPEndpoint struct {
	Name        string
	Description string
	Method      string // GET, POST, PUT, DELETE
	Path        string
	Handler     interface{} // The actual handler function
	InputSchema map[string]interface{}
}

// EndpointTool wraps an API endpoint as an MCP tool
type EndpointTool struct {
	endpoint MCPEndpoint
}

// ExecuteWithContext runs the endpoint handler with authentication context
func (e *EndpointTool) ExecuteWithContext(ctx *MCPToolContext, params json.RawMessage) (interface{}, error) {
	// Parse parameters
	var paramsMap map[string]interface{}
	if len(params) > 0 {
		if err := json.Unmarshal(params, &paramsMap); err != nil {
			return nil, fmt.Errorf("invalid parameters: %v", err)
		}
	}

	// Call handler using reflection
	handlerValue := reflect.ValueOf(e.endpoint.Handler)
	if handlerValue.Kind() != reflect.Func {
		return nil, fmt.Errorf("handler is not a function")
	}

	// For now, call with no args - will expand later
	results := handlerValue.Call([]reflect.Value{})

	if len(results) == 0 {
		return map[string]interface{}{"status": "success"}, nil
	}

	// Return first result
	result := results[0].Interface()
	return result, nil
}

// Name returns the tool name
func (e *EndpointTool) Name() string {
	return e.endpoint.Name
}

// Description returns the tool description
func (e *EndpointTool) Description() string {
	if e.endpoint.Description != "" {
		return e.endpoint.Description
	}
	return fmt.Sprintf("%s %s", e.endpoint.Method, e.endpoint.Path)
}

// InputSchema returns the JSON schema for the tool input
func (e *EndpointTool) InputSchema() map[string]interface{} {
	if e.endpoint.InputSchema != nil {
		return e.endpoint.InputSchema
	}

	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
		"required":   []string{},
	}
}

// MCPEndpointRegistry manages API endpoint registration
type MCPEndpointRegistry struct {
	endpoints []MCPEndpoint
}

// NewEndpointRegistry creates a new endpoint registry
func NewEndpointRegistry() *MCPEndpointRegistry {
	return &MCPEndpointRegistry{
		endpoints: make([]MCPEndpoint, 0),
	}
}

// Register registers an endpoint for MCP exposure
func (r *MCPEndpointRegistry) Register(endpoint MCPEndpoint) {
	// Convert endpoint name to snake_case
	endpoint.Name = toSnakeCase(endpoint.Name)
	r.endpoints = append(r.endpoints, endpoint)
	log.Debugf("MCP: Registered endpoint tool: %s", endpoint.Name)
}

// GetTools returns all registered endpoints as MCP tools
func (r *MCPEndpointRegistry) GetTools() []MCPTool {
	tools := make([]MCPTool, 0, len(r.endpoints))
	for _, ep := range r.endpoints {
		tools = append(tools, &EndpointTool{endpoint: ep})
	}
	return tools
}

// toSnakeCase converts a string to snake_case
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
