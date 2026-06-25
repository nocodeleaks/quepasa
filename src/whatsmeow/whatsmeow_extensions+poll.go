package whatsmeow

import (
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	"go.mau.fi/util/random"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

func GeneratePollMessage(msg *whatsapp.WhatsappMessage) (*waE2E.Message, error) {
	if msg.Poll == nil {
		return nil, fmt.Errorf("poll data is nil")
	}

	// Create poll options
	pollOptions := make([]*waE2E.PollCreationMessage_Option, len(msg.Poll.Options))
	for i, option := range msg.Poll.Options {
		pollOptions[i] = &waE2E.PollCreationMessage_Option{
			OptionName: proto.String(option),
		}
	}

	// Set default max selections if not provided
	if msg.Poll.Selections <= 0 {
		msg.Poll.Selections = 1
	}

	// Ensure max selections is valid
	if int(msg.Poll.Selections) > len(msg.Poll.Options) {
		msg.Poll.Selections = uint(len(msg.Poll.Options))
	}

	// Create poll message
	pollMessage := &waE2E.Message{
		PollCreationMessage: &waE2E.PollCreationMessage{
			Name:                   proto.String(msg.Poll.Question),
			Options:                pollOptions,
			SelectableOptionsCount: proto.Uint32(uint32(msg.Poll.Selections)),
		},
		// Add MessageContextInfo with random message secret for proper poll functionality
		MessageContextInfo: &waE2E.MessageContextInfo{
			MessageSecret: random.Bytes(32),
		},
	}

	return pollMessage, nil
}
