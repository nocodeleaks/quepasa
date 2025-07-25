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

	response := api.UserIdentifierResponse{UserIdentifierRequest: request}

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

	}

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

	// Use convertPhoneToJid to validate and format the phone number
	jid, err := convertPhoneToJid(response.Phone)
	if err != nil {
		response.ParseError(fmt.Errorf("failed to process phone number: %v", err))
		RespondInterface(w, response)
		return
	}

	if len(jid) == 0 {
		response.ParseError(fmt.Errorf("invalid phone number"))
		RespondInterface(w, response)
		return
	}

	// Extract the phone part from the JID for the response
	processedPhone := response.Phone

	// Try to get LID with original phone number first
	lid, err := server.GetLIDFromPhone(processedPhone)

	// If not found and Brazilian 9-digit handling is enabled, try alternative formats
	if err != nil && models.ENV.ShouldRemoveDigit9() {
		// Extract phone number with country code
		phoneWithCountry, phoneErr := library.GetPhoneIfValid(processedPhone)
		if phoneErr == nil {
			// Try to remove the 9th digit if eligible (Brazilian mobile phones)
			phoneWithout9, removeErr := library.RemoveDigit9IfElegible(phoneWithCountry)
			if removeErr == nil {
				// Extract phone number without country code and + sign
				phoneWithout9Clean := strings.TrimPrefix(phoneWithout9, "+")

				// Try with the phone number without the 9th digit
				lid, err = server.GetLIDFromPhone(phoneWithout9Clean)
				if err == nil {
					// Update the processed phone for response
					processedPhone = phoneWithout9Clean
				}
			}
		}

		// If still not found, try the original logic but with + prefix
		if err != nil && !strings.HasPrefix(processedPhone, "+") {
			phoneWithPlus := "+" + processedPhone
			phoneExtracted, extractErr := library.GetPhoneIfValid(phoneWithPlus)
			if extractErr == nil {
				phoneWithout9, removeErr := library.RemoveDigit9IfElegible(phoneExtracted)
				if removeErr == nil {
					phoneWithout9Clean := strings.TrimPrefix(phoneWithout9, "+")
					lid, err = server.GetLIDFromPhone(phoneWithout9Clean)
					if err == nil {
						processedPhone = phoneWithout9Clean
					}
				}
			}
		}
	}

	// If still no LID found, return the original error
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Set response data
	response.Phone = processedPhone
	response.LId = lid
	response.ParseSuccess("LID found successfully")

	RespondSuccess(w, response)
}

// Helper function to convert phone numbers or partial JIDs to full JIDs
func convertPhoneToJid(phone string) ([]string, error) {
	result := make([]string, 0)

	// If it already contains @, assume it's a JID
	if strings.Contains(phone, "@") {
		result = append(result, phone)
	} else {
		// Otherwise, treat as a phone number and convert to JID format
		result = append(result, phone+"@s.whatsapp.net")
	}

	return result, nil
}

//endregion
