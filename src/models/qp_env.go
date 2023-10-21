package models

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

const (
	WEBSOCKETSSL        = "WEBSOCKETSSL"
	ENVIRONMENT         = "APP_ENV"
	MIGRATIONS          = "MIGRATIONS"
	TITLE               = "APP_TITLE"
	DEBUG_REQUESTS      = "DEBUGREQUESTS"
	DEBUG_JSON_MESSAGES = "DEBUGJSONMESSAGES"
	REMOVEDIGIT9        = "REMOVEDIGIT9"
	READRECEIPTS        = "READRECEIPTS"
	SYNOPSISLENGTH      = "SYNOPSISLENGTH"
	WHATSMEOWLOGLEVEL   = "WHATSMEOWLOGLEVEL"
)

type Environment struct{}

var ENV Environment

func (_ *Environment) ShouldConvertWaveToOgg() bool {
	environment, _ := GetEnvBool("CONVERT_WAVE_TO_OGG", true)
	return environment
}

func (_ *Environment) IsDevelopment() bool {
	environment, _ := GetEnvStr(ENVIRONMENT)
	if strings.ToLower(environment) == "development" {
		return true
	} else {
		return false
	}
}

// WEBSOCKETSSL => default false
func (_ *Environment) UseSSLForWebSocket() bool {
	migrations, _ := GetEnvStr(WEBSOCKETSSL)
	boolMigrations, err := strconv.ParseBool(migrations)
	if err == nil {
		return boolMigrations
	} else {
		return false
	}
}

// MIGRATIONS => Path to database migrations folder
func (_ *Environment) Migrate() bool {
	migrations, _ := GetEnvStr(MIGRATIONS)
	boolMigrations, err := strconv.ParseBool(migrations)
	if err == nil {
		return boolMigrations
	} else {
		return true
	}
}

// MIGRATIONS => Path to database migrations folder
func (_ *Environment) MigrationPath() string {
	migrations, _ := GetEnvStr(MIGRATIONS)
	_, err := strconv.ParseBool(migrations)
	if err != nil {
		return migrations
	} else {
		return "" // indicates that should use default path
	}
}

// Force Default Whatsmeow Log Level
func (_ *Environment) WhatsmeowLogLevel() string {
	result, _ := GetEnvStr(WHATSMEOWLOGLEVEL)
	return result
}

func (_ *Environment) AppTitle() string {
	result, _ := GetEnvStr(TITLE)
	return result
}

func (_ *Environment) DEBUGRequests() bool {

	if ENV.IsDevelopment() {
		environment, err := GetEnvBool(DEBUG_REQUESTS, true)
		if err == nil {
			return environment
		}
	}

	return false
}

func (_ *Environment) DEBUGJsonMessages() bool {

	if ENV.IsDevelopment() {
		environment, err := GetEnvBool(DEBUG_JSON_MESSAGES, true)
		if err == nil {
			return environment
		}
	}

	return false
}

var ErrEnvVarEmpty = errors.New("getenv: environment variable empty")

func GetEnvBool(key string, value bool) (bool, error) {
	result := value
	s, err := GetEnvStr(key)
	if err == nil {
		trying, err := strconv.ParseBool(s)
		if err == nil {
			result = trying
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

func (_ *Environment) ShouldRemoveDigit9() bool {
	value, _ := GetEnvBool(REMOVEDIGIT9, false)
	return value
}

func (_ *Environment) ShouldReadReceipts() bool {
	value, _ := GetEnvBool(READRECEIPTS, false)
	return value
}

// MIGRATIONS => Path to database migrations folder
func (_ *Environment) SynopsisLength() uint64 {
	stringValue, err := GetEnvStr(SYNOPSISLENGTH)
	if err == nil {
		value, err := strconv.ParseUint(stringValue, 10, 32)
		if err == nil {
			return value
		}
	}

	return 50
}
