package whatsmeow

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync/atomic"
	"time"

	library "github.com/nocodeleaks/quepasa/library"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
	whatsmeow "go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/appstate"
	types "go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

type WhatsmeowHandlers struct {
	library.LogStruct // logging

	// particular whatsapp options for this handler
	*whatsapp.WhatsappOptions

	// default whatsmeow service global options
	WhatsmeowOptions

	Client     *whatsmeow.Client
	WAHandlers whatsapp.IWhatsappHandlers

	eventHandlerID           uint32
	unregisterRequestedToken bool
	service                  *WhatsmeowServiceModel

	// events counter
	Counter uint64
}

func (source *WhatsmeowHandlers) GetServiceOptions() (options whatsapp.WhatsappOptionsExtended) {
	if source != nil {
		return source.WhatsappOptionsExtended
	}

	return
}

//#region WHATSAPP OPTIONS

func (source WhatsmeowOptions) GetPresence() types.Presence {
	switch source.Presence {
	case string(types.PresenceAvailable):
		return types.PresenceAvailable
	case string(types.PresenceUnavailable):
		return types.PresenceUnavailable
	default:
		return WhatsmeowPresenceDefault
	}
}

func (source *WhatsmeowHandlers) HandleBroadcasts() bool {
	if source == nil {
		return false
	}

	var defaultValue whatsapp.WhatsappBoolean
	if source.WhatsappOptions != nil {
		defaultValue = source.WhatsappOptions.Broadcasts
	}

	serviceOptions := source.GetServiceOptions()
	return serviceOptions.HandleBroadcasts(defaultValue)
}

func (source *WhatsmeowHandlers) HandleGroups() bool {
	if source == nil {
		return false
	}

	var defaultValue whatsapp.WhatsappBoolean
	if source.WhatsappOptions != nil {
		defaultValue = source.WhatsappOptions.Groups
	}

	serviceOptions := source.GetServiceOptions()
	return serviceOptions.HandleGroups(defaultValue)
}

func (source *WhatsmeowHandlers) HandleReadReceipts() bool {
	if source == nil {
		return false
	}

	var defaultValue whatsapp.WhatsappBoolean
	if source.WhatsappOptions != nil {
		defaultValue = source.WhatsappOptions.ReadReceipts
	}

	serviceOptions := source.GetServiceOptions()
	return serviceOptions.HandleReadReceipts(defaultValue)
}

func (source *WhatsmeowHandlers) HandleCalls() bool {
	if source == nil {
		return false
	}

	var defaultValue whatsapp.WhatsappBoolean
	if source.WhatsappOptions != nil {
		defaultValue = source.WhatsappOptions.Calls
	}

	serviceOptions := source.GetServiceOptions()
	return serviceOptions.HandleCalls(defaultValue)
}

//#endregion

func (source WhatsmeowHandlers) HandleHistorySync() bool {
	options := source.GetServiceOptions()
	if options.HistorySync != nil {
		return true
	}

	return whatsapp.WhatsappHistorySync
}

// only affects whatsmeow
func (handler *WhatsmeowHandlers) UnRegister() {
	handler.unregisterRequestedToken = true

	// if is this session
	found := handler.Client.RemoveEventHandler(handler.eventHandlerID)
	if found {
		handler.GetLogger().Infof("handler unregistered, id: %v", handler.eventHandlerID)
	}
}

func (source *WhatsmeowHandlers) Register() (err error) {
	if source.Client.Store == nil {
		err = fmt.Errorf("this client lost the store, probably a logout from whatsapp phone")
		return
	}

	source.unregisterRequestedToken = false
	source.eventHandlerID = source.Client.AddEventHandler(source.EventsHandler)

	logentry := source.GetLogger()
	logentry.Infof("handler registered, id: %v, loglevel: %s", source.eventHandlerID, logentry.Level)

	return
}

func (source *WhatsmeowHandlers) SendPresence(presence types.Presence, from string) {
	logentry := source.GetLogger()
	client := source.Client
	SendPresence(client, presence, from, logentry)
}

var historySyncID int32
var startupTime = time.Now().Unix()

// Define os diferentes tipos de eventos a serem reconhecidos
// Aqui se define se vamos processar mensagens | confirmações de leitura | etc
func (source *WhatsmeowHandlers) EventsHandler(rawEvt interface{}) {
	if source == nil {
		return
	}

	logentry := source.GetLogger()

	if source.unregisterRequestedToken {
		logentry.Info("unregister event handler requested")
		source.Client.RemoveEventHandler(source.eventHandlerID)
		return
	}

	switch evt := rawEvt.(type) {

	case *events.Message:
		go source.Message(*evt, "live")
		return

		//# region CALLS
	case *events.CallOffer:
		logentry.Infof("CallOffer: %v", evt)
		go source.CallMessage(evt.BasicCallMeta)
		return

	case *events.CallOfferNotice:
		logentry.Infof("CallOfferNotice: %v", evt)
		go source.CallMessage(evt.BasicCallMeta)
		return

	/*
		case *events.CallRelayLatency:
			logentry.Infof("CallRelayLatency: %v", evt)
			return
	*/
	//#endregion

	case *events.Receipt:
		go source.Receipt(*evt)
		return

	case *events.Connected:
		if source.Client != nil {
			// zerando contador de tentativas de reconexão
			// importante para zerar o tempo entre tentativas em caso de erro
			source.Client.AutoReconnectErrors = 0

			presence := source.GetPresence()
			source.SendPresence(presence, "'connected' event")
		}

		if source.WAHandlers != nil && !source.WAHandlers.IsInterfaceNil() {
			go source.WAHandlers.OnConnected()
		}
		return

	case *events.PushNameSetting:

		presence := source.GetPresence()
		source.SendPresence(presence, "'push name setting' event")
		return

	case *events.Disconnected:
		msgDisconnected := "disconnected from server"
		if source.Client.EnableAutoReconnect {
			logentry.Infof("%s, dont worry, reconnecting", msgDisconnected)
		} else {
			logentry.Warn(msgDisconnected)
		}

		if source.WAHandlers != nil && !source.WAHandlers.IsInterfaceNil() {
			go source.WAHandlers.OnDisconnected()
		}
		return

	case *events.LoggedOut:
		source.OnLoggedOutEvent(*evt)
		return

	case *events.HistorySync:
		if source.HandleHistorySync() {
			go source.OnHistorySyncEvent(*evt)
		}
		return

	case *events.AppStateSyncComplete:
		if evt.Name == appstate.WAPatchCriticalBlock {
			presence := source.GetPresence()
			source.SendPresence(presence, "'app state sync complete' event")
		}
		return

	case *events.JoinedGroup:
		source.JoinedGroup(*evt)
		return

	case *events.Contact:
		go OnEventContact(source, *evt)
		return

	case
		*events.AppState,
		*events.CallTerminate,
		*events.DeleteChat,
		*events.DeleteForMe,
		*events.MarkChatAsRead,
		*events.Mute,
		*events.OfflineSyncCompleted,
		*events.OfflineSyncPreview,
		*events.PairSuccess,
		*events.Pin,
		*events.PushName,
		*events.GroupInfo,
		*events.QR:
		logentry.Tracef("event ignored: %v", reflect.TypeOf(evt))
		return // ignoring not implemented yet

	default:
		logentry.Debugf("event not handled: %v", reflect.TypeOf(evt))
		return
	}
}

func HistorySyncSaveJSON(evt events.HistorySync) {
	id := atomic.AddInt32(&historySyncID, 1)
	fileName := fmt.Sprintf("history-%d-%d.json", startupTime, id)
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

func (source *WhatsmeowHandlers) OnHistorySyncEvent(evt events.HistorySync) {
	logentry := source.GetLogger()
	logentry.Infof("history sync: %s", evt.Data.SyncType)
	// HistorySyncSaveJSON(evt)

	// whatsmeow service options
	options := source.GetServiceOptions()
	if options.HistorySync == nil {
		return
	}

	conversations := evt.Data.GetConversations()
	for _, conversation := range conversations {
		for _, historyMsg := range conversation.GetMessages() {
			wid, err := types.ParseJID(conversation.GetID())
			if err != nil {
				logentry.Errorf("failed to parse jid at history sync: %v", err)
				return
			}

			// converting to event
			msgInfo := historyMsg.GetMessage()
			msgTime := msgInfo.GetMessageTimestamp()
			if !options.HandleHistory(msgTime) {
				continue
			}

			msgevt, err := source.Client.ParseWebMessage(wid, msgInfo)
			if err != nil {
				logentry.Errorf("failed to parse web message at history sync: %v", err)
				return
			}

			// put here a logic for history sync days filter
			source.Message(*msgevt, "history")
		}
	}
}

//#region EVENT MESSAGE

// Aqui se processar um evento de recebimento de uma mensagem genérica
func (handler *WhatsmeowHandlers) Message(evt events.Message, from string) {
	logentry := handler.GetLogger()
	logentry.Trace("event message received")

	if evt.Message == nil {
		if evt.SourceWebMsg != nil {
			// probably from recover history sync
			logentry.Info("web message cant be full decrypted, ignoring")
			return
		}

		jsonstring, _ := json.Marshal(evt)
		logentry.Errorf("nil message on receiving whatsmeow events | try use rawMessage ! json: %s", string(jsonstring))
		return
	}

	message := &whatsapp.WhatsappMessage{
		Content:        evt.Message,
		InfoForHistory: evt.Info,
		FromHistory:    from == "history",
	}

	// basic information
	message.Id = evt.Info.ID
	message.Timestamp = ImproveTimestamp(evt.Info.Timestamp)
	// fmt.Printf("event timestamp: %v, new timestamp: %v\n", evt.Info.Timestamp, message.Timestamp)

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
			message.Chat.Title = library.GetPhoneByWId(message.Chat.Id)
		}
	}

	// Process diferent message types
	HandleKnowingMessages(handler, message, evt.Message)

	// discard and return
	if message.Type == whatsapp.DiscardMessageType {
		JsonMsg := ToJson(evt)
		logentry.Debugf("debugging and ignoring an discard message: %s", JsonMsg)
		return
	}

	// unknown and continue
	if message.Type == whatsapp.UnknownMessageType {
		message.Text += " :: " + ToJson(evt)
	}

	handler.Follow(message, from)
}

func ToJson(in interface{}) string {
	bytes, err := json.Marshal(in)
	if err == nil {
		return string(bytes)
	}
	return ""
}

//#endregion

/*
<summary>

	Follow throw internal handlers

</summary>
*/

// Append to cache handlers if exists, and then webhook
func (handler *WhatsmeowHandlers) Follow(message *whatsapp.WhatsappMessage, from string) {
	if handler.WAHandlers != nil {

		// following to internal handlers
		go handler.WAHandlers.Message(message, from)

	} else {
		logentry := handler.GetLogger()
		logentry.Warn("no internal handler registered")
	}

	// testing, mark read function
	if handler.WhatsappOptionsExtended.ReadUpdate && !message.FromBroadcast() {
		go handler.MarkRead(message, types.ReceiptTypeRead)
	}
}

func (handler *WhatsmeowHandlers) MarkRead(message *whatsapp.WhatsappMessage, receipt types.ReceiptType) (err error) {
	logentry := handler.GetLogger()

	client := handler.Client
	ids := []string{message.Id}
	chatJID, err := types.ParseJID(message.Chat.Id)
	if err != nil {
		logentry.Errorf("error on mark read, parsing chat jid: %s", err.Error())
		return
	}

	var senderJID types.JID
	if message.Participant != nil {
		senderJID, err = types.ParseJID(message.Participant.Id)
		if err != nil {
			logentry.Errorf("error on mark read, parsing sender jid: %s", err.Error())
			return
		}
	}

	readtime := time.Now()
	err = client.MarkRead(ids, readtime, chatJID, senderJID, receipt)
	if err != nil {
		logentry.Errorf("error on mark read: %s", err.Error())
		return
	}

	logentry.Debugf("marked read chat id: %s, at: %v", message.Chat.Id, readtime)
	return
}

//#region EVENT CALL

func (source *WhatsmeowHandlers) CallMessage(evt types.BasicCallMeta) {
	logentry := source.GetLogger()
	logentry.Trace("event CallMessage !")

	message := &whatsapp.WhatsappMessage{Content: evt}

	// basic information
	message.Id = evt.CallID
	message.Timestamp = evt.Timestamp
	message.FromMe = false

	message.Chat = whatsapp.WhatsappChat{}
	chatID := fmt.Sprint(evt.From.User, "@", evt.From.Server)
	message.Chat.Id = chatID

	message.Type = whatsapp.CallMessageType

	if source.WAHandlers != nil {

		// following to internal handlers
		go source.WAHandlers.Message(message, "call")
	}

	// should reject this call
	if !source.HandleCalls() {
		err := source.Client.RejectCall(evt.From, evt.CallID)
		if err != nil {
			logentry.Errorf("error on rejecting call: %s", err.Error())
		} else {
			logentry.Infof("rejecting incoming call from: %s", evt.From)
		}
	}
}

/*
func (source *WhatsmeowHandlers) AcceptCall(from types.JID) error {
	if source == nil {
		return fmt.Errorf("nil source handler")
	}

	var node = binary.Node{
		Tag: "ack",
		Attrs: binary.Attrs{
			"id":    source.Client.GenerateMessageID(),
			"to":    from,
			"class": "receipt",
			"from":  source.Client.Store.ID.String(),
		},
	}

	logentry := source.GetLogger()
	logentry.Infof("accepting incoming call from: %s", from)

	return source.Client.DangerousInternals().SendNode(node)
}
*/
//#endregion

// #region EVENT READ RECEIPT

func (source *WhatsmeowHandlers) Receipt(evt events.Receipt) {
	eventid := atomic.AddUint64(&source.Counter, 1)
	logentry := source.GetLogger()
	logentry = logentry.WithField(LogFields.EventId, eventid)

	logentry.Trace("event Receipt !")

	chatID := fmt.Sprint(evt.Chat.User, "@", evt.Chat.Server)

	// Ignore chats with @broadcast and @newsletter
	if strings.Contains(chatID, "@broadcast") || strings.Contains(chatID, "@newsletter") {
		return
	}

	statuses := make(map[string]whatsapp.WhatsappMessageStatus)
	for _, id := range evt.MessageIDs {
		last := statuses[id]
		current := GetWhatsappMessageStatus(evt.Type)

		sublogentry := logentry.WithField(LogFields.MessageId, id)
		sublogentry.Tracef("reading receipt event, from: %s, type: %s, status: %s", evt.SourceString(), evt.Type, current)

		if current.Uint32() > last.Uint32() {
			statuses[id] = current
		}
	}

	if source.WAHandlers == nil {
		return
	}

	for id, status := range statuses {
		updated := source.WAHandlers.MessageStatusUpdate(id, status)
		if !updated {
			continue
		}

		sublogentry := logentry.WithField(LogFields.MessageId, id)
		sublogentry.Debugf("updated status: %s", status)

		if status.Uint32() != whatsapp.WhatsappMessageStatusRead.Uint32() {
			continue
		}

		if !source.HandleReadReceipts() {
			continue
		}

		sublogentry.Debugf("dispatching read receipt event, status: %s", status)

		message := &whatsapp.WhatsappMessage{Content: evt}
		message.Id = "readreceipt"

		// basic information
		message.Timestamp = evt.Timestamp
		message.FromMe = false

		message.Chat = whatsapp.WhatsappChat{}
		message.Chat.Id = chatID

		message.Type = whatsapp.SystemMessageType

		// message ids comma separated
		message.Text = id

		// following to internal handlers
		go source.WAHandlers.Receipt(message)
	}
}

//#endregion

//#region HANDLE LOGGED OUT EVENT

func (handler *WhatsmeowHandlers) OnLoggedOutEvent(evt events.LoggedOut) {
	reason := evt.Reason.String()
	handler.GetLogger().Tracef("logged out %s", reason)

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

	handler.Follow(message, "logout")
	handler.UnRegister()
}

//#endregion

//#region HANDLE GROUP JOIN OUT EVENT

func (handler *WhatsmeowHandlers) JoinedGroup(evt events.JoinedGroup) {

	id := evt.CreateKey
	if len(id) == 0 {
		id = handler.Client.GenerateMessageID()
	}

	message := &whatsapp.WhatsappMessage{
		Content: evt,

		Id:        id,
		Timestamp: time.Now().Truncate(time.Second),
		Type:      whatsapp.GroupMessageType,
		Chat:      whatsapp.WhatsappChat{Id: evt.GroupInfo.JID.String(), Title: evt.GroupInfo.GroupName.Name},
		Text:      "joined group",

		Info: GroupJoinInfo{
			Owner:        evt.GroupInfo.OwnerJID.String(),
			Created:      evt.GroupInfo.GroupCreated,
			Participants: len(evt.GroupInfo.Participants),
			Reason:       evt.Reason,
			Type:         evt.Type,
		},
	}

	handler.Follow(message, "group")
}

//#endregion
