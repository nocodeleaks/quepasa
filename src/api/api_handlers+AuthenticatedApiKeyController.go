package api

import (
	"net/http"

	models "github.com/nocodeleaks/quepasa/models"
	runtime "github.com/nocodeleaks/quepasa/runtime"
)

// AuthenticatedApiKeyController manages the authenticated user's personal API key.
//
//	GET    /account/apikey  → status (whether a key is set, last rotation time)
//	POST   /account/apikey  → rotate: generate a new key, return the plaintext ONCE
//	DELETE /account/apikey  → revoke the key
//
// The personal API key lets a user authenticate to their own WhatsApp sessions
// (header X-QUEPASA-USERKEY) without the admin master key. Rotation invalidates
// the previous key immediately.
func AuthenticatedApiKeyController(w http.ResponseWriter, r *http.Request) {
	user, err := GetAuthenticatedUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		apiKeyStatus(w, user)
	case http.MethodPost:
		apiKeyRotate(w, user)
	case http.MethodDelete:
		apiKeyRevoke(w, user)
	default:
		RespondErrorCode(w, nil, http.StatusMethodNotAllowed)
	}
}

func apiKeyStatus(w http.ResponseWriter, user *models.QpUser) {
	response := map[string]interface{}{
		"configured": user.APIKey != nil && *user.APIKey != "",
	}
	if user.APIKeyRotatedAt != nil {
		response["rotated_at"] = user.APIKeyRotatedAt
	}
	RespondSuccess(w, response)
}

func apiKeyRotate(w http.ResponseWriter, user *models.QpUser) {
	plaintext, err := runtime.RotateUserAPIKey(user.Username)
	if err != nil {
		RespondErrorCode(w, err, http.StatusInternalServerError)
		return
	}

	// The plaintext key is returned exactly once; only its hash is stored.
	RespondSuccess(w, map[string]interface{}{
		"apikey":  plaintext,
		"header":  "X-QUEPASA-USERKEY",
		"warning": "store this key now; it cannot be retrieved again",
	})
}

func apiKeyRevoke(w http.ResponseWriter, user *models.QpUser) {
	if err := runtime.RevokeUserAPIKey(user.Username); err != nil {
		RespondErrorCode(w, err, http.StatusInternalServerError)
		return
	}
	RespondSuccess(w, map[string]interface{}{"revoked": true})
}
