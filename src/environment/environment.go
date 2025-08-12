package environment

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

// EnvironmentManager provides centralized access to all environment configurations.
// This is the main class that aggregates all environment variable management.
type EnvironmentManager struct {
	// Embedded structs for organized access to different environment categories
	Database *DatabaseEnvironment
	API      *APIEnvironment
	WhatsApp *WhatsAppEnvironment
	SIPProxy *SIPProxyEnvironment
	Logging  *LoggingEnvironment
	General  *GeneralEnvironment
	RabbitMQ *RabbitMQEnvironment
}

// Settings is the global singleton instance for accessing all environment configurations.
var Settings *EnvironmentManager

// Initialize the global environment manager
func init() {
	Settings = &EnvironmentManager{
		Database: &DatabaseEnvironment{},
		API:      &APIEnvironment{},
		WhatsApp: &WhatsAppEnvironment{},
		SIPProxy: &SIPProxyEnvironment{},
		Logging:  &LoggingEnvironment{},
		General:  &GeneralEnvironment{},
		RabbitMQ: &RabbitMQEnvironment{},
	}
}

// ErrEnvVarEmpty is returned when an environment variable is requested but is empty.
var ErrEnvVarEmpty = errors.New("getenv: environment variable empty")

// --- Helper Functions for Environment Variables ---

// getEnvOrDefaultString fetches an environment variable, returning a default value if not set.
func getEnvOrDefaultString(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(value)
	}
	return defaultValue
}

// getEnvOrDefaultBool fetches a boolean environment variable, returning a default value.
// It logs a warning if the environment variable exists but cannot be parsed as a boolean.
func getEnvOrDefaultBool(key string, defaultValue bool) bool {
	if valueStr, ok := os.LookupEnv(key); ok {
		trimmedValueStr := strings.TrimSpace(valueStr)
		if parsedValue, err := strconv.ParseBool(trimmedValueStr); err == nil {
			return parsedValue
		}
		logrus.Warnf("Invalid boolean value for environment variable %s: '%s'. Using default: %t", key, valueStr, defaultValue)
	}
	return defaultValue
}

// getEnvOrDefaultUint64 fetches an unsigned 64-bit integer environment variable, returning a default value.
// It logs a warning if the environment variable exists but cannot be parsed as a uint64.
func getEnvOrDefaultUint64(key string, defaultValue uint64) uint64 {
	if valueStr, ok := os.LookupEnv(key); ok {
		trimmedValueStr := strings.TrimSpace(valueStr)
		if parsedValue, err := strconv.ParseUint(trimmedValueStr, 10, 64); err == nil {
			return parsedValue
		}
		logrus.Warnf("Invalid unsigned integer value for environment variable %s: '%s'. Using default: %d", key, valueStr, defaultValue)
	}
	return defaultValue
}

// getEnvOrDefaultUint32 fetches an unsigned 32-bit integer environment variable, returning a default value.
// It logs a warning if the environment variable exists but cannot be parsed as a uint32.
func getEnvOrDefaultUint32(key string, defaultValue uint32) uint32 {
	if valueStr, ok := os.LookupEnv(key); ok {
		trimmedValueStr := strings.TrimSpace(valueStr)
		if parsedValue, err := strconv.ParseUint(trimmedValueStr, 10, 32); err == nil {
			return uint32(parsedValue)
		}
		logrus.Warnf("Invalid unsigned integer value for environment variable %s: '%s'. Using default: %d", key, valueStr, defaultValue)
	}
	return defaultValue
}

// getOptionalEnvBool fetches a boolean environment variable where nil indicates "use system default logic".
// It returns:
//   - *bool (true): if the variable is explicitly set to "true", "1", "yes", etc.
//   - *bool (false): if the variable is explicitly set to "false", "0", "no", etc.
//   - nil: if the variable is not set, empty, or its value is not a valid boolean,
//     indicating that the system's internal default logic should apply.
func getOptionalEnvBool(key string) *bool {
	if valueStr, ok := os.LookupEnv(key); ok {
		trimmedValueStr := strings.TrimSpace(valueStr)
		if trimmedValueStr == "" {
			return nil // Empty string means "use default logic"
		}
		if parsedValue, err := strconv.ParseBool(trimmedValueStr); err == nil {
			return &parsedValue
		}
	}
	return nil // Not set or invalid means "use default logic"
}
