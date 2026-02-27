package environment

import (
	"os"
	"strings"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// WhatsApp environment variable names
const (
	ENV_READUPDATE      = "READUPDATE"      // mark chat read when send any msg
	ENV_READRECEIPTS    = "READRECEIPTS"    // trigger dispatch methods for read receipts events
	ENV_CALLS           = "CALLS"           // defines if will be accepted calls
	ENV_GROUPS          = "GROUPS"          // handle groups
	ENV_DIRECT          = "DIRECT"          // handle direct chats (@s.whatsapp.net and @lid)
	ENV_DIRECT_LEGACY   = "INDIVIDUALS"     // deprecated legacy alias for direct chats
	ENV_BROADCASTS      = "BROADCASTS"      // handle broadcasts
	ENV_HISTORYSYNCDAYS = "HISTORYSYNCDAYS" // history sync days
	ENV_PRESENCE        = "PRESENCE"        // presence state
	ENV_WAKEUP_HOUR     = "WAKEUP_HOUR"     // scheduled hour(s) to activate presence (0-23, can be comma-separated for multiple hours)
	ENV_WAKEUP_DURATION = "WAKEUP_DURATION" // duration in seconds to keep presence online during wake up (default: 10)
)

// WhatsAppSettings holds all WhatsApp configuration loaded from environment
type WhatsAppSettings struct {
	ReadUpdate      whatsapp.WhatsappBooleanExtended `json:"read_update"`
	ReadReceipts    whatsapp.WhatsappBooleanExtended `json:"read_receipts"`
	Calls           whatsapp.WhatsappBooleanExtended `json:"calls"`
	Groups          whatsapp.WhatsappBooleanExtended `json:"groups"`
	Direct          whatsapp.WhatsappBooleanExtended `json:"direct"`
	Broadcasts      whatsapp.WhatsappBooleanExtended `json:"broadcasts"`
	HistorySyncDays *uint32                          `json:"history_sync_days"`
	Presence        string                           `json:"presence"`
	WakeUpHour      string                           `json:"wakeup_hour"`     // Hour(s) as integers: 0-23 or 0,8,16 for multiple hours
	WakeUpDuration  int                              `json:"wakeup_duration"` // duration in seconds
}

// NewWhatsAppSettings creates a new WhatsApp settings by loading all values from environment
func NewWhatsAppSettings() WhatsAppSettings {
	return WhatsAppSettings{
		ReadUpdate:      getWhatsappBooleanExtended(ENV_READUPDATE),
		ReadReceipts:    getWhatsappBooleanExtended(ENV_READRECEIPTS),
		Calls:           getWhatsappBooleanExtended(ENV_CALLS),
		Groups:          getWhatsappBooleanExtended(ENV_GROUPS),
		Direct:          getWhatsappBooleanExtendedWithFallback(ENV_DIRECT, ENV_DIRECT_LEGACY),
		Broadcasts:      getWhatsappBooleanExtended(ENV_BROADCASTS),
		HistorySyncDays: getOptionalEnvUint32(ENV_HISTORYSYNCDAYS),
		Presence:        getEnvOrDefaultString(ENV_PRESENCE, "unavailable"),
		WakeUpHour:      getEnvOrDefaultString(ENV_WAKEUP_HOUR, ""),
		WakeUpDuration:  getEnvOrDefaultInt(ENV_WAKEUP_DURATION, 10),
	}
}

// Helper function to convert environment variables to WhatsappBooleanExtended
func getWhatsappBooleanExtended(key string) whatsapp.WhatsappBooleanExtended {
	if valueStr, ok := os.LookupEnv(key); ok {
		formatted := strings.TrimSpace(valueStr)
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
		}
	}
	return whatsapp.WhatsappBooleanExtended(whatsapp.UnSetBooleanType)
}

// Reads a primary environment key and falls back to a deprecated alias when primary is unset.
func getWhatsappBooleanExtendedWithFallback(primary, legacy string) whatsapp.WhatsappBooleanExtended {
	if _, ok := os.LookupEnv(primary); ok {
		return getWhatsappBooleanExtended(primary)
	}
	return getWhatsappBooleanExtended(legacy)
}
