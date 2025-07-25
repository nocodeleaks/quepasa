package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	metrics "github.com/nocodeleaks/quepasa/metrics"
	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
)

const APIVersion2 string = "v2"

var ControllerPrefixV2 string = fmt.Sprintf("/%s/bot/{token}", APIVersion2)

func RegisterAPIV2Controllers(r chi.Router) {

	r.Get(ControllerPrefixV2, InformationHandlerV2)
	r.Post(ControllerPrefixV2+"/send", SendAPIHandlerV2)
	r.Post(ControllerPrefixV2+"/sendtext", SendAPIHandlerV2)
	r.Get(ControllerPrefixV2+"/receive", ReceiveAPIHandlerV2)

	// external for now
	r.Post(ControllerPrefixV2+"/senddocument", SendDocumentAPIHandlerV2)
	r.Post(ControllerPrefixV2+"/attachment", AttachmentAPIHandlerV2)
	r.Post(ControllerPrefixV2+"/webhook", WebHookAPIHandlerV2)
	r.Get(ControllerPrefixV2+"/webhook", WebHookAPIHandlerV2)
	r.Delete(ControllerPrefixV2+"/webhook", WebHookAPIHandlerV2)
}

// InformationController renders route GET "/{version}/bot/{token}"
func InformationHandlerV2(w http.ResponseWriter, r *http.Request) {

	response := &models.QpInfoResponseV2{}

	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	response.Id = server.Token
	response.Number = server.GetNumber()
	response.Username = server.User
	response.FirstName = "John"
	response.LastName = "Doe"
	RespondSuccess(w, response)
}

// SendAPIHandler renders route "/{version}/bot/{token}/send"
func SendAPIHandlerV2(w http.ResponseWriter, r *http.Request) {
	response := &models.QpSendResponseV2{}

	server, err := GetServer(r)
	if err != nil {
		metrics.MessageSendErrors.Inc()
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Declare a new request struct.
	request := &models.QpSendRequestV2{}

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		metrics.MessageSendErrors.Inc()
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	log.Tracef("sending requested: %v", request)
	trackid := GetTrackId(r)
	waMsg, err := whatsapp.ToMessage(request.Recipient, request.Message, trackid)
	if err != nil {
		metrics.MessageSendErrors.Inc()
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// setting source msg participant
	if waMsg.FromGroup() && len(waMsg.Participant.Id) == 0 {
		waMsg.Participant.Id = server.Wid
	}

	// setting wa msg chat title
	if len(waMsg.Chat.Title) == 0 {
		waMsg.Chat.Title = server.GetChatTitle(waMsg.Chat.Id)
	}

	sendResponse, err := server.SendMessage(waMsg)
	if err != nil {
		metrics.MessageSendErrors.Inc()
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	response.Chat.ID = waMsg.Chat.Id
	response.Chat.UserName = waMsg.Chat.Id
	response.Chat.Title = waMsg.Chat.Title
	response.From.ID = server.Wid
	response.From.UserName = server.GetNumber()
	response.ID = sendResponse.GetId()

	// Para manter a compatibilidade
	response.PreviusV1 = models.QPSendResult{
		Source:    server.GetWId(),
		Recipient: waMsg.Chat.Id,
		MessageId: sendResponse.GetId(),
	}

	metrics.MessagesSent.Inc()
	RespondSuccess(w, response)
}

// Renders route GET "/{version}/bot/{token}/receive"
func ReceiveAPIHandlerV2(w http.ResponseWriter, r *http.Request) {
	response := &models.QpReceiveResponseV2{}

	server, err := GetServer(r)
	if err != nil {
		metrics.MessageReceiveErrors.Inc()
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// append server to response
	response.Bot = *models.ToQpServerV2(server.QpServer)

	// Checking for ready state
	status := server.GetStatus()
	if status != whatsapp.Ready {
		metrics.MessageReceiveErrors.Inc()
		err = &ApiServerNotReadyException{Wid: server.Wid, Status: status}
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	queryValues := r.URL.Query()
	timestamp := queryValues.Get("timestamp")

	messages, err := GetMessagesToAPIV2(server, timestamp)
	if err != nil {
		metrics.MessageReceiveErrors.Inc()
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// append messages to response
	response.Messages = messages

	// metrics
	metrics.MessagesReceived.Add(float64(len(messages)))
	RespondSuccess(w, response)
}

// NOT TESTED ----------------------------------
// NOT TESTED ----------------------------------
// NOT TESTED ----------------------------------
// NOT TESTED ----------------------------------

// Usado para envio de documentos, anexos, separados do texto, em caso de imagem, aceita um caption (titulo)
func SendDocumentAPIHandlerV2(w http.ResponseWriter, r *http.Request) {

	// setting default response type as json
	w.Header().Set("Content-Type", "application/json")

	server, err := GetServerRespondOnError(w, r)
	if err != nil {
		metrics.MessageSendErrors.Inc()
		return
	}

	// Declare a new Person struct.
	var requestV2 models.QPSendDocumentRequestV2

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err = json.NewDecoder(r.Body).Decode(&requestV2)
	if err != nil {
		metrics.MessageSendErrors.Inc()
		RespondServerError(server, w, err)
		return
	}

	if requestV2.Attachment == (models.QPAttachmentV1{}) {
		metrics.MessageSendErrors.Inc()
		RespondServerError(server, w, fmt.Errorf("attachment not found"))
		return
	}

	request := requestV2.ToQpSendRequest()
	request.TrackId = GetTrackId(r)

	waMsg, err := request.ToWhatsappMessage()
	if err != nil {
		metrics.MessageSendErrors.Inc()
		return
	}

	atts := request.ToWhatsappAttachment()

	waMsg.Attachment = atts.Attach
	waMsg.Type = whatsapp.GetMessageType(atts.Attach)

	sendResponse, err := server.SendMessage(waMsg)
	if err != nil {
		metrics.MessageSendErrors.Inc()
		RespondServerError(server, w, err)
		return
	}

	response := &models.QpSendResponseV2{}
	response.Chat.ID = waMsg.Chat.Id
	response.Chat.UserName = waMsg.Chat.Id
	response.Chat.Title = server.GetChatTitle(waMsg.Chat.Id)
	response.From.ID = server.Wid
	response.From.UserName = server.GetNumber()
	response.ID = sendResponse.GetId()

	// Para manter a compatibilidade
	response.PreviusV1 = models.QPSendResult{
		Source:    server.GetWId(),
		Recipient: waMsg.Chat.Id,
		MessageId: sendResponse.GetId(),
	}

	metrics.MessagesSent.Inc()
	RespondSuccess(w, response)
}

// AttachmentHandler renders route POST "/v1/bot/{token}/attachment"
func AttachmentAPIHandlerV2(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	server, err := models.GetServerFromToken(token)
	if err != nil {
		RespondNoContentV2(w, fmt.Errorf("token '%s' not found", token))
		return
	}

	// Checking for ready state
	status := server.GetStatus()
	if status != whatsapp.Ready {
		RespondNotReady(w, &ApiServerNotReadyException{Wid: server.GetWId(), Status: status})
		return
	}

	// Declare a new Person struct.
	var p models.QPAttachmentV1

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err = json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		RespondServerError(server, w, err)
	}

	ss := strings.Split(p.Url, "/")
	id := ss[len(ss)-1]

	att, err := server.Download(id, false)
	if err != nil {
		RespondServerError(server, w, err)
		return
	}

	if len(att.FileName) > 0 {
		w.Header().Set("Content-Disposition", "attachment; filename="+att.FileName)
	}

	if len(att.Mimetype) > 0 {
		w.Header().Set("Content-Type", att.Mimetype)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(*att.GetContent())
}

func WebHookAPIHandlerV2(w http.ResponseWriter, r *http.Request) {

	// setting default response type as json
	w.Header().Set("Content-Type", "application/json")

	response := &models.QpWebhookResponse{}

	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	logentry := server.GetLogger()

	// reading body to avoid converting to json if empty
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Declare a new Person struct.
	var webhook *models.QpWebhook

	if len(body) > 0 {

		// Try to decode the request body into the struct. If there is an error,
		// respond to the client with the error message and a 400 status code.
		err = json.Unmarshal(body, &webhook)
		if err != nil {
			jsonError := fmt.Errorf("error converting body to json: %v", err.Error())
			response.ParseError(jsonError)
			RespondInterface(w, response)
			return
		}
	}

	// creating an empty webhook, to filter or clear it all
	if webhook == nil {
		webhook = &models.QpWebhook{}
	}

	// updating wid for logging and response headers
	webhook.Wid = server.Wid

	switch os := r.Method; os {
	case http.MethodPost:
		affected, err := server.WebhookAddOrUpdate(webhook)
		if err != nil {
			response.ParseError(err)
			RespondInterface(w, response)
		} else {
			response.Affected = affected
			response.ParseSuccess("updated with success")
			RespondSuccess(w, response)
			if affected > 0 {
				logentry.Infof("updating webhook url=%s, items affected: %v", webhook.Url, affected)
			}
		}
		return
	case http.MethodDelete:
		affected, err := server.WebhookRemove(webhook.Url)
		if err != nil {
			response.ParseError(err)
			RespondInterface(w, response)
		} else {
			response.Affected = affected
			response.ParseSuccess("deleted with success")
			RespondSuccess(w, response)
			if affected > 0 {
				logentry.Infof("removing webhook url=%s, items affected: %v", webhook.Url, affected)
			}
		}
		return
	default:
		url := r.Header.Get("X-QUEPASA-WHURL")
		response.Webhooks = server.GetWebHooksByUrl(url)
		if len(url) > 0 {
			response.ParseSuccess(fmt.Sprintf("getting with filter, url=%s", url))
		} else {
			response.ParseSuccess("getting without filter")
		}

		RespondSuccess(w, response)
		return
	}
}
