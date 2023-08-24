package whatsmeow

import (
	"fmt"
	"reflect"

	library "github.com/nocodeleaks/quepasa/library"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
	whatsmeow "go.mau.fi/whatsmeow"
	types "go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

type WhatsmeowHandlers struct {
	Client                   *whatsmeow.Client
	WAHandlers               whatsapp.IWhatsappHandlers
	eventHandlerID           uint32
	unregisterRequestedToken bool
	log                      *log.Entry
}

// only affects whatsmeow
func (handler *WhatsmeowHandlers) UnRegister() {
	handler.unregisterRequestedToken = true

	// if is this session
	found := handler.Client.RemoveEventHandler(handler.eventHandlerID)
	if found {
		handler.log.Infof("handler unregistered, id: %v", handler.eventHandlerID)
	}
}

func (handler *WhatsmeowHandlers) Register() (err error) {
	if handler.Client.Store == nil {
		err = fmt.Errorf("this client lost the store, probably a logout from whatsapp phone")
		return
	}

	handler.unregisterRequestedToken = false
	handler.eventHandlerID = handler.Client.AddEventHandler(handler.EventsHandler)
	handler.log.Infof("handler registered, id: %v", handler.eventHandlerID)

	return
}

// Define os diferentes tipos de eventos a serem reconhecidos
// Aqui se define se vamos processar mensagens | confirmações de leitura | etc
func (handler *WhatsmeowHandlers) EventsHandler(evt interface{}) {
	if handler.unregisterRequestedToken {
		handler.log.Info("unregister event handler requested")
		handler.Client.RemoveEventHandler(handler.eventHandlerID)
		return
	}

	switch v := evt.(type) {

	case *events.Message:
		go handler.Message(*v)
		return

	case *events.CallOffer:
		go handler.CallMessage(*&v.BasicCallMeta)
		return

	case *events.CallOfferNotice:
		go handler.CallMessage(*&v.BasicCallMeta)
		return

	case *events.Connected:
		// zerando contador de tentativas de reconexão
		// importante para zerar o tempo entre tentativas em caso de erro
		handler.Client.AutoReconnectErrors = 0
		return

	case *events.Disconnected:
		msgDisconnected := "disconnected from server"
		if handler.Client.EnableAutoReconnect {
			handler.log.Info(msgDisconnected + ", dont worry, reconnecting")
		} else {
			handler.log.Warn(msgDisconnected)
		}
		return

	case *events.LoggedOut:
		reason := v.Reason.String()
		handler.log.Trace("logged out " + reason)

		if handler.WAHandlers != nil {
			handler.WAHandlers.LoggedOut(reason)
		}

		handler.UnRegister()
		return

	case
		*events.AppState,
		*events.AppStateSyncComplete,
		*events.Contact,
		*events.DeleteChat,
		*events.DeleteForMe,
		*events.HistorySync,
		*events.MarkChatAsRead,
		*events.Mute,
		*events.OfflineSyncCompleted,
		*events.OfflineSyncPreview,
		*events.PairSuccess,
		*events.Pin,
		*events.PushName,
		*events.PushNameSetting,
		*events.QR,
		*events.Receipt:
		return // ignoring not implemented yet

	default:
		handler.log.Debugf("event not handled: %v", reflect.TypeOf(v))
		return
	}
}

//#region EVENT MESSAGE

// Aqui se processar um evento de recebimento de uma mensagem genérica
func (handler *WhatsmeowHandlers) Message(evt events.Message) {
	handler.log.Trace("event message received")
	if evt.Message == nil {
		handler.log.Error("nil message on receiving whatsmeow events | try use rawMessage !")
		return
	}

	message := &whatsapp.WhatsappMessage{Content: evt.Message}

	// basic information
	message.Id = evt.Info.ID
	message.Timestamp = evt.Info.Timestamp
	message.FromMe = evt.Info.IsFromMe

	message.Chat = whatsapp.WhatsappChat{}
	chatID := fmt.Sprint(evt.Info.Chat.User, "@", evt.Info.Chat.Server)
	message.Chat.Id = chatID
	message.Chat.Title = GetChatTitle(handler.Client, evt.Info.Chat)

	if evt.Info.IsGroup {
		message.Participant = &whatsapp.WhatsappChat{}

		participantID := fmt.Sprint(evt.Info.Sender.User, "@", evt.Info.Sender.Server)
		message.Participant.Id = participantID
		message.Participant.Title = GetChatTitle(handler.Client, evt.Info.Sender)

		// sugested by hugo sampaio, removing message.FromMe
		if len(message.Participant.Title) == 0 {
			message.Participant.Title = evt.Info.PushName
		}
	} else {
		if len(message.Chat.Title) == 0 && message.FromMe {

			// if you doesnt have the contact in your list, it will bring your name instead
			// message.Chat.Title = evt.Info.PushName

			message.Chat.Title = library.GetPhoneByWId(message.Chat.Id)
		}
	}

	// Process diferent message types
	HandleKnowingMessages(handler, message, evt.Message)
	if message.Type == whatsapp.UnknownMessageType {
		HandleUnknownMessage(handler.log, evt)
	}

	handler.Follow(message)
}

//#endregion

/*
<summary>

	Follow throw internal handlers

</summary>
*/
func (handler *WhatsmeowHandlers) Follow(message *whatsapp.WhatsappMessage) {
	if handler.WAHandlers != nil {

		// following to internal handlers
		go handler.WAHandlers.Message(message)
	} else {
		handler.log.Warn("no internal handler registered")
	}
}

//#region EVENT CALL

func (handler *WhatsmeowHandlers) CallMessage(evt types.BasicCallMeta) {
	handler.log.Trace("event CallMessage !")

	message := &whatsapp.WhatsappMessage{Content: evt}

	// basic information
	message.Id = evt.CallID
	message.Timestamp = evt.Timestamp
	message.FromMe = false

	message.Chat = whatsapp.WhatsappChat{}
	chatID := fmt.Sprint(evt.From.User, "@", evt.From.Server)
	message.Chat.Id = chatID

	message.Type = whatsapp.CallMessageType

	if handler.WAHandlers != nil {

		// following to internal handlers
		go handler.WAHandlers.Message(message)
	}
}

//#endregion
