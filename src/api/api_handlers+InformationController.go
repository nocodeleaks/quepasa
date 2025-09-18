package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	models "github.com/nocodeleaks/quepasa/models"
)

//region CONTROLLER - Information

// InformationController manages bot/server information and settings
// @Summary Manage bot information
// @Description Get, update, or delete bot/server information and settings
// @Tags Information
// @Accept json
// @Produce json
// @Param request body object{settings=object} false "Settings update (for PATCH)"
// @Success 200 {object} models.QpInfoResponse
// @Failure 400 {object} models.QpResponse
// @Security ApiKeyAuth
// @Router /info [get]
// @Router /info [patch]
// @Router /info [delete]
func InformationController(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodPatch:
		InformationPatchRequest(w, r)
	case http.MethodGet:
		InformationGetRequest(w, r)
	case http.MethodDelete:
		InformationDeleteRequest(w, r)
	default:
		err := fmt.Errorf("invalid http method: %s", r.Method)
		RespondErrorCode(w, err, http.StatusMethodNotAllowed)
		return
	}
}

//endregion

func InformationGetRequest(w http.ResponseWriter, r *http.Request) {
	response := &models.QpInfoResponse{}

	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterfaceCode(w, response, http.StatusNoContent)
		return
	}

	response.ParseSuccess(server)
	RespondSuccess(w, response)
}

func InformationPatchRequest(w http.ResponseWriter, r *http.Request) {
	response := &models.QpInfoResponse{}

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

	var request *models.QpInfoPatchRequest

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

	update := ""
	if request.Username != nil {
		jsonUsername := *request.Username
		if len(jsonUsername) > 0 {

			// searching user
			_, err := models.WhatsappService.DB.Users.Find(jsonUsername)
			if err != nil {
				jsonError := fmt.Errorf("user not found: %v", err.Error())
				response.ParseError(jsonError)
				RespondInterface(w, response)
				return
			}
		}

		if !strings.EqualFold(server.User, jsonUsername) {
			server.User = jsonUsername
			update += fmt.Sprintf("username to: {%s}; ", jsonUsername)
		}
	}

	//#region WHATSAPP OPTIONS

	if request.Groups != nil {
		option := *request.Groups

		if server.Groups != option {
			server.Groups = option
			update += fmt.Sprintf("groups to: {%s}; ", option)
		}
	}

	if request.Broadcasts != nil {
		option := *request.Broadcasts

		if server.Broadcasts != option {
			server.Broadcasts = option
			update += fmt.Sprintf("broadcasts to: {%s}; ", option)
		}
	}

	if request.ReadReceipts != nil {
		option := *request.ReadReceipts

		if server.ReadReceipts != option {
			server.ReadReceipts = option
			update += fmt.Sprintf("readreceipts to: {%s}; ", option)
		}
	}

	if request.Calls != nil {
		option := *request.Calls

		if server.Calls != option {
			server.Calls = option
			update += fmt.Sprintf("calls to: {%s}; ", option)
		}
	}

	//#endregion

	if len(update) > 0 {
		err = server.Save("patching info")
		if err != nil {
			response.ParseError(err)
			RespondInterface(w, response)
			return
		}

		info := fmt.Sprintf("server updated: {%s}", update)
		logentry := server.GetLogger()
		logentry.Info(info)

		response.PatchSuccess(server, "server updated")
		RespondSuccess(w, response)
	} else {
		response.PatchSuccess(server, "no update required")
		RespondSuccess(w, response)
	}
}

func InformationDeleteRequest(w http.ResponseWriter, r *http.Request) {
	response := &models.QpResponse{}

	server, err := GetServer(r)
	if err != nil {

		if err == models.ErrServerNotFound {
			RespondNoContent(w)
			return
		}

		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	err = models.WhatsappService.Delete(server)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	response.ParseSuccess("server deleted")
	RespondSuccess(w, response)
}
