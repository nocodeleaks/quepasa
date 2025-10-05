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

// GetInformationController handles GET requests for bot/server information
//
//	@Summary		Get bot information
//	@Description	Get bot/server information and settings
//	@Tags			Information
//	@Accept			json
//	@Produce		json
//	@Success		200		{object}	models.QpInfoResponse
//	@Failure		400		{object}	models.QpResponse
//	@Security		ApiKeyAuth
//	@Router			/info [get]
func GetInformationController(w http.ResponseWriter, r *http.Request) {
	InformationGetRequest(w, r)
}

// UpdateInformationController handles PATCH requests for updating bot/server information
//
//	@Summary		Update bot information
//	@Description	Update bot/server information and settings
//	@Tags			Information
//	@Accept			json
//	@Produce		json
//	@Param			request	body		object{settings=object}	false	"Settings update"
//	@Success		200		{object}	models.QpInfoResponse
//	@Failure		400		{object}	models.QpResponse
//	@Security		ApiKeyAuth
//	@Router			/info [patch]
func UpdateInformationController(w http.ResponseWriter, r *http.Request) {
	InformationPatchRequest(w, r)
}

// DeleteInformationController handles DELETE requests for bot/server information
//
//	@Summary		Delete bot information
//	@Description	Delete bot/server information and settings
//	@Tags			Information
//	@Accept			json
//	@Produce		json
//	@Success		200		{object}	models.QpInfoResponse
//	@Failure		400		{object}	models.QpResponse
//	@Security		ApiKeyAuth
//	@Router			/info [delete]
func DeleteInformationController(w http.ResponseWriter, r *http.Request) {
	InformationDeleteRequest(w, r)
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
