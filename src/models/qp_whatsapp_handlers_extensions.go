package models

import (
	"fmt"
	"reflect"
	"strings"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// IsValidForDispatch validates if a given WhatsApp message payload is suitable for further processing
// and dispatch. It checks for specific conditions that would indicate the message should be ignored.
//
// Parameters:
//
//	payload *whatsapp.WhatsappMessage: A pointer to the WhatsApp message payload to be validated.
//
// Returns:
//
//	string: An empty string if the message is valid for dispatch.
//	        A non-empty string containing an explanation message if the message should be ignored.
func IsValidForDispatch(payload *whatsapp.WhatsappMessage) string {
	// Ignores messages with 'Discard' or 'Unknown' types, as these are typically not meant for
	// application-level processing or indicate an unhandled message format.

	if !ENV.DispatchUnhandled() {

		if payload.Type == whatsapp.UnhandledMessageType {
			return fmt.Sprintf("ignoring unhandled message type on webhook request: %v", reflect.TypeOf(&payload))
		}

		// Empty message validation
		// -----------------------
		// Ignores text messages that are empty or contain only whitespace.
		// Such messages generally don't carry meaningful information for the application.
		//
		// EXCEPTION: Allow empty text when it's a reaction (InReaction=true) because
		// reaction removal sends empty text but still needs to be dispatched
		isTextMessage := payload.Type == whatsapp.TextMessageType
		isEmptyText := len(strings.TrimSpace(payload.Text)) <= 0

		if isTextMessage && isEmptyText {
			// InReaction=true with empty text indicates a reaction removal
			if !payload.InReaction {
				return fmt.Sprintf("ignoring empty text message on webhook request: %s", payload.Id)
			}
			// Empty text with InReaction=true is valid (reaction removal)
		}
	}

	// If none of the above conditions are met, the message is considered valid for dispatch.
	return ""
}
