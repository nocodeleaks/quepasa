package api

import (
	"encoding/json"
	"io"
	"net/http"

	models "github.com/nocodeleaks/quepasa/models"
)

// AuthenticatedUIController handles GET and PATCH for the authenticated user's UI preferences.
func AuthenticatedUIController(w http.ResponseWriter, r *http.Request) {
	user, err := GetAuthenticatedUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		authenticatedUIGet(w, user)
	case http.MethodPatch:
		authenticatedUIPatch(w, r, user)
	default:
		RespondErrorCode(w, nil, http.StatusMethodNotAllowed)
	}
}

func authenticatedUIGet(w http.ResponseWriter, user *models.QpUser) {
	ui := parseUserUI(user)
	RespondSuccess(w, ui)
}

func authenticatedUIPatch(w http.ResponseWriter, r *http.Request, user *models.QpUser) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	// Merge patch into existing prefs
	current := parseUserUI(user)
	if err := json.Unmarshal(body, current); err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	serialized, err := json.Marshal(current)
	if err != nil {
		RespondErrorCode(w, err, http.StatusInternalServerError)
		return
	}

	db := models.GetDatabase()
	if err := db.Users.UpdateUI(user.Username, string(serialized)); err != nil {
		RespondErrorCode(w, err, http.StatusInternalServerError)
		return
	}

	RespondSuccess(w, current)
}

func parseUserUI(user *models.QpUser) *models.QpUserUI {
	ui := &models.QpUserUI{}
	if user.UI != nil && *user.UI != "" {
		json.Unmarshal([]byte(*user.UI), ui) //nolint:errcheck — invalid JSON falls back to zero value
	}
	return ui
}
