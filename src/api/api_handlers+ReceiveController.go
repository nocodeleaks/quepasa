package api

import (
	"fmt"
	"net/http"
	"time"

	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// ReceiveAPIHandler renders route GET "/receive"
// @Summary Receive messages
// @Description Retrieves pending messages from WhatsApp with optional timestamp filtering and dispatch error filtering
// @Tags Message
// @Accept json
// @Produce json
// @Param timestamp query string false "Timestamp filter for messages"
// @Param dispatcherror query string false "Filter by dispatch error status: 'true' for messages with dispatch errors, 'false' for messages without dispatch errors, omit for all messages"
// @Success 200 {object} models.QpReceiveResponse
// @Failure 400 {object} models.QpResponse
// @Security ApiKeyAuth
// @Router /receive [get]

func ReceiveAPIHandler(w http.ResponseWriter, r *http.Request) {
	response := &models.QpReceiveResponse{}

	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Checking for ready state
	status := server.GetStatus()
	if status != whatsapp.Ready {
		err = &ApiServerNotReadyException{Wid: server.GetWId(), Status: status}
		response.ParseError(err)
		RespondInterfaceCode(w, response, http.StatusServiceUnavailable)
		return
	}

	if server.Handler == nil {
		err = fmt.Errorf("handlers not attached")
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	response.Total = server.Handler.Count()

	timestamp, err := GetTimestamp(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Get dispatch error filter parameter
	queryValues := r.URL.Query()
	dispatchErrorFilter := queryValues.Get("dispatcherror")

	messages := GetOrderedMessagesWithDispatchFilter(server, timestamp, dispatchErrorFilter)

	response.Server = server.QpServer
	response.Messages = messages
	response.Total = uint64(len(messages))

	// Build success message with filter information
	var msg string
	if timestamp > 0 {
		searchTime := time.Unix(timestamp, 0)
		msg = fmt.Sprintf("getting with timestamp: %v => %s", timestamp, searchTime)
	} else {
		msg = "getting without timestamp filter"
	}

	if dispatchErrorFilter != "" {
		msg += fmt.Sprintf(", dispatch_error filter: %s", dispatchErrorFilter)
	}

	response.ParseSuccess(msg)
	RespondSuccess(w, response)
}
