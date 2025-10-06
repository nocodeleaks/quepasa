package whatsmeow

import (
	"github.com/nocodeleaks/quepasa/metrics"
)

// WhatsMeow-specific metrics initialized directly using generic factory functions
var (
	MessagesReceived         = metrics.CreateCounterRecorder("quepasa_whatsmeow_messages_received_total", "Total messages received via WhatsMeow")
	MessageReceiveErrors     = metrics.CreateCounterRecorder("quepasa_whatsmeow_message_receive_errors_total", "Total message receive errors via WhatsMeow")
	MessageReceiveUnhandled  = metrics.CreateCounterRecorder("quepasa_whatsmeow_message_receive_unhandled_total", "Total unhandled messages received via WhatsMeow")
	MessageReceiveSyncEvents = metrics.CreateCounterRecorder("quepasa_whatsmeow_message_receive_sync_events_total", "Total sync events received via WhatsMeow")
	MessagesByType           = metrics.CreateCounterVecRecorder("quepasa_whatsmeow_messages_by_type_total", "Total messages by type (text, image, audio, etc)", []string{"type"})
)
