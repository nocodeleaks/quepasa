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
	"go.mau.fi/whatsmeow/appstate"
	types "go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

type WhatsmeowHandlers struct {
	WhatsmeowOptions          // default whatsmeow service global options
	*WhatsmeowConnection      // Connection reference for accessing embedded managers
	*whatsapp.WhatsappOptions // particular whatsapp options for this handler

	WAHandlers whatsapp.IWhatsappHandlers

	eventHandlerID           uint32
	unregisterRequestedToken bool

	// events counter
	Counter uint64
}

func NewWhatsmeowHandlers(conn *WhatsmeowConnection, wmOptions WhatsmeowOptions, waOptions *whatsapp.WhatsappOptions) *WhatsmeowHandlers {
	return &WhatsmeowHandlers{
		WhatsmeowConnection: conn,
		WhatsmeowOptions:    wmOptions,
		WhatsappOptions:     waOptions,
	}
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

func (source WhatsmeowHandlers) ShouldDispatchUnhandled() bool {
	options := source.GetServiceOptions()
	return options.DispatchUnhandled
}

func (source WhatsmeowHandlers) HandleHistorySync() bool {
	options := source.GetServiceOptions()
	if options.HistorySync != nil {
		return true
	}

	return whatsapp.WhatsappHistorySync
}

// only affects whatsmeow
func (handler *WhatsmeowHandlers) UnRegister(reason string) {
	if handler == nil {
		return
	}

	handler.unregisterRequestedToken = true

	logentry := handler.GetLogger()
	logentry.Tracef("unregistering handler, id: %v, reason: %s", handler.eventHandlerID, reason)

	// if client is nil, we can't unregister
	if handler.Client == nil {
		if reason == "dispose" {
			logentry.Tracef("unregister requested, but client is already nil, reason: %s", reason)
		} else {
			logentry.Warnf("unregister requested, but client is nil, reason: %s", reason)
		}
		return
	}

	// if is this session
	found := handler.Client.RemoveEventHandler(handler.eventHandlerID)
	if found {
		logentry.Infof("handler unregistered, id: %v, reason: %s", handler.eventHandlerID, reason)
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

		if source.Client == nil {
			logentry.Debugf("unregister requested, but client is already nil")
			return
		}

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

	case *events.PairError:
		{
			jsonEvt := library.ToJson(evt)
			logentry.Errorf("pair error event: %s", jsonEvt)
		}

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
		logentry.Tracef("event not implemented yet: %v", reflect.TypeOf(evt))
		if source.ShouldDispatchUnhandled() {
			go source.DispatchUnhandledEvent(evt, reflect.TypeOf(rawEvt).String())
		}
		return // ignoring not implemented yet

	default:
		logentry.Debugf("event not handled: %v", reflect.TypeOf(evt))

		// Only dispatch debug events if DEBUGEVENTS is true
		// If DEBUGEVENTS=false or not set, do nothing (no webhook dispatch)
		if source.ShouldDispatchUnhandled() {
			go source.DispatchUnhandledEvent(evt, reflect.TypeOf(rawEvt).String())
		}
		return
	}
}

// DispatchUnhandledEvent creates a debug message for unhandled events and dispatches it
func (source *WhatsmeowHandlers) DispatchUnhandledEvent(evt interface{}, eventType string) {
	logentry := source.GetLogger()
	logentry.Debugf("dispatching debug event: %s", eventType)

	// Clean up the event type by removing the *events. prefix
	cleanEventType := strings.TrimPrefix(eventType, "*events.")

	message := &whatsapp.WhatsappMessage{
		Content:   evt,
		Id:        source.Client.GenerateMessageID(),
		Timestamp: time.Now().Truncate(time.Second),
		Type:      whatsapp.UnhandledMessageType,
		FromMe:    false,
	}

	// Create debug information with the event in JSON format
	message.Debug = &whatsapp.WhatsappMessageDebug{
		Event:  cleanEventType,
		Info:   evt,
		Reason: "event",
	}

	// Try to extract chat information from events that have Info field
	if eventWithInfo, ok := evt.(interface{ GetInfo() types.MessageInfo }); ok {
		info := eventWithInfo.GetInfo()

		// basic information
		message.Id = info.ID
		message.Timestamp = ImproveTimestamp(info.Timestamp)
		message.FromMe = info.IsFromMe

		// Populate chat and participant information
		source.PopulateChatAndParticipant(message, info)

		// Follow the same pattern as other messages
		source.Follow(message, "debug")
		return
	}

	// Fallback to system chat if we can't extract chat information
	message.Chat = whatsapp.WASYSTEMCHAT
	source.Follow(message, "debug")
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

func (handler *WhatsmeowHandlers) PopulateChatAndParticipant(message *whatsapp.WhatsappMessage, info types.MessageInfo) {
	message.Chat = *NewWhatsappChat(handler, info.Chat)

	if info.IsGroup {
		message.Participant = NewWhatsappChat(handler, info.Sender)
	} /* else { // Obsolete
		// If title is empty, use Phone as fallback
		if len(message.Chat.Title) == 0 && message.FromMe {
			message.Chat.Title = library.GetPhoneByWId(message.Chat.Id)
		}
	}
	*/
}

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

	// Populate chat and participant information
	handler.PopulateChatAndParticipant(message, evt.Info)

	// Process diferent message types
	HandleKnowingMessages(handler, message, evt.Message)

	// discard and return
	if message.Type == whatsapp.UnhandledMessageType {
		if message.Debug == nil {
			logentry.Warnf("unhandled message type, no debug information: %s", message.Type)
		}
	}

	handler.Follow(message, from)
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

	message.Chat = *NewWhatsappChat(source, evt.From)
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

		message.Chat = *NewWhatsappChat(source, evt.Chat)
		message.Type = whatsapp.SystemMessageType // message ids comma separated
		message.Text = id

		// following to internal handlers
		go source.WAHandlers.Receipt(message)
	}
}

//#endregion

//#region HANDLE LOGGED OUT EVENT

func (handler *WhatsmeowHandlers) OnLoggedOutEvent(evt events.LoggedOut) {
	reason := evt.Reason.String()

	logentry := handler.GetLogger()
	logentry.Tracef("logged out %s", reason)

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
	handler.UnRegister("logged out event")
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
