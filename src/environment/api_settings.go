package environment

import (
	"strings"
	"time"
)

// API environment variable names
const (
	ENV_WEBSOCKETSSL     = "WEBSOCKETSSL"        // use SSL for websocket qrcode
	ENV_SIGNING_SECRET   = "SIGNING_SECRET"      // token for hash signing cookies
	ENV_MASTER_KEY       = "MASTERKEY"           // used for manage all instances at all
	ENV_WEBHOOK_TIMEOUT  = "WEBHOOK_TIMEOUT"     // timeout in milliseconds for webhook requests
	ENV_API_PREFIX       = "API_PREFIX"          // API routes prefix
	ENV_API_TIMEOUT      = "API_TIMEOUT"         // API request timeout in milliseconds
	ENV_API_DEFAULT_VER  = "API_DEFAULT_VERSION" // default version for unversioned API alias
	ENV_USER             = "USER"                // default user for database seeding
	ENV_PASSWORD         = "PASSWORD"            // default password for database seeding
	ENV_RELAXED_SESSIONS = "RELAXED_SESSIONS"     // when true, authenticated requests can create sessions without master key
	ENV_CORS_ORIGINS     = "CORS_ALLOWED_ORIGINS" // comma-separated browser origins allowed by CORS; empty = same-origin only
)

// APISettings holds all API configuration loaded from environment
type APISettings struct {
	UseSSLWebSocket bool   `json:"use_ssl_websocket"`
	SigningSecret   string `json:"signing_secret"`
	MasterKey       string `json:"master_key"`
	WebhookTimeout  uint32 `json:"webhook_timeout"` // webhook timeout in milliseconds
	Prefix          string `json:"prefix"`
	Timeout         uint32 `json:"timeout"`          // API request timeout in milliseconds
	DefaultVersion  string `json:"default_version"`  // default version for unversioned API alias
	User            string `json:"user"`             // default user for database seeding
	Password        string `json:"password"`         // default password for database seeding
	RelaxedSessions bool   `json:"relaxed_sessions"` // true = any authenticated user can create sessions (default)
	AllowedOrigins  []string `json:"allowed_origins"` // CORS allow-list; empty = no cross-origin (same-origin only)
}

// NewAPISettings creates a new API settings by loading all values from environment
func NewAPISettings() APISettings {
	return APISettings{
		UseSSLWebSocket: getEnvOrDefaultBool(ENV_WEBSOCKETSSL, false),
		SigningSecret:   getEnvOrDefaultString(ENV_SIGNING_SECRET, ""),
		MasterKey:       getEnvOrDefaultString(ENV_MASTER_KEY, ""),
		WebhookTimeout:  getEnvOrDefaultUint32(ENV_WEBHOOK_TIMEOUT, 10000),
		Prefix:          getEnvOrDefaultString(ENV_API_PREFIX, "api"),
		Timeout:         getEnvOrDefaultUint32(ENV_API_TIMEOUT, 30000),
		DefaultVersion:  normalizeAPIDefaultVersion(getEnvOrDefaultString(ENV_API_DEFAULT_VER, CurrentDefaultAPIVersion)),
		User:            getEnvOrDefaultString(ENV_USER, ""),
		Password:        getEnvOrDefaultString(ENV_PASSWORD, ""),
		RelaxedSessions: getEnvOrDefaultBool(ENV_RELAXED_SESSIONS, true),
		AllowedOrigins:  parseOriginList(getEnvOrDefaultString(ENV_CORS_ORIGINS, "")),
	}
}

const CurrentDefaultAPIVersion = "v4"

func normalizeAPIDefaultVersion(version string) string {
	switch strings.ToLower(strings.TrimSpace(version)) {
	case "v5":
		return "v5"
	case "v4":
		fallthrough
	default:
		return CurrentDefaultAPIVersion
	}
}

// GetAPITimeout returns the API timeout as time.Duration
func (settings APISettings) GetAPITimeout() time.Duration {
	return time.Duration(settings.Timeout) * time.Millisecond
}

// parseOriginList splits a comma-separated CORS origin list, trimming blanks.
func parseOriginList(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		if o := strings.TrimSpace(p); o != "" {
			origins = append(origins, o)
		}
	}
	return origins
}
