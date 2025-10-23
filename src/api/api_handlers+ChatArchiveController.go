package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	whatsmeow "github.com/nocodeleaks/quepasa/whatsmeow"
)

// ChatArchiveRequest defines the parameters for archiving/unarchiving a chat
type ChatArchiveRequest struct {
	ChatId  string `json:"chatid"`  // Required: Chat to archive/unarchive
	Archive bool   `json:"archive"` // Required: true to archive, false to unarchive
}

// ArchiveChatController handles API requests for archiving or unarchiving a chat
//
//	@Summary		Archive or unarchive chat
//	@Description	Archives or unarchives a WhatsApp chat. Archiving also unpins the chat automatically.
//	@Tags			Chat
//	@Accept			json
//	@Produce		json
//	@Param			request	body		ChatArchiveRequest	true	"Chat archive request"
//	@Success		200		{object}	models.QpResponse
//	@Failure		400		{object}	models.QpResponse
//	@Failure		503		{object}	models.QpResponse
//	@Security		ApiKeyAuth
//	@Router			/chat/archive [post]
func ArchiveChatController(w http.ResponseWriter, r *http.Request) {
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
	var request *ChatArchiveRequest
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

	// Archive or unarchive chat
	err = whatsmeow.ArchiveChat(conn, request.ChatId, request.Archive)
	if err != nil {
		action := "archive"
		if !request.Archive {
			action = "unarchive"
		}
		err = fmt.Errorf("failed to %s chat: %s", action, err.Error())
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	action := "archived"
	if !request.Archive {
		action = "unarchived"
	}

	logentry.Infof("chat %s: %s", action, request.ChatId)

	// Create successful response
	response.Success = true
	response.ParseSuccess(fmt.Sprintf("chat %s %s successfully", request.ChatId, action))
	RespondInterface(w, response)
}
