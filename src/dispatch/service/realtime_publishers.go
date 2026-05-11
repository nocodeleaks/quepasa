package service

import (
	"sync"
	"time"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// RealtimeServerMessage is the canonical realtime payload for message fanout.
type RealtimeServerMessage struct {
	Token   string                    `json:"token"`
	User    string                    `json:"user"`
	WID     string                    `json:"wid"`
	State   string                    `json:"state"`
	Message *whatsapp.WhatsappMessage `json:"message"`
}

// RealtimeLifecycleEvent is the canonical realtime payload for lifecycle fanout.
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

// RealtimePublisher is a transport-neutral realtime publisher contract used by
// the dispatch module. Payload shape is owned by the caller.
type RealtimePublisher interface {
	PublishMessage(payload interface{})
	PublishLifecycle(payload interface{})
}

var realtimePublishers struct {
	sync.RWMutex
	items []RealtimePublisher
}

// RegisterRealtimePublisher appends one realtime publisher to the dispatch bus.
func RegisterRealtimePublisher(publisher RealtimePublisher) {
	if publisher == nil {
		return
	}

	realtimePublishers.Lock()
	realtimePublishers.items = append(realtimePublishers.items, publisher)
	realtimePublishers.Unlock()
}

// PublishRealtimeMessage forwards one message payload to all registered
// realtime publishers. Each publisher runs in a dedicated goroutine.
func PublishRealtimeMessage(payload interface{}) {
	realtimePublishers.RLock()
	publishers := append([]RealtimePublisher(nil), realtimePublishers.items...)
	realtimePublishers.RUnlock()

	for _, publisher := range publishers {
		if publisher == nil {
			continue
		}

		go publisher.PublishMessage(payload)
	}
}

// PublishRealtimeLifecycle forwards one lifecycle payload to all registered
// realtime publishers. Each publisher runs in a dedicated goroutine.
func PublishRealtimeLifecycle(payload interface{}) {
	realtimePublishers.RLock()
	publishers := append([]RealtimePublisher(nil), realtimePublishers.items...)
	realtimePublishers.RUnlock()

	for _, publisher := range publishers {
		if publisher == nil {
			continue
		}

		go publisher.PublishLifecycle(payload)
	}
}
