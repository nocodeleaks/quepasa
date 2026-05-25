package rabbitmq

import (
	"github.com/nocodeleaks/quepasa/metrics"
)

// RabbitMQ-specific metrics initialized directly using generic factory functions
var (
	// Counters
	MessagesPublished      = metrics.CreateCounterVecRecorder("quepasa_rabbitmq_messages_published_total", "Total messages published to RabbitMQ", []string{"queue"})
	MessagePublishErrors   = metrics.CreateCounterRecorder("quepasa_rabbitmq_message_publish_errors_total", "Total message publish errors to RabbitMQ")
	MessagesCached         = metrics.CreateCounterRecorder("quepasa_rabbitmq_messages_cached_total", "Total messages added to cache when connection is down")
	MessagesCacheProcessed = metrics.CreateCounterRecorder("quepasa_rabbitmq_messages_cache_processed_total", "Total messages processed from cache")
	MessagesCacheDropped   = metrics.CreateCounterRecorder("quepasa_rabbitmq_messages_cache_dropped_total", "Total messages dropped because the cache was full")
	ConnectionsEstablished = metrics.CreateCounterRecorder("quepasa_rabbitmq_connections_established_total", "Total RabbitMQ connections established")
	ConnectionsLost        = metrics.CreateCounterRecorder("quepasa_rabbitmq_connections_lost_total", "Total RabbitMQ connections lost")
	ReconnectionAttempts   = metrics.CreateCounterRecorder("quepasa_rabbitmq_reconnection_attempts_total", "Total reconnection attempts to RabbitMQ")

	// Gauges
	CacheSizeCurrent = metrics.CreateGaugeRecorder("quepasa_rabbitmq_cache_size_current", "Current number of messages in the RabbitMQ retry cache")
)
