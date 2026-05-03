package models

// RabbitMQPublisherClient is the minimal contract required by models when
// publishing to RabbitMQ transport.
type RabbitMQPublisherClient interface {
	EnsureExchangeAndQueues() error
	PublishQuePasaMessage(routingKey string, messageContent any)
}

// GlobalRabbitMQGetClient resolves a RabbitMQ client for a connection string.
var GlobalRabbitMQGetClient = func(connectionString string) RabbitMQPublisherClient {
	return nil
}

func GetRabbitMQPublisherClient(connectionString string) RabbitMQPublisherClient {
	transportServicesMu.RLock()
	resolver := GlobalRabbitMQGetClient
	transportServicesMu.RUnlock()
	return resolver(connectionString)
}

// GlobalRabbitMQCloseClient closes a RabbitMQ client identified by connection string.
var GlobalRabbitMQCloseClient = func(connectionString string) {}

func CloseRabbitMQPublisherClient(connectionString string) {
	transportServicesMu.RLock()
	closer := GlobalRabbitMQCloseClient
	transportServicesMu.RUnlock()
	closer(connectionString)
}

// GlobalRabbitMQInjectQueueBackend allows bootstrap wiring to inject cache backend
// into transport-level RabbitMQ clients after cache service initialization.
var GlobalRabbitMQInjectQueueBackend = func() {}

func InjectRabbitMQQueueBackend() {
	transportServicesMu.RLock()
	injector := GlobalRabbitMQInjectQueueBackend
	transportServicesMu.RUnlock()
	injector()
}

// QuePasa RabbitMQ naming/routing defaults can be overridden by bootstrap wiring.
var GlobalRabbitMQExchangeName = "quepasa.exchange"
var GlobalRabbitMQRoutingKeyProd = "prod"
var GlobalRabbitMQRoutingKeyHistory = "history"
var GlobalRabbitMQRoutingKeyEvents = "events"

// Metric hooks used by models without importing rabbitmq metrics directly.
var GlobalRabbitMQMessagesPublishedInc = func() {}
var GlobalRabbitMQMessagePublishErrorsInc = func() {}

func IncrementRabbitMQMessagesPublished() {
	transportServicesMu.RLock()
	inc := GlobalRabbitMQMessagesPublishedInc
	transportServicesMu.RUnlock()
	inc()
}

func IncrementRabbitMQMessagePublishErrors() {
	transportServicesMu.RLock()
	inc := GlobalRabbitMQMessagePublishErrorsInc
	transportServicesMu.RUnlock()
	inc()
}

func GetRabbitMQExchangeName() string {
	transportServicesMu.RLock()
	name := GlobalRabbitMQExchangeName
	transportServicesMu.RUnlock()
	return name
}

func GetRabbitMQRoutingKeyProd() string {
	transportServicesMu.RLock()
	key := GlobalRabbitMQRoutingKeyProd
	transportServicesMu.RUnlock()
	return key
}

func GetRabbitMQRoutingKeyHistory() string {
	transportServicesMu.RLock()
	key := GlobalRabbitMQRoutingKeyHistory
	transportServicesMu.RUnlock()
	return key
}

func GetRabbitMQRoutingKeyEvents() string {
	transportServicesMu.RLock()
	key := GlobalRabbitMQRoutingKeyEvents
	transportServicesMu.RUnlock()
	return key
}
