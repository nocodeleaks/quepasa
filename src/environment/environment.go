package environment

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	logrus "github.com/sirupsen/logrus"
)

// EnvironmentSettings provides centralized access to all environment configurations.
// This is the main class that aggregates all environment variable management.
type EnvironmentSettings struct {
	// Embedded structs for organized access to different environment categories
	Database  DatabaseSettings
	API       APISettings
	WhatsApp  WhatsAppSettings
	Whatsmeow WhatsmeowSettings
	SIPProxy  SIPProxySettings
	General   GeneralSettings
	RabbitMQ  RabbitMQSettings
}

// Settings is the global singleton instance for accessing all environment configurations.
var Settings EnvironmentSettings

// Initialize the global environment manager
func init() {
	logentry := logrus.NewEntry(logrus.StandardLogger()).WithField("package", "environment")
	logentry.Println("Starting Environment Manager initialization...")

	// loading environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		logentry.Println("Failed to load .env file")
	} else {
		logentry.Println("Successfully loaded .env file")
	}

	Settings = EnvironmentSettings{
		Database:  NewDatabaseSettings(),
		API:       NewAPISettings(),
		WhatsApp:  NewWhatsAppSettings(),
		Whatsmeow: NewWhatsmeowSettings(),
		SIPProxy:  NewSIPProxySettings(),
		General:   NewGeneralSettings(),
		RabbitMQ:  NewRabbitMQSettings(),
	}

	logentry.Println("Environment Manager ready - All configurations loaded!")
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

// Helper function to get optional uint32 from environment (redeclared locally to avoid conflicts)
func getOptionalEnvUint32(key string) *uint32 {
	if valueStr, ok := os.LookupEnv(key); ok {
		trimmedValueStr := strings.TrimSpace(valueStr)
		if trimmedValueStr == "" {
			return nil
		}
		if parsedValue, err := strconv.ParseUint(trimmedValueStr, 10, 32); err == nil {
			result := uint32(parsedValue)
			return &result
		}
	}
	return nil
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

// getEnvOrDefaultInt fetches a signed integer environment variable, returning a default value.
// It logs a warning if the environment variable exists but cannot be parsed as an integer.
func getEnvOrDefaultInt(key string, defaultValue int) int {
	if valueStr, ok := os.LookupEnv(key); ok {
		trimmedValueStr := strings.TrimSpace(valueStr)
		if parsedValue, err := strconv.Atoi(trimmedValueStr); err == nil {
			return parsedValue
		}
		logrus.Warnf("Invalid integer value for environment variable %s: '%s'. Using default: %d", key, valueStr, defaultValue)
	}
	return defaultValue
}

// isEnvVarSet checks if an environment variable is explicitly set
func isEnvVarSet(key string) bool {
	_, exists := os.LookupEnv(key)
	return exists
}

// getOptionalEnvInt fetches an integer environment variable.
// Returns the value if set, or -1 if not set (indicating disabled/not configured)
func getOptionalEnvInt(key string) int {
	if valueStr, ok := os.LookupEnv(key); ok {
		trimmedValueStr := strings.TrimSpace(valueStr)
		if parsedValue, err := strconv.Atoi(trimmedValueStr); err == nil {
			return parsedValue
		}
		logrus.Warnf("Invalid integer value for environment variable %s: '%s'. Feature will be disabled", key, valueStr)
	}
	return -1 // Not set or invalid means "disabled"
}
