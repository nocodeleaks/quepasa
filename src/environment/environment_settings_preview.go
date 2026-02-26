package environment

import (
	"fmt"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// EnvironmentSettingsPreview provides a read-only preview of environment settings
// Used for public endpoints without authentication
type EnvironmentSettingsPreview struct {
	Groups            string `json:"groups"`
	Individuals       string `json:"individuals"`
	Broadcasts        string `json:"broadcasts"`
	ReadReceipts      string `json:"read_receipts"`
	Calls             string `json:"calls"`
	HistorySync       string `json:"history_sync"`
	LogLevel          string `json:"log_level"`
	DBLogLevel        string `json:"db_log_level,omitempty"`
	RetryMessageStore string `json:"retry_message_store,omitempty"`
	Presence          string `json:"presence,omitempty"`
	ReadUpdate        string `json:"read_update,omitempty"`
	WakeUpHour        string `json:"wakeup_hour,omitempty"`
	WakeUpDuration    string `json:"wakeup_duration,omitempty"`
}

// GetPreview returns a read-only preview of current environment settings
func GetPreview() *EnvironmentSettingsPreview {
	preview := &EnvironmentSettingsPreview{
		Groups:            formatBooleanExtended(Settings.WhatsApp.Groups),
		Individuals:       formatBooleanExtended(Settings.WhatsApp.Individuals),
		Broadcasts:        formatBooleanExtended(Settings.WhatsApp.Broadcasts),
		ReadReceipts:      formatBooleanExtended(Settings.WhatsApp.ReadReceipts),
		Calls:             formatBooleanExtended(Settings.WhatsApp.Calls),
		HistorySync:       formatHistorySync(Settings.WhatsApp.HistorySyncDays),
		LogLevel:          formatLogLevel(Settings.Whatsmeow.LogLevel),
		DBLogLevel:        formatLogLevel(Settings.Whatsmeow.DBLogLevel),
		RetryMessageStore: formatBool(Settings.Whatsmeow.UseRetryMessageStore),
		Presence:          Settings.WhatsApp.Presence,
		ReadUpdate:        formatBooleanExtended(Settings.WhatsApp.ReadUpdate),
	}

	// Optional fields
	if Settings.WhatsApp.WakeUpHour != "" {
		preview.WakeUpHour = Settings.WhatsApp.WakeUpHour
	}
	if Settings.WhatsApp.WakeUpDuration > 0 {
		preview.WakeUpDuration = fmt.Sprintf("%d seconds", Settings.WhatsApp.WakeUpDuration)
	}

	return preview
}

// formatBooleanExtended converts WhatsappBooleanExtended to human-readable string
func formatBooleanExtended(value whatsapp.WhatsappBooleanExtended) string {
	switch value {
	case whatsapp.WhatsappBooleanExtended(whatsapp.TrueBooleanType):
		return "true"
	case whatsapp.WhatsappBooleanExtended(whatsapp.FalseBooleanType):
		return "false"
	case whatsapp.ForcedTrueBooleanType:
		return "forced true"
	case whatsapp.ForcedFalseBooleanType:
		return "forced false"
	case whatsapp.WhatsappBooleanExtended(whatsapp.UnSetBooleanType):
		return "unset"
	default:
		return "unknown"
	}
}

// formatHistorySync converts history sync days to human-readable string
func formatHistorySync(days *uint32) string {
	if days == nil {
		return "disabled"
	}
	if *days == 0 {
		return "disabled"
	}
	return fmt.Sprintf("%d days", *days)
}

// formatLogLevel converts log level to human-readable string
func formatLogLevel(level string) string {
	if level == "" {
		return "default"
	}
	return level
}

// formatBool converts boolean to human-readable string
func formatBool(value bool) string {
	if value {
		return "true"
	}
	return "false"
}
