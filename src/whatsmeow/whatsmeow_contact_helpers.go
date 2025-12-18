package whatsmeow

import (
	"context"
	"strings"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types"
)

// GetContactsFromDevice reads contacts directly from a whatsmeow store device
// This is a shared function used by both WhatsmeowContactManager and WhatsmeowStoreContactManager
// to avoid code duplication
func GetContactsFromDevice(device *store.Device) (chats []whatsapp.WhatsappChat, err error) {
	if device == nil || device.Contacts == nil {
		return nil, err
	}

	contacts, err := device.Contacts.GetAllContacts(context.TODO())
	if err != nil {
		return nil, err
	}

	// Map to track contacts by phone number and their ContactInfo
	type contactEntry struct {
		chat whatsapp.WhatsappChat
		info types.ContactInfo
	}
	contactMap := make(map[string]contactEntry)

	for jid, info := range contacts {
		// Use existing ExtractContactName logic for consistent title extraction
		title := ExtractContactName(info)

		var phoneNumber string
		var lid string
		var phoneE164 string

		if strings.Contains(jid.String(), whatsapp.WHATSAPP_SERVERDOMAIN_LID_SUFFIX) {
			// For @lid contacts, get the corresponding phone number
			pnJID, err := device.LIDs.GetPNForLID(context.TODO(), jid)
			if err == nil && !pnJID.IsEmpty() {
				phoneNumber = pnJID.User
				lid = jid.String()
				// Format phone to E164
				if phone, err := whatsapp.GetPhoneIfValid(phoneNumber); err == nil {
					phoneE164 = phone
				}
			} else {
				// If no mapping found, use the LID as unique identifier
				phoneNumber = jid.String()
				lid = ""
			}
		} else {
			// For regular @s.whatsapp.net contacts
			phoneNumber = jid.User
			// Format phone to E164
			if phone, err := whatsapp.GetPhoneIfValid(phoneNumber); err == nil {
				phoneE164 = phone
			}

			// Try to get corresponding LID
			lidJID, err := device.LIDs.GetLIDForPN(context.TODO(), jid)
			if err == nil && !lidJID.IsEmpty() {
				lid = lidJID.String()
			} else {
				lid = ""
			}
		}

		// Check if contact with this phone number already exists
		existingEntry, exists := contactMap[phoneNumber]

		if !exists {
			// First contact with this phone number
			// If we have a valid phone, prefer @s.whatsapp.net format for Id
			contactId := jid.String()
			if len(phoneE164) > 0 && strings.Contains(jid.String(), whatsapp.WHATSAPP_SERVERDOMAIN_LID_SUFFIX) {
				// Has phone and current is @lid, construct @s.whatsapp.net Id
				contactId = phoneNumber + whatsapp.WHATSAPP_SERVERDOMAIN_USER_SUFFIX
			}

			chat := whatsapp.WhatsappChat{
				Id:    contactId,
				Title: title,
				LId:   lid,
				Phone: phoneE164,
			}
			contactMap[phoneNumber] = contactEntry{chat: chat, info: info}
		} else {
			// Merge information: prefer @s.whatsapp.net for Id when phone available, accumulate LId
			existingContact := existingEntry.chat

			if strings.Contains(jid.String(), whatsapp.WHATSAPP_SERVERDOMAIN_LID_SUFFIX) {
				// Current is @lid - update LId
				existingContact.LId = jid.String()
				// If we have phone, ensure Id uses @s.whatsapp.net format
				if len(phoneE164) > 0 {
					existingContact.Id = phoneNumber + whatsapp.WHATSAPP_SERVERDOMAIN_USER_SUFFIX
				} else if len(existingContact.Id) == 0 {
					existingContact.Id = jid.String()
				}
				// Use ExtractContactName priority: compare current info vs existing info
				if len(title) > 0 && (len(existingContact.Title) == 0 || getContactNamePriority(info) < getContactNamePriority(existingEntry.info)) {
					existingContact.Title = title
					existingEntry.info = info
				}
				existingEntry.chat = existingContact
				contactMap[phoneNumber] = existingEntry
			} else {
				// Current is @s.whatsapp.net - this is preferred for Id when phone is available
				existingContact.Id = jid.String()
				if len(existingContact.LId) == 0 {
					existingContact.LId = lid
				}
				// Use ExtractContactName priority: compare current info vs existing info
				if len(title) > 0 && (len(existingContact.Title) == 0 || getContactNamePriority(info) < getContactNamePriority(existingEntry.info)) {
					existingContact.Title = title
					existingEntry.info = info
				}
				if len(phoneE164) > 0 && len(existingContact.Phone) == 0 {
					existingContact.Phone = phoneE164
				}
				existingEntry.chat = existingContact
				contactMap[phoneNumber] = existingEntry
			}
		}
	}

	// Convert map to slice
	for _, entry := range contactMap {
		chats = append(chats, entry.chat)
	}

	return chats, nil
}

// getContactNamePriority returns the priority level of the name extracted from ContactInfo
// Lower numbers = higher priority (1 is highest, 4 is lowest)
// This follows the ExtractContactName logic: FullName > BusinessName > PushName > FirstName
func getContactNamePriority(info types.ContactInfo) int {
	if len(info.FullName) > 0 {
		return 1 // FullName - User's saved name (highest priority)
	}
	if len(info.BusinessName) > 0 {
		return 2 // BusinessName - Business account name
	}
	if len(info.PushName) > 0 {
		return 3 // PushName - Contact's public name
	}
	if len(info.FirstName) > 0 {
		return 4 // FirstName - Generic first name (lowest priority)
	}

	return 999 // No name available
}
