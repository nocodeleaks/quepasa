package whatsmeow

import (
	"context"

	library "github.com/nocodeleaks/quepasa/library"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	whatsmeow "go.mau.fi/whatsmeow"
	types "go.mau.fi/whatsmeow/types"
)

/**
 * CleanJID removes the session suffix from a JID if present.
 * Example: 554792857088:72@s.whatsapp.net -> 554792857088@s.whatsapp.net
 *
 * @param jid The JID to clean
 * @return JID without session suffix
 */
func CleanJID(jid types.JID) types.JID {
	// Always reconstruct the JID using only User and Server
	// This automatically removes any session suffix that might be present in the original string representation
	cleanJID := types.JID{
		User:   jid.User,
		Server: jid.Server,
	}

	return cleanJID
}

/**
 * ExtractContactName extracts the best available contact name following the hierarchy:
 * BusinessName > FullName > PushName > FirstName
 *
 * @param cInfo Contact info from WhatsApp store
 * @return The best available name or empty string if none found
 */
func ExtractContactName(cInfo types.ContactInfo) string {
	if !cInfo.Found {
		return ""
	}
	if len(cInfo.BusinessName) > 0 {
		return cInfo.BusinessName
	}
	if len(cInfo.FullName) > 0 {
		return cInfo.FullName
	}
	if len(cInfo.PushName) > 0 {
		return cInfo.PushName
	}
	if len(cInfo.FirstName) > 0 {
		return cInfo.FirstName
	}
	return ""
}

/**
 * GetContactName retrieves the contact name for a given JID from the WhatsApp store.
 * Performs null checks to avoid errors and uses ExtractContactName for name extraction.
 * Handles JIDs with session suffixes by trying both full JID and base JID (without session).
 *
 * @param client Whatsmeow client instance
 * @param jid WhatsApp JID to look up
 * @return The best available contact name or empty string if not found or on error
 */
func GetContactName(client *whatsmeow.Client, jid types.JID) string {
	if client == nil || client.Store == nil || client.Store.Contacts == nil {
		return ""
	}

	// Always use cleaned JID (without session) for contact lookup
	cleanJID := CleanJID(jid)
	cInfo, err := client.Store.Contacts.GetContact(context.Background(), cleanJID)
	if err != nil {
		return ""
	}

	if !cInfo.Found {
		return ""
	}

	name := ExtractContactName(cInfo)
	return name
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
		gInfo, _ := client.GetGroupInfo(context.Background(), jid)
		if gInfo != nil {
			title = gInfo.Name
			_ = GroupInfoCache.Append(jid.String(), title, "GetChatTitle")
			goto found
		}
	} else {
		title = GetContactName(client, jid)
		if len(title) > 0 {
			goto found
		}
	}
	return ""
found:
	return library.NormalizeForTitle(title)
}

func NewWhatsappChat(handler *WhatsmeowEventHandler, jid types.JID) *whatsapp.WhatsappChat {
	contactManager := handler.GetContactManager()
	return NewWhatsappChatRaw(handler.Client, contactManager, jid)
}

func NewWhatsappChatRaw(client *whatsmeow.Client, contactManager whatsapp.WhatsappContactManagerInterface, jid types.JID) *whatsapp.WhatsappChat {
	chat := &whatsapp.WhatsappChat{}

	// Always use User@Server format WITHOUT session ID
	// The types.JID already separates the user from session suffix
	chat.Id = jid.User + "@" + jid.Server

	chat.Title = GetChatTitle(client, jid)

	switch jid.Server {
	case whatsapp.WHATSAPP_SERVERDOMAIN_USER:
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
	case whatsapp.WHATSAPP_SERVERDOMAIN_LID:
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
