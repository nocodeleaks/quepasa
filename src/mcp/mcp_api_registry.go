package mcp

import (
	api "github.com/nocodeleaks/quepasa/api"
	log "github.com/sirupsen/logrus"
)

// RegisterAPITools registers all API endpoints as MCP tools
func (s *MCPServer) RegisterAPITools() {
	log.Info("MCP: Registering API endpoints as tools...")

	// Send Message
	s.registry.Register(&APIHandlerTool{
		name:        "send_message",
		description: "Send a WhatsApp message (text, image, document, audio, video)",
		method:      "POST",
		path:        "/send",
		handler:     api.SendAny,
		inputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"token": map[string]interface{}{
					"type":        "string",
					"description": "Bot token (required when using master key authentication)",
				},
				"chatId": map[string]interface{}{
					"type":        "string",
					"description": "Chat ID (phone number with country code, e.g., 5521999999999)",
				},
				"text": map[string]interface{}{
					"type":        "string",
					"description": "Message text",
				},
				"attachment": map[string]interface{}{
					"type":        "object",
					"description": "Media attachment",
					"properties": map[string]interface{}{
						"url": map[string]interface{}{
							"type":        "string",
							"description": "Media URL",
						},
						"mimetype": map[string]interface{}{
							"type":        "string",
							"description": "MIME type (e.g., image/jpeg, audio/ogg)",
						},
						"filename": map[string]interface{}{
							"type":        "string",
							"description": "File name",
						},
					},
				},
			},
			"required": []string{"chatId"},
		},
	})

	// Receive Messages
	s.registry.Register(&APIHandlerTool{
		name:        "receive_messages",
		description: "Get received messages from webhook cache",
		method:      "GET",
		path:        "/receive",
		handler:     api.ReceiveAPIHandler,
		inputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"token": map[string]interface{}{
					"type":        "string",
					"description": "Bot token (required when using master key authentication)",
				},
				"timestamp": map[string]interface{}{
					"type":        "string",
					"description": "Get messages after this timestamp (optional)",
				},
			},
			"required": []string{},
		},
	})

	// Get QR Code
	s.registry.Register(&APIHandlerTool{
		name:        "get_qrcode",
		description: "Get QR code for WhatsApp pairing",
		method:      "GET",
		path:        "/scan",
		handler:     api.ScannerController,
		inputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"token": map[string]interface{}{
					"type":        "string",
					"description": "Bot token (required when using master key authentication)",
				},
			},
			"required": []string{},
		},
	})

	// Download Media
	s.registry.Register(&APIHandlerTool{
		name:        "download_media",
		description: "Download media from message",
		method:      "GET",
		path:        "/download/{messageId}",
		handler:     api.DownloadController,
		inputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"token": map[string]interface{}{
					"type":        "string",
					"description": "Bot token (required when using master key authentication)",
				},
				"messageId": map[string]interface{}{
					"type":        "string",
					"description": "Message ID to download media from",
				},
			},
			"required": []string{"messageId"},
		},
	})

	// Get Contacts
	s.registry.Register(&APIHandlerTool{
		name:        "get_contacts",
		description: "Get WhatsApp contacts list",
		method:      "GET",
		path:        "/contacts",
		handler:     api.ContactsController,
		inputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"token": map[string]interface{}{
					"type":        "string",
					"description": "Bot token (required when using master key authentication)",
				},
			},
			"required": []string{},
		},
	})

	// Get Groups
	s.registry.Register(&APIHandlerTool{
		name:        "get_groups",
		description: "Get WhatsApp groups list",
		method:      "GET",
		path:        "/groups/getall",
		handler:     api.FetchAllGroupsController,
		inputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"token": map[string]interface{}{
					"type":        "string",
					"description": "Bot token (required when using master key authentication)",
				},
			},
			"required": []string{},
		},
	})

	// Check if number is on WhatsApp
	s.registry.Register(&APIHandlerTool{
		name:        "is_on_whatsapp",
		description: "Check if a phone number is registered on WhatsApp",
		method:      "GET",
		path:        "/isonwhatsapp/{phone}",
		handler:     api.IsOnWhatsappController,
		inputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"token": map[string]interface{}{
					"type":        "string",
					"description": "Bot token (required when using master key authentication)",
				},
				"phone": map[string]interface{}{
					"type":        "string",
					"description": "Phone number with country code (e.g., 5521999999999)",
				},
			},
			"required": []string{"phone"},
		},
	})

	// Get Profile Picture
	s.registry.Register(&APIHandlerTool{
		name:        "get_picture",
		description: "Get profile picture URL for a contact or group",
		method:      "GET",
		path:        "/picinfo/{chatId}",
		handler:     api.PictureController,
		inputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"token": map[string]interface{}{
					"type":        "string",
					"description": "Bot token (required when using master key authentication)",
				},
				"chatId": map[string]interface{}{
					"type":        "string",
					"description": "Chat ID (phone number or group ID)",
				},
			},
			"required": []string{"chatId"},
		},
	})

	// Mark chat as read
	s.registry.Register(&APIHandlerTool{
		name:        "mark_as_read",
		description: "Mark chat messages as read",
		method:      "POST",
		path:        "/chat/markread",
		handler:     api.MarkChatAsReadController,
		inputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"token": map[string]interface{}{
					"type":        "string",
					"description": "Bot token (required when using master key authentication)",
				},
				"chatId": map[string]interface{}{
					"type":        "string",
					"description": "Chat ID to mark as read",
				},
			},
			"required": []string{"chatId"},
		},
	})

	log.Infof("MCP: Registered %d API endpoint tools", len(s.registry.List()))
}
