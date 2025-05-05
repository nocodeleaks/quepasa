package whatsmeow

import (
	"encoding/base64"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/nocodeleaks/quepasa/library"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
	whatsmeow "go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	types "go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

func GetMediaTypeFromAttachment(source *whatsapp.WhatsappAttachment) whatsmeow.MediaType {
	msgType := whatsapp.GetMessageType(source)
	return GetMediaTypeFromWAMsgType(msgType)
}

// Traz o MediaType para download do whatsapp
func GetMediaTypeFromWAMsgType(msgType whatsapp.WhatsappMessageType) whatsmeow.MediaType {
	switch msgType {
	case whatsapp.ImageMessageType:
		return whatsmeow.MediaImage
	case whatsapp.AudioMessageType:
		return whatsmeow.MediaAudio
	case whatsapp.VideoMessageType:
		return whatsmeow.MediaVideo
	default:
		return whatsmeow.MediaDocument
	}
}

func ToWhatsmeowMessage(source whatsapp.IWhatsappMessage) (msg *waE2E.Message, err error) {
	messageText := source.GetText()

	if !source.HasAttachment() {
		internal := &waE2E.ExtendedTextMessage{Text: &messageText}
		msg = &waE2E.Message{ExtendedTextMessage: internal}
	}

	return
}

func NewWhatsmeowMessageAttachment(response whatsmeow.UploadResponse, waMsg whatsapp.WhatsappMessage, media whatsmeow.MediaType, inreplycontext *waE2E.ContextInfo) (msg *waE2E.Message) {
	attach := waMsg.Attachment

	var seconds *uint32
	if attach.Seconds > 0 {
		seconds = proto.Uint32(attach.Seconds)
	}

	var mimetype *string
	if len(attach.Mimetype) > 0 {
		mimetype = proto.String(attach.Mimetype)
	}

	switch media {
	case whatsmeow.MediaImage:
		internal := &waE2E.ImageMessage{
			URL:           proto.String(response.URL),
			DirectPath:    proto.String(response.DirectPath),
			MediaKey:      response.MediaKey,
			FileEncSHA256: response.FileEncSHA256,
			FileSHA256:    response.FileSHA256,
			FileLength:    proto.Uint64(response.FileLength),
			Mimetype:      mimetype,
			Caption:       proto.String(waMsg.Text),
			ContextInfo:   inreplycontext,
		}
		msg = &waE2E.Message{ImageMessage: internal}
		return
	case whatsmeow.MediaAudio:

		var ptt *bool
		if attach.IsValidPTT() {
			ptt = proto.Bool(true)
		} else if attach.IsPTTCompatible() { // trick to send audio as ptt, "technical resource"
			ptt = proto.Bool(true)
			mimetype = proto.String(whatsapp.WhatsappPTTMime)
		}

		internal := &waE2E.AudioMessage{
			URL:           proto.String(response.URL),
			DirectPath:    proto.String(response.DirectPath),
			MediaKey:      response.MediaKey,
			FileEncSHA256: response.FileEncSHA256,
			FileSHA256:    response.FileSHA256,
			FileLength:    proto.Uint64(response.FileLength),
			Seconds:       seconds,
			Mimetype:      mimetype,
			PTT:           ptt,
			Waveform:      attach.WaveForm,
			ContextInfo:   inreplycontext,
		}
		msg = &waE2E.Message{AudioMessage: internal}
		return
	case whatsmeow.MediaVideo:
		internal := &waE2E.VideoMessage{
			URL:           proto.String(response.URL),
			DirectPath:    proto.String(response.DirectPath),
			MediaKey:      response.MediaKey,
			FileEncSHA256: response.FileEncSHA256,
			FileSHA256:    response.FileSHA256,
			FileLength:    proto.Uint64(response.FileLength),
			Seconds:       seconds,
			Mimetype:      mimetype,
			Caption:       proto.String(waMsg.Text),
			ContextInfo:   inreplycontext,
		}
		msg = &waE2E.Message{VideoMessage: internal}
		return
	default:
		internal := &waE2E.DocumentMessage{
			URL:           proto.String(response.URL),
			DirectPath:    proto.String(response.DirectPath),
			MediaKey:      response.MediaKey,
			FileEncSHA256: response.FileEncSHA256,
			FileSHA256:    response.FileSHA256,
			FileLength:    proto.Uint64(response.FileLength),

			Mimetype:    mimetype,
			FileName:    proto.String(attach.FileName),
			Caption:     proto.String(waMsg.Text),
			ContextInfo: inreplycontext,
		}
		msg = &waE2E.Message{DocumentMessage: internal}
		return
	}
}

func GetStringFromBytes(bytes []byte) string {
	if len(bytes) > 0 {
		return base64.StdEncoding.EncodeToString(bytes)
	}
	return ""
}

// should implement a cache !!! urgent
// returns a valid chat title from local memory store
func GetChatTitle(client *whatsmeow.Client, jid types.JID) (title string) {
	if jid.Server == "g.us" {

		title = GroupInfoCache.Get(jid.String())
		if len(title) > 0 {
			goto found
		}

		// fmt.Printf("getting group info: %s", jid.String())
		gInfo, _ := client.GetGroupInfo(jid)
		if gInfo != nil {
			title = library.NormalizeForTitle(gInfo.Name)
			_ = GroupInfoCache.Append(jid.String(), title, "GetChatTitle")
			goto found
		}
	} else {
		cInfo, _ := client.Store.Contacts.GetContact(jid)
		if cInfo.Found {
			if len(cInfo.BusinessName) > 0 {
				title = cInfo.BusinessName
				goto found
			} else if len(cInfo.FullName) > 0 {
				title = cInfo.FullName
				goto found
			}

			title = cInfo.PushName
			goto found
		}
	}

	return ""

found:
	return library.NormalizeForTitle(title)
}

/*
<summary>

	Send defined presence when connecting and when the pushname is changed.
	This makes sure that outgoing messages always have the right pushname.

<summary/>
*/
func SendPresence(client *whatsmeow.Client, presence types.Presence, from string, logentry *log.Entry) {
	if len(client.Store.PushName) > 0 {
		err := client.SendPresence(presence)
		if err != nil {
			logentry.Warnf("failed to send presence: '%s', error: %s, from: %s", presence, err.Error(), from)
		} else {
			logentry.Debugf("marked self as '%s', from: %s", presence, from)
		}
	}
}

func GetWhatsappMessageStatus(receipt types.ReceiptType) whatsapp.WhatsappMessageStatus {
	switch receipt {
	case types.ReceiptTypeDelivered:
		return whatsapp.WhatsappMessageStatusDelivered
	case types.ReceiptTypeRetry, types.ReceiptTypeServerError:
		return whatsapp.WhatsappMessageStatusError
	case types.ReceiptTypeRead, types.ReceiptTypePlayed:
		return whatsapp.WhatsappMessageStatusRead
	}

	return whatsapp.WhatsappMessageStatusUnknown
}

func ImproveTimestamp(evtTimestamp time.Time) time.Time {

	if evtTimestamp.Nanosecond() == 0 {

		now := time.Now()
		if evtTimestamp.Second() == now.Second() {

			nanos := time.Now().Nanosecond()
			currentNanosecond := time.Duration(nanos)
			duration := currentNanosecond * time.Nanosecond
			return evtTimestamp.Add(duration)
		}
	}

	return evtTimestamp
}
