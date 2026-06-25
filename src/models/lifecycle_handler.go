package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	events "github.com/nocodeleaks/quepasa/events"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// LifecycleHandler manages WhatsApp session lifecycle events (connected, disconnected, logged out, etc).
// This handler is responsible for:
// - Reacting to connection state changes
// - Recording diagnostic information
// - Publishing lifecycle events to webhooks and realtime transports
// - Creating synthetic event messages for session lifecycle transitions
type LifecycleHandler struct {
	dispatcher *DispatchingHandler
}

// NewLifecycleHandler creates a new lifecycle handler for the given dispatcher.
func NewLifecycleHandler(dispatcher *DispatchingHandler) *LifecycleHandler {
	return &LifecycleHandler{
		dispatcher: dispatcher,
	}
}

// OnConnected handles the connected lifecycle event.
func (lh *LifecycleHandler) OnConnected() {
	if lh.dispatcher == nil {
		return
	}

	// one step at a time
	if lh.dispatcher.Server() != nil {

		// Reset server start timestamp on connection (uptime starts from connection moment)
		lh.dispatcher.Server().Timestamps.Start = time.Now().UTC()

		// marking unverified and wait for more analyses
		err := lh.dispatcher.Server().MarkVerified(true)
		if err != nil {
			logger := lh.dispatcher.Server().GetLogger()
			logger.Errorf("error on mark verified after connected: %s", err.Error())
		}
		err = lh.dispatcher.Server().ClearConnectionIssue("connected and cleared connection issue")
		if err != nil {
			logger := lh.dispatcher.Server().GetLogger()
			logger.Errorf("error clearing connection issue after connected: %s", err.Error())
		}
		lh.publishRealtimeLifecycle("connected", "", "")
		lh.publishLifecycleEvent("session.lifecycle.connected", "success", map[string]string{
			"state":    lh.dispatcher.Server().GetState().String(),
			"verified": formatLifecycleBool(lh.dispatcher.Server().Verified),
		})
	}
}

// OnDisconnected handles the disconnected lifecycle event.
func (lh *LifecycleHandler) OnDisconnected(cause string, details string) {
	if lh.dispatcher == nil {
		return
	}

	if lh.dispatcher.Server() == nil {
		return
	}

	logger := lh.dispatcher.GetLogger()
	logger.Infof("dispatching server disconnect event: %s - %s", cause, details)

	if err := lh.dispatcher.Server().RecordDisconnect(cause, details); err != nil {
		logger.Errorf("failed to persist disconnect diagnostic: %s", err.Error())
	}

	// Get phone number and wid from server
	phone := lh.dispatcher.Server().GetNumber()
	wid := lh.dispatcher.Server().GetWId()

	// Create description with cause and details in text
	description := fmt.Sprintf("WhatsApp disconnected: %s", cause)
	if details != "" {
		description = fmt.Sprintf("%s - %s", description, details)
	}

	// Create disconnect event message with JSON details
	eventData := map[string]interface{}{
		"event":     "disconnected",
		"cause":     cause,
		"details":   details,
		"wid":       wid,
		"phone":     phone,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	message := &whatsapp.WhatsappMessage{
		Id:        uuid.New().String(),
		Timestamp: time.Now().UTC(),
		Type:      whatsapp.SystemMessageType,
		FromMe:    false,
		Chat:      whatsapp.WASYSTEMCHAT,
		Text:      description,
		Info:      eventData,
	}

	// Add to cache and send through dispatchers
	lh.dispatcher.AppendMsgToCache(message, "disconnected")

	lh.publishRealtimeLifecycle("disconnected", cause, details)
	lh.publishLifecycleEvent("session.lifecycle.disconnected", "success", map[string]string{
		"cause":    cause,
		"state":    lh.dispatcher.Server().GetState().String(),
		"verified": formatLifecycleBool(lh.dispatcher.Server().Verified),
	})
}

// LoggedOut handles the logged out lifecycle event.
func (lh *LifecycleHandler) LoggedOut(reason string) {
	if lh.dispatcher == nil {
		return
	}

	// one step at a time
	if lh.dispatcher.Server() != nil {

		msg := "logged out !"
		if len(reason) > 0 {
			msg += " reason: " + reason
		}

		logger := lh.dispatcher.GetLogger()
		logger.Warn(msg)

		// Persist a diagnostic so the frontend can explain why the session is
		// now unverified instead of only showing the generic state.
		lh.dispatcher.Server().RecordLogout(reason)

		lh.publishRealtimeLifecycle("logged_out", reason, "")
	}
}

// OnStopped handles the manually stopped lifecycle event.
func (lh *LifecycleHandler) OnStopped(cause string) {
	if lh.dispatcher == nil {
		return
	}

	if lh.dispatcher.Server() == nil {
		return
	}

	logger := lh.dispatcher.GetLogger()
	logger.Infof("dispatching server stop event: %s", cause)

	// Get phone number and wid from server
	phone := lh.dispatcher.Server().GetNumber()
	wid := lh.dispatcher.Server().GetWId()

	// Create description
	description := fmt.Sprintf("WhatsApp server manually stopped: %s", cause)

	// Create stop event message with JSON details
	eventData := map[string]interface{}{
		"event":     "stopped",
		"cause":     cause,
		"wid":       wid,
		"phone":     phone,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	message := &whatsapp.WhatsappMessage{
		Id:        uuid.New().String(),
		Timestamp: time.Now().UTC(),
		Type:      whatsapp.SystemMessageType,
		FromMe:    false,
		Chat:      whatsapp.WASYSTEMCHAT,
		Text:      description,
		Info:      eventData,
	}

	// Add to cache and send through dispatchers
	lh.dispatcher.AppendMsgToCache(message, "stopped")

	lh.publishRealtimeLifecycle("stopped", cause, "")
	lh.publishLifecycleEvent("session.lifecycle.stopped", "success", map[string]string{
		"cause":    cause,
		"state":    lh.dispatcher.Server().GetState().String(),
		"verified": formatLifecycleBool(lh.dispatcher.Server().Verified),
	})
}

// OnDeleted handles the deleted lifecycle event.
func (lh *LifecycleHandler) OnDeleted(cause string) {
	if lh.dispatcher == nil {
		return
	}

	if lh.dispatcher.Server() == nil {
		return
	}

	logger := lh.dispatcher.GetLogger()
	logger.Infof("dispatching server delete event: %s", cause)

	message := NewServerDeletedEvent(lh.dispatcher.Server(), cause, nil)

	// Add to cache and send through dispatchers
	lh.dispatcher.AppendMsgToCache(message, "deleted")

	lh.publishRealtimeLifecycle("deleted", cause, "")
	lh.publishLifecycleEvent("session.lifecycle.deleted", "success", map[string]string{
		"cause":    cause,
		"state":    lh.dispatcher.Server().GetState().String(),
		"verified": formatLifecycleBool(lh.dispatcher.Server().Verified),
	})
}

// publishLifecycleEvent publishes internal lifecycle event for analytics and webhooks.
func (lh *LifecycleHandler) publishLifecycleEvent(name string, status string, attributes map[string]string) {
	if lh.dispatcher == nil || lh.dispatcher.Server() == nil {
		return
	}

	if attributes == nil {
		attributes = map[string]string{}
	}
	attributes["wid"] = lh.dispatcher.Server().GetWId()

	events.Publish(events.Event{
		Name:       name,
		Source:     "models.lifecycle_handler",
		Status:     status,
		Attributes: attributes,
	})
}

// publishRealtimeLifecycle publishes lifecycle event to realtime transports (SignalR, cable).
func (lh *LifecycleHandler) publishRealtimeLifecycle(kind string, cause string, details string) {
	if lh.dispatcher == nil || lh.dispatcher.Server() == nil {
		return
	}

	lh.dispatcher.LifecyclePublisher().PublishLifecycle(&DispatchingLifecycleEvent{
		Kind:      kind,
		Token:     lh.dispatcher.Server().Token,
		User:      lh.dispatcher.Server().GetUser(),
		Wid:       lh.dispatcher.Server().GetWId(),
		Phone:     lh.dispatcher.Server().GetNumber(),
		State:     lh.dispatcher.Server().GetState().String(),
		Verified:  lh.dispatcher.Server().Verified,
		Cause:     cause,
		Details:   details,
		Timestamp: time.Now().UTC(),
	})
}

// formatLifecycleBool formats a boolean as a string for lifecycle events.
func formatLifecycleBool(value bool) string {
	if value {
		return "true"
	}
	return "false"
}
