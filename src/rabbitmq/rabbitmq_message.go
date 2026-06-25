package rabbitmq

import "time"

// RabbitMQMessage represents the structure of the JSON to be sent to RabbitMQ Exchange.
// QuePasa always uses exchange-based routing, never direct queue publishing.
type RabbitMQMessage struct {
	ID         string    `json:"id"`
	Payload    any       `json:"payload"`
	Timestamp  time.Time `json:"timestamp"`
	Exchange   string    `json:"exchange"`    // Exchange name for routing
	RoutingKey string    `json:"routing_key"` // Routing key for exchange routing
}
