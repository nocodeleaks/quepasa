package models

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/nocodeleaks/quepasa/library"
	rabbitmq "github.com/nocodeleaks/quepasa/rabbitmq"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

type QpRabbitMQConfig struct {
	library.LogStruct // logging

	// Optional whatsapp options
	// ------------------------
	whatsapp.WhatsappOptions

	// RabbitMQ Connection Settings
	ConnectionString string `json:"connection_string,omitempty"` // Full RabbitMQ connection string (amqp://user:pass@host:port/vhost)
	ExchangeName     string `json:"exchange_name,omitempty"`     // RabbitMQ exchange name for routing
	RoutingKey       string `json:"routing_key,omitempty"`       // RabbitMQ routing key for exchange routing
	QueueHistory     string `json:"queue_history,omitempty"`     // RabbitMQ history queue name (optional)

	// Configuration Options
	ForwardInternal bool        `json:"forwardinternal,omitempty"` // forward internal msg from api
	TrackId         string      `json:"trackid,omitempty"`         // identifier of remote system to avoid loop
	Extra           interface{} `json:"extra,omitempty"`           // extra info to append on payload

	// Status Tracking
	Failure   *time.Time `json:"failure,omitempty"` // first failure timestamp
	Success   *time.Time `json:"success,omitempty"` // last success timestamp
	Timestamp *time.Time `json:"timestamp,omitempty"`

	// just for logging and response headers
	Wid string `json:"-"`
}

//#region VIEWS TRICKS

// IsFailureMoreRecent checks if the last failure is more recent than the last success
func (source QpRabbitMQConfig) IsFailureMoreRecent() bool {
	if source.Failure == nil {
		return false
	}
	if source.Success == nil {
		return true
	}
	return source.Failure.After(*source.Success)
}

// HasRecentSuccess checks if there's a recent success and no more recent failure
func (source QpRabbitMQConfig) HasRecentSuccess() bool {
	if source.Success == nil {
		return false
	}
	if source.Failure == nil {
		return true
	}
	return source.Success.After(*source.Failure)
}

func (source QpRabbitMQConfig) GetReadReceipts() bool {
	return source.ReadReceipts.Boolean()
}

func (source QpRabbitMQConfig) IsSetReadReceipts() bool {
	return source.ReadReceipts != whatsapp.UnSetBooleanType
}

func (source QpRabbitMQConfig) GetGroups() bool {
	return source.Groups.Boolean()
}

func (source QpRabbitMQConfig) IsSetGroups() bool {
	return source.Groups != whatsapp.UnSetBooleanType
}

func (source QpRabbitMQConfig) GetIndividuals() bool {
	return source.Individuals.Boolean()
}

func (source QpRabbitMQConfig) IsSetIndividuals() bool {
	return source.Individuals != whatsapp.UnSetBooleanType
}

func (source QpRabbitMQConfig) GetBroadcasts() bool {
	return source.Broadcasts.Boolean()
}

func (source QpRabbitMQConfig) IsSetBroadcasts() bool {
	return source.Broadcasts != whatsapp.UnSetBooleanType
}

func (source QpRabbitMQConfig) GetCalls() bool {
	return source.Calls.Boolean()
}

func (source QpRabbitMQConfig) IsSetCalls() bool {
	return source.Calls != whatsapp.UnSetBooleanType
}

func (source QpRabbitMQConfig) IsSetExtra() bool {
	return source.Extra != nil
}

// GetExtraText converts extra field to JSON string for template display
func (source QpRabbitMQConfig) GetExtraText() string {
	if source.Extra != nil {
		if bytes, err := json.Marshal(source.Extra); err == nil {
			return string(bytes)
		}
	}
	return ""
}

//#endregion

// PublishMessage publishes message to RabbitMQ exchange using the specific connection string and routing key
func (source *QpRabbitMQConfig) PublishMessage(message *whatsapp.WhatsappMessage) (err error) {
	startTime := time.Now()

	// updating log
	logentry := source.LogWithField(LogFields.MessageId, message.Id)
	logentry.Infof("publishing to QuePasa Exchange: %s using connection: %s", rabbitmq.QuePasaExchangeName, source.ConnectionString)

	payload := &QpRabbitMQPayload{
		WhatsappMessage: message,
		Extra:           source.Extra,
	}

	// Calculate payload size for metrics
	payloadJson, marshalErr := json.Marshal(&payload)
	var payloadSizeBytes float64
	if marshalErr == nil {
		payloadSizeBytes = float64(len(payloadJson))
	}

	// Get or create RabbitMQ client for this specific connection string
	client := rabbitmq.GetRabbitMQClient(source.ConnectionString)
	if client != nil {
		// Ensure QuePasa Exchange and Queues exist
		err = client.EnsureExchangeAndQueues()
		if err != nil {
			logentry.Errorf("failed to ensure QuePasa exchange and queues: %s", err.Error())

			// Record RabbitMQ publish error
			rabbitmq.MessagePublishErrors.Inc()
			return err
		}

		// Determine routing key based on message type
		routingKey := source.DetermineRoutingKey(message)
		client.PublishQuePasaMessage(routingKey, payload)

		// Always increment RabbitMQ messages published counter
		rabbitmq.MessagesPublished.Inc()

		// Record publish duration
		duration := time.Since(startTime)
		// Note: Publish duration and message size metrics removed - not implemented in rabbitmq module

		currentTime := time.Now().UTC()
		source.Failure = nil
		source.Success = &currentTime

		logentry.Infof("message published to QuePasa exchange: %s with routing key: %s (duration: %v, size: %.0f bytes)", rabbitmq.QuePasaExchangeName, routingKey, duration, payloadSizeBytes)
	} else {
		err = errors.New("failed to get rabbitmq client for connection: " + source.ConnectionString)
		logentry.Errorf("rabbitmq client not available for connection %s: %s", source.ConnectionString, err.Error())

		// Record RabbitMQ publish error
		rabbitmq.MessagePublishErrors.Inc()

		currentTime := time.Now().UTC()
		if source.Failure == nil {
			source.Failure = &currentTime
		}
	}

	return
}

// ToDispatching converts QpRabbitMQConfig to QpDispatching
func (source *QpRabbitMQConfig) ToDispatching() *QpDispatching {
	// Use ConnectionString as the unique identifier, fallback to ExchangeName if not set
	connectionString := source.ConnectionString
	if connectionString == "" {
		connectionString = source.ExchangeName
	}

	return &QpDispatching{
		LogStruct:        source.LogStruct,
		WhatsappOptions:  source.WhatsappOptions,
		ConnectionString: connectionString,
		Type:             DispatchingTypeRabbitMQ,
		ForwardInternal:  source.ForwardInternal,
		TrackId:          source.TrackId,
		Extra:            source.Extra,
		Failure:          source.Failure,
		Success:          source.Success,
		Timestamp:        source.Timestamp,
		Wid:              source.Wid,
	}
}

//#region IMPLEMENTING WHATSAPP OPTIONS INTERFACE

func (source *QpRabbitMQConfig) GetOptions() *whatsapp.WhatsappOptions {
	return &source.WhatsappOptions
}

func (source *QpRabbitMQConfig) Save(reason string) error {
	// Convert to dispatching and save via the server
	// This method would need access to the server to save properly
	// For now, return nil as the saving is handled in the toggle methods
	return nil
}

//#endregion

// GetUniqueIdentifier returns the unique identifier for this RabbitMQ config
// Uses ConnectionString if available, otherwise falls back to ExchangeName
func (source *QpRabbitMQConfig) GetUniqueIdentifier() string {
	if source.ConnectionString != "" {
		return source.ConnectionString
	}
	return source.ExchangeName
}

// ValidateConfig validates the RabbitMQ configuration
func (source *QpRabbitMQConfig) ValidateConfig() error {
	if source.ConnectionString == "" && source.ExchangeName == "" {
		return errors.New("either connection_string or exchange_name is required")
	}

	// If only exchange_name is provided, it will be used as connection_string
	if source.ConnectionString == "" {
		source.ConnectionString = source.ExchangeName
	}

	// Set default exchange name if not provided
	if source.ExchangeName == "" {
		source.ExchangeName = "quepasa.exchange"
	}

	// Set default routing key if not provided (will be overridden by intelligent routing)
	if source.RoutingKey == "" {
		source.RoutingKey = "prod" // Default to production
	}

	return nil
}

// DetermineRoutingKey determines the appropriate routing key based on message type and content
// Returns one of the QuePasa standard routing keys
func (source *QpRabbitMQConfig) DetermineRoutingKey(message *whatsapp.WhatsappMessage) string {
	// Check if message is from history sync
	if message.FromHistory {
		return rabbitmq.QuePasaRoutingKeyHistory
	}

	// Check if message type is unhandled (debug/system messages)
	if message.Type == whatsapp.UnhandledMessageType {
		return rabbitmq.QuePasaRoutingKeyEvents
	}

	// Special-case: read-receipt system payloads created by the handlers use id "readreceipt"
	// Route them to events so consumers receive read receipts in the events queue
	if message.Id == "readreceipt" {
		return rabbitmq.QuePasaRoutingKeyEvents
	}

	// Check if message is a system message
	if message.Type == whatsapp.SystemMessageType {
		return rabbitmq.QuePasaRoutingKeyProd
	}

	// Check if message is a contact message with edited=true and has attachment
	if message.Type == whatsapp.ContactMessageType {
		if message.Edited && message.Attachment != nil {
			return rabbitmq.QuePasaRoutingKeyEvents
		}
	}

	// Check if message is a call message (could be considered events)
	if message.Type == whatsapp.CallMessageType {
		return rabbitmq.QuePasaRoutingKeyProd
	}

	// Check if message is a revoke message
	if message.Type == whatsapp.RevokeMessageType {
		return rabbitmq.QuePasaRoutingKeyProd
	}

	// Default to production queue for normal messages
	return rabbitmq.QuePasaRoutingKeyProd
}
