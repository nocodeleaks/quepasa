package cable

import (
	"testing"

	models "github.com/nocodeleaks/quepasa/models"
)

func TestQueueFrameIgnoresClosedClientWithoutPanic(t *testing.T) {
	hub := NewHub()
	client := &Client{
		id:            "client-closed",
		user:          &models.QpUser{Username: "closed-user"},
		hub:           hub,
		send:          make(chan []byte, 1),
		subscriptions: map[string]struct{}{},
	}

	client.sendMu.Lock()
	client.closed = true
	close(client.send)
	client.sendMu.Unlock()

	hub.queueFrame(client, ServerFrame{Type: "event", Event: "test.closed"})
}

func TestRemoveClientClosesChannelAndQueueFrameReturns(t *testing.T) {
	hub := NewHub()
	client := &Client{
		id:            "client-remove",
		user:          &models.QpUser{Username: "remove-user"},
		hub:           hub,
		send:          make(chan []byte, 1),
		subscriptions: map[string]struct{}{},
	}

	hub.removeClient(client)
	hub.queueFrame(client, ServerFrame{Type: "event", Event: "test.after-remove"})

	if !client.closed {
		t.Fatal("expected client to be marked closed after removal")
	}
}
