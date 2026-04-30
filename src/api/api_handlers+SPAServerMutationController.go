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
)

// SPAServerCreateController creates a new pre-configured server owned by the SPA user.
func SPAServerCreateController(w http.ResponseWriter, r *http.Request) {
	user, err := GetSPAUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
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

	info := &models.QpServer{
		Token: uuid.NewString(),
	}
	info.SetUser(user.Username)

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

	server, err := models.WhatsappService.AppendNewServer(info)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	if err := server.Save("server created via SPA"); err != nil {
		delete(models.WhatsappService.Servers, info.Token)
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
		if err := server.Save("patching info via SPA"); err != nil {
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
		server, err = models.WhatsappService.AppendNewServer(serverRecord)
		if err != nil {
			response.ParseError(err)
			RespondInterfaceCode(w, response, http.StatusInternalServerError)
			return
		}
	}

	if err := models.WhatsappService.Delete(server, "spa"); err != nil {
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

	if _, err := server.ToggleDevel(); err != nil {
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
	switch option {
	case "groups":
		if err := models.ToggleGroups(server); err != nil {
			return false, err
		}
		return server.GetGroups(), nil
	case "broadcasts":
		if err := models.ToggleBroadcasts(server); err != nil {
			return false, err
		}
		return server.GetBroadcasts(), nil
	case "readreceipts":
		if err := models.ToggleReadReceipts(server); err != nil {
			return false, err
		}
		return server.GetReadReceipts(), nil
	case "calls":
		if err := models.ToggleCalls(server); err != nil {
			return false, err
		}
		return server.GetCalls(), nil
	case "readupdate":
		if err := models.ToggleReadUpdate(server); err != nil {
			return false, err
		}
		return server.ReadUpdate.Boolean(), nil
	default:
		return false, fmt.Errorf("unsupported option: %s", option)
	}
}
