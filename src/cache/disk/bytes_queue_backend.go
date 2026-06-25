package disk

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type BytesQueueBackend struct {
	basePath string
	capacity int
	mu       sync.Mutex
}

func NewBytesQueueBackend(basePath string, capacity int) (*BytesQueueBackend, error) {
	if len(strings.TrimSpace(basePath)) == 0 {
		return nil, errors.New("queue disk path is empty")
	}

	if err := os.MkdirAll(basePath, 0o755); err != nil {
		return nil, err
	}

	return &BytesQueueBackend{basePath: basePath, capacity: capacity}, nil
}

func (backend *BytesQueueBackend) filePath() string {
	name := fmt.Sprintf("%020d.queue", time.Now().UnixNano())
	return filepath.Join(backend.basePath, name)
}

func (backend *BytesQueueBackend) listFiles() ([]string, error) {
	entries, err := os.ReadDir(backend.basePath)
	if err != nil {
		return nil, err
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".queue") {
			continue
		}
		files = append(files, filepath.Join(backend.basePath, entry.Name()))
	}

	sort.Strings(files)
	return files, nil
}

func (backend *BytesQueueBackend) Enqueue(payload []byte) (bool, error) {
	backend.mu.Lock()
	defer backend.mu.Unlock()

	files, err := backend.listFiles()
	if err != nil {
		return false, err
	}
	if backend.capacity > 0 && len(files) >= backend.capacity {
		return false, nil
	}

	return true, os.WriteFile(backend.filePath(), payload, 0o644)
}

func (backend *BytesQueueBackend) Dequeue() ([]byte, bool, error) {
	backend.mu.Lock()
	defer backend.mu.Unlock()

	files, err := backend.listFiles()
	if err != nil {
		return nil, false, err
	}
	if len(files) == 0 {
		return nil, false, nil
	}

	payload, err := os.ReadFile(files[0])
	if err != nil {
		return nil, false, err
	}
	if err := os.Remove(files[0]); err != nil {
		return nil, false, err
	}

	return payload, true, nil
}

func (backend *BytesQueueBackend) Len() (int, error) {
	backend.mu.Lock()
	defer backend.mu.Unlock()

	files, err := backend.listFiles()
	if err != nil {
		return 0, err
	}
	return len(files), nil
}

func (backend *BytesQueueBackend) Close() error {
	return nil
}
