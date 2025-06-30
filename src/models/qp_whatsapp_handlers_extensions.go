package models

import (
	"fmt"
	"reflect"
	"strings"

	rabbitmq "github.com/nocodeleaks/quepasa/rabbitmq"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// IsValidForDispatch validates if a given WhatsApp message payload is suitable for further processing
// and dispatch. It checks for specific conditions that would indicate the message should be ignored.
//
// Parameters:
//
//	payload *whatsapp.WhatsappMessage: A pointer to the WhatsApp message payload to be validated.
//
// Returns:
//
//	string: An empty string if the message is valid for dispatch.
//	        A non-empty string containing an explanation message if the message should be ignored.
func IsValidForDispatch(payload *whatsapp.WhatsappMessage) string {
	// Ignores messages with 'Discard' or 'Unknown' types, as these are typically not meant for
	// application-level processing or indicate an unhandled message format.

	if ENV.DebugEvents() && (payload.Type == whatsapp.DiscardMessageType || payload.Type == whatsapp.UnknownMessageType) {
		return ""
	}

	if payload.Type == whatsapp.DiscardMessageType || payload.Type == whatsapp.UnknownMessageType {
		return fmt.Sprintf("ignoring discard|unknown message type on webhook request: %v", reflect.TypeOf(&payload))
	}

	// Ignores text messages that are empty or contain only whitespace.
	// Such messages generally don't carry meaningful information for the application.
	if payload.Type == whatsapp.TextMessageType && len(strings.TrimSpace(payload.Text)) <= 0 {
		return fmt.Sprintf("ignoring empty text message on webhook request: %s", payload.Id)
	}

	// If none of the above conditions are met, the message is considered valid for dispatch.
	return ""
}

// RabbitMQPublish validates a WhatsApp message payload and, if valid,
// publishes it asynchronously to the default RabbitMQ queue.
//
// This function first calls IsValidForDispatch to check if the message
// should be processed. If IsValidForDispatch returns a non-empty string,
// indicating the message should be ignored, the function exits early.
// Otherwise, it dispatches the message to RabbitMQ using the global
// RabbitMQClientInstance in a new goroutine to avoid blocking the caller.
//
// Parameters:
//
//	payload *whatsapp.WhatsappMessage: A pointer to the WhatsApp message
//	                                   payload to be published.
func RabbitMQPublish(payload *whatsapp.WhatsappMessage) {

	// Validate the message payload. If it's not valid for dispatch,
	// IsValidForDispatch will return a reason string.
	reason := IsValidForDispatch(payload)
	if len(reason) > 0 {
		// If a reason is returned, it means the message should be ignored.
		// No further action is needed, so we simply return.
		return
	}

	// If the message is valid, publish it to RabbitMQ.
	// This is done in a new goroutine to ensure the publishing process
	// doesn't block the execution of the calling function, allowing for
	// non-blocking message processing.
	go rabbitmq.RabbitMQClientInstance.PublishMessage(payload)
}
