package cache

import "time"

type RedisConfig struct {
	Host         string
	Port         uint32
	Username     string
	Password     string
	Database     uint32
	KeyPrefix    string
	PoolSize     uint32
	MaxRetries   uint32
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}
