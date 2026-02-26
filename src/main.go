package main

import (
	_ "github.com/nocodeleaks/quepasa/api"
	environment "github.com/nocodeleaks/quepasa/environment"
	_ "github.com/nocodeleaks/quepasa/form"
	library "github.com/nocodeleaks/quepasa/library"
	_ "github.com/nocodeleaks/quepasa/mcp"
	_ "github.com/nocodeleaks/quepasa/metrics"
	models "github.com/nocodeleaks/quepasa/models"
	webserver "github.com/nocodeleaks/quepasa/webserver"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	whatsmeow "github.com/nocodeleaks/quepasa/whatsmeow"

	_ "github.com/nocodeleaks/quepasa/swagger" // Swagger docs
	logrus "github.com/sirupsen/logrus"
)

// @title						QuePasa WhatsApp API
// @version					4.0.0
// @description				QuePasa is a Go-based WhatsApp bot platform that exposes HTTP APIs for WhatsApp messaging integration
// @termsOfService				https://github.com/nocodeleaks/quepasa
// @contact.name				QuePasa Support
// @contact.url				https://github.com/nocodeleaks/quepasa
// @license.name				GNU Affero General Public License v3.0
// @license.url				https://github.com/nocodeleaks/quepasa/blob/main/LICENSE.md
// @BasePath					/
// @schemes					http https
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						X-QUEPASA-TOKEN
func main() {

	loglevel := environment.Settings.General.LogLevelFromLogrus(logrus.InfoLevel)
	logrus.SetLevel(loglevel)

	logentry := library.NewLogEntry("main")
	logentry.Level = loglevel
	logentry.Infof("current log level: %v", logentry.Level)

	// checks for pending database migrations
	err := models.MigrateToLatest(logentry)
	if err != nil {
		logentry.Fatalf("database migration error: %s", err.Error())
	}

	// should became before whatsmeow start
	title := environment.Settings.General.AppTitle
	if len(title) > 0 {
		whatsapp.WhatsappWebAppSystem = title
	}

	whatsappOptions := &whatsapp.WhatsappOptionsExtended{
		Groups:            environment.Settings.WhatsApp.Groups,
		Individuals:       environment.Settings.WhatsApp.Individuals,
		Broadcasts:        environment.Settings.WhatsApp.Broadcasts,
		ReadReceipts:      environment.Settings.WhatsApp.ReadReceipts,
		Calls:             environment.Settings.WhatsApp.Calls,
		ReadUpdate:        environment.Settings.WhatsApp.ReadUpdate,
		HistorySync:       environment.Settings.WhatsApp.HistorySyncDays,
		Presence:          environment.Settings.WhatsApp.Presence,
		DispatchUnhandled: environment.Settings.Whatsmeow.DispatchUnhandled,
		LogLevel:          logentry.Level.String(),
	}

	whatsapp.Options = *whatsappOptions

	options := whatsmeow.WhatsmeowOptions{
		WhatsappOptionsExtended: whatsapp.Options,
		WMLogLevel:              environment.Settings.Whatsmeow.LogLevel,
		DBLogLevel:              environment.Settings.Whatsmeow.DBLogLevel,
		UseRetryMessageStore:    environment.Settings.Whatsmeow.UseRetryMessageStore,
	}

	dbParameters := environment.Settings.Database.GetDBParameters()
	whatsmeow.Start(options, dbParameters, logentry)

	// must execute after whatsmeow started
	for _, element := range models.Running {
		if handler, ok := models.MigrationHandlers[element]; ok {
			handler(element)
		}
	}

	// Inicializando serviço de controle do whatsapp
	// De forma assíncrona
	err = models.QPWhatsappStart(logentry)
	if err != nil {
		logentry.Fatalf("whatsapp service starting error: %s", err.Error())
	}

	err = webserver.WebServerStart(logentry)
	if err != nil {
		logentry.Info("end with errors")
	} else {
		logentry.Info("end")
	}
}
