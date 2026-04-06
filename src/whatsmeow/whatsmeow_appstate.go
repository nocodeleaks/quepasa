package whatsmeow

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/proto"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	"go.mau.fi/whatsmeow/appstate"
	"go.mau.fi/whatsmeow/proto/waSyncAction"
	types "go.mau.fi/whatsmeow/types"
)

// sendAppState sends app state patch to WhatsApp (no retry, returns error as-is)
func sendAppState(conn *WhatsmeowConnection, patch appstate.PatchInfo) error {
	ctx := context.Background()
	return conn.Client.SendAppState(ctx, patch)
}

// MarkChatAsRead marks a chat as read using app state protocol
func MarkChatAsRead(conn *WhatsmeowConnection, chatId string) error {
	if conn.Client == nil {
		return fmt.Errorf("client not defined")
	}

	jid, err := types.ParseJID(chatId)
	if err != nil {
		return fmt.Errorf("invalid chat id format: %v", err)
	}

	patch := appstate.BuildMarkChatAsRead(jid, true, time.Time{}, nil)
	return sendAppState(conn, patch)
}

// MarkChatAsUnread marks a chat as unread using app state protocol
func MarkChatAsUnread(conn *WhatsmeowConnection, chatId string) error {
	if conn.Client == nil {
		return fmt.Errorf("client not defined")
	}

	jid, err := types.ParseJID(chatId)
	if err != nil {
		return fmt.Errorf("invalid chat id format: %v", err)
	}

	patch := appstate.BuildMarkChatAsRead(jid, false, time.Time{}, nil)
	return sendAppState(conn, patch)
}

// ArchiveChat archives or unarchives a chat using app state protocol
func ArchiveChat(conn *WhatsmeowConnection, chatId string, archive bool) error {
	if conn.Client == nil {
		return fmt.Errorf("client not defined")
	}

	jid, err := types.ParseJID(chatId)
	if err != nil {
		return fmt.Errorf("invalid chat id format: %v", err)
	}

	patch := appstate.BuildArchive(jid, archive, time.Time{}, nil)
	return sendAppState(conn, patch)
}

// SaveContact saves a contact to the WhatsApp address book and syncs it to the phone
// when syncToPhone is true, equivalent to "Sync contact with phone" in WhatsApp Web
func SaveContact(conn *WhatsmeowConnection, phone, fullName, firstName string, syncToPhone bool) error {
	if conn.Client == nil {
		return fmt.Errorf("client not defined")
	}

	phoneJID := types.JID{
		User:   phone,
		Server: whatsapp.WHATSAPP_SERVERDOMAIN_USER,
	}

	patch := appstate.PatchInfo{
		Type: appstate.WAPatchRegularHigh,
		Mutations: []appstate.MutationInfo{{
			Index:   []string{appstate.IndexContact, phoneJID.String()},
			Version: 2,
			Value: &waSyncAction.SyncActionValue{
				ContactAction: &waSyncAction.ContactAction{
					FullName:                 proto.String(fullName),
					FirstName:                proto.String(firstName),
					SaveOnPrimaryAddressbook: proto.Bool(syncToPhone),
				},
			},
		}},
	}

	return sendAppState(conn, patch)
}
