package models

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nocodeleaks/quepasa/whatsapp"
)

func TestDispatchingWebhookBlockedAfterSequentialFailures(t *testing.T) {
	hitCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hitCount++
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	}))
	defer server.Close()

	failure := time.Now().UTC().Add(-49 * time.Hour)
	dispatching := &QpDispatching{
		ConnectionString: server.URL,
		Type:             DispatchingTypeWebhook,
		Failure:          &failure,
		Wid:              "test@whatsapp",
	}

	err := dispatching.PostWebhook(&whatsapp.WhatsappMessage{Id: "blocked-message"})
	if err != nil {
		t.Fatalf("expected blocked webhook to return no transport error, got %v", err)
	}

	if hitCount != 0 {
		t.Fatalf("expected blocked webhook to skip HTTP delivery, got %d request(s)", hitCount)
	}

	if dispatching.Failure == nil || !dispatching.Failure.Equal(failure) {
		t.Fatal("expected failure timestamp to remain unchanged while blocked")
	}

	if dispatching.Success != nil {
		t.Fatal("expected success timestamp to remain nil while blocked")
	}
}

func TestDispatchingWebhookRecentSuccessAllowsDelivery(t *testing.T) {
	hitCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hitCount++
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	}))
	defer server.Close()

	failure := time.Now().UTC().Add(-49 * time.Hour)
	success := time.Now().UTC().Add(-1 * time.Hour)
	dispatching := &QpDispatching{
		ConnectionString: server.URL,
		Type:             DispatchingTypeWebhook,
		Failure:          &failure,
		Success:          &success,
		Wid:              "test@whatsapp",
	}

	err := dispatching.PostWebhook(&whatsapp.WhatsappMessage{Id: "retry-message"})
	if err != nil {
		t.Fatalf("expected delivery after recent success, got %v", err)
	}

	if hitCount != 1 {
		t.Fatalf("expected webhook delivery to happen once, got %d request(s)", hitCount)
	}

	if dispatching.Failure != nil {
		t.Fatal("expected failure timestamp to be cleared after successful retry")
	}

	if dispatching.Success == nil {
		t.Fatal("expected success timestamp to be refreshed after successful retry")
	}
}
