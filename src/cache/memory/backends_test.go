package memory

import (
	"testing"
	"time"

	"github.com/nocodeleaks/quepasa/cache"
)

// TestMemoryMessagesBackendBasic verifies basic set/get operations.
func TestMemoryMessagesBackendBasic(t *testing.T) {
	backend := NewMessagesBackend()
	defer backend.Close()

	record := cache.MessageRecord{
		ExpiresAt: time.Now().Add(1 * time.Hour),
		UpdatedAt: time.Now(),
	}

	err := backend.Set("TEST_KEY", record)
	if err != nil {
		t.Errorf("Set() failed: %v", err)
	}

	retrieved, found, err := backend.Get("TEST_KEY")
	if err != nil {
		t.Errorf("Get() failed: %v", err)
	}
	if !found {
		t.Errorf("Get() should find the key")
	}
	if retrieved.ExpiresAt != record.ExpiresAt {
		t.Errorf("Retrieved record should match stored record")
	}
}

// TestMemoryMessagesBackendDelete verifies deletion.
func TestMemoryMessagesBackendDelete(t *testing.T) {
	backend := NewMessagesBackend()
	defer backend.Close()

	record := cache.MessageRecord{
		ExpiresAt: time.Now().Add(1 * time.Hour),
		UpdatedAt: time.Now(),
	}

	backend.Set("TO_DELETE", record)

	err := backend.Delete("TO_DELETE")
	if err != nil {
		t.Errorf("Delete() failed: %v", err)
	}

	_, found, _ := backend.Get("TO_DELETE")
	if found {
		t.Errorf("Key should be deleted")
	}
}

// TestMemoryMessagesBackendList verifies listing.
func TestMemoryMessagesBackendList(t *testing.T) {
	backend := NewMessagesBackend()
	defer backend.Close()

	// Add multiple records
	for i := 0; i < 5; i++ {
		record := cache.MessageRecord{
			ExpiresAt: time.Now().Add(1 * time.Hour),
			UpdatedAt: time.Now(),
		}
		key := "KEY_" + string(rune(i))
		backend.Set(key, record)
	}

	entries, err := backend.List()
	if err != nil {
		t.Errorf("List() failed: %v", err)
	}
	if len(entries) != 5 {
		t.Errorf("List() should return 5 entries, got: %d", len(entries))
	}
}

// TestMemoryQueueBackendBasic verifies queue operations.
func TestMemoryQueueBackendBasic(t *testing.T) {
	backend := NewBytesQueueBackend(100)
	defer backend.Close()

	payload := []byte("test message")

	// Enqueue
	added, err := backend.Enqueue(payload)
	if err != nil {
		t.Errorf("Enqueue() failed: %v", err)
	}
	if !added {
		t.Errorf("Enqueue() should return true")
	}

	// Dequeue
	dequeued, found, err := backend.Dequeue()
	if err != nil {
		t.Errorf("Dequeue() failed: %v", err)
	}
	if !found {
		t.Errorf("Dequeue() should find item")
	}
	if string(dequeued) != string(payload) {
		t.Errorf("Dequeued payload should match enqueued")
	}
}

// TestMemoryQueueBackendLen verifies len operation.
func TestMemoryQueueBackendLen(t *testing.T) {
	backend := NewBytesQueueBackend(100)
	defer backend.Close()

	// Initially empty
	len, _ := backend.Len()
	if len != 0 {
		t.Errorf("Len() should be 0 initially, got: %d", len)
	}

	// Add items
	for i := 0; i < 3; i++ {
		backend.Enqueue([]byte("item"))
	}

	len, _ = backend.Len()
	if len != 3 {
		t.Errorf("Len() should be 3, got: %d", len)
	}
}

// TestMemoryQueueBackendEmpty verifies empty queue detection.
func TestMemoryQueueBackendEmpty(t *testing.T) {
	backend := NewBytesQueueBackend(100)
	defer backend.Close()

	// Initially empty
	len, _ := backend.Len()
	if len != 0 {
		t.Errorf("Queue should be empty initially")
	}

	// Dequeue from empty should not find
	_, found, _ := backend.Dequeue()
	if found {
		t.Errorf("Dequeue from empty queue should not find item")
	}
}

// TestMemoryMessagesBackendClose verifies close operation.
func TestMemoryMessagesBackendClose(t *testing.T) {
	backend := NewMessagesBackend()

	err := backend.Close()
	if err != nil {
		t.Errorf("Close() should not error: %v", err)
	}
}

// TestMemoryQueueBackendClose verifies close operation.
func TestMemoryQueueBackendClose(t *testing.T) {
	backend := NewBytesQueueBackend(100)

	err := backend.Close()
	if err != nil {
		t.Errorf("Close() should not error: %v", err)
	}
}
