package whatsmeow

import (
	"fmt"
	"testing"
	"time"

	qpevents "github.com/nocodeleaks/quepasa/events"
	log "github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow/types/events"
)

// minimalHandlers builds a WhatsmeowHandlers with just enough state for
// event-publishing unit tests: a live logger, no WhatsApp connection.
func minimalHandlers(t *testing.T) *WhatsmeowHandlers {
	t.Helper()
	conn := &WhatsmeowConnection{}
	conn.LogEntry = log.New().WithField("test", t.Name())
	return &WhatsmeowHandlers{
		WhatsmeowConnection: conn,
	}
}

// subscribeOnce registers a one-shot subscriber on the default event bus for
// events matching name and returns a buffered receive-channel plus an unsubscribe
// function the caller must defer.
func subscribeOnce(name string) (<-chan qpevents.Event, func()) {
	ch := make(chan qpevents.Event, 1)
	unsub := qpevents.Subscribe(qpevents.SubscribeOptions{
		Name:       fmt.Sprintf("test-%s", name),
		BufferSize: 1,
		Filter:     func(e qpevents.Event) bool { return e.Name == name },
		Handler: func(e qpevents.Event) {
			select {
			case ch <- e:
			default:
			}
		},
	})
	return ch, unsub
}

// waitEvent asserts that an event arrives within 250 ms.
func waitEvent(t *testing.T, ch <-chan qpevents.Event) qpevents.Event {
	t.Helper()
	select {
	case evt := <-ch:
		return evt
	case <-time.After(250 * time.Millisecond):
		t.Fatal("timed out waiting for event on bus")
		return qpevents.Event{}
	}
}

func TestOnConnectFailureEvent_PublishesEvent(t *testing.T) {
	ch, unsub := subscribeOnce("whatsapp.connect.failure")
	defer unsub()

	h := minimalHandlers(t)
	h.onConnectFailureEvent(&events.ConnectFailure{
		Reason:  events.ConnectFailureGeneric,
		Message: "unit test trigger",
	})

	evt := waitEvent(t, ch)

	if evt.Status != "error" {
		t.Errorf("expected status=error, got %q", evt.Status)
	}
	if evt.Source != "whatsmeow.handlers" {
		t.Errorf("expected source=whatsmeow.handlers, got %q", evt.Source)
	}
	if evt.Attributes["reason"] == "" {
		t.Error("expected non-empty reason attribute")
	}
}

func TestOnStreamErrorEvent_PublishesEvent(t *testing.T) {
	ch, unsub := subscribeOnce("whatsapp.stream.error")
	defer unsub()

	h := minimalHandlers(t)
	h.onStreamErrorEvent(&events.StreamError{Code: "503"})

	evt := waitEvent(t, ch)

	if evt.Status != "error" {
		t.Errorf("expected status=error, got %q", evt.Status)
	}
	if evt.Attributes["code"] != "503" {
		t.Errorf("expected code=503, got %q", evt.Attributes["code"])
	}
}

func TestOnTemporaryBanEvent_PublishesEvent(t *testing.T) {
	ch, unsub := subscribeOnce("whatsapp.ban.temporary")
	defer unsub()

	h := minimalHandlers(t)
	h.onTemporaryBanEvent(&events.TemporaryBan{
		Code:   events.TempBanSentToTooManyPeople, // 101
		Expire: 24 * time.Hour,
	})

	evt := waitEvent(t, ch)

	if evt.Status != "error" {
		t.Errorf("expected status=error, got %q", evt.Status)
	}
	if evt.Attributes["code"] != "101" {
		t.Errorf("expected code=101, got %q", evt.Attributes["code"])
	}
	if evt.Attributes["reason"] != "Sent to too many people" {
		t.Errorf("unexpected reason: %q", evt.Attributes["reason"])
	}
}

func TestOnTemporaryBanEvent_UnknownCodePublishesEvent(t *testing.T) {
	ch, unsub := subscribeOnce("whatsapp.ban.temporary")
	defer unsub()

	h := minimalHandlers(t)
	h.onTemporaryBanEvent(&events.TemporaryBan{
		Code:   events.TempBanReason(999),
		Expire: time.Hour,
	})

	evt := waitEvent(t, ch)

	if evt.Attributes["reason"] != "Unknown" {
		t.Errorf("expected reason=Unknown for unrecognized code, got %q", evt.Attributes["reason"])
	}
}
