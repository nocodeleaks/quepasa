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

// GlobalRabbitMQCloseClient closes a RabbitMQ client identified by connection string.
var GlobalRabbitMQCloseClient = func(connectionString string) {}

// GlobalRabbitMQInjectQueueBackend allows bootstrap wiring to inject cache backend
// into transport-level RabbitMQ clients after cache service initialization.
var GlobalRabbitMQInjectQueueBackend = func() {}

// QuePasa RabbitMQ naming/routing defaults can be overridden by bootstrap wiring.
var GlobalRabbitMQExchangeName = "quepasa.exchange"
var GlobalRabbitMQRoutingKeyProd = "prod"
var GlobalRabbitMQRoutingKeyHistory = "history"
var GlobalRabbitMQRoutingKeyEvents = "events"

// Metric hooks used by models without importing rabbitmq metrics directly.
var GlobalRabbitMQMessagesPublishedInc = func() {}
var GlobalRabbitMQMessagePublishErrorsInc = func() {}
