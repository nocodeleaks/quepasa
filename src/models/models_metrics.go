package models

import (
	metrics "github.com/nocodeleaks/quepasa/metrics"
)

// Models-specific metrics initialized directly using generic factory functions
var (
	WebhooksSent              = metrics.CreateCounterRecorder("quepasa_webhooks_sent_total", "Total webhooks sent")
	WebhookSendErrors         = metrics.CreateCounterRecorder("quepasa_webhook_send_errors_total", "Total webhook send errors")
	MessageProcessingDuration = metrics.CreateHistogramVecRecorder("quepasa_message_processing_duration_seconds", "Time spent processing different types of messages", []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0}, []string{"message_type", "source", "processing_stage"})
	MessageProcessingErrors   = metrics.CreateCounterVecRecorder("quepasa_message_processing_errors_total", "Total message processing errors by type and stage", []string{"message_type", "source", "processing_stage", "error_type"})
	MessageQueueDepth         = metrics.CreateCounterVecRecorder("quepasa_message_queue_depth", "Current depth of internal message processing queues", []string{"queue_type", "priority"}) // Using counter for now, gauge not available
	MessageRetries            = metrics.CreateCounterVecRecorder("quepasa_message_retries_total", "Total message processing retries", []string{"message_type", "retry_reason", "source"})
	WebhookLatency            = metrics.CreateHistogramVecRecorder("quepasa_webhook_duration_seconds", "Webhook request duration in seconds", []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}, []string{})
	WebhookTimeouts           = metrics.CreateCounterRecorder("quepasa_webhook_timeouts_total", "Total webhook timeout errors")
	WebhookHTTPErrors         = metrics.CreateCounterVecRecorder("quepasa_webhook_http_errors_total", "Total webhook HTTP errors by status code", []string{"status_code"})
	WebhookSuccess            = metrics.CreateCounterRecorder("quepasa_webhook_success_total", "Total successful webhooks (HTTP 200)")
)
