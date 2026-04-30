package models

import (
	"testing"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

type nonDispatchingTestSubscriber struct{}

func (nonDispatchingTestSubscriber) HandleDispatching(*whatsapp.WhatsappMessage) {}

type captureLifecyclePublisher struct {
	event *DispatchingLifecycleEvent
}

func (c *captureLifecyclePublisher) PublishLifecycle(event *DispatchingLifecycleEvent) {
	c.event = event
}

func TestHasDispatchingSubscriber_OnlyForMarkedSubscriber(t *testing.T) {
	handler := &DispatchingHandler{}

	handler.Register(nonDispatchingTestSubscriber{})
	if handler.HasDispatchingSubscriber() {
		t.Fatalf("expected no dispatching subscriber when only non-marked subscriber is registered")
	}

	server := &QpWhatsappServer{QpServer: &QpServer{Token: "test-token"}}
	handler.Register(NewOutboundDispatchingSubscriber(server))
	if !handler.HasDispatchingSubscriber() {
		t.Fatalf("expected dispatching subscriber to be detected after registering marked subscriber")
	}
}

func TestPublishRealtimeLifecycle_UsesInjectedPublisher(t *testing.T) {
	previousPublisher := GlobalDispatchingLifecyclePublisher
	defer func() {
		GlobalDispatchingLifecyclePublisher = previousPublisher
	}()

	capture := &captureLifecyclePublisher{}
	GlobalDispatchingLifecyclePublisher = capture

	server := &QpWhatsappServer{QpServer: &QpServer{Token: "token-123"}}
	server.QpServer.SetWId("5511999999999@s.whatsapp.net")
	server.QpServer.SetUser("integration-user")
	server.Verified = true

	handler := &DispatchingHandler{server: server}
	lifecycle := NewLifecycleHandler(handler)
	lifecycle.publishRealtimeLifecycle("connected", "", "")

	if capture.event == nil {
		t.Fatalf("expected lifecycle event to be published")
	}

	if capture.event.Kind != "connected" {
		t.Fatalf("expected lifecycle kind 'connected', got %q", capture.event.Kind)
	}

	if capture.event.Token != "token-123" {
		t.Fatalf("expected token 'token-123', got %q", capture.event.Token)
	}

	if capture.event.Wid != "5511999999999@s.whatsapp.net" {
		t.Fatalf("expected wid from server, got %q", capture.event.Wid)
	}

	if capture.event.User != "integration-user" {
		t.Fatalf("expected user from server, got %q", capture.event.User)
	}
}
