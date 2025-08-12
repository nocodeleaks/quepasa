package environment

import (
	"strconv"
	"strings"
)

// GeneralEnvironment handles general application environment variables
type GeneralEnvironment struct{}

// General environment variable names
const (
	ENV_MIGRATIONS               = "MIGRATIONS"               // enable migrations (can also be a path)
	ENV_TITLE                    = "APP_TITLE"                // application title for whatsapp id
	ENV_REMOVEDIGIT9             = "REMOVEDIGIT9"             // remove digit 9 from phones
	ENV_SYNOPSISLENGTH           = "SYNOPSISLENGTH"           // synopsis length for messages
	ENV_CACHELENGTH              = "CACHELENGTH"              // cache max items
	ENV_CACHEDAYS                = "CACHEDAYS"                // cache max days
	ENV_CONVERT_WAVE_TO_OGG      = "CONVERT_WAVE_TO_OGG"      // convert wave to OGG
	ENV_COMPATIBLE_MIME_AS_AUDIO = "COMPATIBLE_MIME_AS_AUDIO" // treat compatible MIME as audio
	ENV_ACCOUNTSETUP             = "ACCOUNTSETUP"             // enable or disable account creation
	ENV_TESTING                  = "TESTING"                  // testing mode
	ENV_DISPATCH_UNHANDLED       = "DISPATCHUNHANDLED"        // enable or disable dispatch unhandled messages
)

// UseCompatibleMIMEsAsAudio checks if compatible MIME types should be treated as audio.
// Defaults to true.
func (env *GeneralEnvironment) UseCompatibleMIMEsAsAudio() bool {
	convertWave := getEnvOrDefaultBool(ENV_CONVERT_WAVE_TO_OGG, true)
	compatibleMime := getEnvOrDefaultBool(ENV_COMPATIBLE_MIME_AS_AUDIO, true)
	return convertWave || compatibleMime
}

// Migrate checks if database migrations should be enabled. Defaults to true.
func (env *GeneralEnvironment) Migrate() bool {
	return getEnvOrDefaultBool(ENV_MIGRATIONS, true)
}

// MigrationPath returns the custom path for database migrations.
// Returns an empty string if migrations are enabled via boolean flag or no custom path is set.
func (env *GeneralEnvironment) MigrationPath() string {
	rawValue := getEnvOrDefaultString(ENV_MIGRATIONS, "")
	trimmedValue := strings.TrimSpace(rawValue)

	if trimmedValue == "" {
		return ""
	}
	if _, err := strconv.ParseBool(trimmedValue); err == nil {
		return "" // Boolean value means no custom path
	}
	return trimmedValue // Return as custom path
}

// AppTitle returns the application title. Defaults to empty string.
func (env *GeneralEnvironment) AppTitle() string {
	return getEnvOrDefaultString(ENV_TITLE, "")
}

// ShouldRemoveDigit9 checks if the 9th digit should be removed from phone numbers.
// Returns true or false, defaulting to false if the environment variable is not set or invalid.
func (env *GeneralEnvironment) ShouldRemoveDigit9() bool {
	return getEnvOrDefaultBool(ENV_REMOVEDIGIT9, false)
}

// SynopsisLength returns the length for message synopsis. Defaults to 50.
func (env *GeneralEnvironment) SynopsisLength() uint64 {
	return getEnvOrDefaultUint64(ENV_SYNOPSISLENGTH, 50)
}

// CacheLength returns the maximum number of items for the cache. Defaults to 0 (no limit).
func (env *GeneralEnvironment) CacheLength() uint64 {
	return getEnvOrDefaultUint64(ENV_CACHELENGTH, 0)
}

// CacheDays returns the maximum number of days for cached messages. Defaults to 0 (no limit).
func (env *GeneralEnvironment) CacheDays() uint64 {
	return getEnvOrDefaultUint64(ENV_CACHEDAYS, 0)
}

// Testing checks if testing methods should be applied. Defaults to false.
func (env *GeneralEnvironment) Testing() bool {
	return getEnvOrDefaultBool(ENV_TESTING, false)
}

// AccountSetup checks if account creation is enabled. Defaults to true.
func (env *GeneralEnvironment) AccountSetup() bool {
	return getEnvOrDefaultBool(ENV_ACCOUNTSETUP, true)
}

// DispatchUnhandled checks if dispatching unhandled messages is enabled. Defaults to false.
func (env *GeneralEnvironment) DispatchUnhandled() bool {
	return getEnvOrDefaultBool(ENV_DISPATCH_UNHANDLED, false)
}
