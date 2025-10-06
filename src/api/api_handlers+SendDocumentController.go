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

// SendDocument sends a document, forcing document type recognition
//
//	@Summary		Send document with forced document type
//	@Description	Endpoint to send documents via WhatsApp, forcing document type recognition regardless of file mimetype. Useful for sending images, audio files, or other content as documents.
//	@Description	Accepts the same parameters as /send but always treats attachments as documents.
//	@Description
//	@Description	Main fields:
//	@Description	- chatId: chat identifier (can be WID, LID or number with suffix @s.whatsapp.net)
//	@Description	- text: message text (optional)
//	@Description	- url: public URL to download a file
//	@Description	- content: embedded base64 content (e.g.: data:image/png;base64,...)
//	@Description	- fileName: file name (optional, used when name cannot be inferred)
//	@Description
//	@Description	Example:
//	@Description	```json
//	@Description	{
//	@Description	"chatId": "5511999999999@s.whatsapp.net",
//	@Description	"url": "https://example.com/document.pdf",
//	@Description	"text": "Please check this document"
//	@Description	}
//	@Description	```
//	@Tags			Send
//	@Accept			json
//	@Produce		json
//	@Param			request	body		object{chatId=string,text=string,url=string,content=string,fileName=string}	false	"Request body"
//	@Success		200		{object}	models.QpSendResponse
//	@Failure		400		{object}	models.QpSendResponse
//	@Security		ApiKeyAuth
//	@Router			/senddocument [post]
func SendDocument(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	server, err := GetServer(r)
	if err != nil {
		MessageSendErrors.Inc()
		ObserveAPIRequestDuration(r.Method, "/senddocument", "400", time.Since(startTime).Seconds())

		response := &models.QpSendResponse{}
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	SendDocumentWithServer(w, r, server)

	// Record successful API processing time (assuming 200 status for now)
	ObserveAPIRequestDuration(r.Method, "/senddocument", "200", time.Since(startTime).Seconds())
}

func SendDocumentWithServer(w http.ResponseWriter, r *http.Request, server *models.QpWhatsappServer) {
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

	SendDocumentRequest(w, r, &request.QpSendRequest, server)
}

// SendDocumentRequest sends a document request, forcing document type
func SendDocumentRequest(w http.ResponseWriter, r *http.Request, request *models.QpSendRequest, server *models.QpWhatsappServer) {
	response := &models.QpSendResponse{}
	var err error

	att := request.ToWhatsappAttachment()

	// if not set, try to recover "text"
	if len(request.Text) == 0 {
		request.Text = GetTextParameter(r)
		if len(request.Text) > 0 {
			response.Debug = append(response.Debug, "[debug][SendDocumentRequest] 'text' found in parameters")
		}
	}

	// if not set, try to recover "in reply"
	if len(request.InReply) == 0 {
		request.InReply = GetInReplyParameter(r)
		if len(request.InReply) > 0 {
			response.Debug = append(response.Debug, "[debug][SendDocumentRequest] 'inreply' found in parameters")
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
	SendDocumentToServer(server, response, request, w, att.Attach)
}

// SendDocument sends a document to the whatsapp server, forcing document type
func SendDocumentToServer(server *models.QpWhatsappServer, response *models.QpSendResponse, request *models.QpSendRequest, w http.ResponseWriter, attach *whatsapp.WhatsappAttachment) {
	// Use the common send method but force document type
	SendWithMessageType(server, response, request, w, attach, whatsapp.DocumentMessageType)
}
