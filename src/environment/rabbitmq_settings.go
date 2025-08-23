package environment

// RabbitMQ environment variable names
const (
	ENV_RABBITMQ_QUEUE            = "RABBITMQ_QUEUE"            // RabbitMQ queue name
	ENV_RABBITMQ_CONNECTIONSTRING = "RABBITMQ_CONNECTIONSTRING" // RabbitMQ connection string
	ENV_RABBITMQ_CACHELENGTH      = "RABBITMQ_CACHELENGTH"      // RabbitMQ cache length
)

// RabbitMQSettings holds all RabbitMQ configuration loaded from environment
type RabbitMQSettings struct {
	Queue            string `json:"queue"`
	ConnectionString string `json:"connection_string"`
	CacheLength      uint64 `json:"cache_length"`
}

// NewRabbitMQSettings creates a new RabbitMQ settings by loading all values from environment
func NewRabbitMQSettings() RabbitMQSettings {
	return RabbitMQSettings{
		Queue:            getEnvOrDefaultString(ENV_RABBITMQ_QUEUE, ""),
		ConnectionString: getEnvOrDefaultString(ENV_RABBITMQ_CONNECTIONSTRING, ""),
		CacheLength:      getEnvOrDefaultUint64(ENV_RABBITMQ_CACHELENGTH, 0),
	}
}
