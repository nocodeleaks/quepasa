package cache

import "time"

type MessagesConfig struct {
	Backend  string
	MaxItems uint64
	TTL      time.Duration
	DiskPath string
	Redis    RedisConfig
}
