package rabbitmq

import "time"

// RabbitMQMessage represents the structure of the JSON to be sent to the queue.
// The `json:"..."` tags are used to define the field names in the JSON.
type RabbitMQMessage struct {
	ID        string    `json:"id"`
	Payload   any       `json:"payload"` // Alterado de 'interface{}' para 'any'
	Timestamp time.Time `json:"timestamp"`
	// TargetQueue stores the name of the queue where the message should be published.
	// This is crucial for cached messages to know their original destination.
	TargetQueue string `json:"target_queue"`
}
