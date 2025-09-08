package metrics

import(
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

// Webhook metrics
var WebhooksSent = promauto.NewCounter(prometheus.CounterOpts{
	Name: "quepasa_webhooks_sent_total",
	Help: "Total webhooks sent",
})

var WebhookSendErrors = promauto.NewCounter(prometheus.CounterOpts{
	Name: "quepasa_webhook_send_errors_total",
	Help: "Total webhook send errors",
})

var WebhookRetryAttempts = promauto.NewCounter(prometheus.CounterOpts{
	Name: "quepasa_webhook_retry_attempts_total",
	Help: "Total webhook retry attempts",
})

var WebhookRetriesSuccessful = promauto.NewCounter(prometheus.CounterOpts{
	Name: "quepasa_webhook_retries_successful_total",
	Help: "Total successful webhooks after retry",
})

var WebhookRetryFailures = promauto.NewCounter(prometheus.CounterOpts{
	Name: "quepasa_webhook_retry_failures_total",
	Help: "Total webhook failures after all retries",
})

var WebhookLatency = promauto.NewHistogram(prometheus.HistogramOpts{
	Name:    "quepasa_webhook_duration_seconds",
	Help:    "Webhook request duration in seconds",
	Buckets: prometheus.DefBuckets,
})
