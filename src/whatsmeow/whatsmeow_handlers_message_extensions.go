package whatsmeow

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	slug "github.com/gosimple/slug"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
	proto "go.mau.fi/whatsmeow/binary/proto"
)

func HandleKnowingMessages(handler *WhatsmeowHandlers, out *whatsapp.WhatsappMessage, in *proto.Message) {
	log.Tracef("handling knowing message: %v", in)
	if in.ImageMessage != nil {
		HandleImageMessage(handler.log, out, in.ImageMessage)
	} else if in.StickerMessage != nil {
		HandleStickerMessage(handler.log, out, in.StickerMessage)
	} else if in.DocumentMessage != nil {
		HandleDocumentMessage(handler.log, out, in.DocumentMessage)
	} else if in.AudioMessage != nil {
		HandleAudioMessage(handler.log, out, in.AudioMessage)
	} else if in.VideoMessage != nil {
		HandleVideoMessage(handler.log, out, in.VideoMessage)
	} else if in.ExtendedTextMessage != nil {
		HandleExtendedTextMessage(handler.log, out, in.ExtendedTextMessage)
	} else if in.ButtonsResponseMessage != nil {
		HandleButtonsResponseMessage(handler.log, out, in.ButtonsResponseMessage)
	} else if in.LocationMessage != nil {
		HandleLocationMessage(handler.log, out, in.LocationMessage)
	} else if in.LiveLocationMessage != nil {
		HandleLiveLocationMessage(handler.log, out, in.LiveLocationMessage)
	} else if in.ContactMessage != nil {
		HandleContactMessage(handler.log, out, in.ContactMessage)
	} else if in.ReactionMessage != nil {
		HandleReactionMessage(handler.log, out, in.ReactionMessage)
	} else if in.ProtocolMessage != nil || in.SenderKeyDistributionMessage != nil {
		out.Type = whatsapp.DiscardMessageType
	} else if len(in.GetConversation()) > 0 {
		HandleTextMessage(handler.log, out, in)
	} else {
		log.Warnf("message not threated: %v", in)
	}
}

func HandleUnknownMessage(log *log.Entry, in interface{}) {
	log.Info("Received an unknown message !")
	b, err := json.Marshal(in)
	if err != nil {
		log.Error(err)
		return
	}
	log.Debug(string(b))
}

//#region HANDLING TEXT MESSAGES

func HandleTextMessage(log *log.Entry, out *whatsapp.WhatsappMessage, in *proto.Message) {
	log.Debug("Received a text message !")
	out.Type = whatsapp.TextMessageType
	out.Text = in.GetConversation()
}

// Msg em resposta a outra
func HandleExtendedTextMessage(log *log.Entry, out *whatsapp.WhatsappMessage, in *proto.ExtendedTextMessage) {
	log.Debug("Received a text|extended message !")
	out.Type = whatsapp.TextMessageType

	out.Text = in.GetText()

	info := in.ContextInfo
	if info != nil {
		out.ForwardingScore = info.GetForwardingScore()
		out.InReply = info.GetStanzaId()
	}
}

func HandleReactionMessage(log *log.Entry, out *whatsapp.WhatsappMessage, in *proto.ReactionMessage) {
	log.Debug("Received a Reaction message!")

	out.Type = whatsapp.TextMessageType
	out.Text = in.GetText()
	out.InReply = in.Key.GetId()
}

//#endregion

func HandleButtonsResponseMessage(log *log.Entry, out *whatsapp.WhatsappMessage, in *proto.ButtonsResponseMessage) {
	log.Debug("Received a buttons response message !")
	out.Type = whatsapp.TextMessageType

	/*
		b, err := json.Marshal(in)
		if err != nil {
			log.Error(err)
			return
		}
		log.Debug(string(b))
	*/

	out.Text = in.GetSelectedButtonId()

	info := in.ContextInfo
	if info != nil {
		if info.ForwardingScore != nil {
			out.ForwardingScore = *info.ForwardingScore
		}

		if info.StanzaId != nil {
			out.InReply = *info.StanzaId
		}
	}
}

func HandleImageMessage(log *log.Entry, out *whatsapp.WhatsappMessage, in *proto.ImageMessage) {
	log.Debug("Received an image message !")
	out.Content = in
	out.Type = whatsapp.ImageMessageType

	// in case of caption passed
	out.Text = in.GetCaption()

	jpeg := GetStringFromBytes(in.JpegThumbnail)
	out.Attachment = &whatsapp.WhatsappAttachment{
		Mimetype:      *in.Mimetype,
		FileLength:    *in.FileLength,
		JpegThumbnail: jpeg,
	}
}

func HandleStickerMessage(log *log.Entry, out *whatsapp.WhatsappMessage, in *proto.StickerMessage) {
	log.Debug("Received a image|sticker message !")
	out.Content = in

	if in.GetIsAnimated() {
		out.Type = whatsapp.VideoMessageType
	} else {
		out.Type = whatsapp.ImageMessageType
	}

	jpeg := GetStringFromBytes(in.PngThumbnail)
	out.Attachment = &whatsapp.WhatsappAttachment{
		Mimetype:   *in.Mimetype,
		FileLength: *in.FileLength,

		JpegThumbnail: jpeg,
	}
}

func HandleVideoMessage(log *log.Entry, out *whatsapp.WhatsappMessage, in *proto.VideoMessage) {
	log.Debug("Received a video message !")
	out.Content = in
	out.Type = whatsapp.VideoMessageType

	// in case of caption passed
	if in.Caption != nil {
		out.Text = *in.Caption
	}

	jpeg := base64.StdEncoding.EncodeToString(in.JpegThumbnail)
	out.Attachment = &whatsapp.WhatsappAttachment{
		Mimetype:   *in.Mimetype,
		FileLength: *in.FileLength,

		JpegThumbnail: jpeg,
	}
}

func HandleDocumentMessage(log *log.Entry, out *whatsapp.WhatsappMessage, in *proto.DocumentMessage) {
	log.Debug("Received a document message !")
	out.Content = in
	out.Type = whatsapp.DocumentMessageType

	if in.Title != nil {
		out.Text = *in.Title
	}

	jpeg := base64.StdEncoding.EncodeToString(in.JpegThumbnail)
	out.Attachment = &whatsapp.WhatsappAttachment{
		Mimetype:   *in.Mimetype + "; wa-document",
		FileLength: *in.FileLength,

		FileName:      *in.FileName,
		JpegThumbnail: jpeg,
	}
}

func HandleAudioMessage(log *log.Entry, out *whatsapp.WhatsappMessage, in *proto.AudioMessage) {
	log.Debug("Received an audio message !")
	out.Content = in
	out.Type = whatsapp.AudioMessageType

	var seconds uint32
	if in.Seconds != nil {
		seconds = *in.Seconds
	}

	out.Attachment = &whatsapp.WhatsappAttachment{
		Mimetype:   *in.Mimetype,
		FileLength: *in.FileLength,

		Seconds: seconds,
	}
}

func HandleLocationMessage(log *log.Entry, out *whatsapp.WhatsappMessage, in *proto.LocationMessage) {
	log.Debug("Received a Location message !")
	out.Content = in
	out.Type = whatsapp.LocationMessageType

	// in a near future, create a enviroment variavel for that
	defaultUrl := "https://www.google.com/maps?ll={lat},{lon}&q={lat}+{lon}"

	defaultUrl = strings.Replace(defaultUrl, "{lat}", fmt.Sprintf("%f", *in.DegreesLatitude), -1)
	defaultUrl = strings.Replace(defaultUrl, "{lon}", fmt.Sprintf("%f", *in.DegreesLongitude), -1)

	filename := fmt.Sprintf("%f_%f", *in.DegreesLatitude, *in.DegreesLongitude)
	filename = fmt.Sprintf("%s.url", slug.Make(filename))

	content := []byte("[InternetShortcut]\nURL=" + defaultUrl)
	length := uint64(len(content))
	jpeg := base64.StdEncoding.EncodeToString(in.JpegThumbnail)

	out.Attachment = &whatsapp.WhatsappAttachment{
		Mimetype:      "text/x-uri; location",
		Latitude:      *in.DegreesLatitude,
		Longitude:     *in.DegreesLongitude,
		JpegThumbnail: jpeg,
		Url:           defaultUrl,
		FileName:      filename,
		FileLength:    length,
	}

	out.Attachment.SetContent(&content)
}

func HandleLiveLocationMessage(log *log.Entry, out *whatsapp.WhatsappMessage, in *proto.LiveLocationMessage) {
	log.Debug("Received a Live Location message !")
	out.Content = in
	out.Type = whatsapp.LocationMessageType

	// in a near future, create a enviroment variavel for that
	defaultUrl := "https://www.google.com/maps?ll={lat},{lon}&q={lat}+{lon}"

	defaultUrl = strings.Replace(defaultUrl, "{lat}", fmt.Sprintf("%f", *in.DegreesLatitude), -1)
	defaultUrl = strings.Replace(defaultUrl, "{lon}", fmt.Sprintf("%f", *in.DegreesLongitude), -1)

	if in.Caption != nil {
		out.Text = *in.Caption
	}

	filename := out.Text
	if len(filename) == 0 {
		filename = fmt.Sprintf("%f_%f", *in.DegreesLatitude, *in.DegreesLongitude)
	}
	filename = fmt.Sprintf("%s.url", slug.Make(filename))

	content := []byte("[InternetShortcut]\nURL=" + defaultUrl)
	length := uint64(len(content))
	jpeg := base64.StdEncoding.EncodeToString(in.JpegThumbnail)

	out.Attachment = &whatsapp.WhatsappAttachment{
		Mimetype:      "text/x-uri; live location",
		Latitude:      *in.DegreesLatitude,
		Longitude:     *in.DegreesLongitude,
		Sequence:      *in.SequenceNumber,
		JpegThumbnail: jpeg,
		Url:           defaultUrl,
		FileName:      filename,
		FileLength:    length,
	}

	out.Attachment.SetContent(&content)
}

func HandleContactMessage(log *log.Entry, out *whatsapp.WhatsappMessage, in *proto.ContactMessage) {
	log.Debug("Received a Contact message !")
	out.Content = in
	out.Type = whatsapp.ContactMessageType

	out.Text = *in.DisplayName
	filename := *in.DisplayName
	if len(filename) == 0 {
		filename = out.Id
	}
	filename = fmt.Sprintf("%s.vcf", slug.Make(filename))

	content := []byte(*in.Vcard)
	length := uint64(len(content))

	out.Attachment = &whatsapp.WhatsappAttachment{
		Mimetype:   "text/x-vcard",
		FileName:   filename,
		FileLength: length,
	}

	out.Attachment.SetContent(&content)
}
