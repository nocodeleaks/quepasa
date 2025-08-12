package environment

// APIEnvironment handles all API-related environment variables
type APIEnvironment struct{}

// API environment variable names
const (
	ENV_WEBAPIPORT     = "WEBAPIPORT"     // web API port
	ENV_WEBAPIHOST     = "WEBAPIHOST"     // web API host
	ENV_WEBSOCKETSSL   = "WEBSOCKETSSL"   // use SSL for websocket qrcode
	ENV_SIGNING_SECRET = "SIGNING_SECRET" // token for hash signing cookies
	ENV_MASTER_KEY     = "MASTERKEY"      // used for manage all instances at all
	ENV_HTTPLOGS       = "HTTPLOGS"       // log HTTP requests
)

// Port returns the web API port. Defaults to "31000".
func (env *APIEnvironment) Port() string {
	return getEnvOrDefaultString(ENV_WEBAPIPORT, "31000")
}

// Host returns the web API host. Defaults to empty string.
func (env *APIEnvironment) Host() string {
	return getEnvOrDefaultString(ENV_WEBAPIHOST, "")
}

// UseSSLForWebSocket checks if SSL should be used for WebSocket QR code. Defaults to false.
func (env *APIEnvironment) UseSSLForWebSocket() bool {
	return getEnvOrDefaultBool(ENV_WEBSOCKETSSL, false)
}

// SigningSecret returns the signing secret for cookies. Defaults to empty string.
func (env *APIEnvironment) SigningSecret() string {
	return getEnvOrDefaultString(ENV_SIGNING_SECRET, "")
}

// MasterKey returns the master key for super admin methods. Defaults to empty string.
func (env *APIEnvironment) MasterKey() string {
	return getEnvOrDefaultString(ENV_MASTER_KEY, "")
}

// HTTPLogs checks if HTTP requests should be logged. Defaults to false.
func (env *APIEnvironment) HTTPLogs() bool {
	return getEnvOrDefaultBool(ENV_HTTPLOGS, false)
}
