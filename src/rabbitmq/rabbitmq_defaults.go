package rabbitmq

import (
	"sync"
)

// QuePasa RabbitMQ Fixed Configuration
// All bots use the same Exchange and Queue names
const (
	QuePasaExchangeName      = "quepasa.exchange"
	QuePasaQueueProd         = "quepasa.production"
	QuePasaQueueHistory      = "quepasa.history"
	QuePasaQueueEvents       = "quepasa.other"
	QuePasaRoutingKeyProd    = "prod"
	QuePasaRoutingKeyHistory = "history"
	QuePasaRoutingKeyEvents  = "events"
)

// RabbitMQQueueDefault is the default queue name used for RabbitMQ operations
// if a specific queue name is not provided.
var RabbitMQQueueDefault = "Q-QUEPASA" // Fila padrão atualizada!

// RabbitMQClientInstance is the global, singleton instance of the RabbitMQClient.
// It should be accessed via GetRabbitMQClientInstance function.
var RabbitMQClientInstance *RabbitMQClient // Public (exported) variable

// clientOnce ensures that the RabbitMQClientInstance is initialized only once.
var clientOnce sync.Once

// Connection manager for multiple RabbitMQ connections
var (
	clientManager = make(map[string]*RabbitMQClient)
	clientMutex   sync.RWMutex
)

// GetRabbitMQClient returns or creates a RabbitMQ client for the specified connection string
func GetRabbitMQClient(connectionString string) *RabbitMQClient {
	if connectionString == "" {
		return nil
	}

	clientMutex.RLock()
	client, exists := clientManager[connectionString]
	clientMutex.RUnlock()

	if exists {
		return client
	}

	// Create new client if it doesn't exist
	clientMutex.Lock()
	defer clientMutex.Unlock()

	// Double-check in case another goroutine created it while we were waiting for the lock
	if client, exists := clientManager[connectionString]; exists {
		return client
	}

	// Create new client with default cache size
	client = NewRabbitMQClient(connectionString, 0) // 0 means unlimited cache
	clientManager[connectionString] = client

	return client
}

// CloseRabbitMQClient closes and removes a specific RabbitMQ client
func CloseRabbitMQClient(connectionString string) {
	clientMutex.Lock()
	defer clientMutex.Unlock()

	if client, exists := clientManager[connectionString]; exists {
		client.Close()
		delete(clientManager, connectionString)
	}
}

// CloseAllRabbitMQClients closes all RabbitMQ clients
func CloseAllRabbitMQClients() {
	clientMutex.Lock()
	defer clientMutex.Unlock()

	for connectionString, client := range clientManager {
		client.Close()
		delete(clientManager, connectionString)
	}
}

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
