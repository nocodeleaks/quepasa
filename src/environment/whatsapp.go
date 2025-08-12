package environment

import (
	"os"
	"strconv"
	"strings"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// WhatsAppEnvironment handles all WhatsApp-related environment variables
type WhatsAppEnvironment struct{}

// WhatsApp environment variable names
const (
	ENV_READUPDATE      = "READUPDATE"      // mark chat read when send any msg
	ENV_READRECEIPTS    = "READRECEIPTS"    // trigger webhooks for read receipts events
	ENV_CALLS           = "CALLS"           // defines if will be accepted calls
	ENV_GROUPS          = "GROUPS"          // handle groups
	ENV_BROADCASTS      = "BROADCASTS"      // handle broadcasts
	ENV_HISTORYSYNCDAYS = "HISTORYSYNCDAYS" // history sync days
	ENV_PRESENCE        = "PRESENCE"        // presence state
)

// ReadUpdate checks if read updates are enabled.
// Returns true or false, defaulting to false if the environment variable is not set or invalid.
func (env *WhatsAppEnvironment) ReadUpdate() bool {
	return getEnvOrDefaultBool(ENV_READUPDATE, false)
}

// ReadReceipts returns the WhatsappBooleanExtended setting for read receipts.
func (env *WhatsAppEnvironment) ReadReceipts() whatsapp.WhatsappBooleanExtended {
	value := getEnvOrDefaultString(ENV_READRECEIPTS, "false")
	return parseWhatsappBoolean(value)
}

// Calls returns the WhatsappBooleanExtended setting for calls.
func (env *WhatsAppEnvironment) Calls() whatsapp.WhatsappBooleanExtended {
	value := getEnvOrDefaultString(ENV_CALLS, "false")
	return parseWhatsappBoolean(value)
}

// Groups returns the WhatsappBooleanExtended setting for groups.
func (env *WhatsAppEnvironment) Groups() whatsapp.WhatsappBooleanExtended {
	value := getEnvOrDefaultString(ENV_GROUPS, "false")
	return parseWhatsappBoolean(value)
}

// Broadcasts returns the WhatsappBooleanExtended setting for broadcasts.
func (env *WhatsAppEnvironment) Broadcasts() whatsapp.WhatsappBooleanExtended {
	value := getEnvOrDefaultString(ENV_BROADCASTS, "false")
	return parseWhatsappBoolean(value)
}

// HistorySync returns the history sync days. Returns nil if not set or invalid.
// A nil return indicates that the system should use its internal default logic
// for history sync days, rather than a forced value.
func (env *WhatsAppEnvironment) HistorySync() *uint32 {
	return getOptionalEnvUint32(ENV_HISTORYSYNCDAYS)
}

// Presence returns the presence state. Defaults to "unavailable".
func (env *WhatsAppEnvironment) Presence() string {
	return getEnvOrDefaultString(ENV_PRESENCE, "unavailable")
}

// getOptionalEnvUint32 fetches an unsigned 32-bit integer environment variable where nil indicates "use system default logic".
func getOptionalEnvUint32(key string) *uint32 {
	if valueStr, ok := os.LookupEnv(key); ok {
		trimmedValueStr := strings.TrimSpace(valueStr)
		if trimmedValueStr == "" {
			return nil // Empty string means "use default logic"
		}
		if parsedValue, err := strconv.ParseUint(trimmedValueStr, 10, 32); err == nil {
			result := uint32(parsedValue)
			return &result
		}
	}
	return nil // Not set or invalid means "use default logic"
}

// parseWhatsappBoolean parses a string value into a WhatsappBooleanExtended type.
// It handles various string representations of boolean values, including extended ones.
func parseWhatsappBoolean(value string) whatsapp.WhatsappBooleanExtended {
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
		// Return UnSet for unknown values instead of panicking
		return whatsapp.WhatsappBooleanExtended(whatsapp.UnSetBooleanType)
	}
}
