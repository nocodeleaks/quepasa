package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	models "github.com/nocodeleaks/quepasa/models"
)

//region CONTROLLER - Message

// GetMessageController retrieves a specific message by ID
// @Summary Get message
// @Description Retrieves a specific message by its ID
// @Tags Message
// @Accept json
// @Produce json
// @Param messageid path string true "Message ID"
// @Success 200 {object} models.QpMessageResponse
// @Failure 400 {object} models.QpResponse
// @Security ApiKeyAuth
// @Router /message/{messageid} [get]
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

// RevokeController revokes/deletes a message
// @Summary Revoke message
// @Description Revokes or deletes a specific message by its ID
// @Tags Message
// @Accept json
// @Produce json
// @Param messageid path string true "Message ID"
// @Success 200 {object} models.QpMessageResponse
// @Failure 400 {object} models.QpResponse
// @Security ApiKeyAuth
// @Router /message/{messageid} [delete]
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

// EditMessageController edits the content of an existing message
// @Summary Edit message
// @Description Edits the content of an existing message by its ID
// @Tags Message
// @Accept json
// @Produce json
// @Param request body object{content=string,messageId=string} true "Message edit request"
// @Success 200 {object} models.QpResponse
// @Failure 400 {object} models.QpResponse
// @Security ApiKeyAuth
// @Router /edit [put]
func EditMessageController(w http.ResponseWriter, r *http.Request) {

	response := &models.QpResponse{}
	// Declare a new request struct
	request := &EditMessageRequest{}

	if r.ContentLength > 0 && r.Method == http.MethodPut {
		// Try to decode the request body into the struct. If there is an error,
		// respond to the client with the error message and a 400 status code.
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			jsonErr := fmt.Errorf("invalid json body: %s", err.Error())
			response.ParseError(jsonErr)
			RespondInterface(w, response)
			return
		}
	}

	if len(request.Content) == 0 {
		response.ParseError(fmt.Errorf("empty content for edit"))
		RespondInterface(w, response)
		return
	}

	if len(request.MessageId) == 0 {
		response.ParseError(fmt.Errorf("empty message id"))
		RespondInterface(w, response)
		return
	}

	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	err = server.Edit(request.MessageId, request.Content)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}
	response.ParseSuccess("message edited successfully")
	RespondSuccess(w, response)
}

//endregion
