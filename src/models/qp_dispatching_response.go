package models

// Response for dispatching API requests
type QpDispatchingResponse struct {
	QpResponse
	Affected        uint                `json:"affected,omitempty"`    // items affected
	Dispatching     []*QpDispatching    `json:"dispatching,omitempty"` // current dispatching items
	Webhooks        []*QpWebhook        `json:"webhooks,omitempty"`    // webhook items (for backward compatibility)
	RabbitMQConfigs []*QpRabbitMQConfig `json:"rabbitmq,omitempty"`    // rabbitmq items
}
