package models

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/google/uuid"
	library "github.com/nocodeleaks/quepasa/library"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// Serviço que controla os servidores / bots individuais do whatsapp
type QPWhatsappHandlers struct {
	QpWhatsappMessages
	library.LogStruct // logging

	server *QpWhatsappServer

	syncRegister *sync.Mutex

	// Appended events handler
	aeh []QpDispatchingHandlerInterface
}

// Returns whatsapp controller id on E164
// Ex: 5521967609095
func (source *QPWhatsappHandlers) GetWId() string {
	if source == nil || source.server == nil {
		return ""
	}

	return source.server.GetWId()
}

func (source *QPWhatsappHandlers) HandleGroups() bool {
	global := whatsapp.Options

	var local whatsapp.WhatsappBoolean
	if source.server != nil {
		local = source.server.Groups
	}
	return global.HandleGroups(local)
}

func (source *QPWhatsappHandlers) HandleBroadcasts() bool {
	global := whatsapp.Options

	var local whatsapp.WhatsappBoolean
	if source.server != nil {
		local = source.server.Broadcasts
	}
	return global.HandleBroadcasts(local)
}

//#region EVENTS FROM WHATSAPP SERVICE

// Process messages received from whatsapp service
func (source *QPWhatsappHandlers) Message(msg *whatsapp.WhatsappMessage, from string) {

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
func (source *QPWhatsappHandlers) Receipt(msg *whatsapp.WhatsappMessage) {
	// should implement a better method for that !!!!
	// should implement a better method for that !!!!
	// should implement a better method for that !!!!
	// should implement a better method for that !!!!
	// should implement a better method for that !!!!

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
func (source *QPWhatsappHandlers) LoggedOut(reason string) {

	// one step at a time
	if source.server != nil {

		msg := "logged out !"
		if len(reason) > 0 {
			msg += " reason: " + reason
		}

		logger := source.GetLogger()
		logger.Warn(msg)

		// marking unverified and wait for more analyses
		source.server.MarkVerified(false)
	}
}

/*
<summary>

	Event on:
		* When connected to whatsapp servers and authenticated

</summary>
*/
func (source *QPWhatsappHandlers) OnConnected() {

	// one step at a time
	if source.server != nil {

		// marking unverified and wait for more analyses
		err := source.server.MarkVerified(true)
		if err != nil {
			logger := source.server.GetLogger()
			logger.Errorf("error on mark verified after connected: %s", err.Error())
		}
	}
}

/*
<summary>

	Event on:
		* When connected to whatsapp servers and authenticated

</summary>
*/
func (source *QPWhatsappHandlers) OnDisconnected() {

}

//#endregion
//region MESSAGE CONTROL REGION HANDLE A LOCK

// caches and triggers async hooks
func (source *QPWhatsappHandlers) appendMsgToCache(msg *whatsapp.WhatsappMessage, from string) {

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

func (source *QPWhatsappHandlers) GetById(id string) (*whatsapp.WhatsappMessage, error) {
	return source.QpWhatsappMessages.GetById(id)
}

// endregion
// region EVENT HANDLER TO INTERNAL USE, GENERALLY TO WEBHOOK

// sends the message throw external publishers
func (source *QPWhatsappHandlers) Trigger(payload *whatsapp.WhatsappMessage) {
	// If the source is nil, we cannot proceed with dispatching the message.
	// This is a safeguard to prevent nil pointer dereference errors.
	if source == nil {
		return
	}

	// Validate the message payload. If it's not valid for dispatch,
	// IsValidForDispatch will return a reason string.
	reason := IsValidForDispatch(payload)
	if len(reason) > 0 {
		logentry := source.GetLogger()
		logentry.Debug(reason)

		jsonPayload := library.ToJson(payload)
		logentry.Logger.Debugf("unhandled payload: %s", jsonPayload)

		// If a reason is returned, it means the message should be ignored.
		// No further action is needed, so we simply return.
		return
	}

	// Update last message/event timestamps
	if source.server != nil {
		currentTime := time.Now().UTC()

		// Check if this is an event (system messages, unhandled messages, or read receipts)
		isEvent := payload.Type == whatsapp.UnhandledMessageType ||
			payload.Type == whatsapp.SystemMessageType ||
			payload.Id == "readreceipt"

		if isEvent {
			source.server.Timestamps.Event = &currentTime
		} else {
			// Regular message content (text, image, audio, video, etc.) - received messages only
			source.server.Timestamps.Message = &currentTime
		}
	}

	if source.server != nil {
		payload.Wid = source.GetWId()
		go SignalRHub.Dispatch(source.server.Token, payload)
	}

	for _, handler := range source.aeh {
		go handler.HandleDispatching(payload)
	}
}

// Register an event handler that triggers on a new message received on cache
func (handler *QPWhatsappHandlers) Register(evt QpDispatchingHandlerInterface) {
	handler.syncRegister.Lock() // await for avoid simultaneous calls

	if !handler.IsRegistered(evt) {
		handler.aeh = append(handler.aeh, evt)
	}

	handler.syncRegister.Unlock()
}

// Removes an specific event handler
func (handler *QPWhatsappHandlers) UnRegister(evt QpDispatchingHandlerInterface) {
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
func (handler *QPWhatsappHandlers) Clear() {
	handler.syncRegister.Lock() // await for avoid simultaneous calls

	// updating
	handler.aeh = nil

	handler.syncRegister.Unlock()
}

// Indicates that has any event handler registered
func (handler *QPWhatsappHandlers) IsAttached() bool {
	return len(handler.aeh) > 0
}

// Indicates that if an specific handler is registered
func (handler *QPWhatsappHandlers) IsRegistered(evt interface{}) bool {
	for _, v := range handler.aeh {
		if v == evt {
			return true
		}
	}

	return false
}

//endregion

// processUnhandledMessage handles debugging for unhandled message types
// This method can be easily removed when debugging is no longer needed
func (source *QPWhatsappHandlers) processUnhandledMessage(msg *whatsapp.WhatsappMessage) {
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

func (source *QPWhatsappHandlers) IsInterfaceNil() bool {
	return nil == source
}
