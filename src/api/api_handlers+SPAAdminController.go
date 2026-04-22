package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	environment "github.com/nocodeleaks/quepasa/environment"
	library "github.com/nocodeleaks/quepasa/library"
	models "github.com/nocodeleaks/quepasa/models"
	"github.com/nbutton23/zxcvbn-go"
	"github.com/go-chi/chi/v5"
)

type spaUserCreateRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SPAEnvironmentController returns the current environment configuration for
// authenticated SPA users. The payload mirrors the legacy environment response
// shape so the frontend can decide whether to render full settings or preview.
func SPAEnvironmentController(w http.ResponseWriter, r *http.Request) {
	if _, err := GetSPAUser(r); err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	RespondSuccess(w, map[string]interface{}{
		"settings": environment.Settings,
		"preview":  environment.GetPreview(),
	})
}

// SPAPublicUserCreateController creates a user through the SPA setup flow.
// This route intentionally remains public so the alternate frontend can bootstrap
// the first account without falling back to the classic UI.
func SPAPublicUserCreateController(w http.ResponseWriter, r *http.Request) {
	username, err := createSPAUserFromRequest(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	RespondSuccess(w, map[string]interface{}{
		"result":   "success",
		"username": username,
	})
}

// SPAUserDeleteController removes a user account through the authenticated SPA.
func SPAUserDeleteController(w http.ResponseWriter, r *http.Request) {
	user, err := GetSPAUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	username := strings.TrimSpace(chi.URLParam(r, "username"))
	if username == "" {
		RespondErrorCode(w, fmt.Errorf("missing username parameter"), http.StatusBadRequest)
		return
	}

	if strings.EqualFold(username, user.Username) {
		RespondErrorCode(w, fmt.Errorf("cannot delete the current authenticated user"), http.StatusBadRequest)
		return
	}

	count, err := models.WhatsappService.DB.Users.Count()
	if err != nil {
		RespondErrorCode(w, err, http.StatusInternalServerError)
		return
	}

	if count <= 1 {
		RespondErrorCode(w, fmt.Errorf("cannot delete the last remaining user"), http.StatusBadRequest)
		return
	}

	if err := models.WhatsappService.DB.Users.Delete(username); err != nil {
		RespondErrorCode(w, err, http.StatusNotFound)
		return
	}

	RespondSuccess(w, map[string]interface{}{
		"result":   "success",
		"username": username,
	})
}

func createSPAUserFromRequest(r *http.Request) (string, error) {
	if r.Body == nil {
		return "", fmt.Errorf("missing request body")
	}

	var request spaUserCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return "", fmt.Errorf("error converting body to json: %w", err)
	}

	username := strings.TrimSpace(request.Username)
	if username == "" {
		username = strings.TrimSpace(request.Email)
	}

	password := strings.TrimSpace(request.Password)

	if username == "" || password == "" {
		return "", fmt.Errorf("email and password are required")
	}

	if !library.IsValidEMail(username) {
		return "", fmt.Errorf("email is invalid")
	}

	res := zxcvbn.PasswordStrength(password, nil)
	if res.Score < 1 {
		return "", fmt.Errorf("password is too weak")
	}

	exists, err := models.WhatsappService.DB.Users.Exists(username)
	if err != nil {
		return "", err
	}

	if exists {
		return "", fmt.Errorf("user already exists: %s", username)
	}

	if _, err := models.WhatsappService.DB.Users.Create(username, password); err != nil {
		return "", err
	}

	return username, nil
}
