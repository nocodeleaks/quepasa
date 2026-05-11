package models

import (
	"sync"
	"time"

	dispatchservice "github.com/nocodeleaks/quepasa/dispatch/service"
	library "github.com/nocodeleaks/quepasa/library"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// MessageDispatcher manages message triggering and subscriber orchestration.
// This handler is responsible for:
// - Triggering message dispatch flow to registered subscribers
// - Managing subscriber registration and lifecycle
// - Coordinating async message handling
type MessageDispatcher struct {
	dispatcher   *DispatchingHandler
	syncRegister *sync.Mutex
	subscribers  []QpDispatchingHandlerInterface
}

// NewMessageDispatcher creates a new message dispatcher for the given handler.
func NewMessageDispatcher(handler *DispatchingHandler) *MessageDispatcher {
	return &MessageDispatcher{
		dispatcher:   handler,
		syncRegister: &sync.Mutex{},
		subscribers:  make([]QpDispatchingHandlerInterface, 0),
	}
}

// Trigger sends a message through the dispatch pipeline to all registered subscribers.
func (md *MessageDispatcher) Trigger(payload *whatsapp.WhatsappMessage) {
	if md.dispatcher == nil {
		return
	}

	handlerCallbacks := make([]dispatchservice.HandlerSubscriber, 0, len(md.subscribers))
	for _, handler := range md.subscribers {
		handlerCallbacks = append(handlerCallbacks, handler)
	}

	request := &dispatchservice.HandlerFlowRequest{
		Payload: payload,
		Validate: func(message *whatsapp.WhatsappMessage) string {
			return IsValidForDispatch(message)
		},
		OnInvalid: func(reason string, message *whatsapp.WhatsappMessage) {
			logentry := md.dispatcher.GetLogger()
			logentry.Debug(reason)

			jsonPayload := library.ToJson(message)
			logentry.Logger.Debugf("unhandled payload: %s", jsonPayload)
		},
		MarkEventTimestamp: func() {
			if md.dispatcher.Server() == nil {
				return
			}
			currentTime := time.Now().UTC()
			md.dispatcher.Server().Timestamps.Event = &currentTime
		},
		MarkMessageTimestamp: func() {
			if md.dispatcher.Server() == nil {
				return
			}
			currentTime := time.Now().UTC()
			md.dispatcher.Server().Timestamps.Message = &currentTime
		},
		SetMessageWid: func(message *whatsapp.WhatsappMessage) {
			if md.dispatcher.Server() == nil || message == nil {
				return
			}
			message.Wid = md.dispatcher.GetWId()
		},
		PublishRealtime: func(message *whatsapp.WhatsappMessage) {
			if md.dispatcher.Server() == nil {
				return
			}
			enriched := CloneAndEnrichMessageForServer(md.dispatcher.Server(), message)
			dispatchservice.PublishRealtimeMessage(&dispatchservice.RealtimeServerMessage{
				Token:   md.dispatcher.Server().Token,
				User:    md.dispatcher.Server().GetUser(),
				WID:     md.dispatcher.Server().GetWId(),
				State:   md.dispatcher.Server().GetState().String(),
				Message: enriched,
			})
		},
		HandlerCallbacks: handlerCallbacks,
	}

	dispatchservice.GetInstance().DispatchHandlerFlow(request)
}

// Register adds a new event handler to be triggered on message dispatch.
func (md *MessageDispatcher) Register(evt QpDispatchingHandlerInterface) {
	md.syncRegister.Lock()
	defer md.syncRegister.Unlock()

	if !md.isRegisteredLocked(evt) {
		md.subscribers = append(md.subscribers, evt)
	}
}

// UnRegister removes a specific event handler from the dispatcher.
func (md *MessageDispatcher) UnRegister(evt QpDispatchingHandlerInterface) {
	md.syncRegister.Lock()
	defer md.syncRegister.Unlock()

	newHandlers := make([]QpDispatchingHandlerInterface, 0, len(md.subscribers))
	for _, v := range md.subscribers {
		if v != evt {
			newHandlers = append(newHandlers, v)
		}
	}

	md.subscribers = newHandlers
}

// Clear removes all registered event handlers.
func (md *MessageDispatcher) Clear() {
	md.syncRegister.Lock()
	defer md.syncRegister.Unlock()

	md.subscribers = nil
}

// IsAttached reports whether any event handlers are registered.
func (md *MessageDispatcher) IsAttached() bool {
	md.syncRegister.Lock()
	defer md.syncRegister.Unlock()

	return len(md.subscribers) > 0
}

// IsRegistered checks if a specific handler is registered.
func (md *MessageDispatcher) IsRegistered(evt interface{}) bool {
	md.syncRegister.Lock()
	defer md.syncRegister.Unlock()

	return md.isRegisteredLocked(evt)
}

// isRegisteredLocked is the internal version that assumes lock is already held.
func (md *MessageDispatcher) isRegisteredLocked(evt interface{}) bool {
	for _, v := range md.subscribers {
		if v == evt {
			return true
		}
	}

	return false
}

// HasDispatchingSubscriber reports whether the default outbound dispatching
// subscriber is already attached to this dispatcher.
func (md *MessageDispatcher) HasDispatchingSubscriber() bool {
	md.syncRegister.Lock()
	defer md.syncRegister.Unlock()

	for _, subscriber := range md.subscribers {
		if _, ok := subscriber.(dispatchingSubscriber); ok {
			return true
		}
	}

	return false
}

// Subscribers returns a copy of the current subscriber list.
func (md *MessageDispatcher) Subscribers() []QpDispatchingHandlerInterface {
	md.syncRegister.Lock()
	defer md.syncRegister.Unlock()

	result := make([]QpDispatchingHandlerInterface, len(md.subscribers))
	copy(result, md.subscribers)
	return result
}
