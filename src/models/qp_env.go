package models

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	library "github.com/nocodeleaks/quepasa/library"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	"google.golang.org/protobuf/proto"
)

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
	ENV_MASTER_KEY     = "MASTERKEY"

	ENV_WEBSOCKETSSL             = "WEBSOCKETSSL" // use ssl for websocket qrcode
	ENV_MIGRATIONS               = "MIGRATIONS"   // enable migrations
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

	ENV_TESTING = "TESTING"
)

type Environment struct{}

var ENV Environment

//#region DATABASE CONFIG

func (*Environment) GetDBParameters() library.DatabaseParameters {
	parameters := library.DatabaseParameters{}

	parameters.Driver = os.Getenv(ENV_DBDRIVER)
	if len(parameters.Driver) == 0 {
		parameters.Driver = "sqlite3"
	}

	parameters.Host = os.Getenv(ENV_DBHOST)
	parameters.DataBase = os.Getenv(ENV_DBDATABASE)
	parameters.Port = os.Getenv(ENV_DBPORT)
	parameters.User = os.Getenv(ENV_DBUSER)
	parameters.Password = os.Getenv(ENV_DBPASSWORD)
	parameters.SSL = os.Getenv(ENV_DBSSLMODE)
	return parameters
}

//#endregion

func (*Environment) UseCompatibleMIMEsAsAudio() bool {
	environment, err := GetEnvBool(ENV_CONVERT_WAVE_TO_OGG, proto.Bool(true))
	if err != nil {
		return *environment
	}

	environment, _ = GetEnvBool(ENV_COMPATIBLE_MIME_AS_AUDIO, proto.Bool(true))
	return *environment
}

// WEBSOCKETSSL => default false
func (*Environment) UseSSLForWebSocket() bool {
	migrations, _ := GetEnvStr(ENV_WEBSOCKETSSL)
	boolMigrations, err := strconv.ParseBool(migrations)
	if err == nil {
		return boolMigrations
	} else {
		return false
	}
}

// MIGRATIONS => Path to database migrations folder
func (*Environment) Migrate() bool {
	migrations, _ := GetEnvStr(ENV_MIGRATIONS)
	boolMigrations, err := strconv.ParseBool(migrations)
	if err == nil {
		return boolMigrations
	} else {
		return true
	}
}

// MIGRATIONS => Path to database migrations folder
func (*Environment) MigrationPath() string {
	migrations, _ := GetEnvStr(ENV_MIGRATIONS)
	_, err := strconv.ParseBool(migrations)
	if err != nil {
		return migrations
	} else {
		return "" // indicates that should use default path
	}
}

func (*Environment) AppTitle() string {
	result, _ := GetEnvStr(ENV_TITLE)
	return result
}

var ErrEnvVarEmpty = errors.New("getenv: environment variable empty")

func GetEnvBool(key string, value *bool) (*bool, error) {
	result := value
	s, err := GetEnvStr(key)
	if err == nil {
		trying, err := strconv.ParseBool(s)
		if err == nil {
			result = &trying
		}
	}
	return result, err
}

func GetEnvStr(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return v, ErrEnvVarEmpty
	}
	return v, nil
}

func (*Environment) ShouldRemoveDigit9() bool {
	value, _ := GetEnvBool(ENV_REMOVEDIGIT9, proto.Bool(false))
	return *value
}

//#region WHATSAPP SERVICE OPTIONS - WHATSMEOW

func ParseWhatsappBoolean(value string) whatsapp.WhatsappBooleanExtended {

	formatted := strings.TrimSpace(value)
	formatted = strings.Trim(formatted, `"`)
	formatted = strings.ToLower(formatted)

	switch strings.ToLower(formatted) {
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

func (*Environment) Broadcasts() whatsapp.WhatsappBooleanExtended {
	v := os.Getenv(ENV_BROADCASTS)
	return ParseWhatsappBoolean(v)
}

func (*Environment) Groups() whatsapp.WhatsappBooleanExtended {
	v := os.Getenv(ENV_GROUPS)
	return ParseWhatsappBoolean(v)
}

func (*Environment) ReadReceipts() whatsapp.WhatsappBooleanExtended {
	v := os.Getenv(ENV_READRECEIPTS)
	return ParseWhatsappBoolean(v)
}

func (*Environment) Calls() whatsapp.WhatsappBooleanExtended {
	v := os.Getenv(ENV_CALLS)
	return ParseWhatsappBoolean(v)
}

func (*Environment) ReadUpdate() bool {
	value, _ := GetEnvBool(ENV_READUPDATE, proto.Bool(false))
	return *value
}

//#region LOGS

// forces default presence status (lower)
func (*Environment) Presence() string {
	result, _ := GetEnvStr(ENV_PRESENCE)
	result = strings.ToLower(result)
	return result
}

// Force Default Log Level (lower)(trimmed)
func (*Environment) LogLevel() string {
	result, _ := GetEnvStr(ENV_LOGLEVEL)
	result = strings.ToLower(result)   // to lower
	result = strings.TrimSpace(result) // trim white spaces
	return result
}

func (*Environment) HttpLogs() bool {
	value, _ := GetEnvBool(ENV_HTTPLOGS, proto.Bool(false))
	return *value
}

// Force Default Whatsmeow Log Level
func (*Environment) WhatsmeowLogLevel() string {
	result, _ := GetEnvStr(ENV_WHATSMEOWLOGLEVEL)
	return result
}

// Force Default Whatsmeow DataBase Log Level
func (*Environment) WhatsmeowDBLogLevel() string {
	result, _ := GetEnvStr(ENV_WHATSMEOWDBLOGLEVEL)
	return result
}

//#endregion

// Get history sync days, environment whatsapp service global option
func (*Environment) HistorySync() *uint32 {
	stringValue, err := GetEnvStr(ENV_HISTORYSYNCDAYS)
	if err == nil {
		value, err := strconv.ParseUint(stringValue, 10, 32)
		if err == nil {
			return proto.Uint32(uint32(value))
		}
	}

	return nil
}

//#endregion

// Length for synopsis when replied messages
func (*Environment) SynopsisLength() uint64 {
	stringValue, err := GetEnvStr(ENV_SYNOPSISLENGTH)
	if err == nil {
		value, err := strconv.ParseUint(stringValue, 10, 32)
		if err == nil {
			return value
		}
	}

	return 50
}

// Length for cached messages, auto-cleaner old ones if bigger than 0
func (*Environment) CacheLength() uint64 {
	stringValue, err := GetEnvStr(ENV_CACHELENGTH)
	if err == nil {
		value, err := strconv.ParseUint(stringValue, 10, 32)
		if err == nil {
			return value
		}
	}

	return 0
}

// Days for cached messages, auto-cleaner old ones if bigger than 0
func (*Environment) CacheDays() uint64 {
	stringValue, err := GetEnvStr(ENV_CACHEDAYS)
	if err == nil {
		value, err := strconv.ParseUint(stringValue, 10, 32)
		if err == nil {
			return value
		}
	}

	return 0
}

// Master Key for super admin methods
func (*Environment) MasterKey() string {
	result, _ := GetEnvStr(ENV_MASTER_KEY)
	return result
}

// Testing => Apply Testing Methods (not stable)
func (*Environment) Testing() bool {
	text, _ := GetEnvStr(ENV_TESTING)
	value, err := strconv.ParseBool(text)
	if err == nil {
		return value
	} else {
		return false // default return
	}
}
