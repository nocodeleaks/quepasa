package rabbitmq

import (
	"sync"
	"time"
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

// clientManager holds all active RabbitMQ clients keyed by connection string.
var (
	clientManager = make(map[string]*RabbitMQClient)
	clientMutex   sync.RWMutex
)

// GetRabbitMQClient returns an existing RabbitMQ client for the given connection string,
// or creates a new one with an unlimited in-memory cache.
// On first creation it waits up to 15 seconds for the connection and then ensures
// the QuePasa exchange and queues exist.
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

	// Create new client — double-check under write lock.
	clientMutex.Lock()
	defer clientMutex.Unlock()

	if client, exists = clientManager[connectionString]; exists {
		return client
	}

	client = NewRabbitMQClient(connectionString, 0) // 0 = unlimited cache (100 000 default)
	clientManager[connectionString] = client

	// Initialise Exchange and Queues synchronously so they are ready for the first publish.
	if client.WaitForConnection(15 * time.Second) {
		if err := client.EnsureExchangeAndQueues(); err != nil {
			// Non-fatal: processCache will retry on every reconnection cycle.
			_ = err
		}
	}

	return client
}

// CloseRabbitMQClient closes and removes the client for the given connection string.
func CloseRabbitMQClient(connectionString string) {
	clientMutex.Lock()
	defer clientMutex.Unlock()

	if client, exists := clientManager[connectionString]; exists {
		client.Close()
		delete(clientManager, connectionString)
	}
}

// CloseAllRabbitMQClients closes all active RabbitMQ clients.
func CloseAllRabbitMQClients() {
	clientMutex.Lock()
	defer clientMutex.Unlock()

	for connectionString, client := range clientManager {
		client.Close()
		delete(clientManager, connectionString)
	}
}
