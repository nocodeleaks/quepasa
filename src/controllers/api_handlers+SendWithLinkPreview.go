package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	metrics "github.com/nocodeleaks/quepasa/metrics"
	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	whatsmeow "github.com/nocodeleaks/quepasa/whatsmeow"
	waE2E "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

// SendWithLinkPreviewHandler handles the API endpoint for sending messages with link previews
func SendWithLinkPreviewHandler(w http.ResponseWriter, r *http.Request) {
	// Setting default response type as json
	w.Header().Set("Content-Type", "application/json")

	response := &models.QpSendResponse{}

	// Get server
	server, err := GetServer(r)
	if err != nil {
		metrics.MessageSendErrors.Inc()
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Get logger
	logentry := server.GetLogger()

	// Parse request body
	var request models.LinkPreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		metrics.MessageSendErrors.Inc()
		response.ParseError(fmt.Errorf("invalid request: %v", err))
		RespondInterface(w, response)
		return
	}

	// Validate chat_id
	if request.ChatId == "" {
		metrics.MessageSendErrors.Inc()
		response.ParseError(fmt.Errorf("chat_id is required"))
		RespondInterface(w, response)
		return
	}

	// Format and validate the chat ID
	formattedChatId, err := whatsapp.FormatEndpoint(request.ChatId)
	if err != nil {
		metrics.MessageSendErrors.Inc()
		response.ParseError(fmt.Errorf("invalid chat_id: %v", err))
		RespondInterface(w, response)
		return
	}
	request.ChatId = formattedChatId

	// Validate text
	if request.Text == "" {
		metrics.MessageSendErrors.Inc()
		response.ParseError(fmt.Errorf("text is required and must contain a URL"))
		RespondInterface(w, response)
		return
	}

	// Extract URL from text
	urlRegex := regexp.MustCompile(`https?://[^\s]+`)
	urls := urlRegex.FindAllString(request.Text, -1)

	if len(urls) == 0 {
		metrics.MessageSendErrors.Inc()
		response.ParseError(fmt.Errorf("no URL found in text"))
		RespondInterface(w, response)
		return
	}

	// Use first URL for preview
	targetUrl := urls[0]

	logentry.Debugf("Creating message with link preview for URL: %s", targetUrl)

	// Check if WhatsApp server is ready
	status := server.GetStatus()
	if status != whatsapp.Ready {
		err = &ApiServerNotReadyException{Wid: server.GetWId(), Status: status}
		response.ParseError(err)
		RespondInterfaceCode(w, response, http.StatusServiceUnavailable)
		return
	}

	// Get the whatsmeow client
	conn := server.GetConnection()
	if conn == nil {
		metrics.MessageSendErrors.Inc()
		response.ParseError(fmt.Errorf("connection not available"))
		RespondInterface(w, response)
		return
	}

	// Convert to WhatsmeowConnection
	whatsmeowConn, ok := conn.(*whatsmeow.WhatsmeowConnection)
	if !ok {
		metrics.MessageSendErrors.Inc()
		response.ParseError(fmt.Errorf("unsupported connection type"))
		RespondInterface(w, response)
		return
	}

	// Get the client
	client := whatsmeowConn.Client
	if client == nil {
		metrics.MessageSendErrors.Inc()
		response.ParseError(fmt.Errorf("client not available"))
		RespondInterface(w, response)
		return
	}

	// Parse destination JID
	recipientJID, err := types.ParseJID(request.ChatId)
	if err != nil {
		metrics.MessageSendErrors.Inc()
		response.ParseError(fmt.Errorf("invalid recipient JID: %v", err))
		RespondInterface(w, response)
		return
	}

	// Prepare thumbnail if available
	var thumbData []byte
	if request.CustomThumbUrl != "" {
		// Download thumbnail from URL
		thumbResp, err := http.Get(request.CustomThumbUrl)
		if err == nil && thumbResp.StatusCode == http.StatusOK {
			defer thumbResp.Body.Close()
			thumbData, err = ioutil.ReadAll(thumbResp.Body)
			if err != nil {
				logentry.Warnf("Failed to read thumbnail data: %v", err)
			}
		} else {
			logentry.Warnf("Failed to download thumbnail from %s: %v",
				request.CustomThumbUrl, err)
		}
	}

	// Use default values for title and description if not provided
	title := request.CustomTitle
	if title == "" {
		// Extract domain as default title
		title = extractDomain(targetUrl)
	}

	description := request.CustomDesc

	// Create ExtendedTextMessage
	content := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text:        proto.String(request.Text),
			Title:       proto.String(title),
			MatchedText: proto.String(targetUrl),
		},
	}

	// Add description if available
	if description != "" {
		content.ExtendedTextMessage.Description = proto.String(description)
	}

	// Add thumbnail if available
	if len(thumbData) > 0 {
		content.ExtendedTextMessage.JPEGThumbnail = thumbData
		logentry.Debugf("Added thumbnail image (%d bytes)", len(thumbData))
	}

	// Send the message
	resp, err := client.SendMessage(context.Background(), recipientJID, content)
	if err != nil {
		metrics.MessageSendErrors.Inc()
		response.ParseError(fmt.Errorf("error sending message: %v", err))
		RespondInterface(w, response)
		return
	}

	// Log success
	logentry.Infof("Link preview message sent (server timestamp: %s)", resp.Timestamp)
	metrics.MessagesSent.Inc()

	// Prepare success response
	result := &models.QpSendResponseMessage{}
	result.Wid = server.GetWId()
	result.Id = resp.ID
	result.ChatId = request.ChatId
	result.TrackId = request.TrackId

	response.ParseSuccess(result)
	RespondInterface(w, response)
}

// Helper function to extract domain from URL
func extractDomain(url string) string {
	// Remove protocol
	domain := url
	if strings.HasPrefix(domain, "http://") {
		domain = domain[7:]
	} else if strings.HasPrefix(domain, "https://") {
		domain = domain[8:]
	}

	// Remove path and query
	if index := strings.Index(domain, "/"); index > 0 {
		domain = domain[:index]
	}

	return domain
}
