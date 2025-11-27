package api

import (
	"net/http"

	api "github.com/nocodeleaks/quepasa/api/models"
	environment "github.com/nocodeleaks/quepasa/environment"
)

//region CONTROLLER - ENVIRONMENT

// EnvironmentController handles GET requests for environment settings
//
//	@Summary		Get environment settings
//	@Description	Get all environment variables and configurations (master key required for full access, preview available without auth)
//	@Tags			Environment
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	api.EnvironmentResponse
//	@Failure		401	{object}	api.EnvironmentResponse
//	@Failure		403	{object}	api.EnvironmentResponse
//	@Security		ApiKeyAuth
//	@Router			/environment [get]
func EnvironmentController(w http.ResponseWriter, r *http.Request) {
	EnvironmentGetRequest(w, r)
}

//endregion

// EnvironmentGetRequest handles the environment settings request
func EnvironmentGetRequest(w http.ResponseWriter, r *http.Request) {

	response := &api.EnvironmentResponse{}

	// Check if master key is being used
	isMaster := IsMatchForMaster(r)

	if isMaster {
		response.Settings = &environment.Settings
		response.ParseSuccess("successfully retrieved full environment settings")
	} else {
		response.Preview = environment.GetPreview()
		response.ParseSuccess("successfully retrieved public environment settings preview")
	}

	// With master key: return full environment settings
	RespondSuccess(w, response)
}
