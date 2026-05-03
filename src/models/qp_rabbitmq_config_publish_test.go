package models

import (
	"testing"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// mockRabbitMQClient is a minimal test double for RabbitMQPublisherClient.
type mockRabbitMQClient struct {
	ensureErr    error
	publishedKey string
}

func (m *mockRabbitMQClient) EnsureExchangeAndQueues() error {
	return m.ensureErr
}

func (m *mockRabbitMQClient) PublishQuePasaMessage(routingKey string, _ any) {
	m.publishedKey = routingKey
}

// TestPublishMessage_UsesInjectedClientResolver verifies that WithClientResolver
// causes PublishMessage to use the injected resolver instead of the global one.
func TestPublishMessage_UsesInjectedClientResolver(t *testing.T) {
	mock := &mockRabbitMQClient{}
	cfg := newRabbitMQConfig().WithClientResolver(func(_ string) RabbitMQPublisherClient {
		return mock
	})

	msg := &whatsapp.WhatsappMessage{
		Id:   "test-inject-001",
		Type: whatsapp.TextMessageType,
	}

	err := cfg.PublishMessage(msg)
	if err != nil {
		t.Fatalf("expected no error from PublishMessage, got: %v", err)
	}
	if mock.publishedKey == "" {
		t.Error("expected mock client to be called via injected resolver, but PublishQuePasaMessage was never invoked")
	}
}

// TestPublishMessage_InjectedResolverNilClientReturnsError verifies that a nil
// client returned by the injected resolver is handled gracefully (returns an error).
func TestPublishMessage_InjectedResolverNilClientReturnsError(t *testing.T) {
	cfg := newRabbitMQConfig().WithClientResolver(func(_ string) RabbitMQPublisherClient {
		return nil
	})

	msg := &whatsapp.WhatsappMessage{
		Id:   "test-nil-client-001",
		Type: whatsapp.TextMessageType,
	}

	err := cfg.PublishMessage(msg)
	if err == nil {
		t.Error("expected error when injected resolver returns nil client, got nil")
	}
}
