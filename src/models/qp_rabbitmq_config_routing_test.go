package models

import (
	"testing"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// resetRoutingKeyGlobals restores the default routing key values after each test.
func resetRoutingKeyGlobals(t *testing.T) {
	t.Helper()
	prev := [3]string{GlobalRabbitMQRoutingKeyHistory, GlobalRabbitMQRoutingKeyEvents, GlobalRabbitMQRoutingKeyProd}
	t.Cleanup(func() {
		GlobalRabbitMQRoutingKeyHistory = prev[0]
		GlobalRabbitMQRoutingKeyEvents = prev[1]
		GlobalRabbitMQRoutingKeyProd = prev[2]
	})
}

func newRabbitMQConfig() *QpRabbitMQConfig {
	return &QpRabbitMQConfig{
		ConnectionString: "amqp://guest:guest@localhost:5672/",
		ExchangeName:     "quepasa.exchange",
		RoutingKey:       "prod",
	}
}

func TestDetermineRoutingKey_HistorySyncRoutesToHistory(t *testing.T) {
	resetRoutingKeyGlobals(t)
	cfg := newRabbitMQConfig()
	msg := &whatsapp.WhatsappMessage{FromHistory: true}

	got := cfg.DetermineRoutingKey(msg)
	if got != GlobalRabbitMQRoutingKeyHistory {
		t.Errorf("expected %q for history message, got %q", GlobalRabbitMQRoutingKeyHistory, got)
	}
}

func TestDetermineRoutingKey_UnhandledTypeRoutesToEvents(t *testing.T) {
	resetRoutingKeyGlobals(t)
	cfg := newRabbitMQConfig()
	msg := &whatsapp.WhatsappMessage{Type: whatsapp.UnhandledMessageType}

	got := cfg.DetermineRoutingKey(msg)
	if got != GlobalRabbitMQRoutingKeyEvents {
		t.Errorf("expected %q for unhandled message, got %q", GlobalRabbitMQRoutingKeyEvents, got)
	}
}

func TestDetermineRoutingKey_ReadReceiptIdRoutesToEvents(t *testing.T) {
	resetRoutingKeyGlobals(t)
	cfg := newRabbitMQConfig()
	msg := &whatsapp.WhatsappMessage{
		Id:   "readreceipt",
		Type: whatsapp.TextMessageType,
	}

	got := cfg.DetermineRoutingKey(msg)
	if got != GlobalRabbitMQRoutingKeyEvents {
		t.Errorf("expected %q for readreceipt message, got %q", GlobalRabbitMQRoutingKeyEvents, got)
	}
}

func TestDetermineRoutingKey_SystemMessageRoutesToProd(t *testing.T) {
	resetRoutingKeyGlobals(t)
	cfg := newRabbitMQConfig()
	msg := &whatsapp.WhatsappMessage{Type: whatsapp.SystemMessageType}

	got := cfg.DetermineRoutingKey(msg)
	if got != GlobalRabbitMQRoutingKeyProd {
		t.Errorf("expected %q for system message, got %q", GlobalRabbitMQRoutingKeyProd, got)
	}
}

func TestDetermineRoutingKey_CallMessageRoutesToProd(t *testing.T) {
	resetRoutingKeyGlobals(t)
	cfg := newRabbitMQConfig()
	msg := &whatsapp.WhatsappMessage{Type: whatsapp.CallMessageType}

	got := cfg.DetermineRoutingKey(msg)
	if got != GlobalRabbitMQRoutingKeyProd {
		t.Errorf("expected %q for call message, got %q", GlobalRabbitMQRoutingKeyProd, got)
	}
}

func TestDetermineRoutingKey_RevokeMessageRoutesToProd(t *testing.T) {
	resetRoutingKeyGlobals(t)
	cfg := newRabbitMQConfig()
	msg := &whatsapp.WhatsappMessage{Type: whatsapp.RevokeMessageType}

	got := cfg.DetermineRoutingKey(msg)
	if got != GlobalRabbitMQRoutingKeyProd {
		t.Errorf("expected %q for revoke message, got %q", GlobalRabbitMQRoutingKeyProd, got)
	}
}

func TestDetermineRoutingKey_NormalTextMessageRoutesToProd(t *testing.T) {
	resetRoutingKeyGlobals(t)
	cfg := newRabbitMQConfig()
	msg := &whatsapp.WhatsappMessage{
		Id:   "msg-001",
		Type: whatsapp.TextMessageType,
	}

	got := cfg.DetermineRoutingKey(msg)
	if got != GlobalRabbitMQRoutingKeyProd {
		t.Errorf("expected %q for normal text message, got %q", GlobalRabbitMQRoutingKeyProd, got)
	}
}

func TestDetermineRoutingKey_ContactWithEditedAttachmentRoutesToEvents(t *testing.T) {
	resetRoutingKeyGlobals(t)
	cfg := newRabbitMQConfig()
	msg := &whatsapp.WhatsappMessage{
		Type:       whatsapp.ContactMessageType,
		Edited:     true,
		Attachment: &whatsapp.WhatsappAttachment{},
	}

	got := cfg.DetermineRoutingKey(msg)
	if got != GlobalRabbitMQRoutingKeyEvents {
		t.Errorf("expected %q for edited contact with attachment, got %q", GlobalRabbitMQRoutingKeyEvents, got)
	}
}

func TestDetermineRoutingKey_ContactWithoutAttachmentRoutesToProd(t *testing.T) {
	resetRoutingKeyGlobals(t)
	cfg := newRabbitMQConfig()
	msg := &whatsapp.WhatsappMessage{
		Type:   whatsapp.ContactMessageType,
		Edited: true,
		// no attachment
	}

	got := cfg.DetermineRoutingKey(msg)
	if got != GlobalRabbitMQRoutingKeyProd {
		t.Errorf("expected %q for edited contact without attachment, got %q", GlobalRabbitMQRoutingKeyProd, got)
	}
}

// TestDetermineRoutingKey_HistoryTakesPriorityOverType verifies that FromHistory
// short-circuits all subsequent type checks.
func TestDetermineRoutingKey_HistoryTakesPriorityOverType(t *testing.T) {
	resetRoutingKeyGlobals(t)
	cfg := newRabbitMQConfig()
	msg := &whatsapp.WhatsappMessage{
		Type:        whatsapp.UnhandledMessageType,
		FromHistory: true,
	}

	got := cfg.DetermineRoutingKey(msg)
	if got != GlobalRabbitMQRoutingKeyHistory {
		t.Errorf("expected history routing key to win over type check, got %q", got)
	}
}
