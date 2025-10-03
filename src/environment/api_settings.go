package environment

// API environment variable names
const (
	ENV_WEBAPIPORT      = "WEBAPIPORT"      // web API port
	ENV_WEBAPIHOST      = "WEBAPIHOST"      // web API host
	ENV_WEBSOCKETSSL    = "WEBSOCKETSSL"    // use SSL for websocket qrcode
	ENV_SIGNING_SECRET  = "SIGNING_SECRET"  // token for hash signing cookies
	ENV_MASTER_KEY      = "MASTERKEY"       // used for manage all instances at all
	ENV_HTTPLOGS        = "HTTPLOGS"        // log HTTP requests
	ENV_WEBHOOK_TIMEOUT = "WEBHOOK_TIMEOUT" // timeout in seconds for webhook requests
)

// APISettings holds all API configuration loaded from environment
type APISettings struct {
	Port            string `json:"port"`
	Host            string `json:"host"`
	UseSSLWebSocket bool   `json:"use_ssl_websocket"`
	SigningSecret   string `json:"signing_secret"`
	MasterKey       string `json:"master_key"`
	HTTPLogs        bool   `json:"http_logs"`
	WebhookTimeout  int    `json:"webhook_timeout"`
}

// NewAPISettings creates a new API settings by loading all values from environment
func NewAPISettings() APISettings {
	return APISettings{
		Port:            getEnvOrDefaultString(ENV_WEBAPIPORT, "31000"),
		Host:            getEnvOrDefaultString(ENV_WEBAPIHOST, ""),
		UseSSLWebSocket: getEnvOrDefaultBool(ENV_WEBSOCKETSSL, false),
		SigningSecret:   getEnvOrDefaultString(ENV_SIGNING_SECRET, ""),
		MasterKey:       getEnvOrDefaultString(ENV_MASTER_KEY, ""),
		HTTPLogs:        getEnvOrDefaultBool(ENV_HTTPLOGS, false),
		WebhookTimeout:  getOptionalEnvInt(ENV_WEBHOOK_TIMEOUT),
	}
}

// GetWebhookTimeout returns the timeout with default value
func (settings APISettings) GetWebhookTimeout() int {
	if settings.WebhookTimeout >= 0 {
		return settings.WebhookTimeout
	}
	return 10 // Default timeout
}
