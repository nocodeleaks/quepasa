package whatsmeow

import (
	"fmt"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow/proto/waE2E"
)

func HandleListMessage(logentry *log.Entry, out *whatsapp.WhatsappMessage, in *waE2E.ListMessage) {
	logentry.Trace("received a list message !")
	out.Type = whatsapp.TextMessageType

	var text string
	text = fmt.Sprintf("%s*%s:*\n ", text, in.GetTitle())
	text = fmt.Sprintf("%s_%s_\n ", text, in.GetDescription())
	text = fmt.Sprintf("%s%s", text, "\n")
	text = fmt.Sprintf("%s*%s*:\n", text, in.GetButtonText())

	for _, section := range in.Sections {
		if section != nil {
			text = fmt.Sprintf("%s-- *%s:*\n ", text, section.GetTitle())
			for _, row := range section.Rows {
				if row != nil {
					text = fmt.Sprintf("%s* (%s) *%s:* %s\n ", text, row.GetRowID(), row.GetTitle(), row.GetDescription())
				}
			}
		}
	}

	footerText := in.GetFooterText()
	if len(footerText) > 0 {
		text = fmt.Sprintf("%s\n--------------------------------\n%s", text, footerText)
	}

	out.Text = text

	info := in.GetContextInfo()
	if info != nil {
		out.ForwardingScore = info.GetForwardingScore()
		out.InReply = info.GetStanzaID()
	}
}
