package signalr

import dispatchservice "github.com/nocodeleaks/quepasa/dispatch/service"

type realtimeDispatchPublisher struct{}

func (realtimeDispatchPublisher) PublishMessage(payload interface{}) {
	event, ok := payload.(*dispatchservice.RealtimeServerMessage)
	if !ok || event == nil || event.Token == "" || event.Message == nil {
		return
	}

	SignalRHub.Dispatch(event.Token, event.Message)
}

func (realtimeDispatchPublisher) PublishLifecycle(payload interface{}) {
	event, ok := payload.(*dispatchservice.RealtimeLifecycleEvent)
	if !ok || event == nil || event.Token == "" {
		return
	}

	SignalRHub.DispatchLifecycle(event.Token, event)
}
