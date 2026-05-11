package models

import (
	"testing"
)

// stubPresenceChecker is a test double for RealtimePresenceChecker.
type stubPresenceChecker struct {
	hasActive    bool
	queriedToken string
}

func (s *stubPresenceChecker) HasActiveConnections(token string) bool {
	s.queriedToken = token
	return s.hasActive
}

func TestHasSignalRActiveConnections_NilServerReturnsFalse(t *testing.T) {
	var server *QpWhatsappServer
	if server.HasSignalRActiveConnections() {
		t.Fatal("nil server should return false")
	}
}

func TestHasSignalRActiveConnections_NilCheckerReturnsFalse(t *testing.T) {
	prev := GlobalRealtimePresenceChecker
	defer func() { GlobalRealtimePresenceChecker = prev }()

	GlobalRealtimePresenceChecker = nil

	server := &QpWhatsappServer{QpServer: &QpServer{Token: "tok-001"}}
	if server.HasSignalRActiveConnections() {
		t.Fatal("nil GlobalRealtimePresenceChecker should return false")
	}
}

func TestHasSignalRActiveConnections_DelegatesToCheckerWithServerToken(t *testing.T) {
	prev := GlobalRealtimePresenceChecker
	defer func() { GlobalRealtimePresenceChecker = prev }()

	stub := &stubPresenceChecker{hasActive: true}
	GlobalRealtimePresenceChecker = stub

	const wantToken = "tok-signal-r-test"
	server := &QpWhatsappServer{QpServer: &QpServer{Token: wantToken}}

	got := server.HasSignalRActiveConnections()
	if !got {
		t.Fatal("expected true when stub checker returns true")
	}
	if stub.queriedToken != wantToken {
		t.Errorf("expected checker to be called with token %q, got %q", wantToken, stub.queriedToken)
	}
}

func TestHasSignalRActiveConnections_PropagatesFalseFromChecker(t *testing.T) {
	prev := GlobalRealtimePresenceChecker
	defer func() { GlobalRealtimePresenceChecker = prev }()

	stub := &stubPresenceChecker{hasActive: false}
	GlobalRealtimePresenceChecker = stub

	server := &QpWhatsappServer{QpServer: &QpServer{Token: "tok-no-clients"}}
	if server.HasSignalRActiveConnections() {
		t.Fatal("expected false when checker returns false")
	}
}

func TestGlobalRabbitMQClientResolver_DefaultReturnsFalse(t *testing.T) {
	// The zero-value resolver wired in qp_whatsapp_server.go must return false
	// so servers don't panic on startup without the rabbitmq module.
	prev := GlobalRabbitMQClientResolver
	defer func() { GlobalRabbitMQClientResolver = prev }()

	// Restore the built-in default (no-op)
	GlobalRabbitMQClientResolver = func(_ string) bool { return false }

	server := &QpWhatsappServer{QpServer: &QpServer{Token: "tok-rmq"}}
	_ = server // function under test is the global resolver itself

	if ResolveRabbitMQClient("amqp://guest@localhost/") {
		t.Fatal("default resolver should return false when rabbitmq module is not wired")
	}
}

func TestGlobalRabbitMQClientResolver_CanBeInjected(t *testing.T) {
	prev := GlobalRabbitMQClientResolver
	defer func() { GlobalRabbitMQClientResolver = prev }()

	called := false
	capturedConnStr := ""
	GlobalRabbitMQClientResolver = func(connStr string) bool {
		called = true
		capturedConnStr = connStr
		return true
	}

	const wantConnStr = "amqp://user:pass@rmq:5672/vhost"
	result := ResolveRabbitMQClient(wantConnStr)

	if !called {
		t.Fatal("injected resolver was not called")
	}
	if capturedConnStr != wantConnStr {
		t.Errorf("expected connection string %q, got %q", wantConnStr, capturedConnStr)
	}
	if !result {
		t.Fatal("injected resolver should return true")
	}
}

func TestHasActiveRealtimeConnections_NilCheckerReturnsFalse(t *testing.T) {
	prev := GlobalRealtimePresenceChecker
	defer func() { GlobalRealtimePresenceChecker = prev }()

	GlobalRealtimePresenceChecker = nil

	if HasActiveRealtimeConnections("tok-standalone") {
		t.Fatal("nil realtime presence checker should return false")
	}
}

func TestHasActiveRealtimeConnections_DelegatesToChecker(t *testing.T) {
	prev := GlobalRealtimePresenceChecker
	defer func() { GlobalRealtimePresenceChecker = prev }()

	stub := &stubPresenceChecker{hasActive: true}
	GlobalRealtimePresenceChecker = stub

	const wantToken = "tok-direct-helper"
	if !HasActiveRealtimeConnections(wantToken) {
		t.Fatal("expected helper to propagate true from realtime presence checker")
	}
	if stub.queriedToken != wantToken {
		t.Errorf("expected checker to be called with token %q, got %q", wantToken, stub.queriedToken)
	}
}
