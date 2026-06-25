package calls

import (
	"context"
	"sync"
)

// Per-call registry: tracks active CallSessions and their media-task cancel handles
// so a connection teardown can stop every in-flight call. AbortAll is the teardown
// primitive; the integrator owns a CallRegistry and calls AbortAll from their own
// disconnect/reconnect path (it is not auto-wired).

// callEntry is one registered call: its session plus the optional cancel handle for
// the running media goroutine.
type callEntry struct {
	session   *CallSession
	mediaTask context.CancelFunc // nil until a media goroutine is registered
}

// CallRegistry is a thread-safe map of active calls keyed by call-id, each
// optionally holding the cancel handle for its running media task.
type CallRegistry struct {
	mu    sync.Mutex
	calls map[string]*callEntry
}

// NewCallRegistry returns an empty registry.
func NewCallRegistry() *CallRegistry {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/src/voip/registry.rs#L25-L27
	return &CallRegistry{calls: make(map[string]*callEntry)}
}

// Insert registers a new call; returns false if the id already exists.
func (r *CallRegistry) Insert(session *CallSession) bool {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/src/voip/registry.rs#L30-L43
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.calls[session.CallID]; exists {
		return false
	}
	r.calls[session.CallID] = &callEntry{session: session}
	return true
}

// SetMediaTask attaches (or replaces, cancelling the old) the media task's cancel
// handle for a call. If the call is unknown (e.g. already removed), the handle is
// cancelled immediately so its task can't outlive the call.
func (r *CallRegistry) SetMediaTask(callID string, cancel context.CancelFunc) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/src/voip/registry.rs#L47-L61
	r.mu.Lock()
	entry, found := r.calls[callID]
	var old context.CancelFunc
	if found {
		old = entry.mediaTask
		entry.mediaTask = cancel
	}
	r.mu.Unlock()
	// Cancel outside the lock (non-blocking, but keeps the critical section minimal).
	if !found {
		cancel()
		return
	}
	if old != nil {
		old()
	}
}

// Phase returns the call's current phase, and whether the call is known.
func (r *CallRegistry) Phase(callID string) (CallPhase, bool) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/src/voip/registry.rs#L63-L69
	r.mu.Lock()
	defer r.mu.Unlock()
	if entry, ok := r.calls[callID]; ok {
		return entry.session.Phase(), true
	}
	return CallPhaseIdle, false
}

// Transition advances a call's phase; false if unknown or the move is illegal.
func (r *CallRegistry) Transition(callID string, next CallPhase) bool {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/src/voip/registry.rs#L72-L78
	r.mu.Lock()
	defer r.mu.Unlock()
	entry, ok := r.calls[callID]
	return ok && entry.session.TransitionTo(next)
}

// Snapshot returns a copy of the call's session, and whether it is known.
func (r *CallRegistry) Snapshot(callID string) (CallSession, bool) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/src/voip/registry.rs#L81-L87
	r.mu.Lock()
	defer r.mu.Unlock()
	if entry, ok := r.calls[callID]; ok {
		return *entry.session, true
	}
	return CallSession{}, false
}

// ActiveCount returns the number of registered calls.
func (r *CallRegistry) ActiveCount() int {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/src/voip/registry.rs#L89-L91
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.calls)
}

// Remove deletes a call, cancelling its media task; true if it existed.
func (r *CallRegistry) Remove(callID string) bool {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/src/voip/registry.rs#L94-L109
	r.mu.Lock()
	entry, ok := r.calls[callID]
	if ok {
		delete(r.calls, callID)
	}
	r.mu.Unlock()
	if !ok {
		return false
	}
	if entry.mediaTask != nil {
		entry.mediaTask()
	}
	return true
}

// AbortAll cancels every call's media task and clears the registry, returning the
// number cleared. Call on disconnect/reconnect.
func (r *CallRegistry) AbortAll() int {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/src/voip/registry.rs#L113-L123
	r.mu.Lock()
	entries := r.calls
	r.calls = make(map[string]*callEntry)
	r.mu.Unlock()
	for _, entry := range entries {
		if entry.mediaTask != nil {
			entry.mediaTask()
		}
	}
	return len(entries)
}
