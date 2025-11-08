package mcp

import (
	"encoding/json"

	models "github.com/nocodeleaks/quepasa/models"
)

// HealthTool implements the health check tool for MCP
type HealthTool struct {
	server *models.QpWhatsappServer
}

// HealthRequest represents the request for health check
type HealthRequest struct {
	// No parameters needed for health check
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string                         `json:"status"`
	Connected bool                           `json:"connected"`
	Timestamp string                         `json:"timestamp"`
	ServerInfo *models.QpWhatsappServer     `json:"server_info,omitempty"`
}

// Execute runs the health check tool
func (h *HealthTool) Execute(params json.RawMessage) (interface{}, error) {
	if h.server == nil {
		return &HealthResponse{
			Status:    "error",
			Connected: false,
			Timestamp: "",
		}, nil
	}

	status := h.server.GetStatus()
	
	return &HealthResponse{
		Status:     status.String(),
		Connected:  h.server.GetConnection() != nil,
		Timestamp:  h.server.Timestamp.String(),
		ServerInfo: h.server,
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
