package environment

import "time"

// API environment variable names
const (
	ENV_WEBSOCKETSSL    = "WEBSOCKETSSL"    // use SSL for websocket qrcode
	ENV_SIGNING_SECRET  = "SIGNING_SECRET"  // token for hash signing cookies
	ENV_MASTER_KEY      = "MASTERKEY"       // used for manage all instances at all
	ENV_WEBHOOK_TIMEOUT = "WEBHOOK_TIMEOUT" // timeout in milliseconds for webhook requests
	ENV_API_PREFIX      = "API_PREFIX"      // API routes prefix
	ENV_API_TIMEOUT     = "API_TIMEOUT"     // API request timeout in milliseconds
)

// APISettings holds all API configuration loaded from environment
type APISettings struct {
	UseSSLWebSocket bool   `json:"use_ssl_websocket"`
	SigningSecret   string `json:"signing_secret"`
	MasterKey       string `json:"master_key"`
	WebhookTimeout  uint32 `json:"webhook_timeout"` // webhook timeout in milliseconds
	Prefix          string `json:"prefix"`
	Timeout         uint32 `json:"timeout"` // API request timeout in milliseconds
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
	}
}

// GetAPITimeout returns the API timeout as time.Duration
func (settings APISettings) GetAPITimeout() time.Duration {
	return time.Duration(settings.Timeout) * time.Millisecond
}
