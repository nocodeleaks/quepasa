package whatsmeow

import (
	"testing"
	"time"
)

// TestGetTimestamp_MonotonicOrdering verifies that sequential calls to getTimestamp
// produce strictly increasing timestamps within a small burst (under 1000 events).
// Note: getTimestamp uses sequence % 1000, so after 1000 events it may wrap within
// the same microsecond. This test only verifies ordering within the wrap window.
func TestGetTimestamp_MonotonicOrdering(t *testing.T) {
	handler := &WhatsmeowHandlers{}

	const iterations = 999 // Stay under the 1000-event wrap boundary
	timestamps := make([]time.Time, iterations)

	for i := 0; i < iterations; i++ {
		timestamps[i] = handler.getTimestamp()
	}

	// Verify strict monotonic increase within the wrap window
	for i := 1; i < iterations; i++ {
		if !timestamps[i].After(timestamps[i-1]) {
			t.Errorf("timestamp[%d] (%v) is not after timestamp[%d] (%v)",
				i, timestamps[i], i-1, timestamps[i-1])
		}
	}
}

// TestGetTimestamp_NilHandler verifies that getTimestamp handles nil receiver gracefully.
func TestGetTimestamp_NilHandler(t *testing.T) {
	var handler *WhatsmeowHandlers // nil
	ts := handler.getTimestamp()

	if ts.IsZero() {
		t.Error("getTimestamp on nil handler returned zero time")
	}

	// Should return a time close to now (within 1 second tolerance)
	now := time.Now().UTC()
	diff := now.Sub(ts)
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Second {
		t.Errorf("timestamp diff from now is too large: %v", diff)
	}
}

// TestEventRouter_Initialization verifies that the router is lazily initialized.
func TestEventRouter_Initialization(t *testing.T) {
	handler := &WhatsmeowHandlers{}

	if handler.router != nil {
		t.Error("router should be nil before first use")
	}

	router := handler.getRouter()

	if router == nil {
		t.Error("getRouter() should initialize and return non-nil router")
	}

	if handler.router == nil {
		t.Error("handler.router should be set after getRouter() call")
	}

	// Second call should return the same instance
	router2 := handler.getRouter()
	if router2 != router {
		t.Error("getRouter() should return the same router instance on subsequent calls")
	}
}

// TestHasWAHandlers_NilHandlers verifies that hasWAHandlers returns false when WAHandlers is nil.
func TestHasWAHandlers_NilHandlers(t *testing.T) {
	handler := &WhatsmeowHandlers{}

	if handler.hasWAHandlers() {
		t.Error("hasWAHandlers should return false when WAHandlers is nil")
	}
}

// TestEventCounter_Increment verifies that the event counter increments correctly.
func TestEventCounter_Increment(t *testing.T) {
	handler := &WhatsmeowHandlers{}

	if handler.Counter != 0 {
		t.Errorf("initial Counter should be 0, got %d", handler.Counter)
	}

	handler.Counter++
	if handler.Counter != 1 {
		t.Errorf("Counter after increment should be 1, got %d", handler.Counter)
	}
}

// TestOfflineSyncState_InitialState verifies the initial state of offline sync flags.
func TestOfflineSyncState_InitialState(t *testing.T) {
	handler := &WhatsmeowHandlers{}

	if handler.offlineSyncStarted {
		t.Error("offlineSyncStarted should be false initially")
	}

	if handler.offlineSyncCompleted {
		t.Error("offlineSyncCompleted should be false initially")
	}
}

// TestEventSequence_Concurrency verifies that eventSequence increments atomically under concurrent access.
func TestEventSequence_Concurrency(t *testing.T) {
	handler := &WhatsmeowHandlers{}
	const goroutines = 10
	const iterationsPerGoroutine = 100

	done := make(chan bool, goroutines)

	for g := 0; g < goroutines; g++ {
		go func() {
			for i := 0; i < iterationsPerGoroutine; i++ {
				_ = handler.getTimestamp() // This increments eventSequence atomically
			}
			done <- true
		}()
	}

	// Wait for all goroutines to finish
	for g := 0; g < goroutines; g++ {
		<-done
	}

	expectedSequence := uint64(goroutines * iterationsPerGoroutine)
	if handler.eventSequence != expectedSequence {
		t.Errorf("eventSequence should be %d after concurrent increments, got %d",
			expectedSequence, handler.eventSequence)
	}
}

// TestEventHandlerID_Initial verifies that eventHandlerID starts at 0.
func TestEventHandlerID_Initial(t *testing.T) {
	handler := &WhatsmeowHandlers{}

	if handler.eventHandlerID != 0 {
		t.Errorf("eventHandlerID should be 0 initially, got %d", handler.eventHandlerID)
	}
}

// TestUnregisterRequestedToken_Initial verifies that unregisterRequestedToken starts as false.
func TestUnregisterRequestedToken_Initial(t *testing.T) {
	handler := &WhatsmeowHandlers{}

	if handler.unregisterRequestedToken {
		t.Error("unregisterRequestedToken should be false initially")
	}
}

// Benchmark for getTimestamp to verify it remains fast under load.
func BenchmarkGetTimestamp(b *testing.B) {
	handler := &WhatsmeowHandlers{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = handler.getTimestamp()
	}
}

// Benchmark for concurrent getTimestamp calls (simulates real-world event bursts).
func BenchmarkGetTimestamp_Concurrent(b *testing.B) {
	handler := &WhatsmeowHandlers{}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = handler.getTimestamp()
		}
	})
}

// TestEventRouter_HandlerRegistration verifies that event handlers can be registered.
// This is a characterization test — it documents current behavior without asserting correctness.
func TestEventRouter_HandlerRegistration(t *testing.T) {
	handler := &WhatsmeowHandlers{}
	router := handler.getRouter()

	if router == nil {
		t.Fatal("router should not be nil after initialization")
	}

	// Characterization: router should have a dispatch table (implementation detail).
	// This test documents the existence of the router but does not test its dispatch logic,
	// which requires mocking whatsmeow event types.
}

// TestGetServiceOptions_NilSource verifies that GetServiceOptions handles nil receiver.
func TestGetServiceOptions_NilSource(t *testing.T) {
	var handler *WhatsmeowHandlers // nil
	options := handler.GetServiceOptions()

	// Should return zero value (empty WhatsappOptionsExtended struct)
	if options.LogLevel != "" {
		t.Errorf("expected empty LogLevel for nil handler, got %s", options.LogLevel)
	}
}
