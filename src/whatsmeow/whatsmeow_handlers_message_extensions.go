package whatsmeow

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	slug "github.com/gosimple/slug"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow/proto/waE2E"
)

func HandleKnowingMessages(handler *WhatsmeowHandlers, out *whatsapp.WhatsappMessage, in *waE2E.Message) {
	logentry := handler.GetLogger()
	logentry = logentry.WithField(LogFields.ChatId, out.Chat.Id)
	logentry = logentry.WithField(LogFields.MessageId, out.Id)
	logentry.Tracef("handling knowing message: %v", in)

	switch {
	case in.ImageMessage != nil:
		HandleImageMessage(logentry, out, in.ImageMessage)
	case in.StickerMessage != nil:
		HandleStickerMessage(logentry, out, in.StickerMessage)
	case in.DocumentMessage != nil:
		HandleDocumentMessage(logentry, out, in.DocumentMessage)
	case in.AudioMessage != nil:
		HandleAudioMessage(logentry, out, in.AudioMessage)
	case in.VideoMessage != nil:
		HandleVideoMessage(logentry, out, in.VideoMessage)
	case in.ExtendedTextMessage != nil:
		HandleExtendedTextMessage(logentry, out, in.ExtendedTextMessage)
	case in.EphemeralMessage != nil:
		HandleEphemeralMessage(logentry, out, in.EphemeralMessage)
	case in.ButtonsResponseMessage != nil:
		HandleButtonsResponseMessage(logentry, out, in.ButtonsResponseMessage)
	case in.LocationMessage != nil:
		HandleLocationMessage(logentry, out, in.LocationMessage)
	case in.LiveLocationMessage != nil:
		HandleLiveLocationMessage(logentry, out, in.LiveLocationMessage)
	case in.ContactMessage != nil:
		HandleContactMessage(logentry, out, in.ContactMessage)
	case in.ReactionMessage != nil:
		HandleReactionMessage(logentry, out, in.ReactionMessage)
	case in.EditedMessage != nil:
		HandleEditTextMessage(logentry, out, in.EditedMessage)
	case in.ProtocolMessage != nil:
		HandleProtocolMessage(logentry, out, in.ProtocolMessage)
	case in.SenderKeyDistributionMessage != nil:
		out.Type = whatsapp.DiscardMessageType
	case in.StickerSyncRmrMessage != nil:
		out.Type = whatsapp.DiscardMessageType
	case len(in.GetConversation()) > 0:
		HandleTextMessage(logentry, out, in)
	default:
		out.Type = whatsapp.UnknownMessageType
		logentry.Warnf("message not handled: %v", in)
	}
}

//#region HANDLING TEXT MESSAGES

func HandleTextMessage(log *log.Entry, out *whatsapp.WhatsappMessage, in *waE2E.Message) {
	log.Debug("received a text message !")
	out.Type = whatsapp.TextMessageType
	out.Text = in.GetConversation()
}

func HandleEditTextMessage(log *log.Entry, out *whatsapp.WhatsappMessage, in *waE2E.FutureProofMessage) {
	// never throws, obs !!!!
	// it came as a single text msg
	log.Debug("received a edited text message !")
	out.Type = whatsapp.TextMessageType
	out.Text = in.String()
}

func HandleProtocolMessage(logentry *log.Entry, out *whatsapp.WhatsappMessage, in *waE2E.ProtocolMessage) {
	logentry.Trace("received a protocol message !")

	switch v := in.GetType(); {
	case v == waE2E.ProtocolMessage_MESSAGE_EDIT:
		out.Type = whatsapp.TextMessageType
		out.Id = in.Key.GetID()
		out.Text = in.EditedMessage.GetConversation()
		out.Edited = true
		return

	case v == waE2E.ProtocolMessage_REVOKE:
		out.Id = in.Key.GetID()
		out.Type = whatsapp.RevokeMessageType
		return

	case v == waE2E.ProtocolMessage_HISTORY_SYNC_NOTIFICATION:
		var logtext string
		out.Type = whatsapp.UnknownMessageType
		b, err := json.Marshal(in)
		if err != nil {
			logentry.Error(err)
			return
		}

		logtext = "ProtocolMessage :: " + string(b)

		notif := in.GetHistorySyncNotification()
		if notif != nil {
			b, err = json.Marshal(notif)
			if err != nil {
				logentry.Error(err)
				return
			}
			logtext = logtext + "History Sync Notification :: " + string(b)
		}

		out.Text = logtext
		return

	default:
		out.Type = whatsapp.UnknownMessageType
		b, err := json.Marshal(in)
		if err != nil {
			logentry.Error(err)
			return
		}

		out.Text = "ProtocolMessage :: " + string(b)
		return
	}
}

// temporary messages
func HandleEphemeralMessage(logentry *log.Entry, out *whatsapp.WhatsappMessage, in *waE2E.FutureProofMessage) {
	logentry.Warnf("handling ephemeral message not implemented: %v", in)
}

// Msg em resposta a outra
func HandleExtendedTextMessage(logentry *log.Entry, out *whatsapp.WhatsappMessage, in *waE2E.ExtendedTextMessage) {
	logentry.Debug("received a text|extended message !")
	out.Type = whatsapp.TextMessageType

	out.Text = in.GetText()

	matchedText := in.GetMatchedText()
	if len(matchedText) > 0 {
		out.Url = &whatsapp.WhatsappMessageUrl{
			Reference:   matchedText,
			Title:       in.GetTitle(),
			Description: in.GetDescription(),
		}

		thumbnail := in.GetJPEGThumbnail()
		if len(thumbnail) > 0 {
			out.Url.Thumbnail = whatsapp.NewWhatsappMessageThumbnail(thumbnail)
		}
	}

	info := in.GetContextInfo()
	if info != nil {
		out.ForwardingScore = info.GetForwardingScore()
		out.InReply = info.GetStanzaID()
	}

	// ads -------------------
	adreply := info.GetExternalAdReply()
	if adreply != nil {
		ads := &whatsapp.WhatsappMessageAds{
			Id:        adreply.GetCtwaClid(),
			Title:     adreply.GetTitle(),
			SourceId:  adreply.GetSourceID(),
			SourceUrl: adreply.GetSourceURL(),
			App:       adreply.GetSourceApp(),
			Type:      adreply.GetSourceType(),
		}

		thumbnail := adreply.GetThumbnail()
		if len(thumbnail) > 0 {
			ads.Thumbnail = whatsapp.NewWhatsappMessageThumbnail(thumbnail)
		}

		out.Ads = ads
	}
}

func HandleReactionMessage(log *log.Entry, out *whatsapp.WhatsappMessage, in *waE2E.ReactionMessage) {
	log.Debug("received a Reaction message!")

	out.Type = whatsapp.TextMessageType
	out.Text = in.GetText()
	out.InReply = in.Key.GetID()
}

//#endregion

func HandleButtonsResponseMessage(log *log.Entry, out *whatsapp.WhatsappMessage, in *waE2E.ButtonsResponseMessage) {
	log.Debug("received a buttons response message !")
	out.Type = whatsapp.TextMessageType

	/*
		b, err := json.Marshal(in)
		if err != nil {
			log.Error(err)
			return
		}
		log.Debug(string(b))
	*/

	out.Text = in.GetSelectedButtonID()

	info := in.ContextInfo
	if info != nil {
		out.ForwardingScore = info.GetForwardingScore()
		out.InReply = info.GetStanzaID()
	}
}

func HandleImageMessage(logentry *log.Entry, out *whatsapp.WhatsappMessage, in *waE2E.ImageMessage) {
	logentry.Debug("received an image message")
	out.Type = whatsapp.ImageMessageType

	// in case of caption passed
	out.Text = in.GetCaption()

	jpeg := GetStringFromBytes(in.JPEGThumbnail)
	out.Attachment = &whatsapp.WhatsappAttachment{
		CanDownload:   true,
		Mimetype:      in.GetMimetype(),
		FileLength:    in.GetFileLength(),
		JpegThumbnail: jpeg,
	}

	info := in.GetContextInfo()
	if info != nil {
		out.ForwardingScore = info.GetForwardingScore()
		out.InReply = info.GetStanzaID()
	}
}

func HandleStickerMessage(log *log.Entry, out *whatsapp.WhatsappMessage, in *waE2E.StickerMessage) {
	log.Debug("received a image|sticker message !")

	if in.GetIsAnimated() {
		out.Type = whatsapp.VideoMessageType
	} else {
		out.Type = whatsapp.ImageMessageType
	}

	jpeg := GetStringFromBytes(in.PngThumbnail)
	out.Attachment = &whatsapp.WhatsappAttachment{
		CanDownload: true,
		Mimetype:    in.GetMimetype(),
		FileLength:  in.GetFileLength(),

		JpegThumbnail: jpeg,
	}
}

func HandleVideoMessage(log *log.Entry, out *whatsapp.WhatsappMessage, in *waE2E.VideoMessage) {
	log.Debug("received a video message !")
	out.Type = whatsapp.VideoMessageType

	// in case of caption passed
	out.Text = in.GetCaption()

	jpeg := base64.StdEncoding.EncodeToString(in.JPEGThumbnail)
	out.Attachment = &whatsapp.WhatsappAttachment{
		CanDownload: true,
		Mimetype:    in.GetMimetype(),
		FileLength:  in.GetFileLength(),

		JpegThumbnail: jpeg,
	}

	info := in.ContextInfo
	if info != nil {
		out.ForwardingScore = info.GetForwardingScore()
		out.InReply = info.GetStanzaID()
	}
}

func HandleDocumentMessage(log *log.Entry, out *whatsapp.WhatsappMessage, in *waE2E.DocumentMessage) {
	log.Debug("received a document message !")
	out.Type = whatsapp.DocumentMessageType

	// in case of caption passed
	out.Text = in.GetCaption()

	jpeg := base64.StdEncoding.EncodeToString(in.JPEGThumbnail)
	out.Attachment = &whatsapp.WhatsappAttachment{
		CanDownload: true,
		Mimetype:    in.GetMimetype(),
		FileLength:  in.GetFileLength(),

		FileName:      in.GetFileName(),
		JpegThumbnail: jpeg,
	}

	info := in.ContextInfo
	if info != nil {
		out.ForwardingScore = info.GetForwardingScore()
		out.InReply = info.GetStanzaID()
	}
}

func HandleAudioMessage(log *log.Entry, out *whatsapp.WhatsappMessage, in *waE2E.AudioMessage) {
	log.Debug("received an audio message !")
	out.Type = whatsapp.AudioMessageType

	out.Attachment = &whatsapp.WhatsappAttachment{
		CanDownload: true,
		Mimetype:    in.GetMimetype(),
		FileLength:  in.GetFileLength(),
		Seconds:     in.GetSeconds(),
	}

	info := in.ContextInfo
	if info != nil {
		out.ForwardingScore = info.GetForwardingScore()
		out.InReply = info.GetStanzaID()
	}
}

func HandleLocationMessage(log *log.Entry, out *whatsapp.WhatsappMessage, in *waE2E.LocationMessage) {
	log.Debug("received a Location message !")
	out.Type = whatsapp.LocationMessageType

	// in a near future, create a environment variable for that
	defaultUrl := "https://www.google.com/maps?ll={lat},{lon}&q={lat}+{lon}"

	defaultUrl = strings.Replace(defaultUrl, "{lat}", fmt.Sprintf("%f", *in.DegreesLatitude), -1)
	defaultUrl = strings.Replace(defaultUrl, "{lon}", fmt.Sprintf("%f", *in.DegreesLongitude), -1)

	filename := fmt.Sprintf("%f_%f", in.GetDegreesLatitude(), in.GetDegreesLongitude())
	filename = fmt.Sprintf("%s.url", slug.Make(filename))

	content := []byte("[InternetShortcut]\nURL=" + defaultUrl)
	length := uint64(len(content))
	jpeg := base64.StdEncoding.EncodeToString(in.JPEGThumbnail)

	out.Attachment = &whatsapp.WhatsappAttachment{
		CanDownload:   false,
		Mimetype:      "text/x-uri; location",
		Latitude:      in.GetDegreesLatitude(),
		Longitude:     in.GetDegreesLongitude(),
		JpegThumbnail: jpeg,
		Url:           defaultUrl,
		FileName:      filename,
		FileLength:    length,
	}

	out.Attachment.SetContent(&content)
}

func HandleLiveLocationMessage(log *log.Entry, out *whatsapp.WhatsappMessage, in *waE2E.LiveLocationMessage) {
	log.Debug("received a Live Location message !")
	out.Type = whatsapp.LocationMessageType

	// in case of caption passed
	out.Text = in.GetCaption()

	// in a near future, create a environment variable for that
	defaultUrl := "https://www.google.com/maps?ll={lat},{lon}&q={lat}+{lon}"

	defaultUrl = strings.Replace(defaultUrl, "{lat}", fmt.Sprintf("%f", *in.DegreesLatitude), -1)
	defaultUrl = strings.Replace(defaultUrl, "{lon}", fmt.Sprintf("%f", *in.DegreesLongitude), -1)

	filename := out.Text
	if len(filename) == 0 {
		filename = fmt.Sprintf("%f_%f", *in.DegreesLatitude, *in.DegreesLongitude)
	}
	filename = fmt.Sprintf("%s.url", slug.Make(filename))

	content := []byte("[InternetShortcut]\nURL=" + defaultUrl)
	length := uint64(len(content))
	jpeg := base64.StdEncoding.EncodeToString(in.JPEGThumbnail)

	out.Attachment = &whatsapp.WhatsappAttachment{
		CanDownload:   false,
		Mimetype:      "text/x-uri; live location",
		Latitude:      in.GetDegreesLatitude(),
		Longitude:     in.GetDegreesLongitude(),
		Sequence:      in.GetSequenceNumber(),
		JpegThumbnail: jpeg,
		Url:           defaultUrl,
		FileName:      filename,
		FileLength:    length,
	}

	out.Attachment.SetContent(&content)
}

func HandleContactMessage(log *log.Entry, out *whatsapp.WhatsappMessage, in *waE2E.ContactMessage) {
	log.Debug("received a contact message !")
	out.Type = whatsapp.ContactMessageType

	out.Text = in.GetDisplayName()
	filename := out.Text
	if len(filename) == 0 {
		filename = out.Id
	}
	filename = fmt.Sprintf("%s.vcf", slug.Make(filename))

	content := []byte(in.GetVcard())
	out.Attachment = whatsapp.GenerateVCardAttachment(content, filename)
}
