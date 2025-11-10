package mcp

import (
	"encoding/json"

	models "github.com/nocodeleaks/quepasa/models"
)

// HealthTool implements the health check tool for MCP
type HealthTool struct{}

// HealthRequest represents the request for health check
type HealthRequest struct {
	// No parameters needed for health check
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status     string                   `json:"status"`
	Connected  bool                     `json:"connected"`
	Timestamp  string                   `json:"timestamp"`
	ServerInfo *models.QpWhatsappServer `json:"server_info,omitempty"`
}

// ExecuteWithContext runs the health check with authentication context
func (h *HealthTool) ExecuteWithContext(ctx *MCPToolContext, params json.RawMessage) (interface{}, error) {
	if ctx.IsMaster {
		// Master key access - return global system health
		totalServers := len(models.WhatsappService.Servers)
		connectedServers := 0

		for _, srv := range models.WhatsappService.Servers {
			if srv.GetConnection() != nil {
				connectedServers++
			}
		}

		return map[string]interface{}{
			"status":               "ok",
			"access_level":         "master",
			"total_servers":        totalServers,
			"connected_servers":    connectedServers,
			"disconnected_servers": totalServers - connectedServers,
		}, nil
	}

	// Bot token access - return server-specific health
	if ctx.Server == nil {
		return map[string]interface{}{
			"status": "error",
			"error":  "No server context available",
		}, nil
	}

	status := ctx.Server.GetStatus()

	return &HealthResponse{
		Status:     status.String(),
		Connected:  ctx.Server.GetConnection() != nil,
		Timestamp:  ctx.Server.Timestamp.String(),
		ServerInfo: ctx.Server,
	}, nil
}

// Name returns the tool name
func (h *HealthTool) Name() string {
	return "health"
}

// Description returns the tool description
func (h *HealthTool) Description() string {
	return "Check WhatsApp server health and connection status"
}

// InputSchema returns the JSON schema for the tool input
func (h *HealthTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
		"required":   []string{},
	}
}
