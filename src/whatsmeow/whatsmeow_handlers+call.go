package whatsmeow

import (
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
	types "go.mau.fi/whatsmeow/types"
)

// CallIdentifiers holds the resolved JID information for a call event
type CallIdentifiers struct {
	ChatJID types.JID // The JID to use for creating the chat
	LidJID  types.JID // The LID JID if available
	HasLID  bool      // Whether we have LID information
}

// ResolveCallIdentifiers determines the best JID to use for a call event
// Priority: Use CallCreatorAlt (phone) if CallCreator is a LID, otherwise use best available
func ResolveCallIdentifiers(evt types.BasicCallMeta, logentry *log.Entry) CallIdentifiers {
	result := CallIdentifiers{}

	// Log available identifiers for debugging
	logentry.Debugf("identifiers: creator: %s, alt: %s", evt.CallCreator, evt.CallCreatorAlt)

	// Determine best chat JID with priority for non-LID identifiers
	// Strategy: Prefer CallCreatorAlt if CallCreator is a LID, otherwise use CallCreator
	if !evt.CallCreator.IsEmpty() && evt.CallCreator.Server == whatsapp.WHATSAPP_SERVERDOMAIN_LID {
		// CallCreator is a LID, prefer CallCreatorAlt if available
		result.HasLID = true
		result.LidJID = evt.CallCreator

		if !evt.CallCreatorAlt.IsEmpty() {
			result.ChatJID = evt.CallCreatorAlt
			logentry.Debugf("using CallCreatorAlt (phone) because CallCreator is LID: %s", result.ChatJID)
		} else {
			result.ChatJID = evt.CallCreator
			logentry.Debugf("using CallCreator (LID) - no alternative available: %s", result.ChatJID)
		}
	} else if !evt.CallCreatorAlt.IsEmpty() {
		// CallCreatorAlt is available and CallCreator is not LID
		result.ChatJID = evt.CallCreatorAlt
		logentry.Debugf("using CallCreatorAlt as chat JID: %s", result.ChatJID)

		// Check if CallCreator is also available as potential LID
		if !evt.CallCreator.IsEmpty() && evt.CallCreator.Server == whatsapp.WHATSAPP_SERVERDOMAIN_LID {
			result.HasLID = true
			result.LidJID = evt.CallCreator
		}
	} else if !evt.CallCreator.IsEmpty() {
		// Only CallCreator is available
		result.ChatJID = evt.CallCreator
		logentry.Debugf("using CallCreator as chat JID: %s", result.ChatJID)

		if evt.CallCreator.Server == whatsapp.WHATSAPP_SERVERDOMAIN_LID {
			result.HasLID = true
			result.LidJID = evt.CallCreator
		}
	} else {
		// Fallback to From
		result.ChatJID = evt.From
		logentry.Debugf("using From as chat JID: %s", result.ChatJID)
	}

	return result
}

// EnrichCallChat enriches the chat object with LID information and resolves phone/title when needed
func EnrichCallChat(handler *WhatsmeowHandlers, chat *whatsapp.WhatsappChat, identifiers CallIdentifiers, chatJID types.JID, logentry *log.Entry) {
	if !identifiers.HasLID || identifiers.LidJID.IsEmpty() {
		return
	}

	lidStr := identifiers.LidJID.User + "@" + identifiers.LidJID.Server

	// Set LId if not already set
	if len(chat.LId) == 0 {
		chat.LId = lidStr
	}

	// If we used LID as chatJID and don't have phone, try to resolve it
	if chatJID.Server == whatsapp.WHATSAPP_SERVERDOMAIN_LID && len(chat.Phone) == 0 {
		contactManager := handler.GetContactManager()
		if contactManager == nil {
			return
		}

		// Try to resolve LID to phone
		phone, err := contactManager.GetPhoneFromLID(lidStr)
		if err != nil || len(phone) == 0 {
			return
		}

		chat.Phone = phone
		logentry.Debugf("resolved phone from LID: %s -> %s", lidStr, phone)

		// Update chat.Id to use phone-based format for consistency
		chat.Id = phone + "@" + whatsapp.WHATSAPP_SERVERDOMAIN_USER

		// Try to get title if still empty
		if len(chat.Title) == 0 {
			if phoneJID, err := types.ParseJID(chat.Id); err == nil {
				chat.Title = GetChatTitle(handler.Client, phoneJID)
				logentry.Debugf("resolved title from phone JID: %s", chat.Title)
			}
		}
	}
}
