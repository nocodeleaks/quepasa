package oauth

import (
	"fmt"
	"os"
	"strings"

	environment "github.com/nocodeleaks/quepasa/environment"
)

// OAuthProvider abstracts the external OAuth/OIDC identity provider. QuePasa
// ships with a generic OIDC implementation that works with any compliant provider
// (Keycloak, Auth0, Google, Microsoft, Okta, GitLab, etc.).
type OAuthProvider interface {
	// GetAuthURL builds the authorization URL for the initial redirect.
	// codeChallenge is optional (empty string if PKCE not used).
	GetAuthURL(state string, codeChallenge string) string

	// Exchange trades the authorization code for an access token.
	// codeVerifier is optional (empty string if PKCE not used).
	Exchange(code string, codeVerifier string) (accessToken string, err error)

	// GetUserInfo fetches the authenticated user's profile from the provider.
	GetUserInfo(accessToken string) (*OAuthUserInfo, error)
}

// OAuthUserInfo carries the minimal user profile QuePasa needs to create/link
// a local account after successful OAuth authentication.
type OAuthUserInfo struct {
	Email    string // primary identifier; becomes local username if not exists
	Username string // optional; provider's username/handle
	Name     string // optional; display name
}

// OAuthConfig holds the runtime OAuth settings loaded from environment variables.
// The configuration is provider-agnostic: any OIDC-compliant endpoint works.
type OAuthConfig struct {
	Enabled      bool
	ProviderURL  string   // base URL of the OIDC provider (e.g. https://identity.example.com)
	ClientID     string   // OAuth client ID registered with the provider
	ClientSecret string   // OAuth client secret
	RedirectURI  string   // callback URL registered with the provider (e.g. https://quepasa.example.com/oauth/callback)
	Scopes       []string // requested scopes (e.g. openid, email, profile)
}

const (
	envOAuthEnabled      = "OAUTH_ENABLED"
	envOAuthProviderURL  = "OAUTH_PROVIDER_URL"
	envOAuthClientID     = "OAUTH_CLIENT_ID"
	envOAuthClientSecret = "OAUTH_CLIENT_SECRET"
	envOAuthRedirectURI  = "OAUTH_REDIRECT_URI"
	envOAuthScopes       = "OAUTH_SCOPES"
)

var globalOAuthConfig *OAuthConfig

// LoadOAuthConfig reads OAuth settings from the environment. Call once at startup.
func LoadOAuthConfig() *OAuthConfig {
	enabled := strings.ToLower(strings.TrimSpace(os.Getenv(envOAuthEnabled))) == "true"
	if !enabled {
		globalOAuthConfig = &OAuthConfig{Enabled: false}
		return globalOAuthConfig
	}

	cfg := &OAuthConfig{
		Enabled:      true,
		ProviderURL:  strings.TrimSpace(os.Getenv(envOAuthProviderURL)),
		ClientID:     strings.TrimSpace(os.Getenv(envOAuthClientID)),
		ClientSecret: strings.TrimSpace(os.Getenv(envOAuthClientSecret)),
		RedirectURI:  strings.TrimSpace(os.Getenv(envOAuthRedirectURI)),
		Scopes:       parseScopes(os.Getenv(envOAuthScopes)),
	}

	globalOAuthConfig = cfg
	return cfg
}

// GetOAuthConfig returns the loaded OAuth configuration. Returns nil when disabled.
func GetOAuthConfig() *OAuthConfig {
	if globalOAuthConfig == nil {
		return LoadOAuthConfig()
	}
	return globalOAuthConfig
}

// IsEnabled reports whether OAuth authentication is active.
func IsEnabled() bool {
	cfg := GetOAuthConfig()
	return cfg != nil && cfg.Enabled && cfg.ProviderURL != "" && cfg.ClientID != ""
}

func parseScopes(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []string{"openid", "email", "profile"} // OIDC defaults
	}
	parts := strings.Split(raw, ",")
	scopes := make([]string, 0, len(parts))
	for _, s := range parts {
		if s = strings.TrimSpace(s); s != "" {
			scopes = append(scopes, s)
		}
	}
	return scopes
}

// GetBaseURL returns the QuePasa base URL for building callback URIs. Falls back
// to environment QUEPASA_BASE_URL or builds from API host/port.
func GetBaseURL() string {
	if base := strings.TrimSpace(os.Getenv("QUEPASA_BASE_URL")); base != "" {
		return strings.TrimRight(base, "/")
	}
	// Fallback: reconstruct from webserver settings.
	host := environment.Settings.WebServer.Host
	port := environment.Settings.WebServer.Port
	if host == "" {
		host = "localhost"
	}
	scheme := "http"
	if port == 443 {
		scheme = "https"
	}
	if port == 80 || port == 443 {
		return scheme + "://" + host
	}
	return scheme + "://" + host + ":" + fmt.Sprintf("%d", port)
}
