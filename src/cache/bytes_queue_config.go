package cache

type BytesQueueConfig struct {
	Backend  string
	Capacity int
	DiskPath string
	QueueKey string
	Redis    RedisConfig
}
