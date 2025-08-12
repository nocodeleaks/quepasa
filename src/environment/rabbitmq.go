package environment

// RabbitMQEnvironment handles all RabbitMQ-related environment variables
type RabbitMQEnvironment struct{}

// RabbitMQ environment variable names
const (
	ENV_RABBITMQ_QUEUE            = "RABBITMQ_QUEUE"            // RabbitMQ queue name
	ENV_RABBITMQ_CONNECTIONSTRING = "RABBITMQ_CONNECTIONSTRING" // RabbitMQ connection string
	ENV_RABBITMQ_CACHELENGTH      = "RABBITMQ_CACHELENGTH"      // RabbitMQ cache length
)

// Queue returns the name of the RabbitMQ queue.
// Defaults to empty string if the environment variable is not set.
func (env *RabbitMQEnvironment) Queue() string {
	return getEnvOrDefaultString(ENV_RABBITMQ_QUEUE, "")
}

// ConnectionString returns the connection string for RabbitMQ.
// Defaults to empty string if the environment variable is not set.
func (env *RabbitMQEnvironment) ConnectionString() string {
	return getEnvOrDefaultString(ENV_RABBITMQ_CONNECTIONSTRING, "")
}

// CacheLength returns the maximum number of items for the RabbitMQ cache. Defaults to 0 (no limit).
func (env *RabbitMQEnvironment) CacheLength() uint64 {
	return getEnvOrDefaultUint64(ENV_RABBITMQ_CACHELENGTH, 0)
}
