package environment

import (
	"os"
	"strings"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

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

// WhatsAppSettings holds all WhatsApp configuration loaded from environment
type WhatsAppSettings struct {
	ReadUpdate      bool                             `json:"read_update"`
	ReadReceipts    whatsapp.WhatsappBooleanExtended `json:"read_receipts"`
	Calls           whatsapp.WhatsappBooleanExtended `json:"calls"`
	Groups          whatsapp.WhatsappBooleanExtended `json:"groups"`
	Broadcasts      whatsapp.WhatsappBooleanExtended `json:"broadcasts"`
	HistorySyncDays *uint32                          `json:"history_sync_days,omitempty"`
	Presence        string                           `json:"presence"`
}

// NewWhatsAppSettings creates a new WhatsApp settings by loading all values from environment
func NewWhatsAppSettings() WhatsAppSettings {
	return WhatsAppSettings{
		ReadUpdate:      getEnvOrDefaultBool(ENV_READUPDATE, false),
		ReadReceipts:    getWhatsappBooleanExtended(ENV_READRECEIPTS),
		Calls:           getWhatsappBooleanExtended(ENV_CALLS),
		Groups:          getWhatsappBooleanExtended(ENV_GROUPS),
		Broadcasts:      getWhatsappBooleanExtended(ENV_BROADCASTS),
		HistorySyncDays: getOptionalEnvUint32(ENV_HISTORYSYNCDAYS),
		Presence:        getEnvOrDefaultString(ENV_PRESENCE, "unavailable"),
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
