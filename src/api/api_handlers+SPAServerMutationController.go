package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	apiModels "github.com/nocodeleaks/quepasa/api/models"
	models "github.com/nocodeleaks/quepasa/models"
	runtime "github.com/nocodeleaks/quepasa/runtime"
)

// SPAServerCreateController creates a new pre-configured server.
//
// Authentication is always required via SPA JWT (Authorization Bearer or X-QUEPASA-TOKEN).
// When RELAXED_SESSIONS=false, a valid master key is also required.
func SPAServerCreateController(w http.ResponseWriter, r *http.Request) {
	user, err := GetSPAUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	if !isRelaxedSessions() && !isMasterKeyRequest(r) {
		RespondErrorCode(w, fmt.Errorf("master key required to create sessions (RELAXED_SESSIONS=false)"), http.StatusForbidden)
		return
	}

	response := &apiModels.InformationResponse{}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	var request *InfoCreateRequest
	if len(body) > 0 {
		if err := json.Unmarshal(body, &request); err != nil {
			response.ParseError(fmt.Errorf("error converting body to json: %v", err.Error()))
			RespondInterface(w, response)
			return
		}
	}

	patch := buildSessionConfigurationPatch(request)
	info := runtime.BuildSessionRecord(uuid.NewString(), user.Username, patch)

	server, err := runtime.CreateSessionRecord(info, "server created via SPA")
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	response.ParseSuccess(server)
	RespondInterfaceCode(w, response, http.StatusCreated)
}

// SPAServerUpdateController patches persisted server configuration for the SPA user.
func SPAServerUpdateController(w http.ResponseWriter, r *http.Request) {
	user, err := GetSPAUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token, err := GetSPATokenParam(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	server, err := GetSPAOwnedLiveServer(user, token)
	if err != nil {
		respondSPAServerLookupError(w, err)
		return
	}

	response := &apiModels.InformationResponse{}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	if len(body) == 0 {
		response.ParseError(fmt.Errorf("empty body"))
		RespondInterface(w, response)
		return
	}

	var request *InfoPatchRequest
	if err := json.Unmarshal(body, &request); err != nil {
		response.ParseError(fmt.Errorf("error converting body to json: %v", err.Error()))
		RespondInterface(w, response)
		return
	}

	if request == nil {
		response.ParseError(fmt.Errorf("invalid request body: %s", string(body)))
		RespondInterface(w, response)
		return
	}

	username := ""
	if request.Username != nil {
		username = *request.Username
	}

	update, err := updateServerConfiguration(server, username, request)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	if len(update) > 0 {
		if err := runtime.SaveSession(server, "patching info via SPA"); err != nil {
			response.ParseError(err)
			RespondInterface(w, response)
			return
		}
		response.PatchSuccess(server, "server updated")
		RespondSuccess(w, response)
		return
	}

	response.PatchSuccess(server, "no update required")
	RespondSuccess(w, response)
}

// SPAServerDeleteController deletes a server owned by the SPA user.
func SPAServerDeleteController(w http.ResponseWriter, r *http.Request) {
	user, err := GetSPAUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	response := &models.QpResponse{}

	token, err := GetSPATokenParam(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	serverRecord, err := GetSPAOwnedServerRecord(user, token)
	if err != nil {
		respondSPAServerLookupError(w, err)
		return
	}

	server := FindSPALiveServer(serverRecord.Token)
	if server == nil {
		server, err = runtime.LoadSessionRecord(serverRecord)
		if err != nil {
			response.ParseError(err)
			RespondInterfaceCode(w, response, http.StatusInternalServerError)
			return
		}
	}

	if err := runtime.DeleteSessionRecord(server, "spa"); err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	response.ParseSuccess("server deleted")
	RespondSuccess(w, response)
}

// SPAServerDebugToggleController toggles server debug mode through the SPA auth surface.
func SPAServerDebugToggleController(w http.ResponseWriter, r *http.Request) {
	user, err := GetSPAUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token, err := GetSPATokenParam(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	server, err := GetSPAOwnedLiveServer(user, token)
	if err != nil {
		respondSPAServerLookupError(w, err)
		return
	}

	if _, err := runtime.ToggleSessionDebug(server); err != nil {
		RespondServerError(server, w, err)
		return
	}

	RespondSuccess(w, map[string]interface{}{
		"devel":  server.Devel,
		"server": BuildSPAServerSummary(server.QpServer, server),
	})
}

// SPAServerOptionToggleController toggles a persisted server option explicitly by name.
func SPAServerOptionToggleController(w http.ResponseWriter, r *http.Request) {
	user, err := GetSPAUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token, err := GetSPATokenParam(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	server, err := GetSPAOwnedLiveServer(user, token)
	if err != nil {
		respondSPAServerLookupError(w, err)
		return
	}

	option := strings.ToLower(strings.TrimSpace(chi.URLParam(r, "option")))
	if option == "" {
		RespondErrorCode(w, fmt.Errorf("missing option parameter"), http.StatusBadRequest)
		return
	}

	value, err := toggleSPAServerOption(server, option)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	RespondSuccess(w, map[string]interface{}{
		"option": option,
		"value":  value,
		"server": BuildSPAServerSummary(server.QpServer, server),
	})
}

func toggleSPAServerOption(server *models.QpWhatsappServer, option string) (bool, error) {
	return runtime.ToggleSessionOption(server, option)
}
