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
	library "github.com/nocodeleaks/quepasa/library"
	models "github.com/nocodeleaks/quepasa/models"
	runtime "github.com/nocodeleaks/quepasa/runtime"
)

// AuthenticatedServerCreateController creates a new pre-configured server.
//
// Authentication is required via SPA JWT (user scope) or X-QUEPASA-TOKEN (single-session scope).
// When authenticated by JWT, X-QUEPASA-TOKEN is treated as the optional token/identifier
// for the new session and is accepted only when RELAXED_SESSIONS=true.
// When RELAXED_SESSIONS=false, a valid master key is also required.
func AuthenticatedServerCreateController(w http.ResponseWriter, r *http.Request) {
	user, err := GetAuthenticatedUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	requestedSessionToken := strings.TrimSpace(r.Header.Get(library.HeaderToken))
	if requestedSessionToken != "" && !isRelaxedSessions() {
		RespondErrorCode(w, fmt.Errorf("X-QUEPASA-TOKEN is only allowed when RELAXED_SESSIONS=true"), http.StatusForbidden)
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
	sessionToken := uuid.NewString()
	if requestedSessionToken != "" {
		sessionToken = requestedSessionToken
	}
	info := runtime.BuildSessionRecord(sessionToken, user.Username, patch)

	server, err := runtime.CreateSessionRecord(info, "server created via SPA")
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	response.ParseSuccess(server)
	RespondInterfaceCode(w, response, http.StatusCreated)
}

// AuthenticatedServerUpdateController patches persisted server configuration for the SPA user.
func AuthenticatedServerUpdateController(w http.ResponseWriter, r *http.Request) {
	user, err := GetAuthenticatedUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token, err := GetAuthenticatedTokenParam(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	server, err := GetOwnedLiveServer(user, token)
	if err != nil {
		respondServerLookupError(w, err)
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

// AuthenticatedServerDeleteController deletes a server owned by the SPA user.
func AuthenticatedServerDeleteController(w http.ResponseWriter, r *http.Request) {
	user, err := GetAuthenticatedUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	response := &models.QpResponse{}

	token, err := GetAuthenticatedTokenParam(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	serverRecord, err := GetOwnedServerRecord(user, token)
	if err != nil {
		respondServerLookupError(w, err)
		return
	}

	server := FindLiveServer(serverRecord.Token)
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

// AuthenticatedServerDebugToggleController toggles server debug mode through the SPA auth surface.
func AuthenticatedServerDebugToggleController(w http.ResponseWriter, r *http.Request) {
	user, err := GetAuthenticatedUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token, err := GetAuthenticatedTokenParam(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	server, err := GetOwnedLiveServer(user, token)
	if err != nil {
		respondServerLookupError(w, err)
		return
	}

	if _, err := runtime.ToggleSessionDebug(server); err != nil {
		RespondServerError(server, w, err)
		return
	}

	RespondSuccess(w, map[string]interface{}{
		"devel":  server.Devel,
		"server": BuildServerSummary(server.QpServer, server),
	})
}

// AuthenticatedServerOptionToggleController toggles a persisted server option explicitly by name.
func AuthenticatedServerOptionToggleController(w http.ResponseWriter, r *http.Request) {
	user, err := GetAuthenticatedUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token, err := GetAuthenticatedTokenParam(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	server, err := GetOwnedLiveServer(user, token)
	if err != nil {
		respondServerLookupError(w, err)
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
		"server": BuildServerSummary(server.QpServer, server),
	})
}

func toggleSPAServerOption(server *models.QpWhatsappServer, option string) (bool, error) {
	return runtime.ToggleSessionOption(server, option)
}
