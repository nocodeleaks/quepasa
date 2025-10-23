package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	whatsmeow "github.com/nocodeleaks/quepasa/whatsmeow"
)

// ChatReadRequest defines the parameters for marking chat as read/unread
type ChatReadRequest struct {
	ChatId string `json:"chatid"` // Required: Chat to mark as read/unread
}

// MarkChatAsReadController handles API requests for marking a chat as read
//
//	@Summary		Mark chat as read
//	@Description	Marks a WhatsApp chat as read (removes unread badge)
//	@Tags			Chat
//	@Accept			json
//	@Produce		json
//	@Param			request	body		ChatReadRequest	true	"Chat read request"
//	@Success		200		{object}	models.QpResponse
//	@Failure		400		{object}	models.QpResponse
//	@Failure		503		{object}	models.QpResponse
//	@Security		ApiKeyAuth
//	@Router			/chat/markread [post]
func MarkChatAsReadController(w http.ResponseWriter, r *http.Request) {
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
	var request *ChatReadRequest
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

	// Get WhatsmeowConnection
	conn := server.GetConnection().(*whatsmeow.WhatsmeowConnection)

	// Mark chat as read
	err = whatsmeow.MarkChatAsRead(conn, request.ChatId)
	if err != nil {
		err = fmt.Errorf("failed to mark chat as read: %s", err.Error())
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	logentry.Infof("marked chat as read: %s", request.ChatId)

	// Create successful response
	response.Success = true
	response.ParseSuccess(fmt.Sprintf("chat %s marked as read", request.ChatId))
	RespondInterface(w, response)
}

// MarkChatAsUnreadController handles API requests for marking a chat as unread
//
//	@Summary		Mark chat as unread
//	@Description	Marks a WhatsApp chat as unread (shows unread badge)
//	@Tags			Chat
//	@Accept			json
//	@Produce		json
//	@Param			request	body		ChatReadRequest	true	"Chat unread request"
//	@Success		200		{object}	models.QpResponse
//	@Failure		400		{object}	models.QpResponse
//	@Failure		503		{object}	models.QpResponse
//	@Security		ApiKeyAuth
//	@Router			/chat/markunread [post]
func MarkChatAsUnreadController(w http.ResponseWriter, r *http.Request) {
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
	var request *ChatReadRequest
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

	// Get WhatsmeowConnection
	conn := server.GetConnection().(*whatsmeow.WhatsmeowConnection)

	// Mark chat as unread
	err = whatsmeow.MarkChatAsUnread(conn, request.ChatId)
	if err != nil {
		err = fmt.Errorf("failed to mark chat as unread: %s", err.Error())
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	logentry.Infof("marked chat as unread: %s", request.ChatId)

	// Create successful response
	response.Success = true
	response.ParseSuccess(fmt.Sprintf("chat %s marked as unread", request.ChatId))
	RespondInterface(w, response)
}
