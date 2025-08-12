package environment

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// LoggingEnvironment handles all logging-related environment variables
type LoggingEnvironment struct{}

// Logging environment variable names
const (
	ENV_LOGLEVEL            = "LOGLEVEL"             // general log level
	ENV_WHATSMEOWLOGLEVEL   = "WHATSMEOW_LOGLEVEL"   // Whatsmeow log level
	ENV_WHATSMEOWDBLOGLEVEL = "WHATSMEOW_DBLOGLEVEL" // Whatsmeow database log level
)

// LogLevel returns the general log level. Defaults to empty string.
func (env *LoggingEnvironment) LogLevel() string {
	return getEnvOrDefaultString(ENV_LOGLEVEL, "")
}

// LogLevelFromLogrus returns a parsed logrus.Level from the environment.
// If the environment variable is empty, returns the provided default level.
// If the environment variable contains an invalid level, it panics with a detailed error message,
// as this indicates a critical configuration error that must be addressed.
func (env *LoggingEnvironment) LogLevelFromLogrus(defaultLevel logrus.Level) logrus.Level {
	envLevelStr := env.LogLevel()
	if len(envLevelStr) == 0 {
		return defaultLevel
	}

	logrusLevel, err := logrus.ParseLevel(envLevelStr)
	if err != nil {
		panic(fmt.Sprintf("Invalid log level '%s' specified in environment variable %s: %v. Please correct this critical configuration.", envLevelStr, ENV_LOGLEVEL, err))
	}

	return logrusLevel
}

// WhatsmeowLogLevel returns the Whatsmeow log level. Defaults to empty string.
func (env *LoggingEnvironment) WhatsmeowLogLevel() string {
	return getEnvOrDefaultString(ENV_WHATSMEOWLOGLEVEL, "")
}

// WhatsmeowDBLogLevel returns the Whatsmeow database log level. Defaults to empty string.
func (env *LoggingEnvironment) WhatsmeowDBLogLevel() string {
	return getEnvOrDefaultString(ENV_WHATSMEOWDBLOGLEVEL, "")
}
