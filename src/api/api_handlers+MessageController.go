package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

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

// MarkReadController marks one or more messages as read on the WhatsApp connection
// @Summary Mark messages as read
// @Description Marks one or more messages as read by id. Accepts a JSON array in the request body (e.g. ["id1","id2"] or [{"id":"id1"},{"id":"id2"}])
// @Tags Message
// @Accept json
// @Produce json
// @Param request body []string true "Array of message ids or objects with id field"
// @Success 200 {object} models.QpResponse
// @Failure 400 {object} models.QpResponse
// @Security ApiKeyAuth
// @Router /read [post]
func MarkReadController(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	response := &models.QpResponse{}

	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Read body as JSON array. Support two formats for convenience:
	//  - array of strings: ["id1","id2"]
	//  - array of objects: [{"id":"id1"},{"id":"id2"}]

	if r.ContentLength == 0 {
		response.ParseError(fmt.Errorf("empty request body"))
		RespondInterface(w, response)
		return
	}

	var raw any
	err = json.NewDecoder(r.Body).Decode(&raw)
	if err != nil {
		response.ParseError(fmt.Errorf("invalid json body: %s", err.Error()))
		RespondInterface(w, response)
		return
	}

	var ids []string

	switch v := raw.(type) {
	case []any:
		for _, item := range v {
			switch it := item.(type) {
			case string:
				ids = append(ids, it)
			case map[string]any:
				if val, ok := it["id"]; ok {
					if s, ok := val.(string); ok {
						ids = append(ids, s)
					}
				}
			}
		}
	default:
		response.ParseError(fmt.Errorf("expected json array in body"))
		RespondInterface(w, response)
		return
	}

	if len(ids) == 0 {
		response.ParseError(fmt.Errorf("no message ids found in body"))
		RespondInterface(w, response)
		return
	}

	// Iterate and mark each id. Collect errors to return a meaningful response.
	var errs []string
	for _, id := range ids {
		if len(id) == 0 {
			errs = append(errs, "empty id")
			continue
		}

		err := server.MarkRead(id)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %s", id, err.Error()))
		}
	}

	if len(errs) > 0 {
		response.ParseError(fmt.Errorf("errors: %s", strings.Join(errs, "; ")))
		RespondInterface(w, response)
		return
	}

	response.ParseSuccess("marked as read")
	RespondSuccess(w, response)
}
