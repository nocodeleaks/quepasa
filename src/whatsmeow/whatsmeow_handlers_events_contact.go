package whatsmeow

import (
	"bytes"
	"fmt"

	slug "github.com/gosimple/slug"
	"github.com/nocodeleaks/quepasa/library"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	"go.mau.fi/whatsmeow/types/events"
)

func OnEventContact(source *WhatsmeowHandlers, evt events.Contact) {
	if source == nil {
		return
	}

	logentry := source.GetLogger()
	logentry.Debugf("on event contact: %+v", evt)

	// verify if the event has an Action
	if evt.Action == nil {
		logentry.Warn("event contact without action")
		return
	}

	var id string
	if source.Client != nil {
		id = source.Client.GenerateMessageID()
	}

	title := evt.Action.GetFullName()

	chat := *NewWhatsappChat(source, evt.JID)

	chat.Id = evt.JID.String()
	chat.LId = evt.Action.GetLidJID()
	chat.Title = title

	phone := chat.GetPhone()

	vcardtext := new(bytes.Buffer)
	fmt.Fprintln(vcardtext, "BEGIN:VCARD")
	fmt.Fprintln(vcardtext, "VERSION:4.0")
	fmt.Fprintln(vcardtext, "FN:"+title)
	if len(phone) > 0 {
		fmt.Fprintln(vcardtext, "TEL;TYPE=main-number:"+phone)
	}
	fmt.Fprintln(vcardtext, "END:VCARD")

	message := &whatsapp.WhatsappMessage{
		Content:     evt,
		FromHistory: evt.FromFullSync,
		Id:          id,
		Timestamp:   evt.Timestamp,
		Type:        whatsapp.ContactMessageType,
		Chat:        chat,
		Text:        title,
		Edited:      true,
	}

	filename := message.Chat.Title
	if len(filename) == 0 {
		if len(message.Id) > 0 {
			filename = message.Id
		} else {
			filename = "unknown"
		}
	}

	filename = fmt.Sprintf("%s.vcf", slug.Make(filename))

	content := vcardtext.Bytes()
	message.Attachment = whatsapp.GenerateVCardAttachment(content, filename)
	message.Attachment.Checksum = library.GenerateCRC32ChecksumString(content)

	// dispatching to internal handlers
	source.Follow(message, "vcard")
}
