package api

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	models "github.com/nocodeleaks/quepasa/models"
)

// RedispatchAPIHandler forces re-dispatch of a cached message by ID
//
//	@Summary		Re-dispatch message
//	@Description	Forces re-dispatch of a cached message to webhooks/RabbitMQ using the message ID. Applies all original dispatching validations including TrackId, ForwardInternal, message type filters (groups, individuals, broadcasts, calls, read receipts).
//	@Tags			Message
//	@Accept			json
//	@Produce		json
//	@Param			messageid	path		string	true	"Message ID to re-dispatch"
//	@Success		200			{object}	models.QpResponse
//	@Failure		400			{object}	models.QpResponse
//	@Failure		404			{object}	models.QpResponse
//	@Security		ApiKeyAuth
//	@Router			/redispatch/{messageid} [post]
func RedispatchAPIHandler(w http.ResponseWriter, r *http.Request) {
	response := &models.QpResponse{}

	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	if server.Handler == nil {
		err = fmt.Errorf("handlers not attached")
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Get message ID from URL parameter
	messageId := chi.URLParam(r, "messageid")
	if messageId == "" {
		err = fmt.Errorf("message ID is required")
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Get message from cache
	message, err := server.Handler.GetById(messageId)
	if err != nil {
		response.ParseError(err)
		RespondInterfaceCode(w, response, http.StatusNotFound)
		return
	}

	// Force re-dispatch using PostToDispatchingFromServer which applies all original validations:
	// - TrackId validation (avoids loops when ForwardInternal is true)
	// - Message type filters (groups, individuals, broadcasts, calls, read receipts)
	// - Internal message handling
	err = models.PostToDispatchingFromServer(server, message)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	response.ParseSuccess(fmt.Sprintf("message %s re-dispatched successfully", messageId))
	RespondSuccess(w, response)
}
