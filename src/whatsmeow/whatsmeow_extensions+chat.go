package whatsmeow

import (
	"context"

	library "github.com/nocodeleaks/quepasa/library"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	whatsmeow "go.mau.fi/whatsmeow"
	types "go.mau.fi/whatsmeow/types"
)

/**
 * GetChatTitle returns a valid chat title from the local memory store or WhatsApp contact/group info.
 *
 * If the JID is a group, tries to get the name from cache or fetches group info. For contacts, checks business name, full name, or push name.
 *
 * @param client Whatsmeow client instance
 * @param jid WhatsApp JID to look up
 * @return Normalized chat title string
 */
func GetChatTitle(client *whatsmeow.Client, jid types.JID) (title string) {
	if jid.Server == whatsapp.WHATSAPP_SERVERDOMAIN_GROUP {
		title = GroupInfoCache.Get(jid.String())
		if len(title) > 0 {
			goto found
		}
		gInfo, _ := client.GetGroupInfo(jid)
		if gInfo != nil {
			title = gInfo.Name
			_ = GroupInfoCache.Append(jid.String(), title, "GetChatTitle")
			goto found
		}
	} else {
		cInfo, _ := client.Store.Contacts.GetContact(context.Background(), jid)
		if cInfo.Found {
			if len(cInfo.BusinessName) > 0 {
				title = cInfo.BusinessName
				goto found
			} else if len(cInfo.FullName) > 0 {
				title = cInfo.FullName
				goto found
			} else if len(cInfo.PushName) > 0 {
				title = cInfo.PushName
				goto found
			} else if len(cInfo.FirstName) > 0 {
				title = cInfo.FirstName
				goto found
			}
		}
	}
	return ""
found:
	return library.NormalizeForTitle(title)
}

func NewWhatsappChat(handler *WhatsmeowHandlers, jid types.JID) *whatsapp.WhatsappChat {
	contactManager := handler.GetContactManager()
	return NewWhatsappChatRaw(handler.Client, contactManager, jid)
}

func NewWhatsappChatRaw(client *whatsmeow.Client, contactManager whatsapp.WhatsappContactManagerInterface, jid types.JID) *whatsapp.WhatsappChat {
	chat := &whatsapp.WhatsappChat{}

	// Always use User@Server format
	// Remove any session ID if present
	chat.Id = jid.User + "@" + jid.Server

	chat.Title = GetChatTitle(client, jid)

	if jid.Server == whatsapp.WHATSAPP_SERVERDOMAIN_USER {
		phone, err := contactManager.GetPhoneFromContactId(chat.Id)
		if err == nil && len(phone) > 0 {
			chat.Phone = phone

			// Try to get LID from phone number
			// This will populate LId if available
			lid, err := contactManager.GetLIDFromPhone(chat.Phone)
			if err == nil && len(lid) > 0 {
				chat.LId = lid
			}
		}
	}

	return chat
}
