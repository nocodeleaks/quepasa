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
	result := serviceOptions.HandleCalls(defaultValue)

	// Debug logging
	logentry := log.WithField("component", "call_handler")
	logentry.Infof("🔍 HandleCalls() debug:")
	logentry.Infof("   📋 source.WhatsappOptions: %+v", source.WhatsappOptions)
	logentry.Infof("   📋 defaultValue (Calls): %+v", defaultValue)
	logentry.Infof("   📋 serviceOptions: %+v", serviceOptions)
	logentry.Infof("   📋 Final result: %v", result)

	return result
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
	logentry.Infof("🔧 Event handler registered, ID: %v, LogLevel: %s", source.eventHandlerID, logentry.Level)
	logentry.Infof("📞 Call events (CallOffer, CallOfferNotice, CallTerminate) will be captured!")
	logentry.Infof("🎯 SIP Proxy ready to process incoming calls")

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

	// 🔍 LOG ALL CALL-RELATED EVENTS FOR DEBUGGING
	eventType := fmt.Sprintf("%T", rawEvt)
	if strings.Contains(eventType, "Call") {
		logentry.Infof("🔍🔍🔍 CALL-RELATED EVENT DETECTED: %s = %+v", eventType, rawEvt)

		// EXTRA DEBUG: Check event type more specifically
		if eventType == "*events.CallOffer" {
			logentry.Infof("🎯🎯🎯 CALLOFFER DETECTED IN MAIN HANDLER! 🎯🎯🎯")
		}
		if eventType == "*events.CallOfferNotice" {
			logentry.Infof("📢📢📢 CALLOFFERNOTICE DETECTED IN MAIN HANDLER! 📢📢📢")
		}
	}

	// 🔍 EXTRA DEBUG: Log ALL events during call timeframe
	if eventType != "*events.Receipt" && eventType != "*events.HistorySync" && eventType != "*events.Message" {
		logentry.Debugf("📋 Event: %s", eventType)
	}

	switch evt := rawEvt.(type) {

	case *events.Message:
		go source.Message(*evt, "live")
		return

		//# region CALLS - Enhanced debugging for VoIP call events
	case *events.CallOffer:
		// =========================================================================
		// 🚫 IMMEDIATE SIP PROCESSING COMMENTED OUT (CAUSES DUPLICATION)
		// =========================================================================
		// This immediate processing was causing duplicate SIP INVITEs
		// We handle CallOffer in the dedicated CallMessage() function instead

		// COMMENTED OUT: Immediate SIP forwarding (causes duplicate calls)
		// logentry.Infof("🔥🔥🔥 CALLOFFER CAPTURED! 🔥🔥🔥")
		// logentry.Infof("📞 CallOffer received: %+v", evt)
		// logentry.Infof("📞 Call Details - From: %s, CallID: %s, Type: CallOffer", evt.From, evt.CallID)
		// logentry.Infof("🎯 STARTING SIP PROXY PROCESS NOW!")
		// logentry.Infof("🔥🔥🔥 PROCESSING CALL IMMEDIATELY 🔥🔥🔥")
		// ... (rest of immediate processing code commented out)
		// logentry.Infof("📡 Call processing completed - forwarded to voip.sufficit.com.br:26499")

		// Only basic logging for CallOffer detection
		logentry.Infof("📞 CallOffer detected - CallID: %s, From: %s", evt.CallID, evt.From.User)

		// Let CallMessage() handle the actual SIP forwarding
		go source.CallMessage(evt.BasicCallMeta)
		return

	case *events.CallOfferNotice:
		logentry.Infof("�📢📢 CALLOFFERNOTICE CAPTURED! 📢📢📢")
		logentry.Infof("����🔔 CallOfferNotice received: %+v", evt)
		logentry.Infof("📞 Call Details - From: %s, CallID: %s, Type: CallOfferNotice", evt.From, evt.CallID)
		logentry.Infof("🎯 PROCESSING SIP PROXY DATA!")

		// Enhanced debugging with CallManager
		if callManager := source.WhatsmeowConnection.GetCallManager(); callManager != nil {
			logentry.Infof("✅ CallManager created for CallOfferNotice")
			callManager.LogCallEvent("CallOfferNotice", evt)
			logentry.Infof("✅ CallOfferNotice logged successfully")
		} else {
			logentry.Errorf("❌ Failed to create CallManager for CallOfferNotice!")
		}

		go source.CallMessage(evt.BasicCallMeta)
		return

	case *events.CallRelayLatency:
		logentry.Infof("📊 CallRelayLatency: %+v", evt)
		logentry.Infof("📞 Call Performance - From: %s, CallID: %s, Latency: %v", evt.From, evt.CallID, evt.Data)

		// Enhanced debugging with CallManager
		if callManager := source.WhatsmeowConnection.GetCallManager(); callManager != nil {
			callManager.LogCallEvent("CallRelayLatency", evt)
		}
		return

	case *events.CallTerminate:
		logentry.Infof("📞❌ CallTerminate: %+v", evt)
		logentry.Infof("📞 Call Ended - From: %s, CallID: %s, Reason: %v", evt.From, evt.CallID, evt.Reason)

		// Enhanced debugging with CallManager
		if callManager := source.WhatsmeowConnection.GetCallManager(); callManager != nil {
			callManager.LogCallEvent("CallTerminate", evt)
		}

		go source.CallTerminateMessage(evt.BasicCallMeta, evt.Reason)
		return

	case *events.CallAccept:
		logentry.Infof("📞✅ CallAccept: %+v", evt)
		logentry.Infof("📞 Call Accepted - From: %s, CallID: %s", evt.From, evt.CallID)

		// Enhanced debugging with CallManager
		if callManager := source.WhatsmeowConnection.GetCallManager(); callManager != nil {
			callManager.LogCallEvent("CallAccept", evt)
		}

		go source.CallAcceptMessage(evt.BasicCallMeta)
		return
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

	// =========================================================================
	// 🚫 EXPERIMENTAL CALL PROCESSING COMPLETELY DISABLED
	// =========================================================================
	// We only process: CallOffer → SIP Server → Monitor SIP Response
	// NO WhatsApp experimental features, NO call answer managers

	// COMMENTED OUT: Experimental call processing
	// if source.WhatsmeowConnection != nil {
	//	callAnswerManager := NewCallAnswerManager(source.WhatsmeowConnection)

	// =========================================================================
	// 🚫 EXPERIMENTAL WhatsApp CALL PROCESSING COMMENTED OUT
	// =========================================================================
	// We only focus on: CallOffer → SIP Server → Monitor SIP Response
	// NO WhatsApp experimental features, NO call monitoring, NO debugging

	// COMMENTED OUT: Enhanced monitoring and experimental call processing
	// Start enhanced monitoring
	// callAnswerManager.StartCallMonitoring()

	// Log debugging info
	// debugInfo := callAnswerManager.GetCallDebuggingInfo()
	// logentry.Infof("🔍 Call System Debug Info: %+v", debugInfo)

	// Experimental: Try to keep call active for data capture (no auto-accept)
	// if source.HandleCalls() {
	//	logentry.Infof("🧪 EXPERIMENTAL: Maintaining call active for data capture from %s", evt.From)
	//	err := callAnswerManager.ExperimentalAcceptCall(evt.From, evt.CallID)
	//	if err != nil {
	//		logentry.Warnf("🧪 Call persistence experiment failed (expected): %v", err)
	//	}
	// }
	// } // End of experimental call processing

	// =========================================================================
	// 🚫 WhatsApp MESSAGE PROCESSING COMMENTED OUT
	// =========================================================================
	// We don't send messages to internal WhatsApp handlers

	// COMMENTED OUT: WhatsApp message processing
	// message := &whatsapp.WhatsappMessage{Content: evt}
	// message.Id = evt.CallID
	// message.Timestamp = evt.Timestamp
	// message.FromMe = false
	// message.Chat = *NewWhatsappChat(source, evt.From)
	// message.Type = whatsapp.CallMessageType
	// if source.WAHandlers != nil {
	//	go source.WAHandlers.Message(message, "call")
	// }

	// =========================================================================
	// � IMMEDIATE CALL PROCESSING COMMENTED OUT
	// =========================================================================
	// This was duplicating SIP calls - we handle this in the section below

	// COMMENTED OUT: Immediate call processing (was causing duplicate SIP calls)
	// �� PROCESSING CALL IMMEDIATELY - DON'T WAIT FOR HandleCalls()
	// logentry.Infof("🔥🔥🔥 PROCESSING CALL IMMEDIATELY 🔥🔥🔥")
	// logentry.Infof("📞 From: %s", evt.From)
	// logentry.Infof("📞 CallID: %s", evt.CallID)
	// logentry.Infof("📞 Phone: %s", evt.From.User)

	// Create CallManager immediately - don't wait for HandleCalls()
	// var callManager *WhatsmeowCallManager
	// if source.WhatsmeowConnection.CallManager != nil {
	//	callManager = source.WhatsmeowConnection.CallManager
	//	logentry.Infof("🔄 Using existing CallManager")
	// } else {
	//	callManager = NewWhatsmeowCallManager(source.WhatsmeowConnection)
	//	source.WhatsmeowConnection.CallManager = callManager
	//	logentry.Infof("🆕 Created new CallManager")
	// }

	// COMMENTED OUT: Duplicate SIP processing (handled in main section below)
	// logentry.Infof("🎯🎯🎯 FORWARDING CALL TO SIP SERVER IMMEDIATELY 🎯🎯🎯")
	// sipIntegration := callManager.GetSIPProxy()
	// if sipIntegration != nil && sipIntegration.IsReady() {
	//	logentry.Infof("🚀 SIP integration ready, forwarding call...")
	//	statusManager := source.GetStatusManager()
	//	myNumber, err := statusManager.GetWidInternal()
	//	if err != nil {
	//		logentry.Errorf("❌ Failed to get WhatsApp number: %v", err)
	//		myNumber = "unknown"
	//	}
	//	err = sipIntegration.ThrowSIPProxy(evt.CallID, evt.From.User, myNumber)
	//	if err != nil {
	//		logentry.Errorf("❌ Failed to forward call to SIP server: %v", err)
	//	} else {
	//		logentry.Infof("✅ Call successfully forwarded to SIP server!")
	//		logentry.Infof("📞 CallID: %s", evt.CallID)
	//		logentry.Infof("📞 From: %s", evt.From.User)
	//		logentry.Infof("🎯 SIP server will handle call acceptance/rejection")
	//	}
	// } else {
	//	logentry.Errorf("❌ SIP integration not ready - cannot forward call")
	//	if sipIntegration == nil {
	//		logentry.Errorf("   → sipIntegration is nil")
	//	} else {
	//		logentry.Errorf("   → sipIntegration.IsReady() = %v", sipIntegration.IsReady())
	//	}
	// }

	logentry.Infof("📡 Call processing completed")

	// 🚀 MINIMAL CALL PROCESSING: CallOffer → SIP Server Only
	handleCallsResult := source.HandleCalls()
	logentry.Infof("🔍 HandleCalls() result: %v", handleCallsResult)

	if handleCallsResult {
		logentry.Infof("📡 CALL DETECTED - Forwarding to SIP server")
		logentry.Infof("📞 CallID: %s | From: %s", evt.CallID, evt.From.User)

		// Use the internal SIP call manager
		sipCallManager := source.WhatsmeowConnection.GetSIPCallManager()
		if sipCallManager.IsEnabled() {
			statusManager := source.GetStatusManager()
			myNumber, err := statusManager.GetWidInternal()
			if err != nil {
				logentry.Errorf("❌ Failed to get WhatsApp number: %v", err)
				myNumber = "unknown"
			}

			// Process incoming WhatsApp call through internal SIP manager
			logentry.Infof("📞 CALL DETECTED - Processing via internal SIP manager:")
			logentry.Infof("   🔵 From (caller): %s", evt.From.User)
			logentry.Infof("   🟢 To (receiver): %s", myNumber)
			logentry.Infof("   📞 CallID: %s", evt.CallID)

			err = sipCallManager.ProcessIncomingCall(evt.CallID, evt.From.User, myNumber)
			if err != nil {
				logentry.Errorf("❌ SIP call processing failed: %v", err)
			} else {
				logentry.Infof("✅ Call processed via internal SIP manager")
			}
		} else {
			logentry.Warnf("⚠️ SIP call manager not enabled")
		}
	} else {
		logentry.Warnf("⚠️ Call handling disabled - no SIP forwarding")
	}

	// =========================================================================
	// 🚫 ALL WhatsApp INTERACTIONS BELOW ARE COMMENTED OUT
	// =========================================================================
	// We only monitor CallOffer → SIP Server and analyze SIP responses
	// NO auto-reject, NO call state changes, NO WhatsApp API calls

	// COMMENTED OUT: Auto-rejection logic (was terminating calls automatically from database config)
	// if !source.HandleCalls() {
	//	err := source.Client.RejectCall(evt.From, evt.CallID)
	//	if err != nil {
	//		logentry.Errorf("error on rejecting call: %s", err.Error())
	//	} else {
	//		logentry.Infof("rejecting incoming call from: %s", evt.From)
	//	}
	// }
}

// CallTerminateMessage handles call termination events
func (source *WhatsmeowHandlers) CallTerminateMessage(evt types.BasicCallMeta, reason interface{}) {
	logentry := source.GetLogger()
	logentry.Tracef("📞❌ Call terminated - CallID: %s, From: %s, Reason: %v", evt.CallID, evt.From, reason)

	// =========================================================================
	// 🚫 WhatsApp TERMINATION PROCESSING COMMENTED OUT
	// =========================================================================
	// We only monitor call termination, NO WhatsApp message processing

	// COMMENTED OUT: WhatsApp message processing for call termination
	// message := &whatsapp.WhatsappMessage{Content: evt}
	// message.Id = evt.CallID
	// message.Timestamp = evt.Timestamp
	// message.FromMe = false
	// message.Chat = *NewWhatsappChat(source, evt.From)
	// message.Type = whatsapp.CallMessageType
	// message.Text = fmt.Sprintf("Call terminated. Reason: %v", reason)
	// if source.WAHandlers != nil {
	//	go source.WAHandlers.Message(message, "call_terminate")
	// }
}

// CallAcceptMessage handles call accept events
func (source *WhatsmeowHandlers) CallAcceptMessage(evt types.BasicCallMeta) {
	logentry := source.GetLogger()
	logentry.Tracef("📞✅ Call accepted - CallID: %s, From: %s", evt.CallID, evt.From)

	// =========================================================================
	// 🚫 WhatsApp ACCEPTANCE PROCESSING COMMENTED OUT
	// =========================================================================
	// We only monitor call acceptance, NO WhatsApp message processing or SIP notifications

	// COMMENTED OUT: SIP proxy notification for call acceptance
	// logentry.Infof("🎯 SIP PROXY: WhatsApp call was ACCEPTED - SIP integration active")
	// if callManager := source.WhatsmeowConnection.GetCallManager(); callManager != nil {
	//	if sipIntegration := callManager.GetSIPProxy(); sipIntegration != nil {
	//		activeCalls := sipIntegration.GetActiveCalls()
	//		logentry.Infof("📊 SIP integration has %d active calls", len(activeCalls))
	//	}
	// }

	// COMMENTED OUT: WhatsApp message processing for call acceptance
	// message := &whatsapp.WhatsappMessage{Content: evt}
	// message.Id = evt.CallID
	// message.Timestamp = evt.Timestamp
	// message.FromMe = false
	// message.Chat = *NewWhatsappChat(source, evt.From)
	// message.Type = whatsapp.CallMessageType
	// message.Text = "Call accepted"
	// if source.WAHandlers != nil {
	//	go source.WAHandlers.Message(message, "call_accept")
	// }
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
