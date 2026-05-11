package whatsmeow

import (
	"context"
	"fmt"
	"time"
	"unicode/utf8"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	"go.mau.fi/util/random"
	"go.mau.fi/whatsmeow/proto/waCommon"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

// SendReaction sends an emoji reaction to an existing message.
// Pass emoji="" to remove an existing reaction.
func (source *WhatsmeowConnection) SendReaction(chatID, targetMsgID string, fromMe bool, emoji string) error {
	if source == nil || source.Client == nil {
		return fmt.Errorf("connection not available")
	}

	// Validate emoji — allow empty string (removes reaction) or a valid unicode rune sequence
	if emoji != "" && !utf8.ValidString(emoji) {
		return fmt.Errorf("invalid emoji: not valid UTF-8")
	}

	formattedChat, _ := whatsapp.FormatEndpoint(chatID)
	chatJID, err := types.ParseJID(formattedChat)
	if err != nil {
		return fmt.Errorf("invalid chat id: %w", err)
	}

	msg := &waE2E.Message{
		ReactionMessage: &waE2E.ReactionMessage{
			Key: &waCommon.MessageKey{
				RemoteJID: proto.String(formattedChat),
				FromMe:    proto.Bool(fromMe),
				ID:        proto.String(targetMsgID),
			},
			Text:              proto.String(emoji),
			SenderTimestampMS: proto.Int64(time.Now().UnixMilli()),
		},
		MessageContextInfo: &waE2E.MessageContextInfo{
			MessageSecret: random.Bytes(32),
		},
	}

	_, err = source.Client.SendMessage(context.Background(), chatJID, msg)
	return err
}
