package api

import (
	metrics "github.com/nocodeleaks/quepasa/metrics"
)

// API-specific metrics initialized directly using generic factory functions
var (
	MessagesSent         = metrics.CreateCounterRecorder("quepasa_api_messages_sent_total", "Total messages sent via API")
	MessageSendErrors    = metrics.CreateCounterRecorder("quepasa_api_message_send_errors_total", "Total message send errors via API")
	MessagesReceived     = metrics.CreateCounterRecorder("quepasa_api_messages_received_total", "Total messages received via API")
	MessageReceiveErrors = metrics.CreateCounterRecorder("quepasa_api_message_receive_errors_total", "Total message receive errors via API")
)
