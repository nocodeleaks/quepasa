package environment

import "strings"

// OAuth environment variable names
const (
	ENV_OAUTH_ENABLED       = "OAUTH_ENABLED"
	ENV_OAUTH_PROVIDER_URL  = "OAUTH_PROVIDER_URL"
	ENV_OAUTH_CLIENT_ID     = "OAUTH_CLIENT_ID"
	ENV_OAUTH_CLIENT_SECRET = "OAUTH_CLIENT_SECRET"
	ENV_OAUTH_REDIRECT_URI  = "OAUTH_REDIRECT_URI"
	ENV_OAUTH_SCOPES        = "OAUTH_SCOPES"
	ENV_OAUTH_RESOURCE_URL  = "OAUTH_RESOURCE_BASE_URL"
)

// OAuthSettings holds OAuth/OIDC configuration loaded from environment.
type OAuthSettings struct {
	Enabled      bool     `json:"enabled"`
	ProviderURL  string   `json:"provider_url"`
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURI  string   `json:"redirect_uri"`
	Scopes       []string `json:"scopes"`
	ResourceURL  string   `json:"resource_url"`
}

// NewOAuthSettings creates a new OAuth settings object by loading values from environment.
func NewOAuthSettings() OAuthSettings {
	return OAuthSettings{
		Enabled:      getEnvOrDefaultBool(ENV_OAUTH_ENABLED, false),
		ProviderURL:  getEnvOrDefaultString(ENV_OAUTH_PROVIDER_URL, ""),
		ClientID:     getEnvOrDefaultString(ENV_OAUTH_CLIENT_ID, ""),
		ClientSecret: getEnvOrDefaultString(ENV_OAUTH_CLIENT_SECRET, ""),
		RedirectURI:  getEnvOrDefaultString(ENV_OAUTH_REDIRECT_URI, ""),
		Scopes:       parseOAuthScopes(getEnvOrDefaultString(ENV_OAUTH_SCOPES, "")),
		ResourceURL:  strings.TrimRight(getEnvOrDefaultString(ENV_OAUTH_RESOURCE_URL, ""), "/"),
	}
}

func parseOAuthScopes(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []string{"openid", "email", "profile"}
	}

	parts := strings.Split(raw, ",")
	scopes := make([]string, 0, len(parts))
	for _, scope := range parts {
		if scope = strings.TrimSpace(scope); scope != "" {
			scopes = append(scopes, scope)
		}
	}
	return scopes
}
