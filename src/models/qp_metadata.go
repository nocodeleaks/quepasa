package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
)

type QpMetadata map[string]any

func (metadata QpMetadata) Value() (driver.Value, error) {
	if len(metadata) == 0 {
		return nil, nil
	}

	payload, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	return string(payload), nil
}

func (metadata *QpMetadata) Scan(value any) error {
	if metadata == nil {
		return nil
	}

	if value == nil {
		*metadata = nil
		return nil
	}

	var payload []byte

	switch typed := value.(type) {
	case []byte:
		payload = typed
	case string:
		payload = []byte(typed)
	default:
		return fmt.Errorf("unsupported metadata type: %T", value)
	}

	if len(strings.TrimSpace(string(payload))) == 0 {
		*metadata = nil
		return nil
	}

	decoded := QpMetadata{}
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return err
	}

	*metadata = decoded
	return nil
}

func (metadata QpMetadata) Clone() QpMetadata {
	if len(metadata) == 0 {
		return nil
	}

	payload, err := json.Marshal(metadata)
	if err != nil {
		cloned := QpMetadata{}
		for key, value := range metadata {
			cloned[key] = value
		}
		return cloned
	}

	cloned := QpMetadata{}
	if err := json.Unmarshal(payload, &cloned); err != nil {
		fallback := QpMetadata{}
		for key, value := range metadata {
			fallback[key] = value
		}
		return fallback
	}

	return cloned
}

func (server *QpServer) GetMetadataValue(key string) any {
	if server == nil || len(server.Metadata) == 0 {
		return nil
	}

	return server.Metadata[key]
}

func (server *QpServer) SetMetadataValue(key string, value any) {
	if server == nil {
		return
	}

	if server.Metadata == nil {
		server.Metadata = QpMetadata{}
	}

	server.Metadata[key] = value
}

func (server *QpServer) RemoveMetadataValue(key string) {
	if server == nil || len(server.Metadata) == 0 {
		return
	}

	delete(server.Metadata, key)
	if len(server.Metadata) == 0 {
		server.Metadata = nil
	}
}
