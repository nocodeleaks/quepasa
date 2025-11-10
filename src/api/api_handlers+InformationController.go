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

// CreateInformationController handles POST requests for creating a new bot/server
//
//	@Summary		Create bot configuration
//	@Description	Create a new bot/server with configuration before QR code scanning
//	@Tags			Information
//	@Accept			json
//	@Produce		json
//	@Param			request	body		InfoCreateRequest	true	"Server creation request"
//	@Success		200		{object}	models.QpInfoResponse
//	@Failure		400		{object}	models.QpResponse
//	@Security		ApiKeyAuth
//	@Router			/info [post]
func CreateInformationController(w http.ResponseWriter, r *http.Request) {
	InformationPostRequest(w, r)
}

// GetInformationController handles GET requests for bot/server information
//
//	@Summary		Get bot information
//	@Description	Get bot/server information and settings
//	@Tags			Information
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	models.QpInfoResponse
//	@Failure		400	{object}	models.QpResponse
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
//	@Success		200	{object}	models.QpInfoResponse
//	@Failure		400	{object}	models.QpResponse
//	@Security		ApiKeyAuth
//	@Router			/info [delete]
func DeleteInformationController(w http.ResponseWriter, r *http.Request) {
	InformationDeleteRequest(w, r)
}

//endregion

func InformationPostRequest(w http.ResponseWriter, r *http.Request) {
	response := &models.QpInfoResponse{}

	// Get token from header (authentication)
	token := GetToken(r)
	if len(token) == 0 {
		err := fmt.Errorf("create information controller, missing token")
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Get username from authentication
	username, err := ValidateUsername(r)
	if err != nil {
		err.Prepend("create information controller, username validation")
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Read and validate request body
	body, ioErr := io.ReadAll(r.Body)
	if ioErr != nil {
		response.ParseError(ioErr)
		RespondInterface(w, response)
		return
	}

	var request *InfoCreateRequest
	if len(body) > 0 {
		jsonErr := json.Unmarshal(body, &request)
		if jsonErr != nil {
			jsonError := fmt.Errorf("error converting body to json: %v", jsonErr.Error())
			response.ParseError(jsonError)
			RespondInterface(w, response)
			return
		}
	}

	// Check if server already exists (AddOrUpdate pattern)
	server, exists := models.WhatsappService.Servers[token]

	if exists {
		// UPDATE: Server exists, update configuration
		update := ""

		// Update user if provided and different
		if len(username) > 0 && server.User != username {
			server.User = username
			update += fmt.Sprintf("user to: {%s}; ", username)
		}

		// Apply configuration from request
		if request != nil {
			if request.Groups != nil && server.Groups != *request.Groups {
				server.Groups = *request.Groups
				update += fmt.Sprintf("groups to: {%s}; ", *request.Groups)
			}
			if request.Broadcasts != nil && server.Broadcasts != *request.Broadcasts {
				server.Broadcasts = *request.Broadcasts
				update += fmt.Sprintf("broadcasts to: {%s}; ", *request.Broadcasts)
			}
			if request.ReadReceipts != nil && server.ReadReceipts != *request.ReadReceipts {
				server.ReadReceipts = *request.ReadReceipts
				update += fmt.Sprintf("readreceipts to: {%s}; ", *request.ReadReceipts)
			}
			if request.Calls != nil && server.Calls != *request.Calls {
				server.Calls = *request.Calls
				update += fmt.Sprintf("calls to: {%s}; ", *request.Calls)
			}
			if request.Devel != nil && server.Devel != *request.Devel {
				server.Devel = *request.Devel
				update += fmt.Sprintf("devel to: {%t}; ", *request.Devel)
			}
		}

		// Save if there were changes
		if len(update) > 0 {
			err := server.Save("server configuration updated via POST")
			if err != nil {
				response.ParseError(err)
				RespondInterface(w, response)
				return
			}

			logentry := server.GetLogger()
			logentry.Infof("server configuration updated: %s", update)
			response.PatchSuccess(server, "server configuration updated")
		} else {
			response.ParseSuccess(server)
		}

	} else {
		// CREATE: Server doesn't exist, create new one
		info := &models.QpServer{
			Token: token,
			User:  username,
		}

		// Apply configuration from request
		if request != nil {
			if request.Groups != nil {
				info.Groups = *request.Groups
			}
			if request.Broadcasts != nil {
				info.Broadcasts = *request.Broadcasts
			}
			if request.ReadReceipts != nil {
				info.ReadReceipts = *request.ReadReceipts
			}
			if request.Calls != nil {
				info.Calls = *request.Calls
			}
			if request.Devel != nil {
				info.Devel = *request.Devel
			}
		}

		// Create server using existing service method
		server, err := models.WhatsappService.AppendNewServer(info)
		if err != nil {
			response.ParseError(err)
			RespondInterface(w, response)
			return
		}

		// Save to database
		err = server.Save("server created without connection via POST")
		if err != nil {
			// Remove from cache if save fails
			delete(models.WhatsappService.Servers, token)
			response.ParseError(err)
			RespondInterface(w, response)
			return
		}

		response.ParseSuccess(server)
	}

	RespondSuccess(w, response)
}

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

	if request.Devel != nil {
		develValue := *request.Devel

		if server.Devel != develValue {
			server.Devel = develValue
			update += fmt.Sprintf("devel to: {%t}; ", develValue)
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
