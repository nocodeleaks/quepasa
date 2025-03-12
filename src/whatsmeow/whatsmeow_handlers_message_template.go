package whatsmeow

import (
	"fmt"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow/proto/waE2E"
)

func HandleTemplateMessage(logentry *log.Entry, out *whatsapp.WhatsappMessage, in *waE2E.TemplateMessage) {
	logentry.Trace("received a template message !")

	if in.HydratedTemplate != nil {

		info := in.GetContextInfo()
		if info != nil {
			out.ForwardingScore = info.GetForwardingScore()
			out.InReply = info.GetStanzaID()
		}

		HandleHydratedTemplate(logentry, out, in.HydratedTemplate)
	} else {
		logentry.Warnf("template message not handled: %v", in)
	}
}

func HandleTemplateButtonReplyMessage(logentry *log.Entry, out *whatsapp.WhatsappMessage, in *waE2E.TemplateButtonReplyMessage) {
	logentry.Trace("received a template button reply message !")
	out.Type = whatsapp.TextMessageType

	var text string
	text = fmt.Sprintf("%s(%v): ", text, in.GetSelectedIndex())
	text = fmt.Sprintf("%s%s", text, in.GetSelectedDisplayText())
	out.Text = text

	info := in.GetContextInfo()
	if info != nil {
		out.ForwardingScore = info.GetForwardingScore()
		out.InReply = info.GetStanzaID()
	}
}

func HandleHydratedTemplate(logentry *log.Entry, out *whatsapp.WhatsappMessage, in *waE2E.TemplateMessage_HydratedFourRowTemplate) {
	logentry.Trace("received a hydrated four row template !")
	out.Type = whatsapp.TextMessageType

	var text string
	contextText := in.GetHydratedContentText()
	if len(contextText) > 0 {
		text = fmt.Sprintf("%s%s\n", text, contextText)
	}
	out.Text = text
}
