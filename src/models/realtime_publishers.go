package models

import (
	"sync"
	"time"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// RealtimeLifecycleEvent describes a server lifecycle transition that is useful
// to broadcast to live clients independently from the raw WhatsApp message flow.
//
// The event keeps only transport-agnostic data so different realtime adapters
// (SignalR, websocket cable, SSE, etc.) can project the same state change using
// their own protocols without forcing the model layer to know about them.
type RealtimeLifecycleEvent struct {
	Kind      string                 `json:"kind"`
	Token     string                 `json:"token,omitempty"`
	User      string                 `json:"user,omitempty"`
	Wid       string                 `json:"wid,omitempty"`
	Phone     string                 `json:"phone,omitempty"`
	State     string                 `json:"state,omitempty"`
	Verified  bool                   `json:"verified"`
	Cause     string                 `json:"cause,omitempty"`
	Details   string                 `json:"details,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// RealtimePublisher is implemented by transport modules that want to receive live
// message and lifecycle events from the model layer.
//
// Implementations are expected to return quickly. DispatchingHandler calls these
// methods from goroutines so a slow transport does not block WhatsApp processing.
type RealtimePublisher interface {
	PublishServerMessage(server *QpWhatsappServer, payload *whatsapp.WhatsappMessage)
	PublishServerLifecycle(event *RealtimeLifecycleEvent)
}

var realtimePublishers struct {
	sync.RWMutex
	items []RealtimePublisher
}

// RegisterRealtimePublisher appends a realtime publisher if it is not nil.
func RegisterRealtimePublisher(publisher RealtimePublisher) {
	if publisher == nil {
		return
	}

	realtimePublishers.Lock()
	realtimePublishers.items = append(realtimePublishers.items, publisher)
	realtimePublishers.Unlock()
}

// PublishRealtimeServerMessage forwards a message to every registered realtime
// publisher. Each publisher runs in its own goroutine to isolate transports.
func PublishRealtimeServerMessage(server *QpWhatsappServer, payload *whatsapp.WhatsappMessage) {
	if server == nil || payload == nil {
		return
	}

	realtimePublishers.RLock()
	publishers := append([]RealtimePublisher(nil), realtimePublishers.items...)
	realtimePublishers.RUnlock()

	for _, publisher := range publishers {
		if publisher == nil {
			continue
		}

		go publisher.PublishServerMessage(server, payload)
	}
}

// PublishRealtimeLifecycle forwards a lifecycle event to every registered realtime
// publisher. The event is copied by value by the caller before it reaches the
// transport layer, so publishers must treat the payload as read-only.
func PublishRealtimeLifecycle(event *RealtimeLifecycleEvent) {
	if event == nil {
		return
	}

	realtimePublishers.RLock()
	publishers := append([]RealtimePublisher(nil), realtimePublishers.items...)
	realtimePublishers.RUnlock()

	for _, publisher := range publishers {
		if publisher == nil {
			continue
		}

		go publisher.PublishServerLifecycle(event)
	}
}
