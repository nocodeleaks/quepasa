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
