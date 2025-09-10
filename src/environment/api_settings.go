package environment

// API environment variable names
const (
	ENV_WEBAPIPORT            = "WEBAPIPORT"            // web API port
	ENV_WEBAPIHOST            = "WEBAPIHOST"            // web API host
	ENV_WEBSOCKETSSL          = "WEBSOCKETSSL"          // use SSL for websocket qrcode
	ENV_SIGNING_SECRET        = "SIGNING_SECRET"        // token for hash signing cookies
	ENV_MASTER_KEY            = "MASTERKEY"             // used for manage all instances at all
	ENV_HTTPLOGS              = "HTTPLOGS"              // log HTTP requests
	ENV_WEBHOOK_RETRY_COUNT   = "WEBHOOK_RETRY_COUNT"   // number of retry attempts for webhook failures
	ENV_WEBHOOK_RETRY_DELAY   = "WEBHOOK_RETRY_DELAY"   // delay in seconds between retry attempts
	ENV_WEBHOOK_TIMEOUT       = "WEBHOOK_TIMEOUT"       // timeout in seconds for webhook requests
	ENV_WEBHOOK_QUEUE_ENABLED = "WEBHOOK_QUEUE_ENABLED" // enable webhook queue system
	ENV_WEBHOOK_QUEUE_SIZE    = "WEBHOOK_QUEUE_SIZE"    // maximum size of webhook queue
	ENV_WEBHOOK_QUEUE_TIMEOUT = "WEBHOOK_QUEUE_TIMEOUT" // timeout for queue processing
	ENV_WEBHOOK_QUEUE_DELAY   = "WEBHOOK_QUEUE_DELAY"   // delay between queue processing
	ENV_WEBHOOK_WORKERS       = "WEBHOOK_WORKERS"       // number of concurrent webhook workers
)

// APISettings holds all API configuration loaded from environment
type APISettings struct {
	Port                string `json:"port"`
	Host                string `json:"host"`
	UseSSLWebSocket     bool   `json:"use_ssl_websocket"`
	SigningSecret       string `json:"signing_secret"`
	MasterKey           string `json:"master_key"`
	HTTPLogs            bool   `json:"http_logs"`
	WebhookRetryCount   int    `json:"webhook_retry_count"`
	WebhookRetryDelay   int    `json:"webhook_retry_delay"`
	WebhookTimeout      int    `json:"webhook_timeout"`
	WebhookQueueEnabled bool   `json:"webhook_queue_enabled"`
	WebhookQueueSize    int    `json:"webhook_queue_size"`
	WebhookQueueTimeout int    `json:"webhook_queue_timeout"`
	WebhookQueueDelay   int    `json:"webhook_queue_delay"`
	WebhookWorkers      int    `json:"webhook_workers"`
}

// NewAPISettings creates a new API settings by loading all values from environment
func NewAPISettings() APISettings {
	return APISettings{
		Port:                getEnvOrDefaultString(ENV_WEBAPIPORT, "31000"),
		Host:                getEnvOrDefaultString(ENV_WEBAPIHOST, ""),
		UseSSLWebSocket:     getEnvOrDefaultBool(ENV_WEBSOCKETSSL, false),
		SigningSecret:       getEnvOrDefaultString(ENV_SIGNING_SECRET, ""),
		MasterKey:           getEnvOrDefaultString(ENV_MASTER_KEY, ""),
		HTTPLogs:            getEnvOrDefaultBool(ENV_HTTPLOGS, false),
		WebhookRetryCount:   getOptionalEnvInt(ENV_WEBHOOK_RETRY_COUNT),
		WebhookRetryDelay:   getOptionalEnvInt(ENV_WEBHOOK_RETRY_DELAY),
		WebhookTimeout:      getOptionalEnvInt(ENV_WEBHOOK_TIMEOUT),
		WebhookQueueEnabled: getEnvOrDefaultBool(ENV_WEBHOOK_QUEUE_ENABLED, false),
		WebhookQueueSize:    getOptionalEnvInt(ENV_WEBHOOK_QUEUE_SIZE),
		WebhookQueueTimeout: getOptionalEnvInt(ENV_WEBHOOK_QUEUE_TIMEOUT),
		WebhookQueueDelay:   getOptionalEnvInt(ENV_WEBHOOK_QUEUE_DELAY),
		WebhookWorkers:      getOptionalEnvInt(ENV_WEBHOOK_WORKERS),
	}
}

// IsWebhookRetryEnabled checks if webhook retry system is enabled
func (settings APISettings) IsWebhookRetryEnabled() bool {
	return settings.WebhookRetryCount >= 0
}

// GetWebhookRetryCount returns the retry count with default value if retry is enabled
func (settings APISettings) GetWebhookRetryCount() int {
	if settings.WebhookRetryCount >= 0 {
		return settings.WebhookRetryCount
	}
	return 3 // Default when enabled but not specified
}

// GetWebhookRetryDelay returns the retry delay with default value if retry is enabled
func (settings APISettings) GetWebhookRetryDelay() int {
	if settings.WebhookRetryDelay >= 0 {
		return settings.WebhookRetryDelay
	}
	return 1 // Default when enabled but not specified
}

// GetWebhookTimeout returns the timeout with default value
func (settings APISettings) GetWebhookTimeout() int {
	if settings.WebhookTimeout >= 0 {
		return settings.WebhookTimeout
	}
	return 10 // Default timeout
}

// GetWebhookQueueSize returns the queue size with default value and validation
func (settings APISettings) GetWebhookQueueSize() int {
	if settings.WebhookQueueSize > 0 {
		// Validate reasonable limits
		if settings.WebhookQueueSize > 10000 {
			// Prevent excessive memory usage
			return 10000
		}
		return settings.WebhookQueueSize
	}
	return 1000 // Default queue size
}

// GetWebhookQueueTimeout returns the queue timeout with default value
func (settings APISettings) GetWebhookQueueTimeout() int {
	if settings.WebhookQueueTimeout >= 0 {
		return settings.WebhookQueueTimeout
	}
	return 30 // Default queue timeout
}

// GetWebhookQueueDelay returns the queue delay with default value
func (settings APISettings) GetWebhookQueueDelay() int {
	if settings.WebhookQueueDelay >= 0 {
		return settings.WebhookQueueDelay
	}
	return 0 // Default no delay
}

// GetWebhookWorkers returns the number of concurrent webhook workers with default value and validation
func (settings APISettings) GetWebhookWorkers() int {
	if settings.WebhookWorkers > 0 {
		// Validate reasonable limits
		if settings.WebhookWorkers > 50 {
			// Prevent excessive goroutines
			return 50
		}
		return settings.WebhookWorkers
	}
	return 10 // Default ten workers
}
