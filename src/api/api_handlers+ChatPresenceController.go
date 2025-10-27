package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// TypingRequest defines the parameters for controlling typing indicators
type ChatPresenceRequest struct {
	ChatId   string                            `json:"chatid"`             // Required: Chat to show typing in
	Type     whatsapp.WhatsappChatPresenceType `json:"type"`               // Text or audio
	Duration uint                              `json:"duration,omitempty"` // Optional: Auto-stop after duration (ms)
}

// ChatPresenceController handles API requests for typing indicators
//
//	@Summary		Control chat presence
//	@Description	Controls typing indicators and chat presence in WhatsApp conversations
//	@Tags			Chat
//	@Accept			json
//	@Produce		json
//	@Param			request	body		ChatPresenceRequest	true	"Chat presence request"
//	@Success		200		{object}	models.QpResponse
//	@Failure		400		{object}	models.QpResponse
//	@Security		ApiKeyAuth
//	@Router			/chat/presence [post]
func ChatPresenceController(w http.ResponseWriter, r *http.Request) {
	// Setting default response type as json
	w.Header().Set("Content-Type", "application/json")

	response := &models.QpResponse{}

	// Get server
	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	logentry := server.GetLogger()

	// Parse request body
	var request *ChatPresenceRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.ParseError(fmt.Errorf("invalid request: %v", err))
		RespondInterface(w, response)
		return
	}

	// Validate request
	if request.ChatId == "" {
		response.ParseError(fmt.Errorf("chatid is required"))
		RespondInterface(w, response)
		return
	}

	// Format and validate the chat ID
	formattedChatId, err := whatsapp.FormatEndpoint(request.ChatId)
	if err != nil {
		response.ParseError(fmt.Errorf("invalid chatid: %v", err))
		RespondInterface(w, response)
		return
	}

	request.ChatId = formattedChatId

	// updating logentry with chatid
	logentry = logentry.WithField(LogFields.ChatId, request.ChatId)

	// Checking for ready state
	status := server.GetStatus()
	if status != whatsapp.Ready {
		err = &ApiServerNotReadyException{Wid: server.GetWId(), Status: status}
		response.ParseError(err)
		RespondInterfaceCode(w, response, http.StatusServiceUnavailable)
		return
	}

	err = server.SendChatPresence(request.ChatId, request.Type)
	if err != nil {
		err = fmt.Errorf("failed to send presence update: %s", err.Error())
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	//ChatPresenceRequestsController.Cancel(request.ChatId)

	//logentry.Debug("sent presence indicator")
	/*found := ChatPresenceRequestsController.Cancel(request.ChatId)

	// For paused type, just cancel and send a single presence update
	err = server.SendChatPresence(request.ChatId, request.Type)
	if err != nil {
		err = fmt.Errorf("failed to send presence update: %s", err.Error())
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	logentry.Debug("sent paused indicator")

	if request.Type != whatsapp.WhatsappChatPresenceTypePaused {

		// create an async request to send typing indicator
		ChatPresenceRequestsController.Append(request, server)
		logentry.Debugf("started presence indicator %s with duration %d ms", request.Type, request.Duration)
	}

	message := fmt.Sprintf("presence indicator %s, previous: %v", request.Type, found)*/

	message := fmt.Sprintf("presence indicator %s", request.Type)

	// Create successful response
	response.Success = true
	response.ParseSuccess(message)
	RespondInterface(w, response)
}
