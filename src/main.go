package main

import (
	"github.com/joho/godotenv"
	library "github.com/nocodeleaks/quepasa/library"
	models "github.com/nocodeleaks/quepasa/models"
	rabbitmq "github.com/nocodeleaks/quepasa/rabbitmq"
	webserver "github.com/nocodeleaks/quepasa/webserver"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	whatsmeow "github.com/nocodeleaks/quepasa/whatsmeow"

	_ "github.com/nocodeleaks/quepasa/docs" // Swagger docs
	logrus "github.com/sirupsen/logrus"
)

// @title QuePasa WhatsApp API
// @version 4.0.0
// @description QuePasa is a Go-based WhatsApp bot platform that exposes HTTP APIs for WhatsApp messaging integration
// @termsOfService https://github.com/nocodeleaks/quepasa
// @contact.name QuePasa Support
// @contact.url https://github.com/nocodeleaks/quepasa
// @license.name GNU Affero General Public License v3.0
// @license.url https://github.com/nocodeleaks/quepasa/blob/main/LICENSE.md
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-QUEPASA-TOKEN
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
		Groups:            models.ENV.Groups(),
		Broadcasts:        models.ENV.Broadcasts(),
		ReadReceipts:      models.ENV.ReadReceipts(),
		Calls:             models.ENV.Calls(),
		ReadUpdate:        models.ENV.ReadUpdate(),
		HistorySync:       models.ENV.HistorySync(),
		Presence:          models.ENV.Presence(),
		DispatchUnhandled: models.ENV.DispatchUnhandled(),
		LogLevel:          logentry.Level.String(),
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

	rabbitmq_connection_string := models.ENV.RabbitMQConnectionString()
	if len(rabbitmq_connection_string) > 0 {
		rabbitmq_queue := models.ENV.RabbitMQQueue()
		if len(rabbitmq_queue) > 0 {
			rabbitmq.RabbitMQQueueDefault = rabbitmq_queue
		}

		cachelength := models.ENV.RabbitMQCacheLength()
		rabbitmq.InitializeRabbitMQClient(rabbitmq_connection_string, cachelength)
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
