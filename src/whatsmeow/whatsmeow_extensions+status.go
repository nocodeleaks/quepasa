package whatsmeow

import (
	"context"
	"fmt"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	"go.mau.fi/util/random"
	"go.mau.fi/whatsmeow/proto/waE2E"
	types "go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

// PublishStatus sends a WhatsApp status (story) to all contacts.
// For media status, the attachment content must be populated.
// For text-only status, leave attachment nil and provide text.
// Returns the sent message ID on success.
func (source *WhatsmeowConnection) PublishStatus(text string, attachment *whatsapp.WhatsappAttachment) (string, error) {
	if source == nil || source.Client == nil {
		return "", fmt.Errorf("connection not available")
	}

	var newMessage *waE2E.Message

	if attachment != nil && attachment.GetContent() != nil && len(*attachment.GetContent()) > 0 {
		// Media status: upload content and build the appropriate message type
		content := *attachment.GetContent()
		mediaType := GetMediaTypeFromAttachment(attachment)

		response, err := source.Client.Upload(context.Background(), content, mediaType)
		if err != nil {
			return "", fmt.Errorf("failed to upload status media: %w", err)
		}

		waMsg := whatsapp.WhatsappMessage{
			Text:       text,
			Attachment: attachment,
		}
		waMsg.Type = whatsapp.GetMessageType(attachment)

		newMessage = NewWhatsmeowMessageAttachment(response, waMsg, mediaType, nil)
	} else {
		// Text-only status
		if text == "" {
			return "", fmt.Errorf("text is required when no attachment is provided")
		}
		newMessage = &waE2E.Message{
			ExtendedTextMessage: &waE2E.ExtendedTextMessage{
				Text: proto.String(text),
			},
		}
	}

	// Mandatory MessageSecret for all outgoing messages
	if newMessage.MessageContextInfo == nil {
		newMessage.MessageContextInfo = &waE2E.MessageContextInfo{}
	}
	newMessage.MessageContextInfo.MessageSecret = random.Bytes(32)

	resp, err := source.Client.SendMessage(context.Background(), types.StatusBroadcastJID, newMessage)
	if err != nil {
		return "", fmt.Errorf("failed to publish status: %w", err)
	}

	return resp.ID, nil
}
