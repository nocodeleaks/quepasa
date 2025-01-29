package controllers

import (
	"fmt"
	"net/http"
	"time"

	metrics "github.com/nocodeleaks/quepasa/metrics"
	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// ReceiveAPIHandler renders route GET "/{version}/bot/{token}/receive"
func ReceiveAPIHandler(w http.ResponseWriter, r *http.Request) {
	response := &models.QpReceiveResponse{}

	server, err := GetServer(r)
	if err != nil {
		metrics.MessageReceiveErrors.Inc()
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Checking for ready state
	status := server.GetStatus()
	if status != whatsapp.Ready {
		metrics.MessageReceiveErrors.Inc()
		err = &ApiServerNotReadyException{Wid: server.GetWId(), Status: status}
		response.ParseError(err)
		RespondInterfaceCode(w, response, http.StatusServiceUnavailable)
		return
	}

	if server.Handler == nil {
		metrics.MessageReceiveErrors.Inc()
		err = fmt.Errorf("handlers not attached")
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	response.Total = server.Handler.Count()

	timestamp, err := GetTimestamp(r)
	if err != nil {
		metrics.MessageReceiveErrors.Inc()
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	messages := GetOrderedMessages(server, timestamp)
	metrics.MessagesReceived.Add(float64(len(messages)))

	response.Server = server.QpServer
	response.Messages = messages

	if timestamp > 0 {
		searchTime := time.Unix(timestamp, 0)
		msg := fmt.Sprintf("getting with timestamp: %v => %s", timestamp, searchTime)
		response.ParseSuccess(msg)
	} else {
		response.ParseSuccess("getting without filter")
	}

	RespondSuccess(w, response)
}

/*
<summary>

	Renders route POST "/{version}/bot/{token}/sendbinary/{chatid}/{filename}/{text}"

	Any of then, at this order of priority
	Path parameters: {chatid}
	Path parameters: {filename}
	Path parameters: {text} only images
	Url parameters: ?chatid={chatid}
	Url parameters: ?filename={filename}
	Url parameters: ?text={text} only images
	Header parameters: X-QUEPASA-CHATID = {chatid}
	Header parameters: X-QUEPASA-FILENAME = {filename}
	Header parameters: X-QUEPASA-TEXT = {text} only images

</summary>
*/
func SendDocumentFromBinary(w http.ResponseWriter, r *http.Request) {
	response := &models.QpSendResponse{}

	server, err := GetServer(r)
	if err != nil {
		metrics.MessageSendErrors.Inc()
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Declare a new request struct.
	request := &models.QpSendRequest{}

	// Getting ChatId parameter
	err = request.EnsureValidChatId(r)
	if err != nil {
		metrics.MessageSendErrors.Inc()
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	err = request.GenerateBodyContent(r)
	if err != nil {
		metrics.MessageSendErrors.Inc()
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	SendRequest(w, r, request, server)
}
