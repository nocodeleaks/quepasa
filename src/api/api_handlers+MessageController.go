package api

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
	response := &models.QpResponse{}

	// Declare a new request struct
	request := &EditMessageRequest{}

	// Decode the JSON body into the request struct
	err := json.NewDecoder(r.Body).Decode(request)
	if err != nil {
		response.ParseError(fmt.Errorf("invalid JSON in request body: %s", err.Error()))
		RespondInterface(w, response)
		return
	}

	// Validate required fields
	if request.MessageId == "" {
		response.ParseError(fmt.Errorf("messageId is required"))
		RespondInterface(w, response)
		return
	}

	if request.Content == "" {
		response.ParseError(fmt.Errorf("content is required"))
		RespondInterface(w, response)
		return
	}

	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Get the message to be edited
	msg, err := server.Handler.GetById(request.MessageId)
	if err != nil {
		response.ParseError(fmt.Errorf("message not found: %s", err.Error()))
		RespondInterface(w, response)
		return
	}

	// Edit the message
	err = server.GetConnection().Edit(msg, request.Content)
	if err != nil {
		response.ParseError(fmt.Errorf("failed to edit message: %s", err.Error()))
		RespondInterface(w, response)
		return
	}

	response.ParseSuccess("message edited successfully")
	RespondSuccess(w, response)
}

//endregion
