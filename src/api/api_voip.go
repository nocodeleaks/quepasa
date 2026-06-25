package api

import (
	"fmt"
	"net/http"
	"strings"

	models "github.com/nocodeleaks/quepasa/models"
	runtime "github.com/nocodeleaks/quepasa/runtime"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// resolveVoIPModeServer resolves the authenticated, owned instance from the
// "token" query-string parameter shared by the VoIP mode controllers. It writes
// the HTTP error and returns ok=false when resolution fails.
func resolveVoIPModeServer(w http.ResponseWriter, r *http.Request) (*models.QpWhatsappServer, bool) {
	user, err := GetAuthenticatedUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return nil, false
	}

	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		RespondErrorCode(w, fmt.Errorf("missing token parameter"), http.StatusBadRequest)
		return nil, false
	}

	server, err := GetOwnedLiveServer(user, token)
	if err != nil {
		respondServerLookupError(w, err)
		return nil, false
	}

	return server, true
}

// AuthenticatedVoIPModeGetController returns the per-instance VoIP mode.
//
//	GET /api/voip/mode?token=<token>
func AuthenticatedVoIPModeGetController(w http.ResponseWriter, r *http.Request) {
	server, ok := resolveVoIPModeServer(w, r)
	if !ok {
		return
	}

	mode, err := runtime.GetSessionVoIPMode(server)
	if err != nil {
		RespondErrorCode(w, err, http.StatusInternalServerError)
		return
	}

	RespondSuccess(w, voipModeResponse(mode))
}

// AuthenticatedVoIPModeSetController updates the per-instance VoIP mode and
// persists it to the database.
//
//	POST /api/voip/mode?token=<token>&mode=<disabled|exclusive|additional>
func AuthenticatedVoIPModeSetController(w http.ResponseWriter, r *http.Request) {
	server, ok := resolveVoIPModeServer(w, r)
	if !ok {
		return
	}

	raw := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("mode")))
	var mode whatsapp.VoIPMode
	switch raw {
	case string(whatsapp.VoIPModeDisabled), string(whatsapp.VoIPModeExclusive), string(whatsapp.VoIPModeAdditional):
		mode = whatsapp.VoIPMode(raw)
	default:
		RespondErrorCode(w, fmt.Errorf("invalid mode %q; expected disabled, exclusive or additional", raw), http.StatusBadRequest)
		return
	}

	if err := runtime.SetSessionVoIPMode(server, mode); err != nil {
		RespondErrorCode(w, err, http.StatusInternalServerError)
		return
	}

	RespondSuccess(w, voipModeResponse(mode))
}

// voipModeResponse is the shared response body for the VoIP mode endpoints.
func voipModeResponse(mode whatsapp.VoIPMode) map[string]interface{} {
	return map[string]interface{}{
		"mode": mode.String(),
		"options": []string{
			string(whatsapp.VoIPModeDisabled),
			string(whatsapp.VoIPModeExclusive),
			string(whatsapp.VoIPModeAdditional),
		},
	}
}
