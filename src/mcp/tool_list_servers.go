package mcp

import (
	"encoding/json"

	models "github.com/nocodeleaks/quepasa/models"
)

// ListServersTool lists all available WhatsApp servers
type ListServersTool struct{}

// ServerInfo represents simplified server information
type ServerInfo struct {
	Token     string `json:"token"`
	Number    string `json:"number"`
	Status    string `json:"status"`
	Connected bool   `json:"connected"`
}

// ExecuteWithContext lists servers with authentication context
func (t *ListServersTool) ExecuteWithContext(ctx *MCPToolContext, params json.RawMessage) (interface{}, error) {
	if !ctx.IsMaster && ctx.Server != nil {
		// Bot token access - return only this server
		return []ServerInfo{
			{
				Token:     ctx.Server.Token,
				Number:    ctx.Server.GetWId(),
				Status:    ctx.Server.GetStatus().String(),
				Connected: ctx.Server.GetConnection() != nil,
			},
		}, nil
	}

	// Master key access - return all servers
	servers := make([]ServerInfo, 0, len(models.WhatsappService.Servers))
	for token, srv := range models.WhatsappService.Servers {
		servers = append(servers, ServerInfo{
			Token:     token,
			Number:    srv.GetWId(),
			Status:    srv.GetStatus().String(),
			Connected: srv.GetConnection() != nil,
		})
	}

	return servers, nil
}

// Name returns the tool name
func (t *ListServersTool) Name() string {
	return "list_servers"
}

// Description returns the tool description
func (t *ListServersTool) Description() string {
	return "List all WhatsApp servers/bots available. Returns token, number, status, and connection state for each server."
}

// InputSchema returns the JSON schema for the tool input
func (t *ListServersTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
		"required":   []string{},
	}
}
