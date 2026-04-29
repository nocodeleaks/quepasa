package service

import (
	"encoding/json"
	"errors"
	"time"

	rabbitmq "github.com/nocodeleaks/quepasa/rabbitmq"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
)

// RabbitMQRequest is the outbound AMQP contract used by the dispatch module
// to publish events/messages to external queues.
type RabbitMQRequest struct {
	ConnectionString string
	Extra            interface{}
}

type RabbitMQResponse struct {
	RoutingKey       string
	Duration         time.Duration
	PayloadSizeBytes float64
	Cached           bool
}

type rabbitMQPayload struct {
	*whatsapp.WhatsappMessage
	Extra interface{} `json:"extra,omitempty"`
}

// PublishRabbitMQ sends one message to external RabbitMQ infrastructure.
// The caller owns domain-level metrics/state updates.
func PublishRabbitMQ(message *whatsapp.WhatsappMessage, request *RabbitMQRequest, logger *log.Entry) (*RabbitMQResponse, error) {
	if request == nil {
		return &RabbitMQResponse{}, nil
	}

	startTime := time.Now()
	routingKey := DetermineRoutingKey(message)

	if logger != nil {
		logger.Infof("publishing to QuePasa Exchange: %s with routing key: %s using connection: %s", rabbitmq.QuePasaExchangeName, routingKey, request.ConnectionString)
	}

	payload := &rabbitMQPayload{
		WhatsappMessage: message,
		Extra:           request.Extra,
	}

	payloadJSON, marshalErr := json.Marshal(&payload)
	payloadSizeBytes := float64(0)
	if marshalErr == nil {
		payloadSizeBytes = float64(len(payloadJSON))
	}

	result := &RabbitMQResponse{
		RoutingKey:       routingKey,
		PayloadSizeBytes: payloadSizeBytes,
	}

	client := rabbitmq.GetRabbitMQClient(request.ConnectionString)
	if client == nil {
		err := errors.New("failed to get rabbitmq client for connection: " + request.ConnectionString)
		rabbitmq.MessagePublishErrors.Inc()
		return result, err
	}

	if err := client.EnsureExchangeAndQueuesWithRetry(); err != nil {
		if logger != nil {
			logger.Warnf("QuePasa setup not ready yet, message will be cached: %s", err.Error())
		}
		rabbitmq.MessagePublishErrors.Inc()
	}

	client.PublishQuePasaMessage(routingKey, payload)

	result.Duration = time.Since(startTime)
	rabbitmq.MessagesPublished.Inc()

	if !client.IsConnectionReady() {
		rabbitmq.MessagePublishErrors.Inc()
		result.Cached = true
		if logger != nil {
			logger.Infof("message cached for QuePasa exchange: %s with routing key: %s (duration: %v, size: %.0f bytes)", rabbitmq.QuePasaExchangeName, routingKey, result.Duration, payloadSizeBytes)
		}
		return result, errors.New("rabbitmq connection not available, message cached")
	}

	if logger != nil {
		logger.Infof("message published to QuePasa exchange: %s with routing key: %s (duration: %v, size: %.0f bytes)", rabbitmq.QuePasaExchangeName, routingKey, result.Duration, payloadSizeBytes)
	}

	return result, nil
}

// DetermineRoutingKey maps message semantics to external queue routes.
func DetermineRoutingKey(message *whatsapp.WhatsappMessage) string {
	if message != nil && message.FromHistory {
		return rabbitmq.QuePasaRoutingKeyHistory
	}

	if isEventMessage(message) {
		return rabbitmq.QuePasaRoutingKeyEvents
	}

	return rabbitmq.QuePasaRoutingKeyProd
}

// isEventMessage identifies messages that should be routed to the events queue.
func isEventMessage(message *whatsapp.WhatsappMessage) bool {
	if message == nil {
		return false
	}

	eventTypes := map[whatsapp.WhatsappMessageType]bool{
		whatsapp.UnhandledMessageType: true,
	}

	if eventTypes[message.Type] {
		return true
	}

	if message.Type == whatsapp.ContactMessageType {
		return message.Edited && message.Attachment != nil
	}

	return message.Id == "readreceipt"
}
