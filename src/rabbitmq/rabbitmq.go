package rabbitmq

import (
	environment "github.com/nocodeleaks/quepasa/environment"
)

// init pre-warms the RabbitMQ client for the connection string configured via
// RABBITMQ_CONNECTIONSTRING, if present. This is optional — bots that configure
// their own connection string via the API will create their clients on demand.
// The pre-warm runs in a goroutine so it never blocks application startup.
func init() {
	connStr := environment.Settings.RabbitMQ.ConnectionString
	if connStr == "" {
		return
	}

	go GetRabbitMQClient(connStr)
}
