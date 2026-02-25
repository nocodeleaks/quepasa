package environment

// WhatsApp environment variable names
const (
	ENV_DISPATCH_UNHANDLED                = "DISPATCHUNHANDLED"                 // enable or disable dispatch unhandled messages
	ENV_WHATSMEOWLOGLEVEL                 = "WHATSMEOW_LOGLEVEL"                // Whatsmeow log level
	ENV_WHATSMEOWDBLOGLEVEL               = "WHATSMEOW_DBLOGLEVEL"              // Whatsmeow database log level
	ENV_WHATSMEOW_USE_RETRY_MESSAGE_STORE = "WHATSMEOW_USE_RETRY_MESSAGE_STORE" // persist outgoing messages for retry receipts
)

// WhatsmeowSettings holds all WhatsApp configuration loaded from environment
type WhatsmeowSettings struct {
	DispatchUnhandled    bool   `json:"dispatch_unhandled"`
	LogLevel             string `json:"whatsmeow_log_level"`
	DBLogLevel           string `json:"whatsmeow_db_log_level"`
	UseRetryMessageStore bool   `json:"whatsmeow_use_retry_message_store"`
}

// NewWhatsmeowSettings creates a new Whatsmeow settings by loading all values from environment
func NewWhatsmeowSettings() WhatsmeowSettings {
	return WhatsmeowSettings{
		DispatchUnhandled:    getEnvOrDefaultBool(ENV_DISPATCH_UNHANDLED, false),
		LogLevel:             getEnvOrDefaultString(ENV_WHATSMEOWLOGLEVEL, ""),
		DBLogLevel:           getEnvOrDefaultString(ENV_WHATSMEOWDBLOGLEVEL, ""),
		UseRetryMessageStore: getEnvOrDefaultBool(ENV_WHATSMEOW_USE_RETRY_MESSAGE_STORE, false),
	}
}
