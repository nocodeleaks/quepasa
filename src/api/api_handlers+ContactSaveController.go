package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	whatsmeow "github.com/nocodeleaks/quepasa/whatsmeow"
)

// ContactSaveRequest defines the parameters for saving a contact
type ContactSaveRequest struct {
	Phone       string `json:"phone"`                  // Required: phone number (e.g. 5516998824990)
	FullName    string `json:"fullname"`               // Required: full name
	FirstName   string `json:"firstname,omitempty"`    // Optional: first name (defaults to FullName if empty)
	SyncToPhone bool   `json:"synctophone,omitempty"` // Optional: sync to phone address book (default false)
}

// ContactSaveController handles API requests for saving a contact
//
//	@Summary		Save contact
//	@Description	Saves a contact to WhatsApp address book. Set synctophone=true to also sync to the device's contacts (equivalent to "Sync contact with phone" in WhatsApp Web)
//	@Tags			Contacts
//	@Accept			json
//	@Produce		json
//	@Param			request	body		ContactSaveRequest	true	"Contact save request"
//	@Success		200		{object}	models.QpResponse
//	@Failure		400		{object}	models.QpResponse
//	@Failure		503		{object}	models.QpResponse
//	@Security		ApiKeyAuth
//	@Router			/contact/save [post]
func ContactSaveController(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := &models.QpResponse{}

	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	logentry := server.GetLogger()

	var request ContactSaveRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.ParseError(fmt.Errorf("invalid request: %v", err))
		RespondInterface(w, response)
		return
	}

	// Validate required fields
	request.Phone = strings.TrimSpace(request.Phone)
	request.FullName = strings.TrimSpace(request.FullName)
	if request.Phone == "" {
		response.ParseError(fmt.Errorf("phone is required"))
		RespondInterface(w, response)
		return
	}
	if request.FullName == "" {
		response.ParseError(fmt.Errorf("fullname is required"))
		RespondInterface(w, response)
		return
	}

	// Normalize phone: strip leading + and any non-digit chars except leading +
	phone := strings.TrimPrefix(request.Phone, "+")

	// Default firstname to fullname if not provided
	firstName := strings.TrimSpace(request.FirstName)
	if firstName == "" {
		firstName = request.FullName
	}

	// Checking for ready state
	status := server.GetStatus()
	if status != whatsapp.Ready {
		err = &ApiServerNotReadyException{Wid: server.GetWId(), Status: status}
		response.ParseError(err)
		RespondInterfaceCode(w, response, http.StatusServiceUnavailable)
		return
	}

	conn := server.GetConnection().(*whatsmeow.WhatsmeowConnection)

	if err := whatsmeow.SaveContact(conn, phone, request.FullName, firstName, request.SyncToPhone); err != nil {
		response.ParseError(fmt.Errorf("failed to save contact: %v", err))
		RespondInterface(w, response)
		return
	}

	logentry.Infof("saved contact: %s (%s), synctophone=%v", request.FullName, phone, request.SyncToPhone)

	response.Success = true
	response.ParseSuccess(fmt.Sprintf("contact %s (%s) saved", request.FullName, phone))
	RespondInterface(w, response)
}
