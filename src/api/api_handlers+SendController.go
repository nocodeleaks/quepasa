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
//	@Summary		Send any type of message (text, file, poll, base64 content, location, contact)
//	@Description	Endpoint to send messages via WhatsApp. Accepts sending of:
//	@Description	- Plain text (field "text")
//	@Description	- Files by URL (field "url") — server will download and send as attachment
//	@Description	- Base64 content (field "content") — use format data:<mime>;base64,<data>
//	@Description	- Polls (field "poll") — send the poll JSON in the "poll" field
//	@Description	- Location (field "location") — send location with latitude/longitude in the "location" object
//	@Description	- Contact (field "contact") — send contact with phone/name in the "contact" object
//	@Description
//	@Description	Main fields:
//	@Description	- chatId: chat identifier (can be WID, LID or number with suffix @s.whatsapp.net)
//	@Description	- text: message text
//	@Description	- url: public URL to download a file
//	@Description	- content: embedded base64 content (e.g.: data:image/png;base64,...)
//	@Description	- fileName: file name (optional, used when name cannot be inferred)
//	@Description	- poll: JSON object with the poll (question, options, selections)
//	@Description	- location: JSON object with location data (latitude, longitude, name, address, url)
//	@Description	- contact: JSON object with contact data (phone, name, vcard)
//	@Description
//	@Description	Location object fields:
//	@Description	- latitude (float64, required): Location latitude in degrees (e.g.: -23.550520)
//	@Description	- longitude (float64, required): Location longitude in degrees (e.g.: -46.633308)
//	@Description	- name (string, optional): Location name/description
//	@Description	- address (string, optional): Location full address
//	@Description	- url (string, optional): URL with link to the map
//	@Description
//	@Description	Contact object fields:
//	@Description	- phone (string, required): Contact phone number
//	@Description	- name (string, required): Contact display name
//	@Description	- vcard (string, optional): Full vCard string (auto-generated if not provided)
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
//	@Description	Location:
//	@Description	```json
//	@Description	{
//	@Description	"chatId": "5511999999999@s.whatsapp.net",
//	@Description	"location": {
//	@Description	"latitude": -23.550520,
//	@Description	"longitude": -46.633308,
//	@Description	"name": "Avenida Paulista, São Paulo"
//	@Description	}
//	@Description	}
//	@Description	```
//	@Description	Contact:
//	@Description	```json
//	@Description	{
//	@Description	"chatId": "5511999999999@s.whatsapp.net",
//	@Description	"contact": {
//	@Description	"phone": "5511999999999",
//	@Description	"name": "John Doe"
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
//	@Param			request	body		object{chatId=string,text=string,url=string,content=string,fileName=string,poll=object{question=string,options=[]string,selections=int},location=object{latitude=float64,longitude=float64,name=string,address=string,url=string},contact=object{phone=string,name=string,vcard=string}}	false	"Request body. Use 'content' for base64, 'url' for remote files, 'poll' for poll JSON, 'location' for location object, or 'contact' for contact object."
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

	if request.Poll == nil && request.Location == nil && request.Contact == nil && att.Attach == nil && len(request.Text) == 0 {
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
		// Only set text type if type was not already set (e.g., by Poll or Location)
		if waMsg.Type == whatsapp.UnhandledMessageType {
			waMsg.Type = whatsapp.TextMessageType
		}
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

	// Always try to convert to @s.whatsapp.net format for sending
	// WhatsApp messages should preferably be sent to phone-based JIDs (@s.whatsapp.net)
	// But fallback to @lid if phone conversion fails
	originalChatId := waMsg.Chat.Id

	if strings.Contains(waMsg.Chat.Id, whatsapp.WHATSAPP_SERVERDOMAIN_LID_SUFFIX) {
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
				logentry.Infof("successfully sent message directly to LID: %s", originalChatId)
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

		// Successfully converted to phone - use @s.whatsapp.net format
		waMsg.Chat.Id = whatsapp.PhoneToWid(phone)
		logentry.Debugf("converted LID %s to phone-based chat ID: %s (from phone %s)", originalChatId, waMsg.Chat.Id, phone)
	} else {
		// Phone number or @s.whatsapp.net format
		// Try to find LID mapping and use the same conversion flow that works
		// This fixes the PN→LID session migration bug in whatsmeow
		phoneNumber := waMsg.Chat.Id

		// Extract phone number from JID if it has suffix
		if strings.Contains(phoneNumber, whatsapp.WHATSAPP_SERVERDOMAIN_USER_SUFFIX) {
			phoneNumber = strings.TrimSuffix(phoneNumber, whatsapp.WHATSAPP_SERVERDOMAIN_USER_SUFFIX)
		}

		// Remove + prefix if present
		phoneNumber = strings.TrimPrefix(phoneNumber, "+")

		// Try to find LID for this phone number
		lid, lidErr := server.GetLIDFromPhone(phoneNumber)
		if lidErr == nil && len(lid) > 0 {
			// Found LID mapping, now get phone back from LID (this establishes correct session)
			phone, phoneErr := server.GetPhoneFromLID(lid)
			if phoneErr == nil && len(phone) > 0 {
				waMsg.Chat.Id = whatsapp.PhoneToWid(phone)
				logentry.Debugf("phone %s has LID %s, converted via LID to: %s (fixes session migration)", originalChatId, lid, waMsg.Chat.Id)
			} else {
				// Fallback to original phone
				waMsg.Chat.Id = whatsapp.PhoneToWid(phoneNumber)
				logentry.Debugf("phone %s has LID %s but failed to get phone back, using original: %s", originalChatId, lid, waMsg.Chat.Id)
			}
		} else {
			// No LID mapping found, use phone directly
			waMsg.Chat.Id = whatsapp.PhoneToWid(phoneNumber)
			logentry.Debugf("no LID mapping for phone %s, using directly: %s", originalChatId, waMsg.Chat.Id)
		}
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
