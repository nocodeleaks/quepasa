package models

import (
	"log"

	cacheservice "github.com/nocodeleaks/quepasa/cache/service"
	rabbitmq "github.com/nocodeleaks/quepasa/rabbitmq"
)

// InitializeCacheService initializes the global cache service and injects backends
// into existing components. This should be called early in application startup,
// after database migration but before services are started.
func InitializeCacheService() error {
	log.Println("Initializing centralized cache service...")

	// Get the singleton cache service instance
	// This will initialize backends based on environment configuration
	// with automatic fallback to memory if configured backend fails
	_ = cacheservice.GetInstance()

	log.Println("Cache service initialized successfully")
	log.Printf("Messages backend: initialized")
	log.Printf("Queue backend: initialized")

	// Inject cache backend into RabbitMQ client if it has been initialized
	if rabbitmq.RabbitMQClientInstance != nil {
		log.Println("Injecting queue backend into RabbitMQ client...")
		rabbitmq.InjectCacheBackendIntoClient(rabbitmq.RabbitMQClientInstance)
	}

	return nil
}

// InjectCacheBackendIntoHandler injects the cache backend into a DispatchingHandler.
// This is called during handler initialization to provide the backend.
func InjectCacheBackendIntoHandler(handler *DispatchingHandler) {
	if handler == nil {
		return
	}

	cacheService := cacheservice.GetInstance()
	backend := cacheService.GetMessagesBackend()

	if backend == nil {
		log.Println("WARNING: Cache backend is nil, handler will not function properly")
		return
	}

	handler.SetBackend(backend)
}
