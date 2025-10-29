package whatsmeow

import (
	"context"
	"encoding/base64"
	"reflect"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
	whatsmeow "go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	types "go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

/**
 * GetMediaTypeFromAttachment returns the whatsmeow.MediaType for a given WhatsappAttachment.
 *
 * It uses the attachment to determine the message type and maps it to the corresponding MediaType.
 *
 * @param source WhatsappAttachment to check
 * @return whatsmeow.MediaType for the attachment
 */
func GetMediaTypeFromAttachment(source *whatsapp.WhatsappAttachment) whatsmeow.MediaType {
	msgType := whatsapp.GetMessageType(source)
	return GetMediaTypeFromWAMsgType(msgType)
}

/**
 * GetMediaTypeFromWAMsgType maps a WhatsappMessageType to a whatsmeow.MediaType.
 *
 * @param msgType WhatsappMessageType to map
 * @return whatsmeow.MediaType corresponding to the message type
 */
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

/**
 * ToWhatsmeowMessage converts a generic IWhatsappMessage to a waE2E.Message.
 *
 * If the source does not have an attachment, it creates an ExtendedTextMessage.
 * Otherwise, returns nil and error (not implemented for attachments).
 *
 * @param source IWhatsappMessage to convert
 * @return waE2E.Message pointer and error
 */
func ToWhatsmeowMessage(source whatsapp.IWhatsappMessage) (msg *waE2E.Message, err error) {
	messageText := source.GetText()

	if !source.HasAttachment() {
		internal := &waE2E.ExtendedTextMessage{Text: &messageText}
		msg = &waE2E.Message{ExtendedTextMessage: internal}
	}

	return
}

/**
 * NewWhatsmeowMessageAttachment creates a new waE2E.Message with the correct media type and metadata.
 *
 * It builds the internal message (Image, Audio, Video, Document) using the upload response and WhatsappMessage data.
 *
 * @param response UploadResponse containing media upload info
 * @param waMsg WhatsappMessage containing attachment and text
 * @param media MediaType to use (image, audio, video, document)
 * @param inreplycontext Optional context info for replies
 * @return waE2E.Message pointer with the correct media type
 */
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
		} else if attach.IsPTTCompatible() {
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

/**
 * GetStringFromBytes encodes a byte slice to a base64 string.
 *
 * Returns an empty string if the input is empty.
 *
 * @param bytes Byte slice to encode
 * @return Base64 string representation
 */
func GetStringFromBytes(bytes []byte) string {
	if len(bytes) > 0 {
		return base64.StdEncoding.EncodeToString(bytes)
	}
	return ""
}

/*
<summary>

	Send defined presence when connecting and when the pushname is changed.
	This makes sure that outgoing messages always have the right pushname.

<summary/>
*/
/**
 * SendPresence sends the defined presence to WhatsApp when connecting or when the pushname changes.
 *
 * Ensures outgoing messages have the correct pushname. Logs success or failure.
 *
 * @param client Whatsmeow client instance
 * @param presence Presence type to send
 * @param from Source identifier for logging
 * @param logentry Logger instance
 */
func SendPresence(client *whatsmeow.Client, presence types.Presence, from string, logentry *log.Entry) {
	if len(client.Store.PushName) > 0 {
		err := client.SendPresence(context.Background(), presence)
		if err != nil {
			logentry.Warnf("failed to send presence: '%s', error: %s, from: %s", presence, err.Error(), from)
		} else {
			logentry.Debugf("marked self as '%s', from: %s", presence, from)
		}
	}
}

/**
 * GetWhatsappMessageStatus maps a WhatsApp receipt type to a WhatsappMessageStatus value.
 *
 * @param receipt ReceiptType from WhatsApp
 * @return WhatsappMessageStatus corresponding to the receipt
 */
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

/**
 * ImproveTimestamp adds nanoseconds to a timestamp if it has zero nanoseconds and matches the current second.
 *
 * This helps avoid duplicate timestamps in high-frequency events.
 *
 * @param evtTimestamp Time to improve
 * @return Improved time.Time value
 */
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

/**
 * GetDownloadableMessage returns the first message inside a waE2E.Message that implements the DownloadableMessage interface.
 *
 * This function checks each possible message type (ImageMessage, AudioMessage, VideoMessage, DocumentMessage, StickerMessage)
 * and returns the first one that implements whatsmeow.DownloadableMessage. If none are found, it returns nil.
 *
 * Usage:
 *   dm := GetDownloadableMessage(msg)
 *   if dm != nil {
 *       // You can now use dm to download media
 *   }
 *
 * This is useful for generic media handling when you don't know the exact type in advance.
 */
func GetDownloadableMessage(msg *waE2E.Message) whatsmeow.DownloadableMessage {
	// Check each possible message type for DownloadableMessage interface
	if msg == nil {
		return nil
	}

	if msg.ImageMessage != nil {
		return msg.ImageMessage
	}

	if msg.AudioMessage != nil {
		return msg.AudioMessage
	}

	if msg.VideoMessage != nil {
		return msg.VideoMessage
	}

	if msg.StickerMessage != nil {
		return msg.StickerMessage
	}

	if msg.DocumentMessage != nil {
		return msg.DocumentMessage
	}

	return nil
}

/**
 * RemoveMessageContextInfo removes the MessageContextInfo field from a WhatsApp message struct.
 *
 * This function uses reflection to set the MessageContextInfo field to its zero value (nil),
 * effectively removing any context information (such como reply, quoted, etc) que estava presente na mensagem.
 *
 * Útil para sanitizar mensagens antes de processar ou enviar, evitando que informações de contexto sejam propagadas.
 *
 * @param content WhatsApp message struct (pointer)
 * @return The same struct with MessageContextInfo removed (set to zero value)
 */
func RemoveMessageContextInfo(content any) any {
	if content == nil {
		return content
	}

	// Use reflection to check if content has messageContextInfo field
	if reflect.TypeOf(content).Kind() == reflect.Ptr {
		elem := reflect.ValueOf(content).Elem()
		if elem.IsValid() && elem.Kind() == reflect.Struct {
			// Check if the struct has a messageContextInfo field
			field := elem.FieldByName("MessageContextInfo")
			if field.IsValid() && field.CanSet() {
				// Remove context info from the message
				field.Set(reflect.Zero(field.Type()))
			}
		}
	}

	return content
}

/**
 * GetMessageEventType extracts the actual message type from a waE2E.Message protobuf struct.
 *
 * It inspects all fields of the message and returns the name of the first non-nil pointer field,
 * converted to snake_case (protobuf convention). If no field is set, returns "unknown".
 *
 * @param in Pointer to waE2E.Message
 * @return String representing the message type in snake_case, or "unknown"
 */
func GetMessageEventType(in *waE2E.Message) string {
	v := reflect.ValueOf(in).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() == reflect.Ptr && !field.IsNil() {
			fieldName := t.Field(i).Name
			// Convert from Go field name to protobuf field name format
			return toSnakeCase(fieldName)
		}
	}

	return "unknown"
}

/**
 * toSnakeCase converts Go field names to protobuf field naming convention (snake_case).
 *
 * For example, "ImageMessage" becomes "image_message".
 *
 * @param s Go field name string
 * @return String in snake_case format
 */
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

func PhoneToJID(source string) (types.JID, error) {
	wid := whatsapp.PhoneToWid(source)
	return types.ParseJID(wid)
}
