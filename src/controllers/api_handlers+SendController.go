package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	metrics "github.com/nocodeleaks/quepasa/metrics"
	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// -------------------------- PUBLIC METHODS
//region TYPES OF SENDING

// SendAPIHandler renders route "/v3/bot/{token}/send"
func SendAny(w http.ResponseWriter, r *http.Request) {

	server, err := GetServer(r)
	if err != nil {
		metrics.MessageSendErrors.Inc()

		response := &models.QpSendResponse{}
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	SendAnyWithServer(w, r, server)
}

func SendAnyWithServer(w http.ResponseWriter, r *http.Request, server *models.QpWhatsappServer) {
	response := &models.QpSendResponse{}

	// Declare a new request struct.
	request := &models.QpSendAnyRequest{}

	if r.ContentLength > 0 && r.Method == http.MethodPost {
		// Try to decode the request body into the struct. If there is an error,
		// respond to the client with the error message and a 400 status code.
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			jsonErr := fmt.Errorf("invalid json body: %s", err.Error())
			response.ParseError(jsonErr)
			RespondInterface(w, response)
			return
		}
	}

	// Getting ChatId parameter
	err := request.EnsureValidChatId(r)
	if err != nil {
		metrics.MessageSendErrors.Inc()
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	/*
		// if spam, checks if is in group
		if strings.HasSuffix(request.ChatId, "@g.us") && strings.Contains(r.RequestURI, "spam") {
			isInGroup := server.GetConnection().HasChat(request.ChatId)
			if !isInGroup {
				metrics.MessageSendErrors.Inc()
				err = fmt.Errorf("it seams that you don't belongs to this group: %s", request.ChatId)
				response.ParseError(err)
				RespondInterface(w, response)
				return
			}
		}
	*/

	if len(request.Url) == 0 && r.URL.Query().Has("url") {
		request.Url = r.URL.Query().Get("url")
	}

	// trim start and end white spaces
	request.Url = strings.TrimSpace(request.Url)

	if len(request.Url) > 0 {

		// download content to byte array
		err = request.GenerateUrlContent()
		if err != nil {
			metrics.MessageSendErrors.Inc()
			response.ParseError(err)
			RespondInterface(w, response)
			return
		}
	} else if len(request.Content) > 0 {

		// BASE64 content to byte array
		err = request.GenerateEmbedContent()
		if err != nil {
			metrics.MessageSendErrors.Inc()
			response.ParseError(err)
			RespondInterface(w, response)
			return
		}
	}

	filename := models.GetFileName(r)
	if len(filename) > 0 {
		request.FileName = filename
	}

	SendRequest(w, r, &request.QpSendRequest, server)
}

//endregion

// -------------------------- INTERNAL METHODS

// Send a request already validated with chatid and server
func SendRequest(w http.ResponseWriter, r *http.Request, request *models.QpSendRequest, server *models.QpWhatsappServer) {
	response := &models.QpSendResponse{}
	var err error

	att := request.ToWhatsappAttachment()

	// if not set, try to recover "text"
	if len(request.Text) == 0 {
		request.Text = GetTextParameter(r)
		if len(request.Text) > 0 {
			response.Debug = append(response.Debug, "[debug][SendRequest] 'text' found in parameters")
		}
	}

	// if not set, try to recover "in reply"
	if len(request.InReply) == 0 {
		request.InReply = GetInReplyParameter(r)
		if len(request.InReply) > 0 {
			response.Debug = append(response.Debug, "[debug][SendRequest] 'inreply' found in parameters")
		}
	}

	if att.Attach == nil && len(request.Text) == 0 {
		metrics.MessageSendErrors.Inc()
		err = fmt.Errorf("text not found, do not send empty messages")
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// getting trackid if not passed in request
	if len(request.TrackId) == 0 {
		request.TrackId = GetTrackId(r)
	}

	response.Debug = append(response.Debug, att.Debug...)
	Send(server, response, request, w, att.Attach)
}

// finally sends to the whatsapp server
// finally sends to the whatsapp server
func Send(server *models.QpWhatsappServer, response *models.QpSendResponse, request *models.QpSendRequest, w http.ResponseWriter, attach *whatsapp.WhatsappAttachment) {
	logentry := server.GetLogger()

	// Create WhatsApp message from request
	waMsg, err := request.ToWhatsappMessage()
	if err != nil {
		metrics.MessageSendErrors.Inc()
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Handle attachment if present
	if attach != nil {
		waMsg.Attachment = attach
		waMsg.Type = whatsapp.GetMessageType(attach)
		logentry.Debugf("send attachment of type: %v, mime: %s, length: %v, filename: %s",
			waMsg.Type, attach.Mimetype, attach.FileLength, attach.FileName)
	} else {
		// test for poll, already set from ToWhatsappMessage
		waMsg.Type = whatsapp.TextMessageType
	}

	// Validate message type
	if waMsg.Type == whatsapp.UnknownMessageType {
		// correct msg type for texts contents
		if len(waMsg.Text) > 0 {
			waMsg.Type = whatsapp.TextMessageType
		}

		logentry.Errorf("unknown message type: %v", waMsg)
		// *** implement an error here if not found any knowing type
	}

	// Checking for ready state
	status := server.GetStatus()
	if status != whatsapp.Ready {
		err = &ApiServerNotReadyException{Wid: server.GetWId(), Status: status}
		response.ParseError(err)
		RespondInterfaceCode(w, response, http.StatusServiceUnavailable)
		return
	}

	// SYNCHRONOUS: Handle typing indicator if requested AND enabled in environment
	if request.ShowTyping && whatsapp.Options.ShowTyping {
		typingDuration := request.TypingDuration
		if typingDuration <= 0 {
			// Default typing duration based on message length
			messageLen := len(request.Text)
			if messageLen > 0 {
				// Roughly 30ms per character, minimum 1s, maximum 5s
				typingDuration = min(max(messageLen*30, 1000), 5000)
			} else {
				typingDuration = 2000 // Default for media
			}
		}

		mediaType := request.MediaType
		if mediaType == "" && attach != nil {
			// Auto-detect media type based on attachment
			if strings.HasPrefix(attach.Mimetype, "audio/") {
				mediaType = "audio"
			}
		}

		// Send the typing indicator
		err := server.SendChatPresence(request.ChatId, true, mediaType)
		if err != nil {
			logentry.Warnf("Failed to send typing indicator: %v", err)
			// Continue even if typing indicator fails
		} else {
			logentry.Infof("Waiting %dms for typing indicator before sending message (global typing: enabled)", typingDuration)
			// Sleep for the typing duration - this will block the request
			time.Sleep(time.Duration(typingDuration) * time.Millisecond)
		}
	} else if request.ShowTyping && !whatsapp.Options.ShowTyping {
		logentry.Infof("Typing indicator requested but disabled by environment setting (SHOWTYPING=false)")
	}

	// SYNCHRONOUS: Send the message
	logentry.Infof("Sending message synchronously to chat: %s", waMsg.Chat.Id)
	sendResponse, err := server.SendMessage(waMsg)
	if err != nil {
		metrics.MessageSendErrors.Inc()
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// success
	metrics.MessagesSent.Inc()
	logentry.Infof("Message sent successfully: %s", waMsg.Id)

	// Stop typing indicator after message is sent (only if we sent one)
	if request.ShowTyping && whatsapp.Options.ShowTyping {
		_ = server.SendChatPresence(request.ChatId, false, "")
	}

	result := &models.QpSendResponseMessage{}
	result.Wid = server.GetWId()
	result.Id = sendResponse.GetId()
	result.ChatId = waMsg.Chat.Id
	result.TrackId = waMsg.TrackId

	response.ParseSuccess(result)
	RespondInterface(w, response)
}
