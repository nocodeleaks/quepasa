package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	library "github.com/nocodeleaks/quepasa/library"
	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// -------------------------- PUBLIC METHODS
//region TYPES OF SENDING

// SendAPIHandler renders route "/send" and "/sendencoded"
//
//	@Summary		Send any type of message (text, file, poll, base64 content)
//	@Description	Endpoint to send messages via WhatsApp. Accepts sending of:
//	@Description	- Plain text (field "text")
//	@Description	- Files by URL (field "url") — server will download and send as attachment
//	@Description	- Base64 content (field "content") — use format data:<mime>;base64,<data>
//	@Description	- Polls (field "poll") — send the poll JSON in the "poll" field
//	@Description
//	@Description	Main fields:
//	@Description	- chatId: chat identifier (can be WID, LID or number with suffix @s.whatsapp.net)
//	@Description	- text: message text
//	@Description	- url: public URL to download a file
//	@Description	- content: embedded base64 content (e.g.: data:image/png;base64,...)
//	@Description	- fileName: file name (optional, used when name cannot be inferred)
//	@Description	- poll: JSON object with the poll (question, options, selections)
//	@Description
//	@Description	Examples:
//	@Description	Text:
//	@Description	```json
//	@Description	{
//	@Description	"chatId": "5511999999999@s.whatsapp.net",
//	@Description	"text": "Hello, world!"
//	@Description	}
//	@Description	```
//	@Description	Poll:
//	@Description	```json
//	@Description	{
//	@Description	"chatId": "5511999999999@s.whatsapp.net",
//	@Description	"poll": {
//	@Description	"question": "Which languages do you know?",
//	@Description	"options": ["JavaScript","Python","Go","Java","C#","Ruby"],
//	@Description	"selections": 3
//	@Description	}
//	@Description	}
//	@Description	```
//	@Description	Base64:
//	@Description	```json
//	@Description	{
//	@Description	"chatId": "5511999999999@s.whatsapp.net",
//	@Description	"content": "data:image/png;base64,...."
//	@Description	}
//	@Description	```
//	@Description	File by URL:
//	@Description	```json
//	@Description	{
//	@Description	"chatId": "5511999999999@s.whatsapp.net",
//	@Description	"url": "https://example.com/path/to/file.jpg"
//	@Description	}
//	@Description	```
//	@Tags			Send
//	@Accept			json
//	@Produce		json
//	@Param			request	body		object{chatId=string,text=string,url=string,content=string,fileName=string,poll=object{question=string,options=[]string,selections=int}}	false	"Request body. Use 'content' for base64, 'url' for remote files, 'poll' for poll JSON."
//	@Success		200		{object}	models.QpSendResponse
//	@Failure		400		{object}	models.QpSendResponse
//	@Security		ApiKeyAuth
//	@Router			/send [post]
func SendAny(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	server, err := GetServer(r)
	if err != nil {
		MessageSendErrors.Inc()
		ObserveAPIRequestDuration(r.Method, "/send", "400", time.Since(startTime).Seconds())

		response := &models.QpSendResponse{}
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	SendAnyWithServer(w, r, server)

	// Record successful API processing time (assuming 200 status for now)
	ObserveAPIRequestDuration(r.Method, "/send", "200", time.Since(startTime).Seconds())
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
		MessageSendErrors.Inc()
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	if len(request.Url) == 0 && r.URL.Query().Has("url") {
		request.Url = r.URL.Query().Get("url")
	}

	// trim start and end white spaces
	request.Url = strings.TrimSpace(request.Url)

	if len(request.Url) > 0 {

		// download content to byte array
		err = request.GenerateUrlContent()
		if err != nil {
			MessageSendErrors.Inc()
			response.ParseError(err)
			RespondInterface(w, response)
			return
		}
	} else if len(request.Content) > 0 {

		// BASE64 content to byte array
		err = request.GenerateEmbedContent()
		if err != nil {
			MessageSendErrors.Inc()
			response.ParseError(err)
			RespondInterface(w, response)
			return
		}
	}

	filename := library.GetFileName(r)
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

	if request.Poll == nil && att.Attach == nil && len(request.Text) == 0 {
		MessageSendErrors.Inc()
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
func Send(server *models.QpWhatsappServer, response *models.QpSendResponse, request *models.QpSendRequest, w http.ResponseWriter, attach *whatsapp.WhatsappAttachment) {
	SendWithMessageType(server, response, request, w, attach, whatsapp.UnhandledMessageType)
}

// SendWithMessageType sends to the whatsapp server with specified message type
// If messageType is UnhandledMessageType, it will auto-detect the type
func SendWithMessageType(server *models.QpWhatsappServer, response *models.QpSendResponse, request *models.QpSendRequest, w http.ResponseWriter, attach *whatsapp.WhatsappAttachment, messageType whatsapp.WhatsappMessageType) {
	waMsg, err := request.ToWhatsappMessage()

	if err != nil {
		MessageSendErrors.Inc()
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	logentry := server.GetLogger()

	pollText := strings.TrimSpace(waMsg.Text)
	if len(pollText) > 0 {
		if strings.HasPrefix(pollText, "poll:") {
			pollText = pollText[5:]

			var poll *whatsapp.WhatsappPoll
			err = json.Unmarshal([]byte(pollText), &poll)
			if err != nil {
				err = fmt.Errorf("error converting text to json poll: %s", err.Error())
				MessageSendErrors.Inc()
				response.ParseError(err)
				RespondInterface(w, response)
				return
			}

			waMsg.Poll = poll
		}
	}

	if attach != nil {
		waMsg.Attachment = attach
		if messageType == whatsapp.UnhandledMessageType {
			waMsg.Type = whatsapp.GetMessageType(attach)
			logentry.Debugf("send attachment of type: %v, mime: %s, length: %v, filename: %s", waMsg.Type, attach.Mimetype, attach.FileLength, attach.FileName)
		} else {
			waMsg.Type = messageType
			logentry.Debugf("send attachment (forced type: %v): mime: %s, length: %v, filename: %s", waMsg.Type, attach.Mimetype, attach.FileLength, attach.FileName)
		}
	} else {
		// test for poll, already set from ToWhatsappMessage
		waMsg.Type = whatsapp.TextMessageType
	}

	// if not set, try to recover "text"
	if waMsg.Type == whatsapp.UnhandledMessageType {
		// correct msg type for texts contents
		if len(waMsg.Text) > 0 {
			waMsg.Type = whatsapp.TextMessageType
		} else {
			err = fmt.Errorf("unknown message type without text")
			response.ParseError(err)
			RespondInterface(w, response)
			return
		}
	}

	// Checking for ready state
	status := server.GetStatus()
	if status != whatsapp.Ready {
		err = &ApiServerNotReadyException{Wid: server.GetWId(), Status: status}
		response.ParseError(err)
		RespondInterfaceCode(w, response, http.StatusServiceUnavailable)
		return
	}

	// Handle LID to phone mapping for sending
	originalChatId := waMsg.Chat.Id
	if strings.Contains(waMsg.Chat.Id, "@lid") {
		logentry.Debugf("LID detected in chat ID: %s, attempting to get phone number", waMsg.Chat.Id)

		phone, err := server.GetPhoneFromLID(waMsg.Chat.Id)
		if err != nil || len(phone) == 0 {
			if err != nil {
				logentry.Warnf("failed to get phone from LID %s: %v, will try sending directly to LID", waMsg.Chat.Id, err)
			} else {
				logentry.Warnf("empty phone number returned for LID: %s, will try sending directly to LID", waMsg.Chat.Id)
			}

			// Try to send directly to LID first
			sendResponse, err := server.SendMessage(waMsg)
			if err == nil {
				// Success sending to LID directly
				MessagesSent.Inc()

				result := &models.QpSendResponseMessage{}
				result.Wid = server.GetWId()
				result.Id = sendResponse.GetId()
				result.ChatId = originalChatId
				result.TrackId = waMsg.TrackId

				response.ParseSuccess(result)
				RespondInterface(w, response)
				return
			}

			// Failed to send to LID, return error
			logentry.Errorf("failed to find phone number for LID %s and failed to send message to LID: %v", originalChatId, err)
			MessageSendErrors.Inc()
			response.ParseError(fmt.Errorf("failed to resolve LID to phone number and failed to send message to LID: %v", err))
			RespondInterface(w, response)
			return
		}

		waMsg.Chat.Id = whatsapp.PhoneToWid(phone)
		logentry.Debugf("converted LID %s to phone-based chat ID: %s (from phone %s)", originalChatId, waMsg.Chat.Id, phone)
	}

	sendResponse, err := server.SendMessage(waMsg)
	if err != nil {
		MessageSendErrors.Inc()
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// success
	MessagesSent.Inc()

	result := &models.QpSendResponseMessage{}
	result.Wid = server.GetWId()
	result.Id = sendResponse.GetId()
	result.ChatId = waMsg.Chat.Id
	result.TrackId = waMsg.TrackId

	response.ParseSuccess(result)
	RespondInterface(w, response)
}
