package models

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings" // Certifique-se de que "strings" está importado

	library "github.com/nocodeleaks/quepasa/library"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	"github.com/sirupsen/logrus"
)

// Environment variable names
const (
	ENV_WEBAPIPORT = "WEBAPIPORT"
	ENV_WEBAPIHOST = "WEBAPIHOST"

	ENV_DBDRIVER   = "DBDRIVER" // database driver, default sqlite3
	ENV_DBHOST     = "DBHOST"
	ENV_DBDATABASE = "DBDATABASE"
	ENV_DBPORT     = "DBPORT"
	ENV_DBUSER     = "DBUSER"
	ENV_DBPASSWORD = "DBPASSWORD"
	ENV_DBSSLMODE  = "DBSSLMODE"

	ENV_SIGNING_SECRET = "SIGNING_SECRET" // token for hash singing cookies
	ENV_MASTER_KEY     = "MASTERKEY"      // used for manage all instances at all

	ENV_WEBSOCKETSSL             = "WEBSOCKETSSL" // use ssl for websocket qrcode
	ENV_MIGRATIONS               = "MIGRATIONS"   // enable migrations (can also be a path)
	ENV_TITLE                    = "APP_TITLE"    // application title for whatsapp id
	ENV_REMOVEDIGIT9             = "REMOVEDIGIT9"
	ENV_SYNOPSISLENGTH           = "SYNOPSISLENGTH"
	ENV_CACHELENGTH              = "CACHELENGTH" // cache max items
	ENV_CACHEDAYS                = "CACHEDAYS"   // cache max days
	ENV_CONVERT_WAVE_TO_OGG      = "CONVERT_WAVE_TO_OGG"
	ENV_COMPATIBLE_MIME_AS_AUDIO = "COMPATIBLE_MIME_AS_AUDIO"

	ENV_READUPDATE      = "READUPDATE"
	ENV_READRECEIPTS    = "READRECEIPTS"
	ENV_CALLS           = "CALLS"
	ENV_GROUPS          = "GROUPS"
	ENV_BROADCASTS      = "BROADCASTS"
	ENV_HISTORYSYNCDAYS = "HISTORYSYNCDAYS"

	ENV_PRESENCE            = "PRESENCE"
	ENV_LOGLEVEL            = "LOGLEVEL"
	ENV_HTTPLOGS            = "HTTPLOGS"
	ENV_WHATSMEOWLOGLEVEL   = "WHATSMEOW_LOGLEVEL"
	ENV_WHATSMEOWDBLOGLEVEL = "WHATSMEOW_DBLOGLEVEL"

	ENV_ACCOUNTSETUP = "ACCOUNTSETUP" // enable or disable account creation, default true
	ENV_TESTING      = "TESTING"

	ENV_RABBITMQ_QUEUE            = "RABBITMQ_QUEUE"            // Nome da variável de ambiente para a fila
	ENV_RABBITMQ_CONNECTIONSTRING = "RABBITMQ_CONNECTIONSTRING" // Nome da variável de ambiente para a string de conexão
	ENV_RABBITMQ_CACHELENGTH      = "RABBITMQ_CACHELENGTH"

	ENV_DISPATCH_UNHANDLED = "DISPATCHUNHANDLED" // enable or disable dispatch unhandled messages, default false
)

// Environment provides methods to access application configurations from environment variables.
type Environment struct{}

// ENV is the global singleton instance for accessing environment configurations.
var ENV Environment

// ErrEnvVarEmpty is returned when an environment variable is requested but is empty.
var ErrEnvVarEmpty = errors.New("getenv: environment variable empty")

// --- Helper Functions for Environment Variables ---

// getEnvOrDefaultString fetches an environment variable, returning a default value if not set.
func getEnvOrDefaultString(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(value) // Aplicado TrimSpace
	}
	return defaultValue
}

// getEnvOrDefaultBool fetches a boolean environment variable, returning a default value.
// It logs a warning if the environment variable exists but cannot be parsed as a boolean.
func getEnvOrDefaultBool(key string, defaultValue bool) bool {
	if valueStr, ok := os.LookupEnv(key); ok {
		trimmedValueStr := strings.TrimSpace(valueStr) // Aplicado TrimSpace
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
		trimmedValueStr := strings.TrimSpace(valueStr) // Aplicado TrimSpace
		if parsedValue, err := strconv.ParseUint(trimmedValueStr, 10, 64); err == nil {
			return parsedValue
		}
		logrus.Warnf("Invalid unsigned integer value for environment variable %s: '%s'. Using default: %d", key, valueStr, defaultValue)
	}
	return defaultValue
}

// --- DATABASE CONFIGURATION ---

// GetDBParameters retrieves database connection parameters from environment variables.
// It defaults to "sqlite3" if DBDRIVER is not set.
func (*Environment) GetDBParameters() library.DatabaseParameters {
	parameters := library.DatabaseParameters{}

	parameters.Driver = getEnvOrDefaultString(ENV_DBDRIVER, "sqlite3")
	parameters.Host = getEnvOrDefaultString(ENV_DBHOST, "")
	parameters.DataBase = getEnvOrDefaultString(ENV_DBDATABASE, "")
	parameters.Port = getEnvOrDefaultString(ENV_DBPORT, "")
	parameters.User = getEnvOrDefaultString(ENV_DBUSER, "")
	parameters.Password = getEnvOrDefaultString(ENV_DBPASSWORD, "")
	parameters.SSL = getEnvOrDefaultString(ENV_DBSSLMODE, "")
	return parameters
}

// --- GENERAL APPLICATION SETTINGS ---

// UseCompatibleMIMEsAsAudio checks if compatible MIME types should be treated as audio.
// Defaults to true.
func (*Environment) UseCompatibleMIMEsAsAudio() bool {
	convertWave := getEnvOrDefaultBool(ENV_CONVERT_WAVE_TO_OGG, true)
	compatibleMime := getEnvOrDefaultBool(ENV_COMPATIBLE_MIME_AS_AUDIO, true)
	return convertWave || compatibleMime
}

// UseSSLForWebSocket checks if SSL should be used for WebSocket QR code. Defaults to false.
func (*Environment) UseSSLForWebSocket() bool {
	return getEnvOrDefaultBool(ENV_WEBSOCKETSSL, false)
}

// Migrate checks if database migrations should be enabled. Defaults to true.
func (*Environment) Migrate() bool {
	return getEnvOrDefaultBool(ENV_MIGRATIONS, true)
}

// MigrationPath returns the custom path for database migrations.
// Returns an empty string if migrations are enabled via boolean flag or no custom path is set.
func (*Environment) MigrationPath() string {
	// Pega o valor bruto da variável de ambiente primeiro para aplicar TrimSpace.
	rawValue := os.Getenv(ENV_MIGRATIONS)
	trimmedValue := strings.TrimSpace(rawValue) // Aplicado TrimSpace

	// Se o valor trimado for vazio, ou se puder ser parseado como um booleano, retorna string vazia.
	// Isso mantém a lógica de que um valor booleano significa "sem caminho personalizado".
	if trimmedValue == "" {
		return ""
	}
	if _, err := strconv.ParseBool(trimmedValue); err == nil {
		return "" // Indica que deve usar o caminho padrão ou as migrações são desabilitadas/habilitadas por bool
	}
	return trimmedValue // Caso contrário, retorna o valor trimado como um caminho
}

// AppTitle returns the application title. Defaults to an empty string.
func (*Environment) AppTitle() string {
	return getEnvOrDefaultString(ENV_TITLE, "")
}

// ShouldRemoveDigit9 checks if the 9th digit should be removed from phone numbers.
// Returns true or false, defaulting to false if the environment variable is not set or invalid.
func (*Environment) ShouldRemoveDigit9() bool {
	return getEnvOrDefaultBool(ENV_REMOVEDIGIT9, false)
}

// SynopsisLength returns the length for message synopsis. Defaults to 50.
func (*Environment) SynopsisLength() uint64 {
	return getEnvOrDefaultUint64(ENV_SYNOPSISLENGTH, 50)
}

// CacheLength returns the maximum number of items for the cache. Defaults to 0 (no limit).
func (*Environment) CacheLength() uint64 {
	return getEnvOrDefaultUint64(ENV_CACHELENGTH, 0)
}

// CacheDays returns the maximum number of days for cached messages. Defaults to 0 (no limit).
func (*Environment) CacheDays() uint64 {
	return getEnvOrDefaultUint64(ENV_CACHEDAYS, 0)
}

// MasterKey returns the master key for super admin methods. Defaults to an empty string.
func (*Environment) MasterKey() string {
	return getEnvOrDefaultString(ENV_MASTER_KEY, "")
}

// Testing checks if testing methods should be applied. Defaults to false.
func (*Environment) Testing() bool {
	return getEnvOrDefaultBool(ENV_TESTING, false)
}

// AccountSetup checks if account creation is enabled. Defaults to true.
func (*Environment) AccountSetup() bool {
	return getEnvOrDefaultBool(ENV_ACCOUNTSETUP, true)
}

// DispatchUnhandled checks if dispatching unhandled messages is enabled. Defaults to false.
func (*Environment) DispatchUnhandled() bool {
	return getEnvOrDefaultBool(ENV_DISPATCH_UNHANDLED, false)
}

// --- WHATSAPP SERVICE OPTIONS - WHATSMEOW ---

// ParseWhatsappBoolean parses a string value into a WhatsappBooleanExtended type.
// It handles various string representations of boolean values, including extended ones.
func ParseWhatsappBoolean(value string) whatsapp.WhatsappBooleanExtended {
	// A função já faz trim, então não precisa de outro aqui.
	formatted := strings.TrimSpace(value)
	formatted = strings.Trim(formatted, `"`)
	formatted = strings.ToLower(formatted)

	switch formatted {
	case "", "0":
		return whatsapp.WhatsappBooleanExtended(whatsapp.UnSetBooleanType)
	case "1", "t", "true", "yes":
		return whatsapp.WhatsappBooleanExtended(whatsapp.TrueBooleanType)
	case "-1", "f", "false", "no":
		return whatsapp.WhatsappBooleanExtended(whatsapp.FalseBooleanType)
	case "-2", "forcedfalse":
		return whatsapp.ForcedFalseBooleanType
	case "2", "forcedtrue":
		return whatsapp.ForcedTrueBooleanType
	default:
		message := fmt.Sprintf("unknown extended boolean type: {%s}", value)
		panic(message)
	}
}

// Broadcasts returns the WhatsappBooleanExtended setting for broadcasts.
func (*Environment) Broadcasts() whatsapp.WhatsappBooleanExtended {
	v := os.Getenv(ENV_BROADCASTS)
	// Chama ParseWhatsappBoolean que já trima a string
	return ParseWhatsappBoolean(v)
}

// Groups returns the WhatsappBooleanExtended setting for groups.
func (*Environment) Groups() whatsapp.WhatsappBooleanExtended {
	v := os.Getenv(ENV_GROUPS)
	// Chama ParseWhatsappBoolean que já trima a string
	return ParseWhatsappBoolean(v)
}

// ReadReceipts returns the WhatsappBooleanExtended setting for read receipts.
func (*Environment) ReadReceipts() whatsapp.WhatsappBooleanExtended {
	v := os.Getenv(ENV_READRECEIPTS)
	// Chama ParseWhatsappBoolean que já trima a string
	return ParseWhatsappBoolean(v)
}

// Calls returns the WhatsappBooleanExtended setting for calls.
func (*Environment) Calls() whatsapp.WhatsappBooleanExtended {
	v := os.Getenv(ENV_CALLS)
	// Chama ParseWhatsappBoolean que já trima a string
	return ParseWhatsappBoolean(v)
}

// ReadUpdate checks if read updates are enabled.
// Returns true or false, defaulting to false if the environment variable is not set or invalid.
func (*Environment) ReadUpdate() bool {
	return getEnvOrDefaultBool(ENV_READUPDATE, false)
}

// HistorySync returns the history sync days. Returns nil if not set or invalid.
// A nil return indicates that the system should use its internal default logic
// for history sync days, rather than a forced value.
func (*Environment) HistorySync() *uint32 {
	rawValue := os.Getenv(ENV_HISTORYSYNCDAYS)
	stringValue := strings.TrimSpace(rawValue) // Aplicado TrimSpace

	if stringValue == "" {
		return nil
	}

	value, err := strconv.ParseUint(stringValue, 10, 32)
	if err != nil {
		logrus.Warnf("Invalid unsigned integer value for environment variable %s: '%s'. Returning nil (use system default logic). Error: %v", ENV_HISTORYSYNCDAYS, rawValue, err) // Loga o valor original para debug
		return nil
	}

	result := uint32(value)
	return &result
}

// --- LOGGING SETTINGS ---

// Presence returns the forced default presence status (lowercase).
func (*Environment) Presence() string {
	result := getEnvOrDefaultString(ENV_PRESENCE, "")
	return strings.ToLower(result)
}

// LogLevel returns the application log level (lowercase and trimmed).
func (*Environment) LogLevel() string {
	result := getEnvOrDefaultString(ENV_LOGLEVEL, "")
	return strings.ToLower(strings.TrimSpace(result))
}

// LogLevelFromLogrus parses the environment's log level string into a logrus.Level.
// If parsing fails (e.g., due to an invalid environment variable value), it will panic,
// as this indicates a critical configuration error that must be addressed.
func (*Environment) LogLevelFromLogrus(defaultLevel logrus.Level) logrus.Level {
	envLevelStr := ENV.LogLevel()
	if len(envLevelStr) == 0 {
		return defaultLevel
	}

	logrusLevel, err := logrus.ParseLevel(envLevelStr)
	if err != nil {
		panic(fmt.Sprintf("Invalid log level '%s' specified in environment variable %s: %v. Please correct this critical configuration.", envLevelStr, ENV_LOGLEVEL, err))
	}
	return logrusLevel
}

// HttpLogs checks if HTTP logging is enabled. Defaults to false.
func (*Environment) HttpLogs() bool {
	return getEnvOrDefaultBool(ENV_HTTPLOGS, false)
}

// WhatsmeowLogLevel returns the Whatsmeow Log Level. Defaults to an empty string.
func (*Environment) WhatsmeowLogLevel() string {
	return getEnvOrDefaultString(ENV_WHATSMEOWLOGLEVEL, "")
}

// WhatsmeowDBLogLevel returns the Whatsmeow Database Log Level. Defaults to an empty string.
func (*Environment) WhatsmeowDBLogLevel() string {
	return getEnvOrDefaultString(ENV_WHATSMEOWDBLOGLEVEL, "")
}

// --- RABBITMQ SETTINGS ---

// RabbitMQQueue returns the name of the RabbitMQ queue.
// Defaults to an empty string if the environment variable is not set.
func (*Environment) RabbitMQQueue() string {
	return getEnvOrDefaultString(ENV_RABBITMQ_QUEUE, "")
}

// RabbitMQConnectionString returns the connection string for RabbitMQ.
// Defaults to an empty string if the environment variable is not set.
func (*Environment) RabbitMQConnectionString() string {
	return getEnvOrDefaultString(ENV_RABBITMQ_CONNECTIONSTRING, "")
}

// CacheLength returns the maximum number of items for the cache. Defaults to 0 (no limit).
func (*Environment) RabbitMQCacheLength() uint64 {
	return getEnvOrDefaultUint64(ENV_RABBITMQ_CACHELENGTH, 0)
}
