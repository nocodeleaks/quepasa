package rabbitmq

import (
	"log"

	cacheservice "github.com/nocodeleaks/quepasa/cache/service"
)

// InjectCacheBackendIntoClient injects the cache backend into a RabbitMQClient.
// This is called after the RabbitMQ client is initialized to provide the backend.
func InjectCacheBackendIntoClient(client *RabbitMQClient) {
	if client == nil {
		return
	}

	cacheService := cacheservice.GetInstance()
	backend := cacheService.GetQueueBackend()

	if backend == nil {
		log.Println("WARNING: Queue cache backend is nil, RabbitMQ retry cache will not function properly")
		return
	}

	client.SetCacheBackend(backend)
}
