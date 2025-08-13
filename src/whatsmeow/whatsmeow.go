package whatsmeow

import (
	environment "github.com/nocodeleaks/quepasa/environment"
	library "github.com/nocodeleaks/quepasa/library"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	logrus "github.com/sirupsen/logrus"
)

// Variáveis de pré-configuração carregadas automaticamente
var (
	preConfiguredOptions  *WhatsmeowOptions
	preConfiguredLogger   *logrus.Entry
	preConfiguredDBParams *library.DatabaseParameters
	isPreConfigured       bool
)

// init() carrega automaticamente as configurações do environment
func init() {

	env := environment.Settings.General
	envWA := environment.Settings.WhatsApp
	envWM := environment.Settings.Whatsmeow

	// Criar logger pré-configurado primeiro
	logentry := logrus.WithField("package", "whatsmeow")
	if logLevel, err := logrus.ParseLevel(env.LogLevel); err == nil {
		logentry.Level = logLevel
	} else {
		logentry.Level = logrus.InfoLevel
	}

	// Criar WhatsApp options extended a partir do environment
	settings := whatsapp.WhatsappOptionsExtended{
		Groups:            envWA.Groups,
		Broadcasts:        envWA.Broadcasts,
		ReadReceipts:      envWA.ReadReceipts,
		Calls:             envWA.Calls,
		ReadUpdate:        envWA.ReadUpdate,
		HistorySync:       envWA.HistorySyncDays,
		Presence:          envWA.Presence,
		DispatchUnhandled: envWM.DispatchUnhandled,
		LogLevel:          logentry.Level.String(),
	}

	// Configurar globalmente no whatsapp package
	whatsapp.Options = settings

	// Criar WhatsmeowOptions a partir das configurações carregadas
	preConfiguredOptions = &WhatsmeowOptions{
		WhatsappOptionsExtended: whatsapp.Options,
		WMLogLevel:              envWM.LogLevel,
		DBLogLevel:              envWM.DBLogLevel,
	}

	// Carregar parâmetros de banco de dados
	dbParams := environment.Settings.Database.GetDBParameters()
	preConfiguredDBParams = &dbParams

	preConfiguredLogger = logentry
	isPreConfigured = true

	logentry.Debug("🔧 WhatsmeowService pré-configurado via environment settings")
}
