package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
		waMsg.Participant.Id = server.WId
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
	response.From.ID = server.WId
	response.From.UserName = server.GetNumber()
	response.ID = sendResponse.GetId()

	// Para manter a compatibilidade
	response.PreviusV1 = models.QPSendResult{
		Source:    server.GetWid(),
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
		err = &ApiServerNotReadyException{Wid: server.WId, Status: status}
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

	// setting default reponse type as json
	w.Header().Set("Content-Type", "application/json")

	server, err := GetServerRespondOnError(w, r)
	if err != nil {
		metrics.MessageSendErrors.Inc()
		return
	}

	// Declare a new Person struct.
	var request models.QPSendDocumentRequestV2

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		metrics.MessageSendErrors.Inc()
		RespondServerError(server, w, err)
		return
	}

	if request.Attachment == (models.QPAttachmentV1{}) {
		metrics.MessageSendErrors.Inc()
		RespondServerError(server, w, fmt.Errorf("attachment not found"))
		return
	}

	trackid := GetTrackId(r)
	waMsg, err := whatsapp.ToMessage(request.Recipient, request.Message, trackid)
	if err != nil {
		metrics.MessageSendErrors.Inc()
		return
	}

	attach, err := models.ToWhatsappAttachment(&request.Attachment)
	if err != nil {
		metrics.MessageSendErrors.Inc()
		RespondServerError(server, w, err)
		return
	}

	waMsg.Attachment = attach
	waMsg.Type = whatsapp.GetMessageType(attach.Mimetype)

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
	response.From.ID = server.WId
	response.From.UserName = server.GetNumber()
	response.ID = sendResponse.GetId()

	// Para manter a compatibilidade
	response.PreviusV1 = models.QPSendResult{
		Source:    server.GetWid(),
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
		RespondNoContent(w, fmt.Errorf("Token '%s' not found", token))
		return
	}

	// Checking for ready state
	status := server.GetStatus()
	if status != whatsapp.Ready {
		RespondNotReady(w, &ApiServerNotReadyException{Wid: server.GetWid(), Status: status})
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

	// setting default reponse type as json
	w.Header().Set("Content-Type", "application/json")

	response := &models.QpWebhookResponse{}

	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// reading body to avoid converting to json if empty
	body, err := ioutil.ReadAll(r.Body)
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

	switch os := r.Method; os {
	case http.MethodPost:
		affected, err := server.WebhookAdd(webhook)
		if err != nil {
			response.ParseError(err)
			RespondInterface(w, response)
		} else {
			response.Affected = affected
			response.ParseSuccess("updated with success")
			RespondSuccess(w, response)
			if affected > 0 {
				server.Log.Infof("updating webhook url: %s, items affected: %v", webhook.Url, affected)
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
				server.Log.Infof("removing webhook url: %s, items affected: %v", webhook.Url, affected)
			}
		}
		return
	default:
		url := r.Header.Get("X-QUEPASA-WHURL")
		response.Webhooks = filterByUrlV2(server.Webhooks, url)
		if len(url) > 0 {
			response.ParseSuccess(fmt.Sprintf("getting with filter: %s", url))
		} else {
			response.ParseSuccess("getting without filter")
		}

		RespondSuccess(w, response)
		return
	}
}

func filterByUrlV2(source []*models.QpWebhook, filter string) (out []models.QpWebhook) {
	for _, element := range source {
		if strings.Contains(element.Url, filter) {
			out = append(out, *element)
		}
	}
	return
}
