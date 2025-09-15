package models

// Response specifically for RabbitMQ API requests
type QpRabbitMQResponse struct {
	QpResponse
	Affected uint                `json:"affected,omitempty"` // items affected
	RabbitMQ []*QpRabbitMQConfig `json:"rabbitmq,omitempty"` // current rabbitmq items
}
