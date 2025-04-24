package main

import (
	"github.com/joho/godotenv"
	controllers "github.com/nocodeleaks/quepasa/controllers"
	"github.com/nocodeleaks/quepasa/library"
	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	whatsmeow "github.com/nocodeleaks/quepasa/whatsmeow"

	logrus "github.com/sirupsen/logrus"
)

// @title chi-swagger example APIs
// @version 1.0
// @description chi-swagger example APIs
// @BasePath /
func main() {

	// loading environment variables from .env file
	godotenv.Load()

	loglevel := models.ENV.LogLevelFromLogrus(logrus.InfoLevel)
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
	title := models.ENV.AppTitle()
	if len(title) > 0 {
		whatsapp.WhatsappWebAppSystem = title
	}

	whatsappOptions := &whatsapp.WhatsappOptionsExtended{
		Groups:       models.ENV.Groups(),
		Broadcasts:   models.ENV.Broadcasts(),
		ReadReceipts: models.ENV.ReadReceipts(),
		Calls:        models.ENV.Calls(),
		ReadUpdate:   models.ENV.ReadUpdate(),
		HistorySync:  models.ENV.HistorySync(),
		Presence:     models.ENV.Presence(),
		LogLevel:     logentry.Level.String(),
	}

	whatsapp.Options = *whatsappOptions

	options := whatsmeow.WhatsmeowOptions{
		WhatsappOptionsExtended: whatsapp.Options,
		WMLogLevel:              models.ENV.WhatsmeowLogLevel(),
		DBLogLevel:              models.ENV.WhatsmeowDBLogLevel(),
	}

	dbParameters := models.ENV.GetDBParameters()
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

	err = controllers.QPWebServerStart(logentry)
	if err != nil {
		logentry.Info("end with errors")
	} else {
		logentry.Info("end")
	}
}
