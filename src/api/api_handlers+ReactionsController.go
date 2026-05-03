package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// ReactionRequest defines the parameters for sending or removing a message reaction.
type ReactionRequest struct {
	// ChatId is the conversation that contains the target message.
	ChatId string `json:"chatid"`

	// MessageId is the ID of the message to react to.
	MessageId string `json:"messageid"`

	// FromMe indicates whether the target message was sent by the session owner.
	// Required to build the correct WhatsApp message key.
	FromMe bool `json:"fromme"`

	// Emoji is the reaction emoji (e.g. "👍"). Send empty string to remove reaction.
	Emoji string `json:"emoji"`
}

// SendReactionController sends an emoji reaction to a specific message.
//
//	@Summary		Send message reaction
//	@Description	Sends an emoji reaction to a specific WhatsApp message. Send emoji="" to remove an existing reaction.
//	@Tags			Messages
//	@Accept			json
//	@Produce		json
//	@Param			request	body		ReactionRequest	true	"Reaction request"
//	@Success		200		{object}	models.QpResponse
//	@Failure		400		{object}	models.QpResponse
//	@Failure		503		{object}	models.QpResponse
//	@Security		ApiKeyAuth
//	@Router			/messages/react [post]
func SendReactionController(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := &models.QpResponse{}

	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	status := server.GetStatus()
	if status != whatsapp.Ready {
		err = &ApiServerNotReadyException{Wid: server.GetWId(), Status: status}
		response.ParseError(err)
		RespondInterfaceCode(w, response, http.StatusServiceUnavailable)
		return
	}

	var request ReactionRequest
	if err = json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.ParseError(fmt.Errorf("invalid request body: %w", err))
		RespondInterface(w, response)
		return
	}

	if request.ChatId == "" {
		response.ParseError(fmt.Errorf("chatid is required"))
		RespondInterface(w, response)
		return
	}

	if request.MessageId == "" {
		response.ParseError(fmt.Errorf("messageid is required"))
		RespondInterface(w, response)
		return
	}

	conn, err := server.GetValidConnection()
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	err = conn.SendReaction(request.ChatId, request.MessageId, request.FromMe, request.Emoji)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	if request.Emoji == "" {
		response.ParseSuccess("reaction removed")
	} else {
		response.ParseSuccess("reaction sent")
	}

	RespondInterface(w, response)
}

// RemoveReactionController removes the emoji reaction from a specific message.
//
//	@Summary		Remove message reaction
//	@Description	Removes an emoji reaction previously sent to a WhatsApp message.
//	@Tags			Messages
//	@Accept			json
//	@Produce		json
//	@Param			request	body		ReactionRequest	true	"Reaction request (emoji field is ignored)"
//	@Success		200		{object}	models.QpResponse
//	@Failure		400		{object}	models.QpResponse
//	@Failure		503		{object}	models.QpResponse
//	@Security		ApiKeyAuth
//	@Router			/messages/react [delete]
func RemoveReactionController(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := &models.QpResponse{}

	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	status := server.GetStatus()
	if status != whatsapp.Ready {
		err = &ApiServerNotReadyException{Wid: server.GetWId(), Status: status}
		response.ParseError(err)
		RespondInterfaceCode(w, response, http.StatusServiceUnavailable)
		return
	}

	var request ReactionRequest
	if err = json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.ParseError(fmt.Errorf("invalid request body: %w", err))
		RespondInterface(w, response)
		return
	}

	if request.ChatId == "" {
		response.ParseError(fmt.Errorf("chatid is required"))
		RespondInterface(w, response)
		return
	}

	if request.MessageId == "" {
		response.ParseError(fmt.Errorf("messageid is required"))
		RespondInterface(w, response)
		return
	}

	conn, err := server.GetValidConnection()
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Empty emoji removes the reaction on WhatsApp
	if err = conn.SendReaction(request.ChatId, request.MessageId, request.FromMe, ""); err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	response.ParseSuccess("reaction removed")
	RespondInterface(w, response)
}
