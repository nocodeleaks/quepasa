package cable

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	library "github.com/nocodeleaks/quepasa/library"
	media "github.com/nocodeleaks/quepasa/media"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
)

// sendMessageRequest is the cable-local contract used by `message.send`.
//
// It intentionally mirrors only the websocket command needs so the cable
// transport can evolve independently from HTTP request DTOs.
type sendMessageRequest struct {
	ID         string
	ChatID     string
	TrackID    string
	Text       string
	InReply    string
	FileName   string
	FileLength uint64
	MimeType   string
	Seconds    uint32
	Content    []byte

	Poll     *whatsapp.WhatsappPoll
	Location *whatsapp.WhatsappLocation
	Contact  *whatsapp.WhatsappContact

	URL           string
	EmbedContent  string
	TypingDelayMS int
	MediaType     string
}

// newSendMessageRequest normalizes the websocket command payload into the
// internal cable request shape.
func newSendMessageRequest(data *sendCommandData, commandID string) *sendMessageRequest {
	if data == nil {
		return &sendMessageRequest{}
	}

	return &sendMessageRequest{
		ID:            firstNonEmpty(data.ID),
		ChatID:        firstNonEmpty(data.ChatID, data.ChatId),
		TrackID:       firstNonEmpty(data.TrackID, data.TrackId, commandID),
		Text:          data.Text,
		InReply:       firstNonEmpty(data.InReply, data.Inreply),
		FileName:      firstNonEmpty(data.FileName, data.Filename),
		FileLength:    firstNonZero(data.FileLength, data.Filelength),
		MimeType:      firstNonEmpty(data.MimeType, data.Mime),
		Seconds:       data.Seconds,
		Poll:          data.Poll,
		Location:      data.Location,
		Contact:       data.Contact,
		URL:           data.Url,
		EmbedContent:  data.Content,
		TypingDelayMS: data.TypingDuration,
		MediaType:     data.MediaType,
	}
}

// EnsureValidChatID applies the same validation used by the send API without
// forcing the cable module to depend on HTTP request DTOs.
func (request *sendMessageRequest) EnsureValidChatID() error {
	if strings.TrimSpace(request.ChatID) == "" {
		return fmt.Errorf("chat id missing")
	}

	chatID, err := whatsapp.FormatEndpoint(request.ChatID)
	if err != nil {
		return err
	}

	request.ChatID = chatID
	return nil
}

// BuildContent resolves optional attachment content from either a remote URL or
// an embedded base64 payload.
func (request *sendMessageRequest) BuildContent() error {
	if request.URL != "" {
		return request.GenerateURLContent()
	}

	if request.EmbedContent != "" {
		return request.GenerateEmbedContent()
	}

	return nil
}

// ToWhatsAppMessage projects the cable request into the internal message model
// accepted by the WhatsApp service.
func (request *sendMessageRequest) ToWhatsAppMessage() (*whatsapp.WhatsappMessage, error) {
	chatID, err := whatsapp.FormatEndpoint(request.ChatID)
	if err != nil {
		return nil, err
	}

	chat := whatsapp.WhatsappChat{Id: chatID}
	if phone, _ := whatsapp.GetPhoneIfValid(chatID); phone != "" {
		chat.Phone = phone
	}

	message := &whatsapp.WhatsappMessage{
		Id:           strings.ToUpper(request.ID),
		TrackId:      request.TrackID,
		InReply:      request.InReply,
		Text:         request.Text,
		Chat:         chat,
		FromMe:       true,
		FromInternal: true,
		Poll:         request.Poll,
	}

	switch {
	case request.Contact != nil:
		message.Type = whatsapp.ContactMessageType
		message.Contact = request.Contact
		return message, nil

	case request.Location != nil:
		message.Type = whatsapp.LocationMessageType
		message.Attachment = &whatsapp.WhatsappAttachment{
			Latitude:  request.Location.Latitude,
			Longitude: request.Location.Longitude,
			Mimetype:  "text/x-uri; location",
		}
		if request.Location.Name != "" {
			message.Text = request.Location.Name
		}
		return message, nil

	case message.Text != "":
		message.Type = whatsapp.TextMessageType
	default:
		message.Type = whatsapp.TextMessageType
	}

	return message, nil
}

// ToWhatsAppAttachment builds the optional attachment payload and reuses the
// shared attachment hardening helpers already used by the HTTP transport.
func (request *sendMessageRequest) ToWhatsAppAttachment() media.QpToWhatsappAttachment {
	var result media.QpToWhatsappAttachment
	if len(request.Content) == 0 {
		return result
	}

	logger := request.logger()
	attachment := &whatsapp.WhatsappAttachment{
		Mimetype:   request.MimeType,
		FileLength: request.FileLength,
		FileName:   request.FileName,
		Seconds:    request.Seconds,
	}

	contentLength := uint64(len(request.Content))
	if attachment.FileLength != contentLength {
		originalLength := attachment.FileLength
		attachment.FileLength = contentLength

		warn := fmt.Sprintf(
			"invalid attachment length, request length: %v != content length: %v, revalidating for security",
			originalLength,
			contentLength,
		)
		result.Debug = append(result.Debug, "[warn][ToWhatsAppAttachment] "+warn)
		logger.Warnf("%s", warn)
	}

	attachment.SetContent(&request.Content)
	result.Attach = attachment
	result.AttachSecureAndCustomize()
	result.AttachImageTreatment()
	result.AttachAudioTreatment()
	return result
}

// GenerateEmbedContent decodes an inline base64 payload, including data-URI
// prefixed values, into binary attachment content.
func (request *sendMessageRequest) GenerateEmbedContent() error {
	content := request.EmbedContent

	if strings.HasPrefix(content, "data:") {
		parts := strings.SplitN(content, ",", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid data URI format")
		}

		header := parts[0]
		if strings.HasPrefix(header, "data:") && strings.Contains(header, ";base64") {
			mimePart := header[5:]
			mimeType := strings.Split(mimePart, ";")[0]
			if request.MimeType == "" {
				request.MimeType = mimeType
			}
		}

		content = parts[1]
	}

	decoded, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return err
	}

	request.Content = decoded
	request.FileLength = uint64(len(decoded))
	return nil
}

// GenerateURLContent downloads attachment bytes from the informed URL and
// enriches missing metadata such as content type and filename.
func (request *sendMessageRequest) GenerateURLContent() error {
	resp, err := http.Get(request.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("error on generate url content, unexpected status code: %v", resp.StatusCode)
		request.logger().Error(err)
		return err
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	request.Content = content
	if resp.ContentLength > -1 {
		request.FileLength = uint64(resp.ContentLength)
	}

	if request.MimeType == "" {
		request.MimeType = resp.Header.Get("Content-Type")
	}

	if request.FileName == "" {
		request.FileName = path.Base(request.URL)
		if request.FileName != "" {
			filename, unescapeErr := url.QueryUnescape(request.FileName)
			if unescapeErr != nil {
				request.logger().Warnf("fail to unescape from url, filename: %s, err: %s", request.FileName, unescapeErr)
			} else {
				request.FileName = filename
			}
		}
	}

	return nil
}

func (request *sendMessageRequest) logger() *log.Entry {
	return log.WithContext(context.Background()).WithField(library.LogFields.ChatId, request.ChatID)
}
