package runtime

import (
	dispatchservice "github.com/nocodeleaks/quepasa/dispatch/service"
	models "github.com/nocodeleaks/quepasa/models"
)

// lifecyclePublisher adapts models lifecycle events to dispatch realtime transport.
type lifecyclePublisher struct{}

// NewDispatchingLifecyclePublisher creates a runtime adapter for lifecycle notifications.
func NewDispatchingLifecyclePublisher() models.DispatchingLifecyclePublisher {
	return lifecyclePublisher{}
}

func (lifecyclePublisher) PublishLifecycle(event *models.DispatchingLifecycleEvent) {
	if event == nil {
		return
	}

	dispatchservice.PublishRealtimeLifecycle(&dispatchservice.RealtimeLifecycleEvent{
		Kind:      event.Kind,
		Token:     event.Token,
		User:      event.User,
		Wid:       event.Wid,
		Phone:     event.Phone,
		State:     event.State,
		Verified:  event.Verified,
		Cause:     event.Cause,
		Details:   event.Details,
		Timestamp: event.Timestamp,
	})
}
