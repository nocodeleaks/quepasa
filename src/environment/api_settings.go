package environment

// API environment variable names
const (
	ENV_WEBSOCKETSSL    = "WEBSOCKETSSL"    // use SSL for websocket qrcode
	ENV_SIGNING_SECRET  = "SIGNING_SECRET"  // token for hash signing cookies
	ENV_MASTER_KEY      = "MASTERKEY"       // used for manage all instances at all
	ENV_WEBHOOK_TIMEOUT = "WEBHOOK_TIMEOUT" // timeout in seconds for webhook requests
	ENV_API_PREFIX      = "API_PREFIX"      // API routes prefix
)

// APISettings holds all API configuration loaded from environment
type APISettings struct {
	UseSSLWebSocket bool   `json:"use_ssl_websocket"`
	SigningSecret   string `json:"signing_secret"`
	MasterKey       string `json:"master_key"`
	WebhookTimeout  int    `json:"webhook_timeout"`
	Prefix          string `json:"prefix"`
}

// NewAPISettings creates a new API settings by loading all values from environment
func NewAPISettings() APISettings {
	return APISettings{
		UseSSLWebSocket: getEnvOrDefaultBool(ENV_WEBSOCKETSSL, false),
		SigningSecret:   getEnvOrDefaultString(ENV_SIGNING_SECRET, ""),
		MasterKey:       getEnvOrDefaultString(ENV_MASTER_KEY, ""),
		WebhookTimeout:  getOptionalEnvInt(ENV_WEBHOOK_TIMEOUT),
		Prefix:          getEnvOrDefaultString(ENV_API_PREFIX, "api"),
	}
}

// GetWebhookTimeout returns the timeout with default value
func (settings APISettings) GetWebhookTimeout() int {
	if settings.WebhookTimeout >= 0 {
		return settings.WebhookTimeout
	}
	return 10 // Default timeout
}
