package models

import (
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
	library "github.com/nocodeleaks/quepasa/library"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// Serviço que controla os servidores / bots individuais do whatsapp
type DispatchingHandler struct {
	QpWhatsappMessages
	library.LogStruct // logging

	server *QpWhatsappServer

	messageDispatcher  *MessageDispatcher
	lifecyclePublisher DispatchingLifecyclePublisher
}

// Server returns the underlying QpWhatsappServer instance.
func (handler *DispatchingHandler) Server() *QpWhatsappServer {
	return handler.server
}

// AppendMsgToCache adds a message to the cache and triggers async handlers.
func (handler *DispatchingHandler) AppendMsgToCache(msg *whatsapp.WhatsappMessage, from string) {
	handler.appendMsgToCache(msg, from)
}

// GetMessageDispatcher returns the message dispatcher for this handler (lazy init).
func (handler *DispatchingHandler) GetMessageDispatcher() *MessageDispatcher {
	if handler.messageDispatcher == nil {
		handler.messageDispatcher = NewMessageDispatcher(handler)
	}
	return handler.messageDispatcher
}

// LoggedOut handles the logged out lifecycle event (implements IWhatsappHandlers).
func (source *DispatchingHandler) LoggedOut(reason string) {
	NewLifecycleHandler(source).LoggedOut(reason)
}

// OnConnected handles the connected lifecycle event (implements IWhatsappHandlers).
func (source *DispatchingHandler) OnConnected() {
	NewLifecycleHandler(source).OnConnected()
}

// OnDisconnected handles the disconnected lifecycle event (implements IWhatsappHandlers).
func (source *DispatchingHandler) OnDisconnected(cause string, details string) {
	NewLifecycleHandler(source).OnDisconnected(cause, details)
}

// OnStopped handles the manually stopped lifecycle event.
func (source *DispatchingHandler) OnStopped(cause string) {
	NewLifecycleHandler(source).OnStopped(cause)
}

// OnDeleted handles the deleted lifecycle event.
func (source *DispatchingHandler) OnDeleted(cause string) {
	NewLifecycleHandler(source).OnDeleted(cause)
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

func PublishDispatchingLifecycle(event *DispatchingLifecycleEvent) {
	transportServicesMu.RLock()
	publisher := GlobalDispatchingLifecyclePublisher
	transportServicesMu.RUnlock()
	publisher.PublishLifecycle(event)
}

// LifecyclePublisher returns the per-handler lifecycle publisher when present,
// falling back to the globally wired publisher for backward compatibility.
func (handler *DispatchingHandler) LifecyclePublisher() DispatchingLifecyclePublisher {
	if handler != nil && handler.lifecyclePublisher != nil {
		return handler.lifecyclePublisher
	}

	return DefaultDispatchingLifecyclePublisher()
}

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

func (source *DispatchingHandler) HandleDirect() bool {
	global := whatsapp.Options

	var local whatsapp.WhatsappBoolean
	if source.server != nil {
		local = source.server.Direct
	}
	return global.HandleDirect(local)
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

	// should skip direct (individual) messages ?
	if !source.HandleDirect() && msg.FromDirect() {
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
	source.GetMessageDispatcher().Trigger(payload)
}

// Register an event handler that triggers on a new message received on cache
func (handler *DispatchingHandler) Register(evt QpDispatchingHandlerInterface) {
	handler.GetMessageDispatcher().Register(evt)
}

// Removes an specific event handler
func (handler *DispatchingHandler) UnRegister(evt QpDispatchingHandlerInterface) {
	handler.GetMessageDispatcher().UnRegister(evt)
}

// Removes an specific event handler
func (handler *DispatchingHandler) Clear() {
	handler.GetMessageDispatcher().Clear()
}

// Indicates that has any event handler registered
func (handler *DispatchingHandler) IsAttached() bool {
	return handler.GetMessageDispatcher().IsAttached()
}

// Indicates that if an specific handler is registered
func (handler *DispatchingHandler) IsRegistered(evt interface{}) bool {
	return handler.GetMessageDispatcher().IsRegistered(evt)
}

// HasDispatchingSubscriber reports whether the default outbound dispatching
// subscriber is already attached to this server handler.
func (handler *DispatchingHandler) HasDispatchingSubscriber() bool {
	return handler.GetMessageDispatcher().HasDispatchingSubscriber()
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
