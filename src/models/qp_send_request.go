package models

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	library "github.com/nocodeleaks/quepasa/library"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
)

type QpSendRequest struct {
	// (Optional) Used if passed
	Id string `json:"id,omitempty"`

	// Recipient of this message
	ChatId string `json:"chatid"`

	// (Optional) TrackId - less priority (urlparam -> query -> header -> body)
	TrackId string `json:"trackid,omitempty"`

	Text string `json:"text,omitempty"`

	// Msg in reply of another ? Message ID
	InReply string `json:"inreply,omitempty"`

	// (Optional) Sugested filename on user download
	FileName string `json:"filename,omitempty"`

	// (Optional) important to navigate throw content
	FileLength uint64 `json:"filelength,omitempty"`

	// (Optional) mime type to facilitate correct delivery
	Mimetype string `json:"mime,omitempty"`

	// (Optional) time in seconds for audio/video contents
	Seconds uint32 `json:"seconds,omitempty"`

	Content []byte

	TypingDuration int    `json:"typing_duration,omitempty"` // How long to show typing (ms)
	MediaType      string `json:"media_type,omitempty"`      // For audio recording indicator

	Poll     *whatsapp.WhatsappPoll     `json:"poll,omitempty"`     // Poll if exists
	Location *whatsapp.WhatsappLocation `json:"location,omitempty"` // Location if exists
	Contact  *whatsapp.WhatsappContact  `json:"contact,omitempty"`  // Contact if exists
}

// get default log entry, never nil
func (source *QpSendRequest) GetLogger() *log.Entry {
	logentry := log.WithContext(context.Background())
	logentry = logentry.WithField(LogFields.ChatId, source.ChatId)

	return logentry
}

func (source *QpSendRequest) EnsureChatId(r *http.Request) (err error) {

	// already set ?
	if len(source.ChatId) > 0 {
		return
	}

	source.ChatId = library.GetChatId(r)
	if len(source.ChatId) == 0 {
		err = fmt.Errorf("chat id missing")
	}
	return
}

func (source *QpSendRequest) EnsureValidChatId(r *http.Request) (err error) {
	err = source.EnsureChatId(r)
	if err != nil {
		return
	}

	chatid, err := whatsapp.FormatEndpoint(source.ChatId)
	if err != nil {
		return
	}

	source.ChatId = chatid
	return
}

func (source *QpSendRequest) ToWhatsappMessage() (msg *whatsapp.WhatsappMessage, err error) {
	chatId, err := whatsapp.FormatEndpoint(source.ChatId)
	if err != nil {
		return
	}

	chat := whatsapp.WhatsappChat{Id: chatId}

	// Try to populate phone field
	if phone, _ := whatsapp.GetPhoneIfValid(chatId); len(phone) > 0 {
		chat.Phone = phone
	}

	msg = &whatsapp.WhatsappMessage{
		Id:           strings.ToUpper(source.Id), // dont know why, must be upper
		TrackId:      source.TrackId,
		InReply:      source.InReply,
		Text:         source.Text,
		Chat:         chat,
		FromMe:       true,
		FromInternal: true,
	}

	msg.Poll = source.Poll

	// Check if this is a contact message
	if source.Contact != nil {
		msg.Type = whatsapp.ContactMessageType
		// Store contact data in message
		msg.Contact = source.Contact
		return
	}

	// Check if this is a location message
	if source.Location != nil {
		msg.Type = whatsapp.LocationMessageType
		// Create attachment with location data
		msg.Attachment = &whatsapp.WhatsappAttachment{
			Latitude:  source.Location.Latitude,
			Longitude: source.Location.Longitude,
			Mimetype:  "text/x-uri; location",
		}
		// Use location name as text if provided
		if len(source.Location.Name) > 0 {
			msg.Text = source.Location.Name
		}
		return
	}

	// Only set text type if no location was provided
	if len(msg.Text) > 0 {
		// setting default type for text messages
		msg.Type = whatsapp.TextMessageType
	} else {
		// No location, no text - set as text type by default
		msg.Type = whatsapp.TextMessageType
	}

	return
}

func (source *QpSendRequest) ToWhatsappAttachment() (result QpToWhatsappAttachment) {
	contentLength := len(source.Content)
	if contentLength == 0 {
		return
	}

	logentry := source.GetLogger()

	attach := &whatsapp.WhatsappAttachment{
		Mimetype:   source.Mimetype,
		FileLength: source.FileLength,
		FileName:   source.FileName,
		Seconds:    source.Seconds,
	}

	// validating content length
	uIntContentLength := uint64(contentLength)
	if attach.FileLength != uIntContentLength {
		attach.FileLength = uIntContentLength

		warn := fmt.Sprintf("invalid attachment length, request length: %v != content length: %v, revalidating for security", attach.FileLength, contentLength)
		result.Debug = append(result.Debug, "[warn][ToWhatsappAttachment] "+warn)
		logentry.Warnf("%s", warn)
	}

	// end source use and set content
	attach.SetContent(&source.Content)

	result.Attach = attach
	result.AttachSecureAndCustomize()
	result.AttachImageTreatment()
	result.AttachAudioTreatment()

	return
}

// From "body" content (sendbinary)
func (source *QpSendRequest) GenerateBodyContent(r *http.Request) (err error) {
	content, err := io.ReadAll(r.Body)
	if err != nil {
		return
	}

	source.Content = content
	source.Mimetype = r.Header.Get("Content-Type")

	InformedLength := r.Header.Get("Content-Length")
	if len(InformedLength) > 0 {
		length, err := strconv.ParseUint(InformedLength, 10, 64)
		if err == nil {
			source.FileLength = length
		}
	}

	source.FileName = library.GetFileName(r)
	return
}
