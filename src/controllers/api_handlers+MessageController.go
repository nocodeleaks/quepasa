package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	models "github.com/nocodeleaks/quepasa/models"
)

//region CONTROLLER - Message

func GetMessageController(w http.ResponseWriter, r *http.Request) {

	// setting default response type as json
	w.Header().Set("Content-Type", "application/json")

	response := &models.QpMessageResponse{}

	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Default parameters
	messageid := GetMessageId(r)

	if len(messageid) == 0 {
		err = fmt.Errorf("empty message id")
		response.ParseError(err)
		RespondInterface(w, response)
		return
	} else {

		msg, err := server.Handler.GetById(messageid)
		if err != nil {
			response.ParseError(err)
			RespondInterface(w, response)
			return
		}

		response.ParseSuccess("found")
		response.Message = msg
		RespondSuccess(w, response)
	}
}

func RevokeController(w http.ResponseWriter, r *http.Request) {

	// setting default response type as json
	w.Header().Set("Content-Type", "application/json")

	response := &models.QpResponse{}

	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Default parameters
	messageid := GetMessageId(r)

	if len(messageid) == 0 {
		err = fmt.Errorf("empty message id")
		response.ParseError(err)
		RespondInterface(w, response)
		return
	} else {

		if GetMessageIdAsPrefix(r) {
			errs := server.RevokeByPrefix(messageid)
			if len(errs) > 0 {
				err = errors.Join(errs...)
				response.ParseError(err)
				RespondInterface(w, response)
				return
			}
		} else {
			err = server.Revoke(messageid)
			if err != nil {
				response.ParseError(err)
				RespondInterface(w, response)
				return
			}
		}

		response.ParseSuccess("revoked with success")
		RespondSuccess(w, response)
	}
}

func EditMessageController(w http.ResponseWriter, r *http.Request) {
	// setting default response type as json
	w.Header().Set("Content-Type", "application/json")

	response := &models.QpResponse{}

	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	var request *models.EditMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.ParseError(fmt.Errorf("invalid request: %v", err))
		RespondInterface(w, response)
		return
	}

	if len(request.Content) == 0 {
		err = fmt.Errorf("empty content for edit")
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	if len(request.MessageId) == 0 {
		err = fmt.Errorf("empty message id")
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Check if the message exists

	err = server.Edit(request.MessageId, request.Content)
	if err != nil {
		response.ParseError(err)
	} else {
		response.ParseSuccess("message edited successfully")
	}

	RespondInterface(w, response)
}

//endregion
