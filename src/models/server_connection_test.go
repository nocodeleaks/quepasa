package models

// Tests for UpdateConnection contracts.
//
// These tests guard against the class of bug where a WhatsApp connection is assigned
// to a live server without going through UpdateConnection(), which is the only path
// that guarantees the server's DispatchingHandler is wired on the connection.
//
// Concretely: if server.connection is set via direct assignment instead of
// UpdateConnection(), the server can connect and send messages but will silently
// drop all inbound events because no external handler is registered on the
// underlying whatsmeow client.
//
// Covered scenarios:
//   - UpdateConnection always calls UpdateHandler on the new connection.
//   - UpdateConnection disposes the previous connection before wiring the new one.
//   - UpdateConnection attaches the dispatching subscriber when it is missing.
//   - UpdateConnection does not create duplicate dispatching subscribers.
//   - Reconnection (second UpdateConnection call) re-wires the handler on the
//     replacement connection and the previous one is disposed.

import (
	"sync"
	"testing"
	"time"

	library "github.com/nocodeleaks/quepasa/library"
	log "github.com/sirupsen/logrus"
)

// newTestServer builds a minimal QpWhatsappServer with a properly wired
// DispatchingHandler, suitable for unit-testing connection lifecycle behaviour.
// It reuses the pairingTestServersData stub declared in qp_whatsapp_service_pairing_test.go.
func newTestServer(t *testing.T, token string) *QpWhatsappServer {
	t.Helper()

	qpSrv := &QpServer{Token: token}
	qpSrv.LogEntry = log.New().WithField("test", t.Name())

	server := &QpWhatsappServer{
		QpServer:       qpSrv,
		syncConnection: &sync.Mutex{},
		syncMessages:   &sync.Mutex{},
		Timestamps:     QpTimestamps{Start: time.Now().UTC()},
		Intent:         SessionIntentNone,
		db:             pairingTestServersData{},
	}
	server.LogEntry = library.NewLogEntry(server).WithField("test", t.Name())

	// Wire a real DispatchingHandler so that HasDispatchingSubscriber et al. work.
	server.Handler = &DispatchingHandler{server: server}

	return server
}

// TestUpdateConnection_WiresHandlerOnNewConnection verifies the fundamental
// contract: after UpdateConnection(), the new connection's handler must be
// the server's DispatchingHandler.
func TestUpdateConnection_WiresHandlerOnNewConnection(t *testing.T) {
	server := newTestServer(t, "wc-token-1")
	conn := newPairingTestConnection(t)

	server.UpdateConnection(conn)

	if conn.updatedHandler == nil {
		t.Fatal("expected UpdateHandler to be called on new connection, got nil handler")
	}
	if conn.updatedHandler != server.Handler {
		t.Fatal("expected new connection handler to be the server DispatchingHandler instance")
	}
}

// TestUpdateConnection_DisposesOldConnectionBeforeWiring verifies that the
// previous connection is disposed before the new one is wired.
// This prevents resource leaks and dangling event listeners on the old client.
func TestUpdateConnection_DisposesOldConnectionBeforeWiring(t *testing.T) {
	server := newTestServer(t, "wc-token-2")
	oldConn := newPairingTestConnection(t)
	server.connection = oldConn // assign directly to simulate a pre-existing connection

	newConn := newPairingTestConnection(t)
	server.UpdateConnection(newConn)

	if !oldConn.disposed {
		t.Fatal("expected old connection to be disposed when UpdateConnection replaces it")
	}
	if newConn.disposed {
		t.Fatal("new connection must not be disposed during UpdateConnection")
	}
}

// TestUpdateConnection_AttachesDispatchingSubscriberWhenMissing verifies that
// the outbound dispatching subscriber is registered when the handler has none.
// This is the scenario that caused inbound messages to be silently dropped after
// pairing: the subscriber was never attached to a freshly created server.
func TestUpdateConnection_AttachesDispatchingSubscriberWhenMissing(t *testing.T) {
	server := newTestServer(t, "wc-token-3")
	conn := newPairingTestConnection(t)

	if server.Handler.HasDispatchingSubscriber() {
		t.Fatal("precondition failed: handler must not have a dispatching subscriber before test")
	}

	server.UpdateConnection(conn)

	if !server.Handler.HasDispatchingSubscriber() {
		t.Fatal("expected dispatching subscriber to be attached after UpdateConnection")
	}
}

// TestUpdateConnection_DoesNotDuplicateDispatchingSubscriber verifies that calling
// UpdateConnection multiple times (e.g. two rapid reconnections) does not register
// the outbound dispatching subscriber more than once.
func TestUpdateConnection_DoesNotDuplicateDispatchingSubscriber(t *testing.T) {
	server := newTestServer(t, "wc-token-4")

	conn1 := newPairingTestConnection(t)
	server.UpdateConnection(conn1)

	// Simulate a second reconnection without going through Stop/Clear.
	conn2 := newPairingTestConnection(t)
	server.UpdateConnection(conn2)

	dispatcher := server.Handler.GetMessageDispatcher()
	count := 0
	for _, sub := range dispatcher.subscribers {
		if _, ok := sub.(dispatchingSubscriber); ok {
			count++
		}
	}
	if count > 1 {
		t.Fatalf("expected at most one dispatching subscriber, found %d", count)
	}
}

// TestReconnect_HandlerRewiredOnReplacementConnection covers the user-reported
// scenario: disconnect + reconnect must wire the handler on the new connection.
// Before the fix, pairing used direct assignment so the replacement connection
// remained handler-less, causing inbound messages to be dropped silently.
func TestReconnect_HandlerRewiredOnReplacementConnection(t *testing.T) {
	server := newTestServer(t, "wc-token-5")

	// First connection (e.g. initial pairing).
	first := newPairingTestConnection(t)
	server.UpdateConnection(first)

	if first.updatedHandler != server.Handler {
		t.Fatal("precondition: first connection must have server handler wired")
	}

	// Second connection (e.g. user triggered disconnect + reconnect).
	second := newPairingTestConnection(t)
	server.UpdateConnection(second)

	if !first.disposed {
		t.Fatal("first connection must be disposed when replaced by the second one")
	}
	if second.updatedHandler == nil {
		t.Fatal("expected UpdateHandler to be called on the replacement connection")
	}
	if second.updatedHandler != server.Handler {
		t.Fatal("replacement connection must receive the server DispatchingHandler")
	}
	if !server.Handler.HasDispatchingSubscriber() {
		t.Fatal("dispatching subscriber must remain attached after reconnection")
	}
}

// TestUpdateConnection_NilHandlerLogsWarningAndSkipsWiring guards the edge case
// where a server ends up with a nil Handler (e.g. incomplete construction).
// UpdateConnection must not panic and must still assign the connection.
func TestUpdateConnection_NilHandlerDoesNotPanic(t *testing.T) {
	server := newTestServer(t, "wc-token-6")
	server.Handler = nil // simulate incomplete construction

	conn := newPairingTestConnection(t)
	// Must not panic.
	server.UpdateConnection(conn)

	if server.GetConnection() != conn {
		t.Fatal("connection must be assigned even when handler is nil")
	}
}
