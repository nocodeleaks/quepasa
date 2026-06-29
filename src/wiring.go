package main

import (
	models "github.com/nocodeleaks/quepasa/models"
	"github.com/nocodeleaks/quepasa/ports"
	rabbitmq "github.com/nocodeleaks/quepasa/rabbitmq"
	runtime "github.com/nocodeleaks/quepasa/runtime"
	signalr "github.com/nocodeleaks/quepasa/signalr"
	whatsmeow "github.com/nocodeleaks/quepasa/whatsmeow"
)

// wiring.go is the composition root: it groups the startup dependency injection
// by subsystem so main() reads as a short sequence of named steps instead of one
// long block of global assignments (PLAN P1.2 / Roadmap Phase D). These remain
// global injections for now; the grouping is the first step toward constructor
// wiring.

// wireWhatsappDriver injects the whatsmeow driver behind the domain-owned ports
// interfaces, keeping `models` free of any whatsmeow import (PLAN P1.1).
// Must run after whatsmeow.Start.
func wireWhatsappDriver() {
	adapter := &whatsmeow.WhatsmeowDriverAdapter{}
	ports.GlobalWhatsappDriverFactory = adapter
	ports.GlobalWhatsappDriverService = adapter
}

// newTransportServices builds the transport adapter bundle that keeps `models`
// transport-agnostic (realtime presence, dispatch lifecycle, RabbitMQ).
func newTransportServices() models.TransportServices {
	services := models.TransportServices{
		RealtimePresenceChecker:       signalr.SignalRHub,
		DispatchingLifecyclePublisher: runtime.NewDispatchingLifecyclePublisher(),
	}
	applyRabbitMQTransport(&services)
	return services
}

// applyRabbitMQTransport fills in the RabbitMQ-specific transport hooks. Grouping
// this separately keeps the broker wiring in one place and out of main().
func applyRabbitMQTransport(services *models.TransportServices) {
	services.RabbitMQGetClient = func(connectionString string) models.RabbitMQPublisherClient {
		return rabbitmq.GetRabbitMQClient(connectionString)
	}
	services.RabbitMQCloseClient = rabbitmq.CloseRabbitMQClient
	services.RabbitMQInjectQueueBackend = func() {
		if rabbitmq.RabbitMQClientInstance != nil {
			rabbitmq.InjectCacheBackendIntoClient(rabbitmq.RabbitMQClientInstance)
		}
	}
	services.RabbitMQExchangeName = rabbitmq.QuePasaExchangeName
	services.RabbitMQRoutingKeyProd = rabbitmq.QuePasaRoutingKeyProd
	services.RabbitMQRoutingKeyHistory = rabbitmq.QuePasaRoutingKeyHistory
	services.RabbitMQRoutingKeyEvents = rabbitmq.QuePasaRoutingKeyEvents
	services.RabbitMQMessagesPublishedInc = func(queue string) {
		rabbitmq.MessagesPublished.WithLabelValues(queue).Inc()
	}
	services.RabbitMQMessagePublishErrorsInc = func() {
		rabbitmq.MessagePublishErrors.Inc()
	}
	services.RabbitMQClientResolver = func(connectionString string) bool {
		return rabbitmq.GetRabbitMQClient(connectionString) != nil
	}
}
