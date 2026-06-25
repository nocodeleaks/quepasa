package disk

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	cache "github.com/nocodeleaks/quepasa/cache"
)

type MessagesBackend struct {
	basePath string
	mu       sync.Mutex
}

func NewMessagesBackend(basePath string) (*MessagesBackend, error) {
	if len(strings.TrimSpace(basePath)) == 0 {
		return nil, errors.New("cache disk path is empty")
	}

	if err := os.MkdirAll(basePath, 0o755); err != nil {
		return nil, err
	}

	return &MessagesBackend{basePath: basePath}, nil
}

func normalizeFileName(key string) string {
	replacer := strings.NewReplacer("/", "_", "\\", "_", ":", "_", "*", "_", "?", "_", "\"", "_", "<", "_", ">", "_", "|", "_")
	return replacer.Replace(key)
}

func (backend *MessagesBackend) filePath(key string) string {
	return filepath.Join(backend.basePath, normalizeFileName(key)+".json")
}

func (backend *MessagesBackend) Get(key string) (cache.MessageRecord, bool, error) {
	backend.mu.Lock()
	defer backend.mu.Unlock()

	data, err := os.ReadFile(backend.filePath(key))
	if err != nil {
		if os.IsNotExist(err) {
			return cache.MessageRecord{}, false, nil
		}
		return cache.MessageRecord{}, false, err
	}

	var record cache.MessageRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return cache.MessageRecord{}, false, err
	}

	return record, true, nil
}

func (backend *MessagesBackend) Set(key string, record cache.MessageRecord) error {
	backend.mu.Lock()
	defer backend.mu.Unlock()

	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	return os.WriteFile(backend.filePath(key), data, 0o644)
}

func (backend *MessagesBackend) Delete(key string) error {
	backend.mu.Lock()
	defer backend.mu.Unlock()

	err := os.Remove(backend.filePath(key))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (backend *MessagesBackend) List() ([]cache.MessageRecordEntry, error) {
	backend.mu.Lock()
	defer backend.mu.Unlock()

	entries, err := os.ReadDir(backend.basePath)
	if err != nil {
		return nil, err
	}

	items := make([]cache.MessageRecordEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".json") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(backend.basePath, entry.Name()))
		if err != nil {
			return nil, err
		}

		var record cache.MessageRecord
		if err := json.Unmarshal(data, &record); err != nil {
			return nil, err
		}

		key := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		items = append(items, cache.MessageRecordEntry{Key: key, Record: record})
	}

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
