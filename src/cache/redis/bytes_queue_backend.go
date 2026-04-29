package redis

import (
	"context"
	"fmt"
	"strings"

	cache "github.com/nocodeleaks/quepasa/cache"
	redis "github.com/redis/go-redis/v9"
)

type BytesQueueBackend struct {
	client *redis.Client
	ctx    context.Context
	key    string
	limit  int
}

func NewBytesQueueBackend(config cache.BytesQueueConfig) (*BytesQueueBackend, error) {
	address := strings.TrimSpace(config.Redis.Host)
	if len(address) == 0 {
		return nil, fmt.Errorf("redis host is empty")
	}
	if config.Redis.Port > 0 {
		address = fmt.Sprintf("%s:%d", address, config.Redis.Port)
	}

	client := redis.NewClient(&redis.Options{
		Addr:         address,
		Username:     config.Redis.Username,
		Password:     config.Redis.Password,
		DB:           int(config.Redis.Database),
		PoolSize:     int(config.Redis.PoolSize),
		MaxRetries:   int(config.Redis.MaxRetries),
		DialTimeout:  config.Redis.DialTimeout,
		ReadTimeout:  config.Redis.ReadTimeout,
		WriteTimeout: config.Redis.WriteTimeout,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	queueKey := strings.TrimSpace(config.QueueKey)
	if len(queueKey) == 0 {
		queueKey = "bytes_queue"
	}
	if len(config.Redis.KeyPrefix) > 0 {
		queueKey = config.Redis.KeyPrefix + ":queue:" + queueKey
	}

	return &BytesQueueBackend{client: client, ctx: ctx, key: queueKey, limit: config.Capacity}, nil
}

func (backend *BytesQueueBackend) Enqueue(payload []byte) (bool, error) {
	if backend.limit > 0 {
		length, err := backend.client.LLen(backend.ctx, backend.key).Result()
		if err != nil {
			return false, err
		}
		if int(length) >= backend.limit {
			return false, nil
		}
	}

	if err := backend.client.RPush(backend.ctx, backend.key, payload).Err(); err != nil {
		return false, err
	}
	return true, nil
}

func (backend *BytesQueueBackend) Dequeue() ([]byte, bool, error) {
	result, err := backend.client.LPop(backend.ctx, backend.key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, false, nil
		}
		return nil, false, err
	}
	return result, true, nil
}

func (backend *BytesQueueBackend) Len() (int, error) {
	length, err := backend.client.LLen(backend.ctx, backend.key).Result()
	if err != nil {
		return 0, err
	}
	return int(length), nil
}

func (backend *BytesQueueBackend) Close() error {
	return backend.client.Close()
}
