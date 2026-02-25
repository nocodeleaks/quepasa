package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	api "github.com/nocodeleaks/quepasa/api/models"
	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

//region CONTROLLER - Information

// CreateInformationController handles POST requests for creating a new bot/server
//
//	@Summary		Create bot configuration
//	@Description	Create a new bot/server with configuration before QR code scanning
//	@Tags			Information
//	@Accept			json
//	@Produce		json
//	@Param			request	body		InfoCreateRequest		true	"Server creation request"
//	@Success		200		{object}	api.InformationResponse	"Server updated"
//	@Success		201		{object}	api.InformationResponse	"Server created"
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
//	@Success		200	{object}	api.InformationResponse
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
//	@Success		200		{object}	api.InformationResponse
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
//	@Success		200	{object}	api.InformationResponse
//	@Failure		400	{object}	models.QpResponse
//	@Security		ApiKeyAuth
//	@Router			/info [delete]
func DeleteInformationController(w http.ResponseWriter, r *http.Request) {
	InformationDeleteRequest(w, r)
}

//endregion

// updateServerConfiguration applies configuration changes to an existing server
// Returns update summary string and error if any
func updateServerConfiguration(server *models.QpWhatsappServer, username string, request interface{}) (string, error) {
	update := ""

	// Update user if provided and different
	if len(username) > 0 && server.User != username {
		// Validate user exists in database
		_, err := models.WhatsappService.DB.Users.Find(username)
		if err != nil {
			return "", fmt.Errorf("user not found: %v", err.Error())
		}

		server.User = username
		update += fmt.Sprintf("user to: {%s}; ", username)
	}

	// Apply configuration from request
	// Handle both InfoCreateRequest and QpInfoPatchRequest
	var groups, broadcasts, readReceipts, calls, readUpdate *whatsapp.WhatsappBoolean
	var devel *bool

	switch req := request.(type) {
	case *InfoCreateRequest:
		if req != nil {
			groups = req.Groups
			broadcasts = req.Broadcasts
			readReceipts = req.ReadReceipts
			calls = req.Calls
			readUpdate = req.ReadUpdate
			devel = req.Devel
		}
	case *models.QpInfoPatchRequest:
		if req != nil {
			// QpInfoPatchRequest may have Username field, handle groups/broadcasts/etc
			groups = req.Groups
			broadcasts = req.Broadcasts
			readReceipts = req.ReadReceipts
			calls = req.Calls
			readUpdate = req.ReadUpdate
			devel = req.Devel
		}
	}

	if groups != nil && server.Groups != *groups {
		server.Groups = *groups
		update += fmt.Sprintf("groups to: {%s}; ", *groups)
	}

	if broadcasts != nil && server.Broadcasts != *broadcasts {
		server.Broadcasts = *broadcasts
		update += fmt.Sprintf("broadcasts to: {%s}; ", *broadcasts)
	}

	if readReceipts != nil && server.ReadReceipts != *readReceipts {
		server.ReadReceipts = *readReceipts
		update += fmt.Sprintf("readreceipts to: {%s}; ", *readReceipts)
	}

	if calls != nil && server.Calls != *calls {
		server.Calls = *calls
		update += fmt.Sprintf("calls to: {%s}; ", *calls)
	}

	if readUpdate != nil && server.ReadUpdate != *readUpdate {
		server.ReadUpdate = *readUpdate
		update += fmt.Sprintf("readupdate to: {%s}; ", *readUpdate)
	}

	if devel != nil && server.Devel != *devel {
		server.Devel = *devel
		update += fmt.Sprintf("devel to: {%t}; ", *devel)
	}

	return update, nil
}

func InformationPostRequest(w http.ResponseWriter, r *http.Request) {
	response := &api.InformationResponse{}

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
		// UPDATE: Server exists, use shared update logic
		update, err := updateServerConfiguration(server, username, request)
		if err != nil {
			response.ParseError(err)
			RespondInterface(w, response)
			return
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

		// Return 200 for updates
		RespondSuccess(w, response)
		return

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
			if request.ReadUpdate != nil {
				info.ReadUpdate = *request.ReadUpdate
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

		// Return 201 for creation
		RespondInterfaceCode(w, response, http.StatusCreated)
		return
	}
}

func InformationGetRequest(w http.ResponseWriter, r *http.Request) {
	response := &api.InformationResponse{}

	// Check authentication
	token := GetToken(r)
	isMaster := IsMatchForMaster(r)

	// Case 1: No authentication - Return basic server (application) info
	if token == "" && !isMaster {
		response.Success = true
		response.Status = "application information"
		response.Version = models.QpVersion
		RespondSuccess(w, response)
		return
	}

	// Case 2 & 3: With authentication - Get specific server
	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterfaceCode(w, response, http.StatusNoContent)
		return
	}

	// Populate response with server information
	response.ParseSuccess(server)
	response.Version = models.QpVersion
	// Server uptime is in response.Server.Uptime (via MarshalJSON)

	RespondSuccess(w, response)
}

func InformationPatchRequest(w http.ResponseWriter, r *http.Request) {
	response := &api.InformationResponse{}

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

	// Get username from request (PATCH allows changing username)
	username := ""
	if request.Username != nil {
		username = *request.Username
	}

	// Use shared update logic
	update, err := updateServerConfiguration(server, username, request)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

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

	err = models.WhatsappService.Delete(server, "api")
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	response.ParseSuccess("server deleted")
	RespondSuccess(w, response)
}
