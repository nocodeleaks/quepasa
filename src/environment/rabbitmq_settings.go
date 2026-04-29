package environment

// RabbitMQ environment variable names
const (
	ENV_RABBITMQ_QUEUE            = "RABBITMQ_QUEUE"            // RabbitMQ queue name
	ENV_RABBITMQ_CONNECTIONSTRING = "RABBITMQ_CONNECTIONSTRING" // RabbitMQ connection string
	ENV_RABBITMQ_CACHELENGTH      = "RABBITMQ_CACHELENGTH"      // RabbitMQ cache length
	ENV_RABBITMQ_CACHE_BACKEND    = "RABBITMQ_CACHE_BACKEND"    // RabbitMQ retry cache backend
	ENV_RABBITMQ_CACHE_DISK_PATH  = "RABBITMQ_CACHE_DISK_PATH"  // RabbitMQ retry disk cache path
	ENV_RABBITMQ_CACHE_QUEUE_KEY  = "RABBITMQ_CACHE_QUEUE_KEY"  // RabbitMQ retry cache queue key
)

// RabbitMQSettings holds all RabbitMQ configuration loaded from environment
type RabbitMQSettings struct {
	Queue            string `json:"queue"`
	ConnectionString string `json:"connection_string"`
	CacheLength      uint64 `json:"cache_length"`
	CacheBackend     string `json:"cache_backend"`
	CacheDiskPath    string `json:"cache_disk_path"`
	CacheQueueKey    string `json:"cache_queue_key"`
}

// NewRabbitMQSettings creates a new RabbitMQ settings by loading all values from environment
func NewRabbitMQSettings() RabbitMQSettings {
	return RabbitMQSettings{
		Queue:            getEnvOrDefaultString(ENV_RABBITMQ_QUEUE, ""),
		ConnectionString: getEnvOrDefaultString(ENV_RABBITMQ_CONNECTIONSTRING, ""),
		CacheLength:      getEnvOrDefaultUint64(ENV_RABBITMQ_CACHELENGTH, 0),
		CacheBackend:     getEnvOrDefaultString(ENV_RABBITMQ_CACHE_BACKEND, ""),
		CacheDiskPath:    getEnvOrDefaultString(ENV_RABBITMQ_CACHE_DISK_PATH, ""),
		CacheQueueKey:    getEnvOrDefaultString(ENV_RABBITMQ_CACHE_QUEUE_KEY, "rabbitmq_retry"),
	}
}
