package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// BlockContactRequest defines the parameters for blocking or unblocking a contact.
type BlockContactRequest struct {
	// Wid is the WhatsApp ID (JID) of the contact to block or unblock.
	// Example: "5511999999999@s.whatsapp.net"
	Wid string `json:"wid"`
}

// BlockContactController blocks a contact, preventing them from sending messages to this session.
//
//	@Summary		Block contact
//	@Description	Blocks a WhatsApp contact by their JID/WID so they cannot send messages to this session.
//	@Tags			Contacts
//	@Accept			json
//	@Produce		json
//	@Param			request	body		BlockContactRequest	true	"Contact to block"
//	@Success		200		{object}	models.QpResponse
//	@Failure		400		{object}	models.QpResponse
//	@Failure		503		{object}	models.QpResponse
//	@Security		ApiKeyAuth
//	@Router			/contacts/block [post]
func BlockContactController(w http.ResponseWriter, r *http.Request) {
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

	var request BlockContactRequest
	if err = json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.ParseError(fmt.Errorf("invalid request body: %w", err))
		RespondInterface(w, response)
		return
	}

	if request.Wid == "" {
		response.ParseError(fmt.Errorf("wid is required"))
		RespondInterface(w, response)
		return
	}

	manager := server.GetContactManager()
	if manager == nil {
		response.ParseError(fmt.Errorf("contact manager not available"))
		RespondInterface(w, response)
		return
	}

	if err = manager.BlockContact(request.Wid); err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	response.ParseSuccess("contact blocked")
	RespondInterface(w, response)
}

// UnblockContactController removes a previously placed block from a contact.
//
//	@Summary		Unblock contact
//	@Description	Removes a previously placed block from a WhatsApp contact.
//	@Tags			Contacts
//	@Accept			json
//	@Produce		json
//	@Param			request	body		BlockContactRequest	true	"Contact to unblock"
//	@Success		200		{object}	models.QpResponse
//	@Failure		400		{object}	models.QpResponse
//	@Failure		503		{object}	models.QpResponse
//	@Security		ApiKeyAuth
//	@Router			/contacts/block [delete]
func UnblockContactController(w http.ResponseWriter, r *http.Request) {
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

	var request BlockContactRequest
	if err = json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.ParseError(fmt.Errorf("invalid request body: %w", err))
		RespondInterface(w, response)
		return
	}

	if request.Wid == "" {
		response.ParseError(fmt.Errorf("wid is required"))
		RespondInterface(w, response)
		return
	}

	manager := server.GetContactManager()
	if manager == nil {
		response.ParseError(fmt.Errorf("contact manager not available"))
		RespondInterface(w, response)
		return
	}

	if err = manager.UnblockContact(request.Wid); err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	response.ParseSuccess("contact unblocked")
	RespondInterface(w, response)
}
