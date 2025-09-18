package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	models "github.com/nocodeleaks/quepasa/models"
)

//region CONTROLLER - CONTACTS

// IsOnWhatsappController checks if phone numbers are registered on WhatsApp
// @Summary Check WhatsApp registration
// @Description Checks if provided phone numbers are registered on WhatsApp
// @Tags Contacts
// @Accept json
// @Produce json
// @Param request body object{phones=[]string} true "Phone numbers to check"
// @Success 200 {object} models.QpIsOnWhatsappResponse
// @Failure 400 {object} models.QpResponse
// @Security ApiKeyAuth
// @Router /isonwhatsapp [post]
func IsOnWhatsappController(w http.ResponseWriter, r *http.Request) {

	// setting default response type as json
	w.Header().Set("Content-Type", "application/json")

	response := &models.QpIsOnWhatsappResponse{}

	// reading body to avoid converting to json if empty
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	if len(body) == 0 {
		err = fmt.Errorf("empty body")
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	var request []string

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err = json.Unmarshal(body, &request)
	if err != nil {
		jsonError := fmt.Errorf("error converting body to json: %v", err.Error())
		response.ParseError(jsonError)
		RespondInterface(w, response)
		return
	}

	if request == nil {
		jsonErr := fmt.Errorf("invalid request body: %s", string(body))
		response.ParseError(jsonErr)
		RespondInterface(w, response)
		return
	}

	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	registered, err := server.IsOnWhatsApp(request...)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	response.Total = len(registered)
	response.Registered = registered
	RespondSuccess(w, response)
}

//endregion
