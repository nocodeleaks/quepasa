package api

import (
	"fmt"
	"net/http"
	"time"

	metrics "github.com/nocodeleaks/quepasa/metrics"
	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// ReceiveAPIHandler renders route GET "/{version}/bot/{token}/receive"
// @Summary Receive messages
// @Description Retrieves pending messages from WhatsApp with optional timestamp filtering
// @Tags Message
// @Accept json
// @Produce json
// @Param timestamp query string false "Timestamp filter for messages"
// @Success 200 {object} models.QpReceiveResponse
// @Failure 400 {object} models.QpResponse
// @Security ApiKeyAuth
// @Router /v3/bot/{token}/receive [get]
// @Router /v2/bot/{token}/receive [get]
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

// SendDocumentFromBinary handles route "/sendbinary"
// @Summary Send binary file directly from request body
// @Description Send any binary file (audio, video, image, document) using raw binary data in request body. Supports multiple parameter methods (path, query, headers).
// @Tags Send
// @Accept application/octet-stream,audio/mpeg,video/mp4,image/jpeg,image/png,application/pdf
// @Produce json
// @Param chatid path string false "Chat ID (path parameter)"
// @Param filename path string false "File name (path parameter)"
// @Param text path string false "Caption text for images (path parameter)"
// @Param chatId query string false "Chat ID (query parameter)"
// @Param filename query string false "File name (query parameter)"
// @Param text query string false "Caption text for images (query parameter)"
// @Param inreply query string false "Message ID to reply to"
// @Param X-QUEPASA-CHATID header string false "Chat ID (header parameter)"
// @Param X-QUEPASA-FILENAME header string false "File name (header parameter)"
// @Param X-QUEPASA-TEXT header string false "Caption text for images (header parameter)"
// @Param X-QUEPASA-TRACKID header string false "Track ID for message tracking"
// @Param Content-Type header string true "MIME type of the binary file (e.g., audio/mpeg, video/mp4, image/jpeg)"
// @Success 200 {object} models.QpSendResponse
// @Failure 400 {object} models.QpSendResponse
// @Security ApiKeyAuth
// @Router /sendbinary [post]
// @Router /sendbinary/{chatid} [post]
// @Router /sendbinary/{chatid}/{filename} [post]
// @Router /sendbinary/{chatid}/{filename}/{text} [post]
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
