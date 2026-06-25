package api

import (
	"testing"
	"time"

	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

func TestRecoverAPIValueReturnsFallbackOnPanic(t *testing.T) {
	got := recoverAPIValue(42, "test panic recovery", nil, func() int {
		panic("boom")
	})

	if got != 42 {
		t.Fatalf("expected fallback value 42, got %d", got)
	}
}

func TestBuildFallbackServerSummaryIncludesStableDefaults(t *testing.T) {
	now := time.Now().UTC()
	server := &models.QpServer{
		Token:     "server-token",
		Verified:  true,
		Devel:     true,
		Timestamp: now,
	}
	server.SetWId("5511999999999@s.whatsapp.net")
	server.SetUser("tester@example.com")

	summary := buildFallbackServerSummary(server, serverRuntimeSnapshot{
		state:         whatsapp.Disconnected,
		dispatchCount: 3,
		webhookCount:  2,
		rabbitmqCount: 1,
	})

	if summary["token"] != server.Token {
		t.Fatalf("expected token %q, got %#v", server.Token, summary["token"])
	}
	if summary["state"] != "Disconnected" {
		t.Fatalf("expected disconnected state, got %#v", summary["state"])
	}
	if summary["dispatchCount"] != 3 {
		t.Fatalf("expected dispatchCount 3, got %#v", summary["dispatchCount"])
	}
	if summary["hasDispatching"] != true {
		t.Fatalf("expected hasDispatching true, got %#v", summary["hasDispatching"])
	}
	if summary["user"] != server.GetUser() {
		t.Fatalf("expected user %q, got %#v", server.GetUser(), summary["user"])
	}
}

func TestBuildServerSummaryPrefersLiveWidWhenPersistedWidIsEmpty(t *testing.T) {
	dbServer := &models.QpServer{Token: "server-token"}
	liveState := &models.QpServer{Token: "server-token"}
	liveState.SetWId("5511999999999@s.whatsapp.net")

	liveServer := &models.QpWhatsappServer{
		QpServer: liveState,
	}

	summary := BuildServerSummary(dbServer, liveServer)

	if got := summary["wid"]; got != "5511999999999@s.whatsapp.net" {
		t.Fatalf("expected live wid to be returned, got %#v", got)
	}
}
