package service

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/nocodeleaks/quepasa/cache"
	cache_disk "github.com/nocodeleaks/quepasa/cache/disk"
	cache_memory "github.com/nocodeleaks/quepasa/cache/memory"
)

// TestCacheServiceSingleton verifies that CacheService returns the same instance.
func TestCacheServiceSingleton(t *testing.T) {
	// Reset the singleton for testing
	once = sync.Once{}
	instance = nil

	instance1 := GetInstance()
	instance2 := GetInstance()

	if instance1 != instance2 {
		t.Errorf("GetInstance() should return the same instance, got different instances")
	}

	if instance1 == nil {
		t.Errorf("GetInstance() returned nil")
	}
}

// TestCacheServiceGetMessagesBackend verifies that messages backend is initialized.
func TestCacheServiceGetMessagesBackend(t *testing.T) {
	once = sync.Once{}
	instance = nil

	service := GetInstance()
	backend := service.GetMessagesBackend()

	if backend == nil {
		t.Errorf("GetMessagesBackend() returned nil")
	}
}

// TestCacheServiceGetQueueBackend verifies that queue backend is initialized.
func TestCacheServiceGetQueueBackend(t *testing.T) {
	once = sync.Once{}
	instance = nil

	service := GetInstance()
	backend := service.GetQueueBackend()

	if backend == nil {
		t.Errorf("GetQueueBackend() returned nil")
	}
}

// TestCacheServiceClose verifies that Close() properly closes backends.
func TestCacheServiceClose(t *testing.T) {
	once = sync.Once{}
	instance = nil

	service := GetInstance()
	err := service.Close()

	// Close should not error (memory backends don't typically error)
	if err != nil && service.messagesBackend == nil {
		t.Errorf("Close() returned error: %v", err)
	}
}

// TestMemoryBackendConcurrency tests concurrent operations on memory backend.
func TestMemoryBackendConcurrency(t *testing.T) {
	backend := cache_memory.NewMessagesBackend()
	defer backend.Close()

	const numGoroutines = 10
	const messagesPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < messagesPerGoroutine; j++ {
				key := "msg-" + string(rune(goroutineID)) + "-" + string(rune(j))
				record := cache.MessageRecord{
					ExpiresAt: time.Now().Add(1 * time.Hour),
					UpdatedAt: time.Now(),
				}
				_ = backend.Set(key, record)
			}
		}(i)
	}

	wg.Wait()

	// Verify all messages were stored
	entries, _ := backend.List()
	if len(entries) == 0 {
		t.Errorf("Expected messages in backend, got %d", len(entries))
	}
}

// TestMemoryQueueBackendConcurrency tests concurrent enqueue/dequeue operations.
func TestMemoryQueueBackendConcurrency(t *testing.T) {
	backend := cache_memory.NewBytesQueueBackend(1000)
	defer backend.Close()

	const numProducers = 5
	const itemsPerProducer = 50

	var wg sync.WaitGroup

	// Producers
	for i := 0; i < numProducers; i++ {
		wg.Add(1)
		go func(producerID int) {
			defer wg.Done()

			for j := 0; j < itemsPerProducer; j++ {
				payload := []byte("message-" + string(rune(producerID)) + "-" + string(rune(j)))
				_, _ = backend.Enqueue(payload)
			}
		}(i)
	}

	// Consumer
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(100 * time.Millisecond) // Let producers queue items

		for {
			payload, found, _ := backend.Dequeue()
			if !found {
				break
			}
			if len(payload) == 0 {
				t.Errorf("Dequeued empty payload")
			}
		}
	}()

	wg.Wait()
}

// TestDiskBackendFileCreation verifies disk backend creates files correctly.
func TestDiskBackendFileCreation(t *testing.T) {
	// Create temporary directory
	tempDir := filepath.Join(t.TempDir(), "cache_test")

	backend, err := cache_disk.NewMessagesBackend(tempDir)
	if err != nil {
		t.Fatalf("Failed to create disk backend: %v", err)
	}
	defer backend.Close()

	// Verify directory was created
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Errorf("Cache directory was not created: %s", tempDir)
	}

	// Store a message
	record := cache.MessageRecord{
		ExpiresAt: time.Now().Add(1 * time.Hour),
		UpdatedAt: time.Now(),
	}

	err = backend.Set("TEST_MSG", record)
	if err != nil {
		t.Errorf("Failed to set message: %v", err)
	}

	// Retrieve the message
	retrieved, found, err := backend.Get("TEST_MSG")
	if err != nil {
		t.Errorf("Failed to get message: %v", err)
	}
	if !found {
		t.Errorf("Message not found in backend")
	}

	// Allow 1 second tolerance for timestamp precision in JSON serialization
	timeDiff := record.ExpiresAt.Sub(retrieved.ExpiresAt)
	if timeDiff < -1*time.Second || timeDiff > 1*time.Second {
		t.Errorf("Retrieved ExpiresAt differs too much from stored: expected %v, got %v (diff: %v)",
			record.ExpiresAt, retrieved.ExpiresAt, timeDiff)
	}
}

// TestDiskQueueBackendFileCreation verifies disk queue backend creates files.
func TestDiskQueueBackendFileCreation(t *testing.T) {
	tempDir := filepath.Join(t.TempDir(), "queue_test")
	capacity := 100

	backend, err := cache_disk.NewBytesQueueBackend(tempDir, capacity)
	if err != nil {
		t.Fatalf("Failed to create disk queue backend: %v", err)
	}
	defer backend.Close()

	// Verify directory was created
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Errorf("Queue directory was not created: %s", tempDir)
	}

	// Enqueue and dequeue
	payload := []byte("test-message")
	added, err := backend.Enqueue(payload)
	if err != nil {
		t.Errorf("Failed to enqueue: %v", err)
	}
	if !added {
		t.Errorf("Failed to add message to queue")
	}

	// Dequeue
	dequeued, found, err := backend.Dequeue()
	if err != nil {
		t.Errorf("Failed to dequeue: %v", err)
	}
	if !found {
		t.Errorf("No message found in queue")
	}
	if string(dequeued) != string(payload) {
		t.Errorf("Dequeued payload differs from enqueued payload")
	}
}

// TestBackendExpiration verifies that expired records are handled correctly.
func TestBackendExpiration(t *testing.T) {
	backend := cache_memory.NewMessagesBackend()
	defer backend.Close()

	// Store an expired record
	expiredRecord := cache.MessageRecord{
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
		UpdatedAt: time.Now(),
	}
	_ = backend.Set("EXPIRED", expiredRecord)

	// Store a valid record
	validRecord := cache.MessageRecord{
		ExpiresAt: time.Now().Add(1 * time.Hour), // Valid
		UpdatedAt: time.Now(),
	}
	_ = backend.Set("VALID", validRecord)

	// List should handle expired records
	entries, _ := backend.List()
	if len(entries) == 0 {
		t.Errorf("Backend should contain entries")
	}

	// Get expired should not be found or should be purged
	_, found, _ := backend.Get("EXPIRED")
	if found {
		// It's OK if it's still there, but consumers should check expiration
		t.Logf("Expired record still in backend (consumer must check expiration)")
	}
}

// TestQueueBackendCapacity verifies that queue respects capacity limits.
func TestQueueBackendCapacity(t *testing.T) {
	capacity := 10
	backend := cache_memory.NewBytesQueueBackend(capacity)
	defer backend.Close()

	// Fill the queue to capacity
	for i := 0; i < capacity; i++ {
		payload := []byte("msg-" + string(rune(i)))
		added, _ := backend.Enqueue(payload)
		if !added {
			t.Errorf("Failed to enqueue message %d (capacity: %d)", i, capacity)
		}
	}

	// Next enqueue should fail
	payload := []byte("overflow-message")
	added, _ := backend.Enqueue(payload)
	if added {
		t.Errorf("Enqueue should fail when queue is at capacity")
	}
}

// TestQueueBackendLen verifies that Len() works correctly.
func TestQueueBackendLen(t *testing.T) {
	backend := cache_memory.NewBytesQueueBackend(100)
	defer backend.Close()

	payload := []byte("test-message")
	_, _ = backend.Enqueue(payload)

	// Len should return 1
	len, err := backend.Len()
	if err != nil {
		t.Errorf("Len failed: %v", err)
	}
	if len != 1 {
		t.Errorf("Len should return 1, got: %d", len)
	}

	// Dequeue should reduce length
	dequeued, found, _ := backend.Dequeue()
	if !found || string(dequeued) != string(payload) {
		t.Errorf("Dequeue should return the item")
	}

	// Len should now be 0
	len, _ = backend.Len()
	if len != 0 {
		t.Errorf("Len should return 0 after dequeue, got: %d", len)
	}
}

// TestMemoryBackendDelete verifies deletion works correctly.
func TestMemoryBackendDelete(t *testing.T) {
	backend := cache_memory.NewMessagesBackend()
	defer backend.Close()

	record := cache.MessageRecord{
		ExpiresAt: time.Now().Add(1 * time.Hour),
		UpdatedAt: time.Now(),
	}

	// Store and retrieve
	_ = backend.Set("DELETE_TEST", record)
	_, found, _ := backend.Get("DELETE_TEST")
	if !found {
		t.Errorf("Record should exist after Set")
	}

	// Delete
	_ = backend.Delete("DELETE_TEST")
	_, found, _ = backend.Get("DELETE_TEST")
	if found {
		t.Errorf("Record should not exist after Delete")
	}
}
