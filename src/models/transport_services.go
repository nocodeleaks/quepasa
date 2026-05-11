package models

import "sync"

var transportServicesMu sync.RWMutex

// TransportServices groups startup-time transport hooks so bootstrap wiring can
// be applied in one place while models remain transport-agnostic.
type TransportServices struct {
	RealtimePresenceChecker         RealtimePresenceChecker
	DispatchingLifecyclePublisher   DispatchingLifecyclePublisher
	RabbitMQGetClient               func(connectionString string) RabbitMQPublisherClient
	RabbitMQCloseClient             func(connectionString string)
	RabbitMQInjectQueueBackend      func()
	RabbitMQExchangeName            string
	RabbitMQRoutingKeyProd          string
	RabbitMQRoutingKeyHistory       string
	RabbitMQRoutingKeyEvents        string
	RabbitMQMessagesPublishedInc    func()
	RabbitMQMessagePublishErrorsInc func()
	RabbitMQClientResolver          func(connectionString string) bool
}

// ApplyTransportServices updates the global transport adapters currently used by
// models. It is intentionally additive so bootstrap can migrate incrementally.
func ApplyTransportServices(services TransportServices) {
	transportServicesMu.Lock()
	defer transportServicesMu.Unlock()

	if services.RealtimePresenceChecker != nil {
		GlobalRealtimePresenceChecker = services.RealtimePresenceChecker
	}
	if services.DispatchingLifecyclePublisher != nil {
		GlobalDispatchingLifecyclePublisher = services.DispatchingLifecyclePublisher
	}
	if services.RabbitMQGetClient != nil {
		GlobalRabbitMQGetClient = services.RabbitMQGetClient
	}
	if services.RabbitMQCloseClient != nil {
		GlobalRabbitMQCloseClient = services.RabbitMQCloseClient
	}
	if services.RabbitMQInjectQueueBackend != nil {
		GlobalRabbitMQInjectQueueBackend = services.RabbitMQInjectQueueBackend
	}
	if services.RabbitMQExchangeName != "" {
		GlobalRabbitMQExchangeName = services.RabbitMQExchangeName
	}
	if services.RabbitMQRoutingKeyProd != "" {
		GlobalRabbitMQRoutingKeyProd = services.RabbitMQRoutingKeyProd
	}
	if services.RabbitMQRoutingKeyHistory != "" {
		GlobalRabbitMQRoutingKeyHistory = services.RabbitMQRoutingKeyHistory
	}
	if services.RabbitMQRoutingKeyEvents != "" {
		GlobalRabbitMQRoutingKeyEvents = services.RabbitMQRoutingKeyEvents
	}
	if services.RabbitMQMessagesPublishedInc != nil {
		GlobalRabbitMQMessagesPublishedInc = services.RabbitMQMessagesPublishedInc
	}
	if services.RabbitMQMessagePublishErrorsInc != nil {
		GlobalRabbitMQMessagePublishErrorsInc = services.RabbitMQMessagePublishErrorsInc
	}
	if services.RabbitMQClientResolver != nil {
		GlobalRabbitMQClientResolver = services.RabbitMQClientResolver
	}
}
