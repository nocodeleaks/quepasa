package oauth

import (
	"strings"
	"sync"
	"time"
)

type accessTokenEntry struct {
	token     string
	expiresAt time.Time
}

var oauthAccessTokens sync.Map

// StoreAccessTokenForSession keeps the provider access token server-side for
// the lifetime of the local QuePasa JWT session.
func StoreAccessTokenForSession(sessionToken string, accessToken string, ttl time.Duration) {
	sessionToken = strings.TrimSpace(sessionToken)
	accessToken = strings.TrimSpace(accessToken)
	if sessionToken == "" || accessToken == "" {
		return
	}
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}

	oauthAccessTokens.Store(sessionToken, accessTokenEntry{
		token:     accessToken,
		expiresAt: time.Now().Add(ttl),
	})
}

// AccessTokenForSession returns the provider access token associated with a
// local QuePasa JWT, when the OAuth login happened in this process.
func AccessTokenForSession(sessionToken string) (string, bool) {
	sessionToken = strings.TrimSpace(sessionToken)
	if sessionToken == "" {
		return "", false
	}

	value, ok := oauthAccessTokens.Load(sessionToken)
	if !ok {
		return "", false
	}

	entry, ok := value.(accessTokenEntry)
	if !ok || entry.token == "" {
		oauthAccessTokens.Delete(sessionToken)
		return "", false
	}
	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		oauthAccessTokens.Delete(sessionToken)
		return "", false
	}

	return entry.token, true
}
