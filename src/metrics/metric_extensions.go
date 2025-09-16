package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var MessagesSent = promauto.NewCounter(prometheus.CounterOpts{
	Name: "quepasa_sent_messages_total",
	Help: "Total sent messages",
})

var MessageSendErrors = promauto.NewCounter(prometheus.CounterOpts{
	Name: "quepasa_send_message_errors_total",
	Help: "Total message send errors",
})

var MessagesReceived = promauto.NewCounter(prometheus.CounterOpts{
	Name: "quepasa_received_messages_total",
	Help: "Total messages received",
})

var MessageReceiveErrors = promauto.NewCounter(prometheus.CounterOpts{
	Name: "quepasa_receive_message_errors_total",
	Help: "Total message receive errors",
})

var MessageReceiveUnhandled = promauto.NewCounter(prometheus.CounterOpts{
	Name: "quepasa_receive_message_unhandled_total",
	Help: "Total unhandled messages received",
})

var MessageReceiveSyncEvents = promauto.NewCounter(prometheus.CounterOpts{
	Name: "quepasa_receive_sync_events_total",
	Help: "Total sync events received",
})

// Webhook metrics
var WebhooksSent = promauto.NewCounter(prometheus.CounterOpts{
	Name: "quepasa_webhooks_sent_total",
	Help: "Total webhooks sent",
})

var WebhookSendErrors = promauto.NewCounter(prometheus.CounterOpts{
	Name: "quepasa_webhook_send_errors_total",
	Help: "Total webhook send errors",
})

// Additional webhook metrics for better monitoring
var WebhookLatency = promauto.NewHistogram(prometheus.HistogramOpts{
	Name:    "quepasa_webhook_duration_seconds",
	Help:    "Webhook request duration in seconds",
	Buckets: prometheus.DefBuckets,
})

var WebhookTimeouts = promauto.NewCounter(prometheus.CounterOpts{
	Name: "quepasa_webhook_timeouts_total",
	Help: "Total webhook timeout errors",
})

var WebhookHTTPErrors = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "quepasa_webhook_http_errors_total",
	Help: "Total webhook HTTP errors by status code",
}, []string{"status_code"})

var WebhookSuccess = promauto.NewCounter(prometheus.CounterOpts{
	Name: "quepasa_webhook_success_total",
	Help: "Total successful webhooks (HTTP 200)",
})

// Connection and server metrics
var ConnectionsActive = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "quepasa_connections_active",
	Help: "Number of active WhatsApp connections",
})

var ConnectionsConnected = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "quepasa_connections_connected",
	Help: "Number of connected WhatsApp connections",
})

var ConnectionsDisconnected = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "quepasa_connections_disconnected",
	Help: "Number of disconnected WhatsApp connections",
})

// Message type metrics
var MessagesByType = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "quepasa_messages_by_type_total",
	Help: "Total messages by type (text, image, audio, etc)",
}, []string{"type"})

// HTTP/API processing time metrics
var APIProcessingTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "quepasa_api_request_duration_seconds",
	Help:    "Time spent processing API requests",
	Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
}, []string{"method", "endpoint", "status_code"})

// RabbitMQ metrics
var RabbitMQMessagesPublished = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "quepasa_rabbitmq_messages_published_total",
	Help: "Total messages published to RabbitMQ",
}, []string{"queue", "exchange", "routing_key"})

var RabbitMQMessagesConsumed = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "quepasa_rabbitmq_messages_consumed_total",
	Help: "Total messages consumed from RabbitMQ",
}, []string{"queue", "consumer_tag"})

var RabbitMQMessagesAcknowledged = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "quepasa_rabbitmq_messages_acknowledged_total",
	Help: "Total messages acknowledged in RabbitMQ",
}, []string{"queue"})

var RabbitMQMessagesRejected = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "quepasa_rabbitmq_messages_rejected_total",
	Help: "Total messages rejected in RabbitMQ",
}, []string{"queue", "reason"})

var RabbitMQMessagesRequeued = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "quepasa_rabbitmq_messages_requeued_total",
	Help: "Total messages requeued in RabbitMQ",
}, []string{"queue"})

var RabbitMQPublishErrors = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "quepasa_rabbitmq_publish_errors_total",
	Help: "Total RabbitMQ publish errors",
}, []string{"queue", "exchange", "error_type"})

var RabbitMQConsumeErrors = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "quepasa_rabbitmq_consume_errors_total",
	Help: "Total RabbitMQ consume errors",
}, []string{"queue", "error_type"})

var RabbitMQConnectionStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "quepasa_rabbitmq_connection_status",
	Help: "RabbitMQ connection status (1 = connected, 0 = disconnected)",
}, []string{"connection_name", "host"})

var RabbitMQChannelStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "quepasa_rabbitmq_channel_status",
	Help: "RabbitMQ channel status (1 = open, 0 = closed)",
}, []string{"channel_name", "connection_name"})

var RabbitMQQueueLength = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "quepasa_rabbitmq_queue_length",
	Help: "Number of messages in RabbitMQ queue",
}, []string{"queue"})

var RabbitMQQueueConsumers = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "quepasa_rabbitmq_queue_consumers",
	Help: "Number of consumers for each queue",
}, []string{"queue"})

var RabbitMQPublishDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "quepasa_rabbitmq_publish_duration_seconds",
	Help:    "Time spent publishing messages to RabbitMQ",
	Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5},
}, []string{"queue", "exchange"})

var RabbitMQConsumeDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "quepasa_rabbitmq_consume_duration_seconds",
	Help:    "Time spent processing consumed messages from RabbitMQ",
	Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
}, []string{"queue", "message_type"})

var RabbitMQMessageSize = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "quepasa_rabbitmq_message_size_bytes",
	Help:    "Size of messages published to RabbitMQ in bytes",
	Buckets: []float64{100, 500, 1000, 5000, 10000, 50000, 100000, 500000, 1000000},
}, []string{"queue", "message_type"})

var RabbitMQConnectionReconnects = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "quepasa_rabbitmq_connection_reconnects_total",
	Help: "Total number of RabbitMQ connection reconnections",
}, []string{"connection_name", "reason"})

// Message processing metrics
var MessageProcessingDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "quepasa_message_processing_duration_seconds",
	Help:    "Time spent processing different types of messages",
	Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
}, []string{"message_type", "source", "processing_stage"})

var MessageProcessingErrors = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "quepasa_message_processing_errors_total",
	Help: "Total message processing errors by type and stage",
}, []string{"message_type", "source", "processing_stage", "error_type"})

var MessageQueueDepth = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "quepasa_message_queue_depth",
	Help: "Current depth of internal message processing queues",
}, []string{"queue_type", "priority"})

var MessageRetries = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "quepasa_message_retries_total",
	Help: "Total message processing retries",
}, []string{"message_type", "retry_reason", "source"})

// Helper functions for recording metrics

// ObserveAPIProcessingTime records API request processing time
func ObserveAPIProcessingTime(method, endpoint, statusCode string, duration float64) {
	APIProcessingTime.WithLabelValues(method, endpoint, statusCode).Observe(duration)
}

// RabbitMQ helper functions

// RecordRabbitMQMessagePublished increments the published messages counter
func RecordRabbitMQMessagePublished(queue, exchange, routingKey string) {
	RabbitMQMessagesPublished.WithLabelValues(queue, exchange, routingKey).Inc()
}

// RecordRabbitMQMessageConsumed increments the consumed messages counter
func RecordRabbitMQMessageConsumed(queue, consumerTag string) {
	RabbitMQMessagesConsumed.WithLabelValues(queue, consumerTag).Inc()
}

// RecordRabbitMQMessageAcknowledged increments the acknowledged messages counter
func RecordRabbitMQMessageAcknowledged(queue string) {
	RabbitMQMessagesAcknowledged.WithLabelValues(queue).Inc()
}

// RecordRabbitMQMessageRejected increments the rejected messages counter
func RecordRabbitMQMessageRejected(queue, reason string) {
	RabbitMQMessagesRejected.WithLabelValues(queue, reason).Inc()
}

// RecordRabbitMQMessageRequeued increments the requeued messages counter
func RecordRabbitMQMessageRequeued(queue string) {
	RabbitMQMessagesRequeued.WithLabelValues(queue).Inc()
}

// RecordRabbitMQPublishError increments the publish error counter
func RecordRabbitMQPublishError(queue, exchange, errorType string) {
	RabbitMQPublishErrors.WithLabelValues(queue, exchange, errorType).Inc()
}

// RecordRabbitMQConsumeError increments the consume error counter
func RecordRabbitMQConsumeError(queue, errorType string) {
	RabbitMQConsumeErrors.WithLabelValues(queue, errorType).Inc()
}

// SetRabbitMQConnectionStatus sets the connection status (1 = connected, 0 = disconnected)
func SetRabbitMQConnectionStatus(connectionName, host string, status float64) {
	RabbitMQConnectionStatus.WithLabelValues(connectionName, host).Set(status)
}

// SetRabbitMQChannelStatus sets the channel status (1 = open, 0 = closed)
func SetRabbitMQChannelStatus(channelName, connectionName string, status float64) {
	RabbitMQChannelStatus.WithLabelValues(channelName, connectionName).Set(status)
}

// SetRabbitMQQueueLength sets the current queue length
func SetRabbitMQQueueLength(queue string, length float64) {
	RabbitMQQueueLength.WithLabelValues(queue).Set(length)
}

// SetRabbitMQQueueConsumers sets the number of consumers for a queue
func SetRabbitMQQueueConsumers(queue string, consumers float64) {
	RabbitMQQueueConsumers.WithLabelValues(queue).Set(consumers)
}

// ObserveRabbitMQPublishDuration records the time spent publishing messages
func ObserveRabbitMQPublishDuration(queue, exchange string, duration float64) {
	RabbitMQPublishDuration.WithLabelValues(queue, exchange).Observe(duration)
}

// ObserveRabbitMQConsumeDuration records the time spent processing consumed messages
func ObserveRabbitMQConsumeDuration(queue, messageType string, duration float64) {
	RabbitMQConsumeDuration.WithLabelValues(queue, messageType).Observe(duration)
}

// ObserveRabbitMQMessageSize records the size of published messages
func ObserveRabbitMQMessageSize(queue, messageType string, sizeBytes float64) {
	RabbitMQMessageSize.WithLabelValues(queue, messageType).Observe(sizeBytes)
}

// RecordRabbitMQConnectionReconnect increments the reconnection counter
func RecordRabbitMQConnectionReconnect(connectionName, reason string) {
	RabbitMQConnectionReconnects.WithLabelValues(connectionName, reason).Inc()
}

// Message processing helper functions

// ObserveMessageProcessingDuration records time spent in different processing stages
func ObserveMessageProcessingDuration(messageType, source, stage string, duration float64) {
	MessageProcessingDuration.WithLabelValues(messageType, source, stage).Observe(duration)
}

// RecordMessageProcessingError increments processing error counter
func RecordMessageProcessingError(messageType, source, stage, errorType string) {
	MessageProcessingErrors.WithLabelValues(messageType, source, stage, errorType).Inc()
}

// SetMessageQueueDepth sets the current depth of internal queues
func SetMessageQueueDepth(queueType, priority string, depth float64) {
	MessageQueueDepth.WithLabelValues(queueType, priority).Set(depth)
}

// RecordMessageRetry increments the retry counter
func RecordMessageRetry(messageType, retryReason, source string) {
	MessageRetries.WithLabelValues(messageType, retryReason, source).Inc()
}
