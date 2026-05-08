package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/nbutton23/zxcvbn-go"
	environment "github.com/nocodeleaks/quepasa/environment"
	library "github.com/nocodeleaks/quepasa/library"
	runtime "github.com/nocodeleaks/quepasa/runtime"
)

type spaUserCreateRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthenticatedEnvironmentController returns the current environment configuration for
// authenticated SPA users. The payload mirrors the legacy environment response
// shape so the frontend can decide whether to render full settings or preview.
func AuthenticatedEnvironmentController(w http.ResponseWriter, r *http.Request) {
	if _, err := GetAuthenticatedUser(r); err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	if !isMasterKeyRequest(r) {
		RespondErrorCode(w, fmt.Errorf("master key required"), http.StatusForbidden)
		return
	}

	RespondSuccess(w, map[string]interface{}{
		"settings": environment.Settings,
		"preview":  environment.GetPreview(),
	})
}

// PublicUserCreateController creates a user through the SPA setup flow.
// First user can be created without master key (bootstrap).
// Subsequent users require X-Master-Key header.
// The route remains reachable without JWT to support bootstrap (first user).
func PublicUserCreateController(w http.ResponseWriter, r *http.Request) {
	// Count existing users to determine if this is the first user
	count, err := runtime.CountPersistedUsers()
	if err != nil {
		RespondErrorCode(w, err, http.StatusInternalServerError)
		return
	}

	// First user (count == 0) can be created without master key
	// Subsequent users (count > 0) require master key
	if count > 0 && !IsMatchForMaster(r) {
		RespondErrorCode(w, fmt.Errorf("master key required to create additional users"), http.StatusForbidden)
		return
	}

	username, err := createAuthenticatedUserFromRequest(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	RespondSuccess(w, map[string]interface{}{
		"result":   "success",
		"username": username,
	})
}

// AuthenticatedUserDeleteController removes a user account through the authenticated SPA.
// Requires a valid X-Master-Key header.
func AuthenticatedUserDeleteController(w http.ResponseWriter, r *http.Request) {
	_, err := GetAuthenticatedUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	if !isMasterKeyRequest(r) {
		RespondErrorCode(w, fmt.Errorf("master key required"), http.StatusForbidden)
		return
	}

	username := strings.TrimSpace(chi.URLParam(r, "username"))
	if username == "" {
		RespondErrorCode(w, fmt.Errorf("missing username parameter"), http.StatusBadRequest)
		return
	}

	if strings.EqualFold(username, "") {
		RespondErrorCode(w, fmt.Errorf("username cannot be empty"), http.StatusBadRequest)
		return
	}

	count, err := runtime.CountPersistedUsers()
	if err != nil {
		RespondErrorCode(w, err, http.StatusInternalServerError)
		return
	}

	if count <= 1 {
		RespondErrorCode(w, fmt.Errorf("cannot delete the last remaining user"), http.StatusBadRequest)
		return
	}

	if err := runtime.DeletePersistedUser(username); err != nil {
		RespondErrorCode(w, err, http.StatusNotFound)
		return
	}

	RespondSuccess(w, map[string]interface{}{
		"result":   "success",
		"username": username,
	})
}

func createAuthenticatedUserFromRequest(r *http.Request) (string, error) {
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

	if _, err := findPersistedUser(username); err == nil {
		return "", fmt.Errorf("user already exists: %s", username)
	}

	if _, err := runtime.CreatePersistedUser(username, password); err != nil {
		return "", err
	}

	return username, nil
}
