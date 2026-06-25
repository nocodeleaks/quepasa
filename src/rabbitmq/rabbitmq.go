package rabbitmq

import (
	environment "github.com/nocodeleaks/quepasa/environment"
)

// Automatically registers the RabbitMQ configuration
func init() {
	rabbitmq_connection_string := environment.Settings.RabbitMQ.ConnectionString
	if len(rabbitmq_connection_string) > 0 {
		rabbitmq_queue := environment.Settings.RabbitMQ.Queue
		if len(rabbitmq_queue) > 0 {
			RabbitMQQueueDefault = rabbitmq_queue
		}

		cachelength := environment.Settings.RabbitMQ.CacheLength
		InitializeRabbitMQClient(rabbitmq_connection_string, cachelength)
	}
}
