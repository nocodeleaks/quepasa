package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// PresenceRequest defines the parameters for controlling global presence
type PresenceRequest struct {
	Presence string `json:"presence"` // "available" or "unavailable"
}

// PresenceController handles API requests for global presence status
//
//	@Summary		Control global presence
//	@Description	Controls the bot's global presence status (available/unavailable) on WhatsApp
//	@Tags			Presence
//	@Accept			json
//	@Produce		json
//	@Param			request	body		PresenceRequest	true	"Presence request (available or unavailable)"
//	@Success		200		{object}	models.QpResponse
//	@Failure		400		{object}	models.QpResponse
//	@Failure		503		{object}	models.QpResponse
//	@Security		ApiKeyAuth
//	@Router			/presence [post]
func PresenceController(w http.ResponseWriter, r *http.Request) {
	// Setting default response type as json
	w.Header().Set("Content-Type", "application/json")

	response := &models.QpResponse{}

	// Get server
	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	logentry := server.GetLogger()

	// Parse request body
	var request *PresenceRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.ParseError(fmt.Errorf("invalid request: %v", err))
		RespondInterface(w, response)
		return
	}

	// Validate presence value
	if request.Presence == "" {
		response.ParseError(fmt.Errorf("presence is required (available or unavailable)"))
		RespondInterface(w, response)
		return
	}

	// Validate presence value
	if request.Presence != "available" && request.Presence != "unavailable" {
		response.ParseError(fmt.Errorf("invalid presence value: %s (must be 'available' or 'unavailable')", request.Presence))
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

	// Send presence update
	err = server.SendPresence(request.Presence)
	if err != nil {
		err = fmt.Errorf("failed to send presence update: %s", err.Error())
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	logentry.Infof("global presence set to: %s", request.Presence)

	message := fmt.Sprintf("presence updated to %s", request.Presence)

	// Create successful response
	response.Success = true
	response.ParseSuccess(message)
	RespondInterface(w, response)
}
