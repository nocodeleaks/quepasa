package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	library "github.com/nocodeleaks/quepasa/library"
	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
)

// SendRequest is the HTTP transport contract used by send endpoints after the
// request has been parsed from query/body input.
//
// It lives in the API boundary because its field names, body parsing helpers,
// and fallback rules are specific to HTTP request handling.
type SendRequest struct {
	// Optional client-supplied id.
	Id string `json:"id,omitempty"`

	// Recipient of this message.
	ChatId string `json:"chatid"`

	// Optional track id used to correlate outbound messages.
	TrackId string `json:"trackid,omitempty"`

	Text string `json:"text,omitempty"`

	// Message id this outbound message is replying to.
	InReply string `json:"inreply,omitempty"`

	// Suggested filename shown to the recipient when relevant.
	FileName string `json:"filename,omitempty"`

	// Declared content length, revalidated against decoded content.
	FileLength uint64 `json:"filelength,omitempty"`

	// MIME type used to classify outbound attachments.
	Mimetype string `json:"mime,omitempty"`

	// Media duration for audio/video attachments when known.
	Seconds uint32 `json:"seconds,omitempty"`

	// Binary content resolved from request body, base64, or remote URL.
	Content []byte

	TypingDuration int    `json:"typing_duration,omitempty"` // How long to show typing (ms).
	MediaType      string `json:"media_type,omitempty"`      // Recording indicator for some media.

	Poll     *whatsapp.WhatsappPoll     `json:"poll,omitempty"`     // Poll payload when present.
	Location *whatsapp.WhatsappLocation `json:"location,omitempty"` // Location payload when present.
	Contact  *whatsapp.WhatsappContact  `json:"contact,omitempty"`  // Contact payload when present.
}

// SendAnyRequest extends SendRequest with the additional HTTP fields accepted
// by `/send` and `/senddocument` for attachment sources.
type SendAnyRequest struct {
	SendRequest

	// Public URL downloaded by the server before sending.
	Url string `json:"url,omitempty"`

	// Base64-encoded or data-URI encoded payload.
	Content string `json:"content,omitempty"`
}

// GetLogger returns a request-scoped logger with chat id context attached.
func (source *SendRequest) GetLogger() *log.Entry {
	logentry := log.WithContext(context.Background())
	logentry = logentry.WithField(library.LogFields.ChatId, source.ChatId)
	return logentry
}

// EnsureChatId populates the chat id from the HTTP request when it was omitted
// from the JSON payload.
func (source *SendRequest) EnsureChatId(r *http.Request) (err error) {
	if len(source.ChatId) > 0 {
		return
	}

	source.ChatId = library.GetChatId(r)
	if len(source.ChatId) == 0 {
		err = fmt.Errorf("chat id missing")
	}
	return
}

// EnsureValidChatId resolves and normalizes the final destination JID.
func (source *SendRequest) EnsureValidChatId(r *http.Request) (err error) {
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

// ToWhatsappMessage projects the API transport request into the internal
// WhatsappMessage model accepted by the runtime service layer.
func (source *SendRequest) ToWhatsappMessage() (msg *whatsapp.WhatsappMessage, err error) {
	chatId, err := whatsapp.FormatEndpoint(source.ChatId)
	if err != nil {
		return
	}

	chat := whatsapp.WhatsappChat{Id: chatId}
	if phone, _ := whatsapp.GetPhoneIfValid(chatId); len(phone) > 0 {
		chat.Phone = phone
	}

	msg = &whatsapp.WhatsappMessage{
		Id:           strings.ToUpper(source.Id),
		TrackId:      source.TrackId,
		InReply:      source.InReply,
		Text:         source.Text,
		Chat:         chat,
		FromMe:       true,
		FromInternal: true,
	}

	msg.Poll = source.Poll

	if source.Contact != nil {
		msg.Type = whatsapp.ContactMessageType
		msg.Contact = source.Contact
		return
	}

	if source.Location != nil {
		msg.Type = whatsapp.LocationMessageType
		msg.Attachment = &whatsapp.WhatsappAttachment{
			Latitude:  source.Location.Latitude,
			Longitude: source.Location.Longitude,
			Mimetype:  "text/x-uri; location",
		}
		if len(source.Location.Name) > 0 {
			msg.Text = source.Location.Name
		}
		return
	}

	if len(msg.Text) > 0 {
		msg.Type = whatsapp.TextMessageType
	} else {
		msg.Type = whatsapp.TextMessageType
	}

	return
}

// ToWhatsappAttachment builds and hardens the optional outbound attachment from
// the resolved binary content.
func (source *SendRequest) ToWhatsappAttachment() (result models.QpToWhatsappAttachment) {
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

	uIntContentLength := uint64(contentLength)
	if attach.FileLength != uIntContentLength {
		originalFileLength := attach.FileLength
		attach.FileLength = uIntContentLength

		warn := fmt.Sprintf(
			"invalid attachment length, request length: %v != content length: %v, revalidating for security",
			originalFileLength,
			contentLength,
		)
		result.Debug = append(result.Debug, "[warn][ToWhatsappAttachment] "+warn)
		logentry.Warnf("%s", warn)
	}

	attach.SetContent(&source.Content)
	result.Attach = attach
	result.AttachSecureAndCustomize()
	result.AttachImageTreatment()
	result.AttachAudioTreatment()
	return
}

// GenerateBodyContent reads binary content directly from the HTTP request body.
func (source *SendRequest) GenerateBodyContent(r *http.Request) (err error) {
	content, err := io.ReadAll(r.Body)
	if err != nil {
		return
	}

	source.Content = content
	source.Mimetype = r.Header.Get("Content-Type")

	informedLength := r.Header.Get("Content-Length")
	if len(informedLength) > 0 {
		length, parseErr := strconv.ParseUint(informedLength, 10, 64)
		if parseErr == nil {
			source.FileLength = length
		}
	}

	source.FileName = library.GetFileName(r)
	return
}

// GenerateEmbedContent decodes base64 or data-URI content into binary payload.
func (source *SendAnyRequest) GenerateEmbedContent() (err error) {
	content := source.Content

	if strings.HasPrefix(content, "data:") {
		parts := strings.SplitN(content, ",", 2)
		if len(parts) != 2 {
			err = fmt.Errorf("invalid data URI format")
			return
		}

		header := parts[0]
		if strings.HasPrefix(header, "data:") && strings.Contains(header, ";base64") {
			mimePart := header[5:]
			mimeType := strings.Split(mimePart, ";")[0]
			if len(source.Mimetype) == 0 {
				source.Mimetype = mimeType
			}
		}

		content = parts[1]
	}

	decoded, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return
	}

	source.SendRequest.Content = decoded
	source.FileLength = uint64(len(decoded))
	return
}

// GenerateUrlContent downloads attachment content from a public URL.
func (source *SendAnyRequest) GenerateUrlContent() (err error) {
	resp, err := http.Get(source.Url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = fmt.Errorf("error on generate url content, unexpected status code: %v", resp.StatusCode)
		logentry := source.GetLogger()
		logentry.Error(err)
		return
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	source.SendRequest.Content = content

	if resp.ContentLength > -1 {
		source.FileLength = uint64(resp.ContentLength)
	}

	if len(source.Mimetype) == 0 {
		source.Mimetype = resp.Header.Get("Content-Type")
	}

	if len(source.FileName) == 0 {
		source.FileName = path.Base(source.Url)
		if len(source.FileName) > 0 {
			filename, unescapeErr := url.QueryUnescape(source.FileName)
			if unescapeErr != nil {
				logentry := source.GetLogger()
				logentry.Warnf("fail to unescape from url, filename: %s, err: %s", source.FileName, unescapeErr)
			} else {
				source.FileName = filename
			}
		}
	}

	return
}
