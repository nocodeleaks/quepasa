package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	metrics "github.com/nocodeleaks/quepasa/metrics"
	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

func SendPollHandler(w http.ResponseWriter, r *http.Request) {
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
	var request models.PollRequest
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

	// Validate question
	if request.Question == "" {
		metrics.MessageSendErrors.Inc()
		response.ParseError(fmt.Errorf("question is required"))
		RespondInterface(w, response)
		return
	}

	// Validate options
	if len(request.Options) < 2 {
		metrics.MessageSendErrors.Inc()
		response.ParseError(fmt.Errorf("at least 2 options are required"))
		RespondInterface(w, response)
		return
	}

	// Set default max selections if not provided
	if request.MaxSelections <= 0 {
		request.MaxSelections = 1
	}

	// Ensure max selections is valid
	if request.MaxSelections > len(request.Options) {
		request.MaxSelections = len(request.Options)
	}

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

	// Simply use the high-level API to send the poll
	err = server.Sendpoll(request.ChatId, request.Question, request.Options, request.MaxSelections)
	if err != nil {
		metrics.MessageSendErrors.Inc()
		response.ParseError(fmt.Errorf("error sending poll: %v", err))
		RespondInterface(w, response)
		return
	}

	// Log success
	logentry.Infof("Poll created successfully with question: %s", request.Question)
	metrics.MessagesSent.Inc()

	// Prepare success response
	result := &models.QpSendResponseMessage{}
	result.Wid = server.GetWId()
	result.ChatId = request.ChatId
	result.TrackId = request.TrackId

	response.ParseSuccess(result)
	RespondInterface(w, response)
}
