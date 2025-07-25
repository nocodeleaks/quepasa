package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

//region CONTROLLER - USER INFO

type UserInfoRequest struct {
	JIDs []string `json:"jids"`
}

type UserInfoResponse struct {
	models.QpResponse
	Total     int           `json:"total"`
	UserInfos []interface{} `json:"userinfos"`
}

func UserInfoController(w http.ResponseWriter, r *http.Request) {
	// setting default response type as json
	w.Header().Set("Content-Type", "application/json")

	response := &UserInfoResponse{}

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

	var request UserInfoRequest

	// Try to decode the request body into the struct
	err = json.Unmarshal(body, &request)
	if err != nil {
		jsonError := fmt.Errorf("error converting body to json: %v", err.Error())
		response.ParseError(jsonError)
		RespondInterface(w, response)
		return
	}

	if len(request.JIDs) == 0 {
		err = fmt.Errorf("jids array is required and cannot be empty")
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Validate and format JIDs - replace original JIDs with formatted ones
	for i, jid := range request.JIDs {
		// Format the JID if it doesn't contain @
		formattedJID, err := whatsapp.FormatEndpoint(jid)
		if err != nil {
			response.ParseError(fmt.Errorf("invalid JID format for %s: %v", jid, err))
			RespondInterface(w, response)
			return
		}
		if formattedJID == "" {
			response.ParseError(fmt.Errorf("JID cannot be empty after formatting: %s", jid))
			RespondInterface(w, response)
			return
		}
		// Replace the original JID with the formatted one
		request.JIDs[i] = formattedJID
	}

	// Get user info from the server using the formatted JIDs
	userInfos, err := server.GetUserInfo(request.JIDs)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	response.Total = len(userInfos)
	response.UserInfos = userInfos
	response.ParseSuccess("User info retrieved successfully")

	RespondSuccess(w, response)
}

//endregion
