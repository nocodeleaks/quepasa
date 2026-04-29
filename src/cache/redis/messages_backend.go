package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	cache "github.com/nocodeleaks/quepasa/cache"
	redis "github.com/redis/go-redis/v9"
)

type MessagesBackend struct {
	client    *redis.Client
	ctx       context.Context
	keyPrefix string
}

func NewMessagesBackend(config cache.RedisConfig) (*MessagesBackend, error) {
	address := strings.TrimSpace(config.Host)
	if len(address) == 0 {
		return nil, fmt.Errorf("redis host is empty")
	}

	if config.Port > 0 {
		address = fmt.Sprintf("%s:%d", address, config.Port)
	}

	client := redis.NewClient(&redis.Options{
		Addr:         address,
		Username:     config.Username,
		Password:     config.Password,
		DB:           int(config.Database),
		PoolSize:     int(config.PoolSize),
		MaxRetries:   int(config.MaxRetries),
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &MessagesBackend{client: client, ctx: ctx, keyPrefix: strings.TrimSpace(config.KeyPrefix)}, nil
}

func (backend *MessagesBackend) key(key string) string {
	if len(backend.keyPrefix) == 0 {
		return key
	}
	return backend.keyPrefix + ":messages:" + key
}

func (backend *MessagesBackend) Get(key string) (cache.MessageRecord, bool, error) {
	data, err := backend.client.Get(backend.ctx, backend.key(key)).Bytes()
	if err != nil {
		if err == redis.Nil {
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
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	ttl := time.Until(record.ExpiresAt)
	if ttl < 0 {
		ttl = 0
	}

	return backend.client.Set(backend.ctx, backend.key(key), data, ttl).Err()
}

func (backend *MessagesBackend) Delete(key string) error {
	return backend.client.Del(backend.ctx, backend.key(key)).Err()
}

func (backend *MessagesBackend) List() ([]cache.MessageRecordEntry, error) {
	keys, err := backend.client.Keys(backend.ctx, backend.key("*")).Result()
	if err != nil {
		return nil, err
	}

	items := make([]cache.MessageRecordEntry, 0, len(keys))
	for _, key := range keys {
		data, err := backend.client.Get(backend.ctx, key).Bytes()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			return nil, err
		}

		var record cache.MessageRecord
		if err := json.Unmarshal(data, &record); err != nil {
			return nil, err
		}

		trimmedKey := key
		if len(backend.keyPrefix) > 0 {
			trimmedKey = strings.TrimPrefix(key, backend.keyPrefix+":messages:")
		}
		items = append(items, cache.MessageRecordEntry{Key: trimmedKey, Record: record})
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
	return backend.client.Close()
}
