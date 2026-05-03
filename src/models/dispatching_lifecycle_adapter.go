package models

// DefaultDispatchingLifecyclePublisher returns the currently wired lifecycle
// publisher used as compatibility fallback during the transport migration.
func DefaultDispatchingLifecyclePublisher() DispatchingLifecyclePublisher {
	transportServicesMu.RLock()
	publisher := GlobalDispatchingLifecyclePublisher
	transportServicesMu.RUnlock()
	return publisher
}
