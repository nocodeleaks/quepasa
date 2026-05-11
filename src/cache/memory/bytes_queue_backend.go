package memory

import "sync"

type BytesQueueBackend struct {
	mu       sync.Mutex
	items    [][]byte
	capacity int
}

func NewBytesQueueBackend(capacity int) *BytesQueueBackend {
	return &BytesQueueBackend{capacity: capacity}
}

func (backend *BytesQueueBackend) Enqueue(payload []byte) (bool, error) {
	backend.mu.Lock()
	defer backend.mu.Unlock()

	if backend.capacity > 0 && len(backend.items) >= backend.capacity {
		return false, nil
	}

	clone := append([]byte(nil), payload...)
	backend.items = append(backend.items, clone)
	return true, nil
}

func (backend *BytesQueueBackend) Dequeue() ([]byte, bool, error) {
	backend.mu.Lock()
	defer backend.mu.Unlock()

	if len(backend.items) == 0 {
		return nil, false, nil
	}

	payload := append([]byte(nil), backend.items[0]...)
	backend.items = backend.items[1:]
	return payload, true, nil
}

func (backend *BytesQueueBackend) Len() (int, error) {
	backend.mu.Lock()
	defer backend.mu.Unlock()

	return len(backend.items), nil
}

func (backend *BytesQueueBackend) Close() error {
	backend.mu.Lock()
	defer backend.mu.Unlock()

	backend.items = nil
	return nil
}
