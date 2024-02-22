package whatsmeow

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	library "github.com/nocodeleaks/quepasa/library"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
	whatsmeow "go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/binary"
	types "go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

type WhatsmeowHandlers struct {
	Client                   *whatsmeow.Client
	WAHandlers               whatsapp.IWhatsappHandlers
	HistorySyncDays          uint
	eventHandlerID           uint32
	unregisterRequestedToken bool
	log                      *log.Entry

	ReadReceipt bool // Should follow Read Receipts to webhook
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

	case *events.Receipt:
		if handler.ReadReceipt {
			go handler.Receipt(*v)
		}
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
		handler.HandleLoggedOut(*v)
		return

	case *events.HistorySync:
		if handler.HistorySyncDays > 0 {
			go handler.HistorySync(*v)
		}
		return

	case
		*events.AppState,
		*events.AppStateSyncComplete,
		*events.CallTerminate,
		*events.Contact,
		*events.DeleteChat,
		*events.DeleteForMe,
		*events.MarkChatAsRead,
		*events.Mute,
		*events.OfflineSyncCompleted,
		*events.OfflineSyncPreview,
		*events.PairSuccess,
		*events.Pin,
		*events.PushName,
		*events.PushNameSetting,
		*events.GroupInfo,
		*events.QR:
		handler.log.Tracef("event ignored: %v", reflect.TypeOf(v))
		return // ignoring not implemented yet

	default:
		handler.log.Debugf("event not handled: %v", reflect.TypeOf(v))
		return
	}
}

func HistorySyncSaveJSON(evt events.HistorySync) {
	fileName := "history-sync.json"
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Errorf("Failed to open file to write history sync: %v", err)
		return
	}
	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	err = enc.Encode(evt.Data)
	if err != nil {
		log.Errorf("Failed to write history sync: %v", err)
		return
	}
	log.Infof("Wrote history sync to %s", fileName)
	_ = file.Close()
}

func (handler *WhatsmeowHandlers) HistorySync(evt events.HistorySync) {
	handler.log.Infof("history sync: %s", evt.Data.SyncType)
	//HistorySyncSaveJSON(evt)

	conversations := evt.Data.Conversations
	for _, conversation := range conversations {
		for _, historyMsg := range conversation.GetMessages() {
			wid, err := types.ParseJID(conversation.GetId())
			if err != nil {
				log.Errorf("failed to parse jid at history sync: %v", err)
				return
			}

			evt, err := handler.Client.ParseWebMessage(wid, historyMsg.GetMessage())
			if err != nil {
				log.Errorf("failed to parse web message at history sync: %v", err)
				return
			}

			handler.Message(*evt)
		}
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

	message := &whatsapp.WhatsappMessage{
		Content: evt.Message,
		Info:    evt.Info,
	}

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

// Append to cache handlers if exists, and then webhook
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

	_ = handler.RejectCall(evt)
}

func (handler *WhatsmeowHandlers) RejectCall(v types.BasicCallMeta) (err error) {
	// Verificar se a variável de ambiente REJECTCALL é verdadeira
	rejectCallEnv := os.Getenv("REJECTCALL")
	rejectCall, err := strconv.ParseBool(rejectCallEnv)
	if err != nil {
		// Se houver um erro ao converter a variável de ambiente, trate-o
		return err
	}

	// Se REJECTCALL for verdadeiro, rejeite a chamada
	if rejectCall {
		var node = binary.Node{
			Tag: "call",
			Attrs: binary.Attrs{
				"to": v.From,
				"id": handler.Client.GenerateMessageID(),
			},
			Content: []binary.Node{
				{
					Tag: "reject",
					Attrs: binary.Attrs{
						"call-id":      v.CallID,
						"call-creator": v.CallCreator,
						"count":        0,
					},
					Content: nil,
				},
			},
		}

		handler.log.Infof("rejecting incoming call from: %s", v.From)
		return handler.Client.DangerousInternals().SendNode(node)
	}

	// Se REJECTCALL for falso, não faça nada
	handler.log.Infof("REJECTCALL is false, not rejecting incoming call from: %s", v.From)
	return nil
}

//#endregion

//#region EVENT READ RECEIPT

func (handler *WhatsmeowHandlers) Receipt(evt events.Receipt) {
	handler.log.Trace("event Receipt !")

	message := &whatsapp.WhatsappMessage{Content: evt}
	message.Id = "readreceipt"

	// basic information
	message.Timestamp = evt.Timestamp
	message.FromMe = false

	message.Chat = whatsapp.WhatsappChat{}
	chatID := fmt.Sprint(evt.Chat.User, "@", evt.Chat.Server)
	message.Chat.Id = chatID

	message.Type = whatsapp.SystemMessageType

	// message ids comma separated
	message.Text = strings.Join(evt.MessageIDs, ",")

	if handler.WAHandlers != nil {

		// following to internal handlers
		go handler.WAHandlers.Receipt(message)
	}
}

//#endregion

//#region HANDLE LOGGED OUT EVENT

func (handler *WhatsmeowHandlers) HandleLoggedOut(evt events.LoggedOut) {
	reason := evt.Reason.String()
	handler.log.Tracef("logged out %s", reason)

	if handler.WAHandlers != nil {
		handler.WAHandlers.LoggedOut(reason)
	}

	message := &whatsapp.WhatsappMessage{
		Timestamp: time.Now().Truncate(time.Second),
		Type:      whatsapp.SystemMessageType,
		Id:        handler.Client.GenerateMessageID(),
		Chat:      whatsapp.WASYSTEMCHAT,
		Text:      reason,
	}

	handler.Follow(message)
	handler.UnRegister()
}

//#endregion
