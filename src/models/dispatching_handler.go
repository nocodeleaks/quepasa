package models

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/google/uuid"
	dispatchservice "github.com/nocodeleaks/quepasa/dispatch/service"
	events "github.com/nocodeleaks/quepasa/events"
	library "github.com/nocodeleaks/quepasa/library"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// Serviço que controla os servidores / bots individuais do whatsapp
type DispatchingHandler struct {
	QpWhatsappMessages
	library.LogStruct // logging

	server *QpWhatsappServer

	syncRegister *sync.Mutex

	// Appended events handler
	aeh []QpDispatchingHandlerInterface
}

type dispatchingSubscriber interface {
	QpDispatchingHandlerInterface
	isDispatchingSubscriber()
}

// DispatchingLifecycleEvent defines transport-agnostic lifecycle payload emitted by models.
type DispatchingLifecycleEvent struct {
	Kind      string
	Token     string
	User      string
	Wid       string
	Phone     string
	State     string
	Verified  bool
	Cause     string
	Details   string
	Timestamp time.Time
}

// DispatchingLifecyclePublisher sends lifecycle events to transport adapters.
type DispatchingLifecyclePublisher interface {
	PublishLifecycle(event *DispatchingLifecycleEvent)
}

type noopDispatchingLifecyclePublisher struct{}

func (noopDispatchingLifecyclePublisher) PublishLifecycle(event *DispatchingLifecycleEvent) {}

// GlobalDispatchingLifecyclePublisher is injected by runtime/bootstrap to keep
// models independent from concrete lifecycle transport implementations.
var GlobalDispatchingLifecyclePublisher DispatchingLifecyclePublisher = noopDispatchingLifecyclePublisher{}

// Returns whatsapp controller id on E164
// Ex: 5521967609494
func (source *DispatchingHandler) GetWId() string {
	if source == nil || source.server == nil {
		return ""
	}

	return source.server.GetWId()
}

func (source *DispatchingHandler) HandleGroups() bool {
	global := whatsapp.Options

	var local whatsapp.WhatsappBoolean
	if source.server != nil {
		local = source.server.Groups
	}
	return global.HandleGroups(local)
}

func (source *DispatchingHandler) HandleBroadcasts() bool {
	global := whatsapp.Options

	var local whatsapp.WhatsappBoolean
	if source.server != nil {
		local = source.server.Broadcasts
	}
	return global.HandleBroadcasts(local)
}

//#region EVENTS FROM WHATSAPP SERVICE

// Process messages received from whatsapp service
func (source *DispatchingHandler) Message(msg *whatsapp.WhatsappMessage, from string) {

	// should skip groups ?
	if !source.HandleGroups() && msg.FromGroup() {
		return
	}

	// should skip broadcast ?
	if !source.HandleBroadcasts() && msg.FromBroadcast() {
		return
	}

	// messages sended with chat title
	if len(msg.Chat.Title) == 0 {
		msg.Chat.Title = source.server.GetChatTitle(msg.Chat.Id)
	}

	if len(msg.InReply) > 0 {
		cached, err := source.QpWhatsappMessages.GetById(msg.InReply)
		if err == nil {
			maxlength := ENV.SynopsisLength() - 4
			if uint64(len(cached.Text)) > maxlength {
				msg.Synopsis = cached.Text[0:maxlength] + " ..."
			} else {
				msg.Synopsis = cached.Text
			}
		}
	}

	// Handle unhandled message type
	if msg.Type == whatsapp.UnhandledMessageType {
		source.processUnhandledMessage(msg)
	}

	logentry := source.GetLogger()
	loglevel := logentry.Level
	logentry = logentry.WithField(LogFields.MessageId, msg.Id)
	logentry = logentry.WithField(LogFields.ChatId, msg.Chat.Id)
	logentry.Level = loglevel

	logentry.Debugf("appending message to cache, from: %s", from)
	source.appendMsgToCache(msg, from)
}

// region STATUS AND RECEIPTS

// does not cache msg, only update status and webhook dispatch
func (source *DispatchingHandler) Receipt(msg *whatsapp.WhatsappMessage) {
	if msg == nil {
		return
	}

	// Receipt payloads must remain distinguishable from regular inbound messages.
	if msg.Type == whatsapp.UnhandledMessageType {
		msg.Type = whatsapp.SystemMessageType
	}

	// triggering external publishers
	source.Trigger(msg)
}

//endregion

/*
<summary>

	Event on:
		* User Logged Out from whatsapp app
		* Maximum numbers of devices reached
		* Banned
		* Token Expired

</summary>
*/
func (source *DispatchingHandler) LoggedOut(reason string) {

	// one step at a time
	if source.server != nil {

		msg := "logged out !"
		if len(reason) > 0 {
			msg += " reason: " + reason
		}

		logger := source.GetLogger()
		logger.Warn(msg)

		// Persist a diagnostic so the frontend can explain why the session is
		// now unverified instead of only showing the generic state.
		source.server.RecordLogout(reason)

		source.publishRealtimeLifecycle("logged_out", reason, "")
	}
}

/*
<summary>

	Event on:
		* When connected to whatsapp servers and authenticated

</summary>
*/
func (source *DispatchingHandler) OnConnected() {

	// one step at a time
	if source.server != nil {

		// Reset server start timestamp on connection (uptime starts from connection moment)
		source.server.Timestamps.Start = time.Now().UTC()

		// marking unverified and wait for more analyses
		err := source.server.MarkVerified(true)
		if err != nil {
			logger := source.server.GetLogger()
			logger.Errorf("error on mark verified after connected: %s", err.Error())
		}
		err = source.server.ClearConnectionIssue("connected and cleared connection issue")
		if err != nil {
			logger := source.server.GetLogger()
			logger.Errorf("error clearing connection issue after connected: %s", err.Error())
		}
		source.publishRealtimeLifecycle("connected", "", "")
		source.publishLifecycleEvent("session.lifecycle.connected", "success", map[string]string{
			"state":    source.server.GetState().String(),
			"verified": formatLifecycleBool(source.server.Verified),
		})
	}
}

/*
<summary>

	Event on:
		* When disconnected from whatsapp servers with specific cause

</summary>
*/
func (source *DispatchingHandler) OnDisconnected(cause string, details string) {
	if source.server == nil {
		return
	}

	logger := source.GetLogger()
	logger.Infof("dispatching server disconnect event: %s - %s", cause, details)

	if err := source.server.RecordDisconnect(cause, details); err != nil {
		logger.Errorf("failed to persist disconnect diagnostic: %s", err.Error())
	}

	// Get phone number and wid from server
	phone := source.server.GetNumber()
	wid := source.server.GetWId()

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
	source.appendMsgToCache(message, "disconnected")

	source.publishRealtimeLifecycle("disconnected", cause, details)
	source.publishLifecycleEvent("session.lifecycle.disconnected", "success", map[string]string{
		"cause":    cause,
		"state":    source.server.GetState().String(),
		"verified": formatLifecycleBool(source.server.Verified),
	})
}

/*
<summary>

	Event on:
		* When server is manually stopped

</summary>
*/
func (source *DispatchingHandler) OnStopped(cause string) {
	if source.server == nil {
		return
	}

	logger := source.GetLogger()
	logger.Infof("dispatching server stop event: %s", cause)

	// Get phone number and wid from server
	phone := source.server.GetNumber()
	wid := source.server.GetWId()

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
	source.appendMsgToCache(message, "stopped")

	source.publishRealtimeLifecycle("stopped", cause, "")
	source.publishLifecycleEvent("session.lifecycle.stopped", "success", map[string]string{
		"cause":    cause,
		"state":    source.server.GetState().String(),
		"verified": formatLifecycleBool(source.server.Verified),
	})
}

/*
<summary>

	Event on:
		* When server is deleted

</summary>
*/
func (source *DispatchingHandler) OnDeleted(cause string) {
	if source.server == nil {
		return
	}

	logger := source.GetLogger()
	logger.Infof("dispatching server delete event: %s", cause)

	message := NewServerDeletedEvent(source.server, cause, nil)

	// Add to cache and send through dispatchers
	source.appendMsgToCache(message, "deleted")

	source.publishRealtimeLifecycle("deleted", cause, "")
	source.publishLifecycleEvent("session.lifecycle.deleted", "success", map[string]string{
		"cause":    cause,
		"state":    source.server.GetState().String(),
		"verified": formatLifecycleBool(source.server.Verified),
	})
}

func (source *DispatchingHandler) publishLifecycleEvent(name string, status string, attributes map[string]string) {
	if source == nil || source.server == nil {
		return
	}

	if attributes == nil {
		attributes = map[string]string{}
	}
	attributes["wid"] = source.server.GetWId()

	events.Publish(events.Event{
		Name:       name,
		Source:     "models.dispatching_handler",
		Status:     status,
		Attributes: attributes,
	})
}

func (source *DispatchingHandler) publishRealtimeLifecycle(kind string, cause string, details string) {
	if source == nil || source.server == nil {
		return
	}

	GlobalDispatchingLifecyclePublisher.PublishLifecycle(&DispatchingLifecycleEvent{
		Kind:      kind,
		Token:     source.server.Token,
		User:      source.server.GetUser(),
		Wid:       source.server.GetWId(),
		Phone:     source.server.GetNumber(),
		State:     source.server.GetState().String(),
		Verified:  source.server.Verified,
		Cause:     cause,
		Details:   details,
		Timestamp: time.Now().UTC(),
	})
}

func formatLifecycleBool(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

//#endregion
//region MESSAGE CONTROL REGION HANDLE A LOCK

// caches and triggers async hooks
func (source *DispatchingHandler) appendMsgToCache(msg *whatsapp.WhatsappMessage, from string) {

	// saving on local normalized cache, do not affect remote msgs
	valid := source.QpWhatsappMessages.Append(msg, from)

	// cache changed, continue to external dispatchers
	if valid {

		// should cleanup old messages ?
		length := ENV.CacheLength()
		source.QpWhatsappMessages.CleanUp(length)

		source.Trigger(msg)
	}
}

func (source *DispatchingHandler) GetById(id string) (*whatsapp.WhatsappMessage, error) {
	return source.QpWhatsappMessages.GetById(id)
}

// endregion
// region EVENT HANDLER TO INTERNAL USE, GENERALLY TO WEBHOOK

// sends the message throw external publishers
func (source *DispatchingHandler) Trigger(payload *whatsapp.WhatsappMessage) {
	if source == nil {
		return
	}

	handlerCallbacks := make([]dispatchservice.HandlerSubscriber, 0, len(source.aeh))
	for _, handler := range source.aeh {
		handlerCallbacks = append(handlerCallbacks, handler)
	}

	request := &dispatchservice.HandlerFlowRequest{
		Payload: payload,
		Validate: func(message *whatsapp.WhatsappMessage) string {
			return IsValidForDispatch(message)
		},
		OnInvalid: func(reason string, message *whatsapp.WhatsappMessage) {
			logentry := source.GetLogger()
			logentry.Debug(reason)

			jsonPayload := library.ToJson(message)
			logentry.Logger.Debugf("unhandled payload: %s", jsonPayload)
		},
		MarkEventTimestamp: func() {
			if source.server == nil {
				return
			}
			currentTime := time.Now().UTC()
			source.server.Timestamps.Event = &currentTime
		},
		MarkMessageTimestamp: func() {
			if source.server == nil {
				return
			}
			currentTime := time.Now().UTC()
			source.server.Timestamps.Message = &currentTime
		},
		SetMessageWid: func(message *whatsapp.WhatsappMessage) {
			if source.server == nil || message == nil {
				return
			}
			message.Wid = source.GetWId()
		},
		PublishRealtime: func(message *whatsapp.WhatsappMessage) {
			if source.server == nil {
				return
			}
			enriched := CloneAndEnrichMessageForServer(source.server, message)
			dispatchservice.PublishRealtimeMessage(&dispatchservice.RealtimeServerMessage{
				Token:   source.server.Token,
				User:    source.server.GetUser(),
				WID:     source.server.GetWId(),
				State:   source.server.GetState().String(),
				Message: enriched,
			})
		},
		HandlerCallbacks: handlerCallbacks,
	}

	dispatchservice.GetInstance().DispatchHandlerFlow(request)
}

// Register an event handler that triggers on a new message received on cache
func (handler *DispatchingHandler) Register(evt QpDispatchingHandlerInterface) {
	handler.syncRegister.Lock() // await for avoid simultaneous calls

	if !handler.IsRegistered(evt) {
		handler.aeh = append(handler.aeh, evt)
	}

	handler.syncRegister.Unlock()
}

// Removes an specific event handler
func (handler *DispatchingHandler) UnRegister(evt QpDispatchingHandlerInterface) {
	handler.syncRegister.Lock() // await for avoid simultaneous calls

	newHandlers := []QpDispatchingHandlerInterface{}
	for _, v := range handler.aeh {
		if v != evt {
			newHandlers = append(newHandlers, v)
		}
	}

	// updating
	handler.aeh = newHandlers

	handler.syncRegister.Unlock()
}

// Removes an specific event handler
func (handler *DispatchingHandler) Clear() {
	handler.syncRegister.Lock() // await for avoid simultaneous calls

	// updating
	handler.aeh = nil

	handler.syncRegister.Unlock()
}

// Indicates that has any event handler registered
func (handler *DispatchingHandler) IsAttached() bool {
	return len(handler.aeh) > 0
}

// Indicates that if an specific handler is registered
func (handler *DispatchingHandler) IsRegistered(evt interface{}) bool {
	for _, v := range handler.aeh {
		if v == evt {
			return true
		}
	}

	return false
}

// HasDispatchingSubscriber reports whether the default outbound dispatching
// subscriber is already attached to this server handler.
func (handler *DispatchingHandler) HasDispatchingSubscriber() bool {
	for _, subscriber := range handler.aeh {
		if _, ok := subscriber.(dispatchingSubscriber); ok {
			return true
		}
	}

	return false
}

//endregion

// processUnhandledMessage handles debugging for unhandled message types
// This method can be easily removed when debugging is no longer needed
func (source *DispatchingHandler) processUnhandledMessage(msg *whatsapp.WhatsappMessage) {
	// Generate a unique UUID to prevent duplicate message IDs
	uniqueID := uuid.New().String()
	msg.Id = msg.Id + "-unhandled-" + uniqueID

	if len(msg.Text) == 0 && msg.Content != nil {
		// Get the type information using reflection
		contentType := reflect.TypeOf(msg.Content)
		var typeInfo string

		if contentType != nil {
			// Get full type name including package
			typeInfo = contentType.String()

			// If it's a pointer, get the element type
			if contentType.Kind() == reflect.Ptr {
				if contentType.Elem() != nil {
					typeInfo = fmt.Sprintf("*%s", contentType.Elem().String())
				}
			}
		} else {
			typeInfo = "<nil>"
		}

		// Include type information and content in the text
		contentJson := library.ToJson(msg.Content)
		msg.Text = fmt.Sprintf("[Type: %s] %s", typeInfo, contentJson)
	}
}

func (source *DispatchingHandler) IsInterfaceNil() bool {
	return nil == source
}

func NewServerDeletedEvent(server *QpWhatsappServer, cause string, previousState *whatsapp.WhatsappConnectionState) *whatsapp.WhatsappMessage {
	if server == nil {
		return nil
	}

	phone := server.GetNumber()
	wid := server.GetWId()
	currentState := server.GetState()
	now := time.Now().UTC()

	eventData := map[string]interface{}{
		"event":     "deleted",
		"cause":     cause,
		"wid":       wid,
		"phone":     phone,
		"state":     currentState.String(),
		"timestamp": now.Format(time.RFC3339),
	}

	if previousState != nil {
		eventData["previous_state"] = previousState.String()
	}

	description := fmt.Sprintf("WhatsApp server was deleted: %s", cause)

	return &whatsapp.WhatsappMessage{
		Id:        uuid.New().String(),
		Timestamp: now,
		Type:      whatsapp.SystemMessageType,
		FromMe:    false,
		Chat:      whatsapp.WASYSTEMCHAT,
		Text:      description,
		Info:      eventData,
	}
}
