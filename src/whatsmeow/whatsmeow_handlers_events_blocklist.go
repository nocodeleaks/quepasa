package whatsmeow

import (
	"fmt"
	"strings"

	qpevents "github.com/nocodeleaks/quepasa/events"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	"go.mau.fi/whatsmeow/types/events"
)

func OnEventBlocklist(source *WhatsmeowHandlers, evt events.Blocklist) {
	if source == nil {
		return
	}

	logentry := source.GetLogger()
	logentry.Debugf("on event blocklist: %+v", evt)

	var changes []string
	for _, change := range evt.Changes {
		changes = append(changes, fmt.Sprintf("%s:%s", change.JID.String(), change.Action))
	}

	text := fmt.Sprintf("blocklist updated (%s)", evt.Action)
	if len(changes) > 0 {
		text = fmt.Sprintf("%s - %s", text, strings.Join(changes, ", "))
	}

	message := &whatsapp.WhatsappMessage{
		Content:   evt,
		Id:        fmt.Sprintf("blocklist_%s", source.Client.GenerateMessageID()),
		Timestamp: source.getTimestamp(),
		Type:      whatsapp.SystemMessageType,
		Chat:      whatsapp.WASYSTEMCHAT,
		Text:      text,
		FromMe:    true,
		Debug: &whatsapp.WhatsappMessageDebug{
			Event:  "Blocklist",
			Info:   evt,
			Reason: "blocklist",
		},
	}

	source.Follow(message, "blocklist")

	qpevents.Publish(qpevents.Event{
		Name:   "whatsapp.blocklist.updated",
		Source: "whatsmeow.handlers",
		Status: "success",
		Attributes: map[string]string{
			"action": string(evt.Action),
		},
	})
}
