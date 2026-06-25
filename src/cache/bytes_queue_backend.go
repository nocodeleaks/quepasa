package cache

type BytesQueueBackend interface {
	Enqueue(payload []byte) (bool, error)
	Dequeue() ([]byte, bool, error)
	Len() (int, error)
	Close() error
}
