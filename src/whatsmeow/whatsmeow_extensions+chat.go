package whatsmeow

import (
	"context"
	"reflect"

	library "github.com/nocodeleaks/quepasa/library"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	whatsmeow "go.mau.fi/whatsmeow"
	types "go.mau.fi/whatsmeow/types"
)

/**
 * ExtractContactName extracts the best available contact name following the hierarchy:
 * BusinessName > FullName > PushName > FirstName
 *
 * @param cInfo Contact info from WhatsApp store
 * @return The best available name or empty string if none found
 */
func ExtractContactName(cInfo interface{}) string {
	// Use reflection to access the fields dynamically
	v := reflect.ValueOf(cInfo)
	if v.Kind() == reflect.Struct {
		// Try BusinessName
		if businessName := v.FieldByName("BusinessName"); businessName.IsValid() && businessName.Kind() == reflect.String {
			if name := businessName.String(); len(name) > 0 {
				return name
			}
		}
		// Try FullName
		if fullName := v.FieldByName("FullName"); fullName.IsValid() && fullName.Kind() == reflect.String {
			if name := fullName.String(); len(name) > 0 {
				return name
			}
		}
		// Try PushName
		if pushName := v.FieldByName("PushName"); pushName.IsValid() && pushName.Kind() == reflect.String {
			if name := pushName.String(); len(name) > 0 {
				return name
			}
		}
		// Try FirstName
		if firstName := v.FieldByName("FirstName"); firstName.IsValid() && firstName.Kind() == reflect.String {
			if name := firstName.String(); len(name) > 0 {
				return name
			}
		}
	}
	return ""
}

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
	if client == nil {
		return ""
	}

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
		if client.Store == nil || client.Store.Contacts == nil {
			return ""
		}

		cInfo, _ := client.Store.Contacts.GetContact(context.Background(), jid)
		if cInfo.Found {
			title = ExtractContactName(cInfo)
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
	} else if jid.Server == whatsapp.WHATSAPP_SERVERDOMAIN_LID {
		// For @lid contacts, get the corresponding phone number
		phone, err := contactManager.GetPhoneFromContactId(chat.Id)
		if err == nil && len(phone) > 0 {
			chat.Phone = phone
			// LId is already the chat.Id for @lid contacts
			chat.LId = chat.Id
		}
	}

	return chat
}
