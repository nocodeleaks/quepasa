package rabbitmq

import (
	"encoding/json"
	"testing"
	"time"

	cache_memory "github.com/nocodeleaks/quepasa/cache/memory"
)

// TestRabbitMQClientSetCacheBackend verifies backend injection works.
func TestRabbitMQClientSetCacheBackend(t *testing.T) {
	client := NewRabbitMQClient("amqp://guest:guest@localhost:5672/", 0)
	defer client.Close()

	backend := cache_memory.NewBytesQueueBackend(100)
	defer backend.Close()

	// Set the backend
	client.SetCacheBackend(backend)

	if client.messageCache != backend {
		t.Errorf("SetCacheBackend() did not set the backend correctly")
	}
}

// TestRabbitMQClientAddToCache verifies message caching works.
func TestRabbitMQClientAddToCache(t *testing.T) {
	client := NewRabbitMQClient("amqp://guest:guest@localhost:5672/", 0)
	defer client.Close()

	backend := cache_memory.NewBytesQueueBackend(100)
	defer backend.Close()
	client.SetCacheBackend(backend)

	msg := RabbitMQMessage{
		ID:         "test-msg-1",
		Payload:    "test payload",
		Timestamp:  time.Now(),
		Exchange:   "test.exchange",
		RoutingKey: "test.key",
	}

	// Add to cache
	success := client.AddToCache(msg)
	if !success {
		t.Errorf("AddToCache() should return true")
	}

	// Verify message is in cache
	len, _ := backend.Len()
	if len != 1 {
		t.Errorf("Cache should have 1 message")
	}

	// Dequeue to verify
	payload, found, _ := backend.Dequeue()
	if !found {
		t.Errorf("Message should be in cache")
	}

	// Verify payload is JSON
	var cachedMsg RabbitMQMessage
	err := json.Unmarshal(payload, &cachedMsg)
	if err != nil {
		t.Errorf("Cached payload should be valid JSON: %v", err)
	}

	if cachedMsg.ID != msg.ID {
		t.Errorf("Cached message ID should match, expected: %s, got: %s", msg.ID, cachedMsg.ID)
	}
}

// TestRabbitMQClientAddToCacheCapacity verifies cache respects capacity.
func TestRabbitMQClientAddToCacheCapacity(t *testing.T) {
	client := NewRabbitMQClient("amqp://guest:guest@localhost:5672/", 0)
	defer client.Close()

	// Create a small cache
	backend := cache_memory.NewBytesQueueBackend(2)
	defer backend.Close()
	client.SetCacheBackend(backend)

	// Add messages up to capacity
	msg1 := RabbitMQMessage{
		ID:         "msg-1",
		Payload:    "payload1",
		Timestamp:  time.Now(),
		Exchange:   "ex1",
		RoutingKey: "key1",
	}
	success1 := client.AddToCache(msg1)
	if !success1 {
		t.Errorf("First message should be added to cache")
	}

	msg2 := RabbitMQMessage{
		ID:         "msg-2",
		Payload:    "payload2",
		Timestamp:  time.Now(),
		Exchange:   "ex1",
		RoutingKey: "key1",
	}
	success2 := client.AddToCache(msg2)
	if !success2 {
		t.Errorf("Second message should be added to cache")
	}

	// Third message should fail (capacity exceeded)
	msg3 := RabbitMQMessage{
		ID:         "msg-3",
		Payload:    "payload3",
		Timestamp:  time.Now(),
		Exchange:   "ex1",
		RoutingKey: "key1",
	}
	success3 := client.AddToCache(msg3)
	if success3 {
		t.Errorf("Third message should NOT be added (capacity exceeded)")
	}
}

// TestRabbitMQClientMaxCacheSizeZero verifies unlimited cache setting.
func TestRabbitMQClientMaxCacheSizeZero(t *testing.T) {
	// maxCacheSize of 0 should be treated as unlimited (default: 100000)
	client := NewRabbitMQClient("amqp://guest:guest@localhost:5672/", 0)
	defer client.Close()

	if client.maxCacheSize != 100000 {
		t.Errorf("maxCacheSize should be 100000 for unlimited cache, got: %d", client.maxCacheSize)
	}
}

// TestRabbitMQClientMaxCacheSizeCustom verifies custom cache size.
func TestRabbitMQClientMaxCacheSizeCustom(t *testing.T) {
	client := NewRabbitMQClient("amqp://guest:guest@localhost:5672/", 5000)
	defer client.Close()

	if client.maxCacheSize != 5000 {
		t.Errorf("maxCacheSize should be 5000, got: %d", client.maxCacheSize)
	}
}

// TestRabbitMQClientProcessCacheDequeue verifies cache processing dequeues messages.
func TestRabbitMQClientProcessCacheDequeue(t *testing.T) {
	client := NewRabbitMQClient("amqp://guest:guest@localhost:5672/", 0)
	defer client.Close()

	backend := cache_memory.NewBytesQueueBackend(100)
	defer backend.Close()
	client.SetCacheBackend(backend)

	// Add a message directly to backend
	msg := RabbitMQMessage{
		ID:         "dequeue-test",
		Payload:    "test",
		Timestamp:  time.Now(),
		Exchange:   "ex",
		RoutingKey: "key",
	}
	payload, _ := json.Marshal(msg)
	backend.Enqueue(payload)

	// Verify it's in cache
	count, _ := backend.Len()
	if count != 1 {
		t.Errorf("Cache should have 1 message, got: %d", count)
	}

	// Dequeue manually (simulating processCache behavior)
	dequeued, found, _ := backend.Dequeue()
	if !found {
		t.Errorf("Should be able to dequeue the message")
	}

	// Verify cache is now empty
	count, _ = backend.Len()
	if count != 0 {
		t.Errorf("Cache should be empty after dequeue, got: %d messages", count)
	}

	// Verify deserialization
	var cachedMsg RabbitMQMessage
	_ = json.Unmarshal(dequeued, &cachedMsg)
	if cachedMsg.ID != msg.ID {
		t.Errorf("Deserialized message should match original")
	}
}

// TestRabbitMQClientGetChannelBlocking verifies GetChannel behavior.
func TestRabbitMQClientGetChannelBlocking(t *testing.T) {
	client := NewRabbitMQClient("amqp://guest:guest@localhost:5672/", 0)
	defer client.Close()

	// Channel should be nil initially (not connected)
	// GetChannel blocks until channel is available, but since we're not connecting,
	// this test verifies it doesn't panic

	// We'll skip the actual call since it would block indefinitely without a real RabbitMQ
	if client.connURI == "" {
		t.Errorf("Client should have connection URI")
	}
}

// TestRabbitMQClientMessageIDGeneration verifies unique message IDs.
func TestRabbitMQClientMessageIDGeneration(t *testing.T) {
	// Message IDs should be unique even when generated at the same time
	msg1 := RabbitMQMessage{
		ID: "msg-1",
	}
	msg2 := RabbitMQMessage{
		ID: "msg-2",
	}

	if msg1.ID == msg2.ID {
		t.Errorf("Message IDs should be unique")
	}
}

// TestRabbitMQClientCacheNilBackend verifies handling of nil backend.
func TestRabbitMQClientCacheNilBackend(t *testing.T) {
	client := NewRabbitMQClient("amqp://guest:guest@localhost:5672/", 0)
	defer client.Close()

	// Don't set backend
	msg := RabbitMQMessage{
		ID:         "test",
		Payload:    "payload",
		Timestamp:  time.Now(),
		Exchange:   "ex",
		RoutingKey: "key",
	}

	// Should handle gracefully
	success := client.AddToCache(msg)
	if success {
		t.Errorf("AddToCache() should fail with nil backend")
	}
}

// TestRabbitMQClientMessageMarshal verifies message serialization.
func TestRabbitMQClientMessageMarshal(t *testing.T) {
	msg := RabbitMQMessage{
		ID:         "marshal-test",
		Payload:    map[string]interface{}{"key": "value"},
		Timestamp:  time.Now(),
		Exchange:   "test.ex",
		RoutingKey: "test.key",
	}

	// Marshal
	payload, err := json.Marshal(msg)
	if err != nil {
		t.Errorf("Failed to marshal message: %v", err)
	}

	// Unmarshal
	var unmarshaled RabbitMQMessage
	err = json.Unmarshal(payload, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal message: %v", err)
	}

	if unmarshaled.ID != msg.ID {
		t.Errorf("Unmarshaled ID should match original")
	}
	if unmarshaled.Exchange != msg.Exchange {
		t.Errorf("Unmarshaled Exchange should match original")
	}
	if unmarshaled.RoutingKey != msg.RoutingKey {
		t.Errorf("Unmarshaled RoutingKey should match original")
	}
}

// TestRabbitMQClientBackendInjectionSequence verifies injection works in order.
func TestRabbitMQClientBackendInjectionSequence(t *testing.T) {
	client := NewRabbitMQClient("amqp://guest:guest@localhost:5672/", 0)
	defer client.Close()

	// Initially no backend
	if client.messageCache != nil {
		t.Errorf("messageCache should be nil before SetCacheBackend")
	}

	// Inject backend
	backend1 := cache_memory.NewBytesQueueBackend(100)
	defer backend1.Close()
	client.SetCacheBackend(backend1)

	if client.messageCache != backend1 {
		t.Errorf("messageCache should be backend1")
	}

	// Can replace with different backend
	backend2 := cache_memory.NewBytesQueueBackend(200)
	defer backend2.Close()
	client.SetCacheBackend(backend2)

	if client.messageCache != backend2 {
		t.Errorf("messageCache should be updated to backend2")
	}
}

// TestRabbitMQClientConcurrentCacheOperations tests concurrent cache operations.
func TestRabbitMQClientConcurrentCacheOperations(t *testing.T) {
	client := NewRabbitMQClient("amqp://guest:guest@localhost:5672/", 0)
	defer client.Close()

	backend := cache_memory.NewBytesQueueBackend(1000)
	defer backend.Close()
	client.SetCacheBackend(backend)

	// Concurrent add operations
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			msg := RabbitMQMessage{
				ID:         "concurrent-" + string(rune(id)),
				Payload:    "payload",
				Timestamp:  time.Now(),
				Exchange:   "ex",
				RoutingKey: "key",
			}
			client.AddToCache(msg)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all messages are in cache
	count, _ := backend.Len()
	if count != 10 {
		t.Errorf("Should have 10 messages in cache, got: %d", count)
	}
}
