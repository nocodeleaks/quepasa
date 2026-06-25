package events

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestPublishDoesNotBlockSlowSubscribers(t *testing.T) {
	bus := NewBus()
	release := make(chan struct{})
	unsubscribe := bus.Subscribe(SubscribeOptions{
		BufferSize: 1,
		Handler: func(event Event) {
			<-release
		},
	})
	defer unsubscribe()

	start := time.Now()
	for index := 0; index < 256; index++ {
		bus.Publish(Event{Name: "test.publish", Source: "events_test", Status: "queued"})
	}
	elapsed := time.Since(start)

	close(release)

	if elapsed > 50*time.Millisecond {
		t.Fatalf("publish blocked for too long: %v", elapsed)
	}
}

func TestSubscribeFilterReceivesOnlyMatchingEvents(t *testing.T) {
	bus := NewBus()
	var received int32
	done := make(chan struct{})
	unsubscribe := bus.Subscribe(SubscribeOptions{
		Filter: func(event Event) bool {
			return event.Status == "success"
		},
		Handler: func(event Event) {
			atomic.AddInt32(&received, 1)
			close(done)
		},
	})
	defer unsubscribe()

	bus.Publish(Event{Name: "test.publish", Status: "error"})
	bus.Publish(Event{Name: "test.publish", Status: "success"})

	select {
	case <-done:
	case <-time.After(250 * time.Millisecond):
		t.Fatal("expected filtered subscriber to receive the success event")
	}

	if got := atomic.LoadInt32(&received); got != 1 {
		t.Fatalf("expected to receive exactly one matching event, got %d", got)
	}
}
