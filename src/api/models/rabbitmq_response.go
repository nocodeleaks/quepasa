package api

import models "github.com/nocodeleaks/quepasa/models"

// RabbitMQResponse is the API transport shape for RabbitMQ configuration endpoints.
type RabbitMQResponse struct {
	models.QpResponse
	Affected uint                       `json:"affected,omitempty"`
	RabbitMQ []*models.QpRabbitMQConfig `json:"rabbitmq,omitempty"`
}
