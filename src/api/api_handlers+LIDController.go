package api

import (
	"fmt"
	"net/http"
	"strings"

	api "github.com/nocodeleaks/quepasa/api/models"
	library "github.com/nocodeleaks/quepasa/library"
	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

//region CONTROLLER - LID

type LIDRequest struct {
	Phone string `json:"phone"`
}

type LIDResponse struct {
	models.QpResponse
	Phone string `json:"phone,omitempty"`
	LID   string `json:"lid,omitempty"`
}

// GetPhoneController retrieves LID (Local Identifier) for a phone number
//
//	@Summary		Get user identifier (LID)
//	@Description	Retrieves the Local Identifier (LID) for a given phone number
//	@Tags			Contacts
//	@Accept			json
//	@Produce		json
//	@Param			phone	query		string	false	"Phone number"
//	@Param			lid		query		string	false	"Local identifier"
//	@Success		200		{object}	LIDResponse
//	@Failure		400		{object}	models.QpResponse
//	@Security		ApiKeyAuth
//	@Router			/useridentifier [get]
func GetPhoneController(w http.ResponseWriter, r *http.Request) {
	response := &LIDResponse{}
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

	// Get lid from query parameter
	lid := library.GetRequestParameter(r, "lid")
	// Validate lid parameter
	if lid == "" {
		response.ParseError(fmt.Errorf("lid parameter is required"))
		RespondInterface(w, response)
		return
	}

	// validate if the lid has the correct suffix
	if !strings.HasSuffix(lid, "@lid") {
		response.ParseError(fmt.Errorf("lid must have @lid suffix"))
		RespondInterface(w, response)
		return
	}

	if len(lid) == 0 {
		response.ParseError(fmt.Errorf("invalid lid"))
		RespondInterface(w, response)
		return
	}

	// use the method GetPhoneFromLID to return the contact phone, lid, // and any other information
	processedPhone, err := server.GetPhoneFromLID(lid)

	// If still no LID found, return the original error
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}
	// Set response data
	response.Phone = processedPhone
	response.LID = lid
	response.ParseSuccess("LID found successfully")
	RespondSuccess(w, response)
}

func GetUserIdentifierController(w http.ResponseWriter, r *http.Request) {

	request := api.UserIdentifierRequest{}
	request.Phone = library.GetRequestParameter(r, "phone")
	request.LId = library.GetRequestParameter(r, "lid")

	response := api.UserIdentifierResponse{}

	if len(request.Phone) == 0 && len(request.LId) == 0 {
		err := fmt.Errorf("get user identifier controller, missing phone or lid")
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	if len(request.Phone) > 0 {
		phone, err := whatsapp.GetPhoneIfValid(request.Phone)
		response.Phone = phone
		if err != nil {
			response.ParseError(fmt.Errorf("invalid phone number: %v", err))
			RespondInterface(w, response)
			return
		}

		phone = strings.TrimPrefix(phone, "+")

		lid, err := GetLIdFromPhone(r, phone)
		if err != nil {
			response.ParseError(fmt.Errorf("failed to get LId from phone: %v", err))
			RespondInterface(w, response)
			return
		}
		response.LId = lid
	} else if len(request.LId) > 0 {
		phone, err := GetPhoneFromLId(r, request.LId)
		if err != nil {
			response.ParseError(fmt.Errorf("failed to get phone from LId: %v", err))
			RespondInterface(w, response)
			return
		}

		response.LId = request.LId

		if !strings.HasPrefix(phone, "+") {
			phone = "+" + phone
		}
		response.Phone = phone
	}

	response.ParseSuccess("found successfully")
	RespondSuccess(w, response)
}

func GetLIdFromPhone(r *http.Request, phone string) (response string, err error) {
	server, err := GetServer(r)
	if err != nil {
		return response, err
	}

	// Checking for ready state
	status := server.GetStatus()
	if status != whatsapp.Ready {
		err = &ApiServerNotReadyException{Wid: server.GetWId(), Status: status}
		return response, err
	}

	contactManager := server.GetContactManager()
	return contactManager.GetLIDFromPhone(phone)
}

func GetPhoneFromLId(r *http.Request, lid string) (response string, err error) {
	server, err := GetServer(r)
	if err != nil {
		return response, err
	}

	// Checking for ready state
	status := server.GetStatus()
	if status != whatsapp.Ready {
		err = &ApiServerNotReadyException{Wid: server.GetWId(), Status: status}
		return response, err
	}

	contactManager := server.GetContactManager()
	return contactManager.GetPhoneFromLID(lid)
}

//endregion
