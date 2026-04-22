package api

import (
	"testing"
	"time"

	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

func TestRecoverSPAValueReturnsFallbackOnPanic(t *testing.T) {
	got := recoverSPAValue(42, "test panic recovery", nil, func() int {
		panic("boom")
	})

	if got != 42 {
		t.Fatalf("expected fallback value 42, got %d", got)
	}
}

func TestBuildSPAFallbackServerSummaryIncludesStableDefaults(t *testing.T) {
	now := time.Now().UTC()
	server := &models.QpServer{
		Token:     "server-token",
		Wid:       "5511999999999@s.whatsapp.net",
		Verified:  true,
		Devel:     true,
		User:      "tester@example.com",
		Timestamp: now,
	}

	summary := buildSPAFallbackServerSummary(server, spaServerRuntimeSnapshot{
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
	if summary["user"] != server.User {
		t.Fatalf("expected user %q, got %#v", server.User, summary["user"])
	}
}
