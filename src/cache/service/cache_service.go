package service

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/nocodeleaks/quepasa/cache"
	cache_disk "github.com/nocodeleaks/quepasa/cache/disk"
	cache_memory "github.com/nocodeleaks/quepasa/cache/memory"
	cache_redis "github.com/nocodeleaks/quepasa/cache/redis"
	environment "github.com/nocodeleaks/quepasa/environment"
)

// CacheService is the centralized cache service for the entire application.
// It provides singleton access to a configured backend (memory, redis, or disk)
// with automatic fallback to memory if the configured backend fails.
type CacheService struct {
	messagesBackend cache.MessagesBackend
	queueBackend    cache.BytesQueueBackend
	mu              sync.RWMutex
}

var (
	instance *CacheService
	once     sync.Once
)

// GetInstance returns the singleton CacheService instance.
// It initializes the service on first call using environment configuration.
func GetInstance() *CacheService {
	once.Do(func() {
		var err error
		instance, err = newCacheService()
		if err != nil {
			log.Fatalf("Failed to initialize CacheService: %v", err)
		}
	})
	return instance
}

// newCacheService creates a new CacheService with configured backends.
// It attempts to initialize the configured backend (redis/disk) for both messages and queue.
// If initialization fails and CACHE_INIT_FALLBACK is true, both fallback to memory.
func newCacheService() (*CacheService, error) {
	service := &CacheService{}

	// Initialize messages backend with fallback
	messagesBackend, err := initMessagesBackend()
	if err != nil {
		if !environment.Settings.Cache.InitFallback {
			return nil, fmt.Errorf("messages backend initialization failed: %w", err)
		}
		log.Printf("Messages backend initialization failed, falling back to memory: %v", err)
		messagesBackend = cache_memory.NewMessagesBackend()
	}
	service.messagesBackend = messagesBackend

	// Initialize queue backend with fallback
	queueBackend, err := initQueueBackend()
	if err != nil {
		if !environment.Settings.Cache.InitFallback {
			return nil, fmt.Errorf("queue backend initialization failed: %w", err)
		}
		log.Printf("Queue backend initialization failed, falling back to memory: %v", err)
		queueBackend = cache_memory.NewBytesQueueBackend(100000)
	}
	service.queueBackend = queueBackend

	return service, nil
}

// initMessagesBackend creates the messages backend based on environment configuration.
// Respects CACHE_BACKEND setting: "memory", "redis", or "disk".
func initMessagesBackend() (cache.MessagesBackend, error) {
	switch environment.Settings.Cache.Backend {
	case cache.BackendMemory:
		return cache_memory.NewMessagesBackend(), nil

	case cache.BackendRedis:
		config := cache.RedisConfig{
			Host:         environment.Settings.Redis.Host,
			Port:         environment.Settings.Redis.Port,
			Username:     environment.Settings.Redis.Username,
			Password:     environment.Settings.Redis.Password,
			Database:     environment.Settings.Redis.Database,
			KeyPrefix:    environment.Settings.Redis.KeyPrefix,
			MaxRetries:   environment.Settings.Redis.MaxRetries,
			PoolSize:     environment.Settings.Redis.PoolSize,
			DialTimeout:  time.Duration(environment.Settings.Redis.DialTimeoutSeconds) * time.Second,
			ReadTimeout:  time.Duration(environment.Settings.Redis.ReadTimeoutSeconds) * time.Second,
			WriteTimeout: time.Duration(environment.Settings.Redis.WriteTimeoutSeconds) * time.Second,
		}
		backend, err := cache_redis.NewMessagesBackend(config)
		if err != nil {
			return nil, fmt.Errorf("redis messages backend: %w", err)
		}
		return backend, nil

	case cache.BackendDisk:
		backend, err := cache_disk.NewMessagesBackend(environment.Settings.Cache.DiskPath)
		if err != nil {
			return nil, fmt.Errorf("disk messages backend: %w", err)
		}
		return backend, nil

	default:
		return nil, fmt.Errorf("unknown cache backend: %s", environment.Settings.Cache.Backend)
	}
}

// initQueueBackend creates the queue backend based on environment configuration.
// For RabbitMQ retry, respects RABBITMQ_CACHE_BACKEND or uses CACHE_BACKEND as default.
func initQueueBackend() (cache.BytesQueueBackend, error) {
	// Determine which backend to use for queue
	backendName := environment.Settings.Cache.Backend
	if environment.Settings.RabbitMQ.CacheBackend != "" {
		backendName = environment.Settings.RabbitMQ.CacheBackend
	}

	// Capacity follows RabbitMQ cache length when configured; 0 means unlimited fallback.
	capacity := int(environment.Settings.RabbitMQ.CacheLength)
	if capacity == 0 {
		capacity = 100000
	}

	queueKey := environment.Settings.RabbitMQ.CacheQueueKey
	if queueKey == "" {
		queueKey = "rabbitmq_retry"
	}

	switch backendName {
	case cache.BackendMemory:
		return cache_memory.NewBytesQueueBackend(capacity), nil

	case cache.BackendRedis:
		config := cache.BytesQueueConfig{
			Capacity: capacity,
			Backend:  cache.BackendRedis,
			QueueKey: queueKey,
			Redis: cache.RedisConfig{
				Host:         environment.Settings.Redis.Host,
				Port:         environment.Settings.Redis.Port,
				Username:     environment.Settings.Redis.Username,
				Password:     environment.Settings.Redis.Password,
				Database:     environment.Settings.Redis.Database,
				KeyPrefix:    environment.Settings.Redis.KeyPrefix,
				MaxRetries:   environment.Settings.Redis.MaxRetries,
				PoolSize:     environment.Settings.Redis.PoolSize,
				DialTimeout:  time.Duration(environment.Settings.Redis.DialTimeoutSeconds) * time.Second,
				ReadTimeout:  time.Duration(environment.Settings.Redis.ReadTimeoutSeconds) * time.Second,
				WriteTimeout: time.Duration(environment.Settings.Redis.WriteTimeoutSeconds) * time.Second,
			},
		}
		queueBackend, err := cache_redis.NewBytesQueueBackend(config)
		if err != nil {
			return nil, fmt.Errorf("redis queue backend: %w", err)
		}
		return queueBackend, nil

	case cache.BackendDisk:
		diskPath := environment.Settings.RabbitMQ.CacheDiskPath
		if diskPath == "" {
			// Fallback to cache disk path if RabbitMQ cache disk path not specified
			diskPath = environment.Settings.Cache.DiskPath
		}
		queueBackend, err := cache_disk.NewBytesQueueBackend(diskPath, capacity)
		if err != nil {
			return nil, fmt.Errorf("disk queue backend: %w", err)
		}
		return queueBackend, nil

	default:
		return nil, fmt.Errorf("unknown queue backend: %s", backendName)
	}
}

// GetMessagesBackend returns the configured messages backend.
// This backend is used by QpWhatsappMessages and other message-related cache consumers.
func (cs *CacheService) GetMessagesBackend() cache.MessagesBackend {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.messagesBackend
}

// GetQueueBackend returns the configured queue backend.
// This backend is used by RabbitMQClient and other queue-related cache consumers.
func (cs *CacheService) GetQueueBackend() cache.BytesQueueBackend {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.queueBackend
}

// Close closes all backends. Should be called during application shutdown.
func (cs *CacheService) Close() error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	var errs []error

	if cs.messagesBackend != nil {
		if err := cs.messagesBackend.Close(); err != nil {
			errs = append(errs, fmt.Errorf("error closing messages backend: %w", err))
		}
	}

	if cs.queueBackend != nil {
		if err := cs.queueBackend.Close(); err != nil {
			errs = append(errs, fmt.Errorf("error closing queue backend: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing backends: %v", errs)
	}

	return nil
}
