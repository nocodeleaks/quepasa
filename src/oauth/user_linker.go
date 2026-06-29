package oauth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	models "github.com/nocodeleaks/quepasa/models"
	log "github.com/nocodeleaks/quepasa/qplog"
)

// FindOrCreateUser resolves an OAuth-authenticated user to a local QuePasa account.
// If the user exists (by email as username), it is returned. If not, a new account
// is created with a random password (OAuth users authenticate via the provider, not
// local password). Account creation respects the ENV_ACCOUNTSETUP gate.
func FindOrCreateUser(userInfo *OAuthUserInfo) (*models.QpUser, error) {
	if userInfo.Email == "" {
		return nil, fmt.Errorf("oauth user info missing email")
	}

	username := userInfo.Email

	// Check if user already exists.
	if models.WhatsappService == nil || models.WhatsappService.DB == nil || models.WhatsappService.DB.Users == nil {
		return nil, fmt.Errorf("user service not initialized")
	}

	existing, err := models.WhatsappService.DB.Users.Find(username)
	if err == nil && existing != nil {
		log.Infof("oauth: linked existing user %s", username)
		return existing, nil
	}

	// User does not exist; create a new one if account setup is enabled.
	if !models.ENV.AccountSetup() {
		return nil, fmt.Errorf("oauth user %s does not exist and account creation is disabled", username)
	}

	// Generate a random password. OAuth users authenticate via the provider, so
	// this password is never used; it exists to satisfy the schema.
	password, err := generateRandomPassword(32)
	if err != nil {
		return nil, fmt.Errorf("generate random password: %w", err)
	}

	user, err := models.WhatsappService.DB.Users.Create(username, password)
	if err != nil {
		return nil, fmt.Errorf("create oauth user: %w", err)
	}

	log.Infof("oauth: created new user %s from provider", username)
	return user, nil
}

func generateRandomPassword(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
