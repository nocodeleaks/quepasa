package environment

// WebServer environment variable names
const (
	ENV_WEBSERVER_PORT            = "WEBSERVER_PORT"            // web server port (fallback: WEBAPIPORT)
	ENV_WEBSERVER_HOST            = "WEBSERVER_HOST"            // web server host (fallback: WEBAPIHOST)
	ENV_WEBSERVER_LOGS            = "WEBSERVER_LOGS"            // web server HTTP logs (fallback: HTTPLOGS)
	ENV_QUEPASA_BASE_URL          = "QUEPASA_BASE_URL"          // canonical external base URL
	ENV_QUEPASA_DEV_FRONTEND      = "QUEPASA_DEV_FRONTEND"      // enable frontend dev reverse proxy
	ENV_QUEPASA_FRONTEND_HOST     = "QUEPASA_FRONTEND_HOST"     // frontend dev origin host
	ENV_QUEPASA_FRONTEND_DEV_PORT = "QUEPASA_FRONTEND_DEV_PORT" // frontend dev origin port
)

// WebServerSettings holds all WebServer configuration loaded from environment
type WebServerSettings struct {
	Port            uint32 `json:"port"`
	Host            string `json:"host"`
	Logs            bool   `json:"logs"`
	BaseURL         string `json:"base_url"`
	DevFrontend     bool   `json:"dev_frontend"`
	FrontendHost    string `json:"frontend_host"`
	FrontendDevPort string `json:"frontend_dev_port"`
}

// NewWebServerSettings creates a new WebServer settings by loading all values from environment
// with fallback compatibility to old variable names
func NewWebServerSettings() WebServerSettings {
	return WebServerSettings{
		Port:            getWebServerPort(),
		Host:            getWebServerHost(),
		Logs:            getWebServerLogs(),
		BaseURL:         getEnvOrDefaultString(ENV_QUEPASA_BASE_URL, ""),
		DevFrontend:     getEnvOrDefaultBool(ENV_QUEPASA_DEV_FRONTEND, false),
		FrontendHost:    getEnvOrDefaultString(ENV_QUEPASA_FRONTEND_HOST, "http://127.0.0.1"),
		FrontendDevPort: getEnvOrDefaultString(ENV_QUEPASA_FRONTEND_DEV_PORT, "5173"),
	}
}

// getWebServerPort gets the web server port with fallback compatibility
func getWebServerPort() uint32 {
	// Try new variable first
	if isEnvVarSet(ENV_WEBSERVER_PORT) {
		return getEnvOrDefaultUint32(ENV_WEBSERVER_PORT, 31000)
	}
	// Fallback to old variable
	return getEnvOrDefaultUint32("WEBAPIPORT", 31000)
}

// getWebServerHost gets the web server host with fallback compatibility
func getWebServerHost() string {
	// Try new variable first
	if host := getEnvOrDefaultString(ENV_WEBSERVER_HOST, ""); host != "" {
		return host
	}
	// Fallback to old variable
	return getEnvOrDefaultString("WEBAPIHOST", "")
}

// getWebServerHTTPLogs gets the web server HTTP logs setting with fallback compatibility
func getWebServerLogs() bool {
	// Try new variable first
	if isEnvVarSet(ENV_WEBSERVER_LOGS) {
		return getEnvOrDefaultBool(ENV_WEBSERVER_LOGS, false)
	}
	// Fallback to old variable
	return getEnvOrDefaultBool("HTTPLOGS", false)
}
