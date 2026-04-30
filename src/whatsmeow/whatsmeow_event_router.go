package whatsmeow

import (
	"reflect"

	"go.mau.fi/whatsmeow/types/events"
)

// EventRouter maps whatsmeow event types to handler functions for a specific
// WhatsmeowHandlers instance. It enables registering new event types without
// touching EventsHandler.
type EventRouter struct {
	handlers map[reflect.Type]func(rawEvt interface{})
}

func newEventRouter() *EventRouter {
	return &EventRouter{
		handlers: make(map[reflect.Type]func(rawEvt interface{})),
	}
}

// register associates an event type (via a zero-value example) with its handler.
func (r *EventRouter) register(evtType reflect.Type, fn func(rawEvt interface{})) {
	r.handlers[evtType] = fn
}

// Dispatch looks up and calls the handler for rawEvt. Returns false when no
// handler is registered for that event type (caller should apply default logic).
func (r *EventRouter) Dispatch(rawEvt interface{}) bool {
	fn, ok := r.handlers[reflect.TypeOf(rawEvt)]
	if !ok {
		return false
	}
	fn(rawEvt)
	return true
}

// buildRouter creates and populates the event router for this handler instance.
// Closures capture source so each registered function is fully self-contained.
func (source *WhatsmeowHandlers) buildRouter() *EventRouter {
	r := newEventRouter()

	r.register(reflect.TypeOf(&events.Message{}), func(raw interface{}) {
		evt := raw.(*events.Message)
		go source.Message(*evt, "live")
	})

	// Calls
	r.register(reflect.TypeOf(&events.CallOffer{}), func(raw interface{}) {
		evt := raw.(*events.CallOffer)
		source.GetLogger().Infof("CallOffer: %v", evt)
		go source.CallMessage(evt.BasicCallMeta)
	})
	r.register(reflect.TypeOf(&events.CallOfferNotice{}), func(raw interface{}) {
		evt := raw.(*events.CallOfferNotice)
		source.GetLogger().Infof("CallOfferNotice: %v", evt)
		go source.CallMessage(evt.BasicCallMeta)
	})

	r.register(reflect.TypeOf(&events.Receipt{}), func(raw interface{}) {
		evt := raw.(*events.Receipt)
		go source.Receipt(*evt)
	})

	r.register(reflect.TypeOf(&events.Connected{}), func(_ interface{}) {
		source.onConnectedEvent()
	})

	r.register(reflect.TypeOf(&events.PushNameSetting{}), func(_ interface{}) {
		presence := source.GetPresence()
		source.SendPresence(presence, "'push name setting' event")
	})

	r.register(reflect.TypeOf(&events.Disconnected{}), func(_ interface{}) {
		logentry := source.GetLogger()
		if source.Client.EnableAutoReconnect {
			logentry.Infof("disconnected from server, dont worry, reconnecting")
		} else {
			logentry.Warn("disconnected from server")
		}
		if source.hasWAHandlers() {
			go source.WAHandlers.OnDisconnected("network", "Server closed connection")
		}
	})

	r.register(reflect.TypeOf(&events.ConnectFailure{}), func(raw interface{}) {
		evt := raw.(*events.ConnectFailure)
		source.onConnectFailureEvent(evt)
	})

	r.register(reflect.TypeOf(&events.StreamError{}), func(raw interface{}) {
		evt := raw.(*events.StreamError)
		source.onStreamErrorEvent(evt)
	})

	r.register(reflect.TypeOf(&events.TemporaryBan{}), func(raw interface{}) {
		evt := raw.(*events.TemporaryBan)
		source.onTemporaryBanEvent(evt)
	})

	r.register(reflect.TypeOf(&events.StreamReplaced{}), func(_ interface{}) {
		source.GetLogger().Warn("stream replaced by another client")
		if source.hasWAHandlers() {
			go source.WAHandlers.OnDisconnected("stream_replaced", "Another client connected with same session")
		}
	})

	r.register(reflect.TypeOf(&events.LoggedOut{}), func(raw interface{}) {
		evt := raw.(*events.LoggedOut)
		source.OnLoggedOutEvent(*evt)
	})

	r.register(reflect.TypeOf(&events.HistorySync{}), func(raw interface{}) {
		evt := raw.(*events.HistorySync)
		if source.HandleHistorySync() {
			go source.OnHistorySyncEvent(*evt)
		}
	})

	r.register(reflect.TypeOf(&events.AppStateSyncComplete{}), func(raw interface{}) {
		evt := raw.(*events.AppStateSyncComplete)
		source.onAppStateSyncCompleteEvent(evt)
	})

	r.register(reflect.TypeOf(&events.JoinedGroup{}), func(raw interface{}) {
		evt := raw.(*events.JoinedGroup)
		source.JoinedGroup(*evt)
	})

	r.register(reflect.TypeOf(&events.Contact{}), func(raw interface{}) {
		evt := raw.(*events.Contact)
		go OnEventContact(source, *evt)
	})

	r.register(reflect.TypeOf(&events.PairError{}), func(raw interface{}) {
		evt := raw.(*events.PairError)
		source.GetLogger().Errorf("pair error event: %v", evt)
	})

	r.register(reflect.TypeOf(&events.OfflineSyncPreview{}), func(raw interface{}) {
		evt := raw.(*events.OfflineSyncPreview)
		go source.OnOfflineSyncPreview(*evt)
	})

	r.register(reflect.TypeOf(&events.OfflineSyncCompleted{}), func(raw interface{}) {
		evt := raw.(*events.OfflineSyncCompleted)
		go source.OnOfflineSyncCompleted(*evt)
	})

	r.register(reflect.TypeOf(&events.UndecryptableMessage{}), func(raw interface{}) {
		evt := raw.(*events.UndecryptableMessage)
		go source.UndecryptableMessage(*evt)
	})

	// Unimplemented events — trace-logged and optionally dispatched as unhandled
	unimplementedHandler := func(raw interface{}) {
		source.GetLogger().Tracef("event not implemented yet: %v", reflect.TypeOf(raw))
		if source.ShouldDispatchUnhandled() {
			go source.DispatchUnhandledEvent(raw, reflect.TypeOf(raw).String())
		}
	}
	r.register(reflect.TypeOf(&events.AppState{}), unimplementedHandler)
	r.register(reflect.TypeOf(&events.CallTerminate{}), unimplementedHandler)
	r.register(reflect.TypeOf(&events.DeleteChat{}), unimplementedHandler)
	r.register(reflect.TypeOf(&events.DeleteForMe{}), unimplementedHandler)
	r.register(reflect.TypeOf(&events.MarkChatAsRead{}), unimplementedHandler)
	r.register(reflect.TypeOf(&events.Mute{}), unimplementedHandler)
	r.register(reflect.TypeOf(&events.Pin{}), unimplementedHandler)
	r.register(reflect.TypeOf(&events.PushName{}), unimplementedHandler)
	r.register(reflect.TypeOf(&events.GroupInfo{}), unimplementedHandler)

	r.register(reflect.TypeOf(&events.QR{}), func(raw interface{}) {
		evt := raw.(*events.QR)
		source.OnQREvent(evt)
	})

	r.register(reflect.TypeOf(&events.PairSuccess{}), func(raw interface{}) {
		evt := raw.(*events.PairSuccess)
		source.OnPairSuccessEvent(evt)
	})

	return r
}
