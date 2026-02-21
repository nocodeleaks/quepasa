package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// ReceiveAPIHandler renders route GET "/receive"
//
//	@Summary		Receive messages
//	@Description	Retrieves pending messages from WhatsApp with optional cache filters
//	@Tags			Message
//	@Accept			json
//	@Produce		json
//	@Param			timestamp	query		string	false	"Timestamp filter for messages"
//	@Param			exceptions	query		string	false	"Filter by exceptions error status: 'true' for messages with exceptions errors, 'false' for messages without exceptions errors, omit for all messages"
//	@Param			type		query		string	false	"Filter by message type (supports comma-separated list)"
//	@Param			category	query		string	false	"Filter by category: sent, received, sync, unhandled, events"
//	@Param			search		query		string	false	"Search text in id, chat, text, trackid, participant and exceptions"
//	@Param			fromme		query		string	false	"Filter by fromme boolean: true or false"
//	@Param			fromhistory	query		string	false	"Filter by fromhistory boolean: true or false"
//	@Param			chatid		query		string	false	"Filter by chat id (contains)"
//	@Param			messageid	query		string	false	"Filter by message id (contains)"
//	@Param			trackid		query		string	false	"Filter by track id (contains)"
//	@Success		200			{object}	models.QpReceiveResponse
//	@Failure		400			{object}	models.QpResponse
//	@Security		ApiKeyAuth
//	@Router			/receive [get]
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

	filters := GetReceiveMessageFilters(r)

	messages := GetOrderedMessagesWithFilters(server, timestamp, filters)

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

	appliedFilters := []string{}
	if filters.Exceptions != "" {
		appliedFilters = append(appliedFilters, "exceptions="+filters.Exceptions)
	}
	if filters.Type != "" {
		appliedFilters = append(appliedFilters, "type="+filters.Type)
	}
	if filters.Category != "" {
		appliedFilters = append(appliedFilters, "category="+filters.Category)
	}
	if filters.Search != "" {
		appliedFilters = append(appliedFilters, "search="+filters.Search)
	}
	if filters.FromMe != "" {
		appliedFilters = append(appliedFilters, "fromme="+filters.FromMe)
	}
	if filters.FromHistory != "" {
		appliedFilters = append(appliedFilters, "fromhistory="+filters.FromHistory)
	}
	if filters.ChatID != "" {
		appliedFilters = append(appliedFilters, "chatid="+filters.ChatID)
	}
	if filters.MessageID != "" {
		appliedFilters = append(appliedFilters, "messageid="+filters.MessageID)
	}
	if filters.TrackID != "" {
		appliedFilters = append(appliedFilters, "trackid="+filters.TrackID)
	}
	if len(appliedFilters) > 0 {
		msg += ", filters: " + strings.Join(appliedFilters, ", ")
	}

	response.ParseSuccess(msg)
	RespondSuccess(w, response)
}
