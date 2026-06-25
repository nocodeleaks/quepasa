package runtime

import (
	"testing"
	"time"

	dispatchservice "github.com/nocodeleaks/quepasa/dispatch/service"
	models "github.com/nocodeleaks/quepasa/models"
)

type lifecycleCapturePublisher struct {
	lifecycle chan interface{}
}

func (p *lifecycleCapturePublisher) PublishMessage(payload interface{}) {}

func (p *lifecycleCapturePublisher) PublishLifecycle(payload interface{}) {
	select {
	case p.lifecycle <- payload:
	default:
	}
}

func TestNewDispatchingLifecyclePublisher_PublishLifecycle(t *testing.T) {
	capture := &lifecycleCapturePublisher{lifecycle: make(chan interface{}, 1)}
	dispatchservice.RegisterRealtimePublisher(capture)

	publisher := NewDispatchingLifecyclePublisher()
	event := &models.DispatchingLifecycleEvent{
		Kind:      "connected",
		Token:     "token-1",
		User:      "user-1",
		Wid:       "5511999999999@s.whatsapp.net",
		Phone:     "5511999999999",
		State:     "ready",
		Verified:  true,
		Cause:     "",
		Details:   "",
		Timestamp: time.Now().UTC(),
	}

	publisher.PublishLifecycle(event)

	select {
	case payload := <-capture.lifecycle:
		realtimeEvent, ok := payload.(*dispatchservice.RealtimeLifecycleEvent)
		if !ok {
			t.Fatalf("expected *RealtimeLifecycleEvent payload, got %T", payload)
		}

		if realtimeEvent.Kind != event.Kind {
			t.Fatalf("expected kind %q, got %q", event.Kind, realtimeEvent.Kind)
		}

		if realtimeEvent.Token != event.Token {
			t.Fatalf("expected token %q, got %q", event.Token, realtimeEvent.Token)
		}

		if realtimeEvent.Wid != event.Wid {
			t.Fatalf("expected wid %q, got %q", event.Wid, realtimeEvent.Wid)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for lifecycle payload")
	}
}
