package rabbitmq

import (
	"sync"
)

// RabbitMQQueueDefault is the default queue name used for RabbitMQ operations
// if a specific queue name is not provided.
var RabbitMQQueueDefault = "Q-QUEPASA" // Fila padrão atualizada!

// RabbitMQClientInstance is the global, singleton instance of the RabbitMQClient.
// It should be accessed via GetRabbitMQClientInstance function.
var RabbitMQClientInstance *RabbitMQClient // Public (exported) variable

// clientOnce ensures that the RabbitMQClientInstance is initialized only once.
var clientOnce sync.Once

// InitializeRabbitMQClient connects to RabbitMQ and sets up the global client instance.
// It uses environment variables for connection string and queue name.
// Errors during connection or setup are logged and will likely cause the application to panic or exit.
// This function doesn't return a value as its purpose is to initialize a global state.
func InitializeRabbitMQClient(connURI string, maxCacheSize uint64) {
	clientOnce.Do(func() {
		// Initialize the global instance using the NewRabbitMQClient constructor.
		RabbitMQClientInstance = NewRabbitMQClient(connURI, maxCacheSize)
	})
}
