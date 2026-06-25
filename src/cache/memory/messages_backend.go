package memory

import (
	"sort"
	"sync"

	cache "github.com/nocodeleaks/quepasa/cache"
)

type MessagesBackend struct {
	items sync.Map
}

func NewMessagesBackend() *MessagesBackend {
	return &MessagesBackend{}
}

func (backend *MessagesBackend) Get(key string) (cache.MessageRecord, bool, error) {
	if value, ok := backend.items.Load(key); ok {
		record, ok := value.(cache.MessageRecord)
		if !ok {
			return cache.MessageRecord{}, false, nil
		}
		return record, true, nil
	}
	return cache.MessageRecord{}, false, nil
}

func (backend *MessagesBackend) Set(key string, record cache.MessageRecord) error {
	backend.items.Store(key, record)
	return nil
}

func (backend *MessagesBackend) Delete(key string) error {
	backend.items.Delete(key)
	return nil
}

func (backend *MessagesBackend) List() ([]cache.MessageRecordEntry, error) {
	items := make([]cache.MessageRecordEntry, 0)
	backend.items.Range(func(key, value any) bool {
		record, ok := value.(cache.MessageRecord)
		if !ok {
			return true
		}
		keyString, ok := key.(string)
		if !ok {
			return true
		}
		items = append(items, cache.MessageRecordEntry{Key: keyString, Record: record})
		return true
	})

	sort.Slice(items, func(i, j int) bool {
		if items[i].Record.ExpiresAt.Equal(items[j].Record.ExpiresAt) {
			return items[i].Key < items[j].Key
		}
		return items[i].Record.ExpiresAt.Before(items[j].Record.ExpiresAt)
	})

	return items, nil
}

func (backend *MessagesBackend) Close() error {
	return nil
}
