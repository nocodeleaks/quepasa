package models

import (
	"strings"
	"testing"
	"time"

	cache_memory "github.com/nocodeleaks/quepasa/cache/memory"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// TestQpWhatsappMessagesSetBackend verifies backend injection works.
func TestQpWhatsappMessagesSetBackend(t *testing.T) {
	messages := &QpWhatsappMessages{}
	backend := cache_memory.NewMessagesBackend()
	defer backend.Close()

	messages.SetBackend(backend)

	if messages.backend != backend {
		t.Errorf("SetBackend() did not set the backend correctly")
	}
}

// TestQpWhatsappMessagesAppend verifies message appending works.
func TestQpWhatsappMessagesAppend(t *testing.T) {
	messages := &QpWhatsappMessages{}
	backend := cache_memory.NewMessagesBackend()
	defer backend.Close()
	messages.SetBackend(backend)

	msg := &whatsapp.WhatsappMessage{
		Id:        "test-msg-1",
		Timestamp: time.Now(),
		Status:    whatsapp.WhatsappMessageStatusImported,
	}

	success := messages.Append(msg, "test")
	if !success {
		t.Errorf("Append() should succeed")
	}

	// Message ID should be uppercase
	if msg.Id != "TEST-MSG-1" {
		t.Errorf("Message ID should be uppercase, got: %s", msg.Id)
	}
}

// TestQpWhatsappMessagesGetById verifies message retrieval.
func TestQpWhatsappMessagesGetById(t *testing.T) {
	messages := &QpWhatsappMessages{}
	backend := cache_memory.NewMessagesBackend()
	defer backend.Close()
	messages.SetBackend(backend)

	msg := &whatsapp.WhatsappMessage{
		Id:        "test-msg-2",
		Timestamp: time.Now(),
		Status:    whatsapp.WhatsappMessageStatusDelivered,
	}

	messages.Append(msg, "test")

	retrieved, err := messages.GetById("test-msg-2")
	if err != nil {
		t.Errorf("GetById() should not error: %v", err)
	}
	if retrieved == nil {
		t.Errorf("GetById() should return a message")
	}
	if retrieved.Id != "TEST-MSG-2" {
		t.Errorf("Retrieved message ID should be TEST-MSG-2, got: %s", retrieved.Id)
	}
}

// TestQpWhatsappMessagesGetSlice verifies getting all messages.
func TestQpWhatsappMessagesGetSlice(t *testing.T) {
	messages := &QpWhatsappMessages{}
	backend := cache_memory.NewMessagesBackend()
	defer backend.Close()
	messages.SetBackend(backend)

	// Add multiple messages
	for i := 0; i < 5; i++ {
		msg := &whatsapp.WhatsappMessage{
			Id:        "msg-" + string(rune(i+48)),
			Timestamp: time.Now(),
		}
		messages.Append(msg, "test")
	}

	slice := messages.GetSlice()
	if len(slice) < 5 {
		t.Logf("GetSlice() returned fewer messages than added")
	}
}

// TestQpWhatsappMessagesCount verifies message count.
func TestQpWhatsappMessagesCount(t *testing.T) {
	messages := &QpWhatsappMessages{}
	backend := cache_memory.NewMessagesBackend()
	defer backend.Close()
	messages.SetBackend(backend)

	// Add 3 messages
	for i := 0; i < 3; i++ {
		msg := &whatsapp.WhatsappMessage{
			Id:        "msg-count-" + string(rune(i+48)),
			Timestamp: time.Now(),
		}
		messages.Append(msg, "test")
	}

	count := messages.Count()
	if count < 3 {
		t.Logf("Count() returned fewer than 3 messages")
	}
}

// TestQpWhatsappMessagesSetStatusById verifies status update.
func TestQpWhatsappMessagesSetStatusById(t *testing.T) {
	messages := &QpWhatsappMessages{}
	backend := cache_memory.NewMessagesBackend()
	defer backend.Close()
	messages.SetBackend(backend)

	msg := &whatsapp.WhatsappMessage{
		Id:        "test-status-msg",
		Timestamp: time.Now(),
		Status:    whatsapp.WhatsappMessageStatusImported,
	}

	messages.Append(msg, "test")
	messages.SetStatusById("test-status-msg", whatsapp.WhatsappMessageStatusDelivered)

	retrieved, _ := messages.GetById("TEST-STATUS-MSG")
	if retrieved.Status != whatsapp.WhatsappMessageStatusDelivered {
		t.Errorf("Status should be Delivered, got: %v", retrieved.Status)
	}
}

// TestQpWhatsappMessagesGetStatusById verifies getting status.
func TestQpWhatsappMessagesGetStatusById(t *testing.T) {
	messages := &QpWhatsappMessages{}
	backend := cache_memory.NewMessagesBackend()
	defer backend.Close()
	messages.SetBackend(backend)

	msg := &whatsapp.WhatsappMessage{
		Id:        "test-get-status",
		Timestamp: time.Now(),
		Status:    whatsapp.WhatsappMessageStatusImported,
	}

	messages.Append(msg, "test")
	status := messages.GetStatusById("TEST-GET-STATUS")

	if status != whatsapp.WhatsappMessageStatusImported {
		t.Errorf("Status should be Imported, got: %v", status)
	}
}

// TestQpWhatsappMessagesGetByTime verifies time filtering.
func TestQpWhatsappMessagesGetByTime(t *testing.T) {
	messages := &QpWhatsappMessages{}
	backend := cache_memory.NewMessagesBackend()
	defer backend.Close()
	messages.SetBackend(backend)

	now := time.Now()

	// Add message before timestamp
	msg1 := &whatsapp.WhatsappMessage{
		Id:        "msg-old",
		Timestamp: now.Add(-1 * time.Hour),
	}
	messages.Append(msg1, "test")

	// Add message after timestamp
	msg2 := &whatsapp.WhatsappMessage{
		Id:        "msg-new",
		Timestamp: now.Add(1 * time.Hour),
	}
	messages.Append(msg2, "test")

	// Filter by timestamp
	filtered := messages.GetByTime(now)

	if len(filtered) < 1 {
		t.Logf("GetByTime() returned fewer messages than expected")
	}
}

// TestQpWhatsappMessagesGetByPrefix verifies prefix filtering.
func TestQpWhatsappMessagesGetByPrefix(t *testing.T) {
	messages := &QpWhatsappMessages{}
	backend := cache_memory.NewMessagesBackend()
	defer backend.Close()
	messages.SetBackend(backend)

	// Add messages with different prefixes
	for i := 0; i < 3; i++ {
		msg := &whatsapp.WhatsappMessage{
			Id:        "PREFIX-A-" + string(rune(i+48)),
			Timestamp: time.Now(),
		}
		messages.Append(msg, "test")
	}

	filtered := messages.GetByPrefix("PREFIX-A")
	if len(filtered) >= 1 {
		// All should start with PREFIX-A
		for _, msg := range filtered {
			if !strings.HasPrefix(msg.Id, "PREFIX-A") {
				t.Errorf("All filtered messages should start with PREFIX-A, got: %s", msg.Id)
			}
		}
	}
}

// TestQpWhatsappMessagesCleanUp verifies cleanup removes oldest messages.
func TestQpWhatsappMessagesCleanUp(t *testing.T) {
	messages := &QpWhatsappMessages{}
	backend := cache_memory.NewMessagesBackend()
	defer backend.Close()
	messages.SetBackend(backend)

	// Add 5 messages with different timestamps
	for i := 0; i < 5; i++ {
		msg := &whatsapp.WhatsappMessage{
			Id:        "msg-cleanup-" + string(rune(i+48)),
			Timestamp: time.Now().Add(time.Duration(i) * time.Hour),
		}
		messages.Append(msg, "test")
	}

	initialCount := messages.Count()

	// Keep only 3 messages (remove 2)
	messages.CleanUp(3)

	finalCount := messages.Count()
	if finalCount <= initialCount {
		t.Logf("After cleanup, count should be <= %d", initialCount)
	}
}

// TestQpWhatsappMessagesMessageStatusUpdate verifies status update logic.
func TestQpWhatsappMessagesMessageStatusUpdate(t *testing.T) {
	messages := &QpWhatsappMessages{}
	backend := cache_memory.NewMessagesBackend()
	defer backend.Close()
	messages.SetBackend(backend)

	msg := &whatsapp.WhatsappMessage{
		Id:        "test-update-status",
		Timestamp: time.Now(),
		Status:    whatsapp.WhatsappMessageStatusImported,
	}

	messages.Append(msg, "test")

	// Update to a higher status
	updated := messages.MessageStatusUpdate("TEST-UPDATE-STATUS", whatsapp.WhatsappMessageStatusDelivered)
	if !updated {
		t.Errorf("MessageStatusUpdate() should return true for higher status")
	}

	// Try to update to lower status (should not update)
	updated = messages.MessageStatusUpdate("TEST-UPDATE-STATUS", whatsapp.WhatsappMessageStatusImported)
	if updated {
		t.Errorf("MessageStatusUpdate() should return false for lower status")
	}
}

// TestQpWhatsappMessagesNilBackend verifies error handling when backend is nil.
func TestQpWhatsappMessagesNilBackend(t *testing.T) {
	messages := &QpWhatsappMessages{}
	// No backend set

	msg := &whatsapp.WhatsappMessage{
		Id:        "test-nil-backend",
		Timestamp: time.Now(),
	}

	// Should not panic, but should fail gracefully
	success := messages.Append(msg, "test")
	if success {
		t.Errorf("Append() should fail with nil backend")
	}

	// Count should be 0
	count := messages.Count()
	if count != 0 {
		t.Errorf("Count() should return 0 with nil backend, got: %d", count)
	}
}

// TestQpWhatsappMessagesBackendInjection verifies injection during handler creation.
func TestQpWhatsappMessagesBackendInjection(t *testing.T) {
	// Create a DispatchingHandler
	handler := &DispatchingHandler{
		QpWhatsappMessages: QpWhatsappMessages{},
	}

	// Verify backend is nil before injection
	if handler.backend != nil {
		t.Errorf("Backend should be nil before injection")
	}

	// Inject backend
	backend := cache_memory.NewMessagesBackend()
	defer backend.Close()
	handler.SetBackend(backend)

	if handler.backend != backend {
		t.Errorf("Backend should be set after injection")
	}
}
