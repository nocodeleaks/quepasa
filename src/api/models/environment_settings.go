package api

import (
	"fmt"

	environment "github.com/nocodeleaks/quepasa/environment"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// EnvironmentSettings represents WhatsApp service environment configuration
type EnvironmentSettings struct {
	Groups         string `json:"groups"`
	Broadcasts     string `json:"broadcasts"`
	ReadReceipts   string `json:"read_receipts"`
	Calls          string `json:"calls"`
	HistorySync    string `json:"history_sync"`
	LogLevel       string `json:"log_level"`
	DBLogLevel     string `json:"db_log_level,omitempty"`
	Presence       string `json:"presence,omitempty"`
	ReadUpdate     string `json:"read_update,omitempty"`
	WakeUpHour     string `json:"wakeup_hour,omitempty"`
	WakeUpDuration string `json:"wakeup_duration,omitempty"`
}

// NewEnvironmentSettings creates environment settings from global configuration
func NewEnvironmentSettings() *EnvironmentSettings {
	settings := &EnvironmentSettings{
		Groups:       formatBooleanExtended(environment.Settings.WhatsApp.Groups),
		Broadcasts:   formatBooleanExtended(environment.Settings.WhatsApp.Broadcasts),
		ReadReceipts: formatBooleanExtended(environment.Settings.WhatsApp.ReadReceipts),
		Calls:        formatBooleanExtended(environment.Settings.WhatsApp.Calls),
		HistorySync:  formatHistorySync(environment.Settings.WhatsApp.HistorySyncDays),
		LogLevel:     formatLogLevel(environment.Settings.Whatsmeow.LogLevel),
		DBLogLevel:   formatLogLevel(environment.Settings.Whatsmeow.DBLogLevel),
		Presence:     environment.Settings.WhatsApp.Presence,
		ReadUpdate:   formatBool(environment.Settings.WhatsApp.ReadUpdate),
	}

	// Optional fields
	if environment.Settings.WhatsApp.WakeUpHour != "" {
		settings.WakeUpHour = environment.Settings.WhatsApp.WakeUpHour
	}
	if environment.Settings.WhatsApp.WakeUpDuration > 0 {
		settings.WakeUpDuration = fmt.Sprintf("%d seconds", environment.Settings.WhatsApp.WakeUpDuration)
	}

	return settings
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
