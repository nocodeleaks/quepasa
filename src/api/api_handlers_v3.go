package api

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	legacy "github.com/nocodeleaks/quepasa/api/legacy"

	api "github.com/nocodeleaks/quepasa/api/models"
)

const APIVersion3 string = "v3"

var ControllerPrefixV3 string = fmt.Sprintf("/%s/bot/{token}", APIVersion3)

func RegisterAPIV3Controllers(r chi.Router) {
	legacy.RegisterAPIV3Controllers(r, legacy.Config{APIVersion3: APIVersion3}, legacyHandlers())
}

//region CONTROLLER - INFORMATION

// InformationController renders route GET "/{version}/info"
func InformationControllerV3(w http.ResponseWriter, r *http.Request) {
	// setting default response type as json
	w.Header().Set("Content-Type", "application/json")

	response := &api.InformationResponse{}

	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterfaceCode(w, response, http.StatusNoContent)
		return
	}

	response.ParseSuccess(server)
	RespondSuccess(w, response)
}

//endregion
