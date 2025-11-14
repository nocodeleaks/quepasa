package mcp

import (
	"fmt"
	"net/http"

	api "github.com/nocodeleaks/quepasa/api"
)

// HandlerRegistry maps controller function names to actual handlers
var HandlerRegistry = map[string]http.HandlerFunc{
	// Account & Authentication
	"AccountController":     api.AccountController,
	"HealthController":      api.HealthController,
	"BasicHealthController": api.BasicHealthController,

	// Information & Control
	"CreateInformationController": api.CreateInformationController,
	"GetInformationController":    api.GetInformationController,
	"UpdateInformationController": api.UpdateInformationController,
	"DeleteInformationController": api.DeleteInformationController,
	"ScannerController":           api.ScannerController,
	"PairCodeController":          api.PairCodeController,
	"CommandController":           api.CommandController,

	// Message Operations
	"GetMessageController":   api.GetMessageController,
	"RevokeController":       api.RevokeController,
	"MarkReadController":     api.MarkReadController,
	"SendAny":                api.SendAny,
	"SendDocument":           api.SendDocument,
	"SendDocumentFromBinary": api.SendDocumentFromBinary,

	// Receive & Download
	"ReceiveAPIHandler":  api.ReceiveAPIHandler,
	"DownloadController": api.DownloadController,

	// Picture Operations
	"PictureController": api.PictureController,

	// Contacts & Groups
	"ContactsController":                api.ContactsController,
	"GetGroupController":                api.GetGroupController,
	"FetchAllGroupsController":          api.FetchAllGroupsController,
	"CreateGroupController":             api.CreateGroupController,
	"SetGroupNameController":            api.SetGroupNameController,
	"SetGroupPhotoController":           api.SetGroupPhotoController,
	"UpdateGroupParticipantsController": api.UpdateGroupParticipantsController,
	"GroupMembershipRequestsController": api.GroupMembershipRequestsController,
	"SetGroupTopicController":           api.SetGroupTopicController,
	"LeaveGroupController":              api.LeaveGroupController,

	// Invite
	"InviteController": api.InviteController,

	// User Operations
	"IsOnWhatsappController":      api.IsOnWhatsappController,
	"UserInfoController":          api.UserInfoController,
	"UserController":              api.UserController,
	"GetPhoneController":          api.GetPhoneController,
	"GetUserIdentifierController": api.GetUserIdentifierController,

	// Chat Operations
	"ArchiveChatController":      api.ArchiveChatController,
	"ChatPresenceController":     api.ChatPresenceController,
	"MarkChatAsReadController":   api.MarkChatAsReadController,
	"MarkChatAsUnreadController": api.MarkChatAsUnreadController,
	"EditMessageController":      api.EditMessageController,

	// Other
	"WebhookController":  api.WebhookController,
	"RabbitMQController": api.RabbitMQController,
}

// GetHandlerByName retrieves handler function by controller name
func GetHandlerByName(name string) (http.HandlerFunc, error) {
	handler, exists := HandlerRegistry[name]
	if !exists {
		return nil, fmt.Errorf("handler not found: %s", name)
	}
	return handler, nil
}

// GetHandlerNameByFunc retrieves controller name by handler function (for debugging)
func GetHandlerNameByFunc(handler http.HandlerFunc) string {
	// This is primarily for debugging/logging
	for name, h := range HandlerRegistry {
		// Note: Function comparison in Go is limited, this is best-effort
		if fmt.Sprintf("%p", h) == fmt.Sprintf("%p", handler) {
			return name
		}
	}
	return "unknown"
}
