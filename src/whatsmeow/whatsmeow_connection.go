package whatsmeow

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"
	"unicode"

	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"

	library "github.com/nocodeleaks/quepasa/library"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	whatsmeow "go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	types "go.mau.fi/whatsmeow/types"
)

// Must Implement IWhatsappConnection
type WhatsmeowConnection struct {
	library.LogStruct // logging
	Client            *whatsmeow.Client

	Handlers       *WhatsmeowHandlers       // composition for handlers
	GroupManager   *WhatsmeowGroupManager   // composition for group operations
	StatusManager  *WhatsmeowStatusManager  // composition for status operations
	ContactManager *WhatsmeowContactManager // composition for contact operations

	failedToken  bool
	paired       func(string)
	IsConnecting bool `json:"isconnecting"` // used to avoid multiple connection attempts
}

//#region IMPLEMENT WHATSAPP CONNECTION OPTIONS INTERFACE

// get default log entry, never nil
func (source *WhatsmeowConnection) GetLogger() *log.Entry {
	if source != nil && source.LogEntry != nil {
		return source.LogEntry
	}

	logentry := library.NewLogEntry(source)
	if source != nil {
		statusManager := source.GetStatusManager()
		wid, _ := statusManager.GetWidInternal()
		if len(wid) > 0 {
			logentry = logentry.WithField(LogFields.WId, wid)
		}
		source.LogEntry = logentry
	}

	logentry.Level = log.ErrorLevel
	logentry.Infof("generating new log entry for %s, with level: %s", reflect.TypeOf(source), logentry.Level)
	return logentry
}

//#endregion

//region IMPLEMENT INTERFACE WHATSAPP CONNECTION

// returns a valid chat title from local memory store
func (conn *WhatsmeowConnection) GetChatTitle(wid string) string {
	return GetChatTitleFromWId(conn, wid)
}

// Connect to websocket only, dot not authenticate yet, errors come after
func (source *WhatsmeowConnection) Connect() (err error) {
	source.GetLogger().Info("starting whatsmeow connection")

	if source.IsConnecting {
		return
	}

	source.IsConnecting = true

	err = source.Client.Connect()
	source.IsConnecting = false

	if err != nil {
		source.failedToken = true
		return
	}

	// waits 2 seconds for loggedin
	// not required
	_ = source.Client.WaitForConnection(time.Millisecond * 2000)

	source.failedToken = false
	return
}

// func (cli *Client) Download(msg DownloadableMessage) (data []byte, err error)
func (source *WhatsmeowConnection) DownloadData(imsg whatsapp.IWhatsappMessage) (data []byte, err error) {
	msg := imsg.GetSource()
	logentry := source.GetLogger().WithField(LogFields.MessageId, imsg.GetId())

	// Try direct downloadable message first
	downloadable, ok := msg.(whatsmeow.DownloadableMessage)
	if ok {
		logentry.Trace("Message implements DownloadableMessage directly, using Client.Download()")
		return source.Client.Download(context.Background(), downloadable)
	}

	waMsg, ok := msg.(*waE2E.Message)
	if ok {
		downloadable = GetDownloadableMessage(waMsg)
		if downloadable != nil {
			logentry.Trace("waMsg implements DownloadableMessage, using Client.Download()")
			return source.Client.Download(context.Background(), downloadable)
		}
	}

	// If internal content as VCard or Localization
	attach := imsg.GetAttachment()
	if attach != nil {
		data := attach.GetContent()
		if data != nil {
			logentry.Trace("no waMsg, found attachment, returning content")
			return *data, err
		}
	}
	// If we reach here, it means we have a waE2E.Message but no DownloadableMessage interface
	return nil, fmt.Errorf("message (%s) is not downloadable", imsg.GetId())
}

func (conn *WhatsmeowConnection) Download(imsg whatsapp.IWhatsappMessage, cache bool) (att *whatsapp.WhatsappAttachment, err error) {
	logentry := conn.GetLogger().WithField(LogFields.MessageId, imsg.GetId())
	logentry.Tracef("Download() method called, Cache: %v", cache)

	att = imsg.GetAttachment()
	if att == nil {
		return nil, fmt.Errorf("message (%s) does not contains attachment info", imsg.GetId())
	}

	if cache && att.HasContent() {
		logentry.Debugf("Download() using cached content - HasContent: %v", att.HasContent())
		return att, nil
	}

	data, err := conn.DownloadData(imsg)
	if err != nil {
		return nil, fmt.Errorf("failed to download data for message (%s): %v", imsg.GetId(), err)
	}

	if !cache {
		newAtt := *att
		att = &newAtt
	}

	att.SetContent(&data)
	return
}

func (source *WhatsmeowConnection) Revoke(msg whatsapp.IWhatsappMessage) error {
	logentry := source.GetLogger()

	jid, err := types.ParseJID(msg.GetChatId())
	if err != nil {
		logentry.Infof("revoke error on get jid: %s", err)
		return err
	}

	participantJid, err := types.ParseJID(msg.GetParticipantId())
	if err != nil {
		logentry.Infof("revoke error on get jid: %s", err)
		return err
	}

	newMessage := source.Client.BuildRevoke(jid, participantJid, msg.GetId())
	_, err = source.Client.SendMessage(context.Background(), jid, newMessage)
	if err != nil {
		logentry.Infof("revoke error: %s", err)
		return err
	}

	return nil
}

// Edit edits an existing message with new content
func (source *WhatsmeowConnection) Edit(msg whatsapp.IWhatsappMessage, newContent string) error {
	logentry := source.GetLogger()

	jid, err := types.ParseJID(msg.GetChatId())
	if err != nil {
		logentry.Infof("edit message error on get jid: %s", err)
		return err
	}

	// Build text message with new content
	textMessage := &waE2E.Message{
		Conversation: &newContent,
	}

	// Build edit message
	editMessage := source.Client.BuildEdit(jid, msg.GetId(), textMessage)
	_, err = source.Client.SendMessage(context.Background(), jid, editMessage)
	if err != nil {
		logentry.Infof("edit message error: %s", err)
		return err
	}

	return nil
}

func isASCII(s string) bool {
	for _, c := range s {
		if c > unicode.MaxASCII {
			return false
		}
	}
	return true
}

func (source *WhatsmeowConnection) GetContextInfo(msg whatsapp.WhatsappMessage) *waE2E.ContextInfo {

	var contextInfo *waE2E.ContextInfo
	if len(msg.InReply) > 0 {
		contextInfo = source.GetInReplyContextInfo(msg)
	}

	// mentions ---------------------------------------
	if msg.FromGroup() {
		messageText := msg.GetText()
		mentions := GetMentions(messageText)
		if len(mentions) > 0 {

			if contextInfo == nil {
				contextInfo = &waE2E.ContextInfo{}
			}

			contextInfo.MentionedJID = mentions
		}
	}

	// disapering messages, not implemented yet
	if contextInfo != nil {
		contextInfo.Expiration = proto.Uint32(0)
		contextInfo.EphemeralSettingTimestamp = proto.Int64(0)
		contextInfo.DisappearingMode = &waE2E.DisappearingMode{Initiator: waE2E.DisappearingMode_CHANGED_IN_CHAT.Enum()}
	}

	return contextInfo
}

func (source *WhatsmeowConnection) GetInReplyContextInfo(msg whatsapp.WhatsappMessage) *waE2E.ContextInfo {
	logentry := source.GetLogger()

	// default information for cached messages
	var info types.MessageInfo

	// getting quoted message if available on cache
	// (optional) another devices will process anyway, but our devices will show quoted only if it exists on cache
	var quoted *waE2E.Message

	cached, _ := source.GetHandlers().WAHandlers.GetById(msg.InReply)
	if cached != nil {

		// update cached info
		info, _ = cached.InfoForHistory.(types.MessageInfo)

		if cached.Content != nil {
			if content, ok := cached.Content.(*waE2E.Message); ok {

				// update quoted message content
				quoted = content

			} else {
				logentry.Warnf("content has an invalid type (%s), on reply to msg id: %s", reflect.TypeOf(cached.Content), msg.InReply)
			}
		} else {
			logentry.Warnf("message content not cached, on reply to msg id: %s", msg.InReply)
		}
	} else {
		logentry.Warnf("message not cached, on reply to msg id: %s", msg.InReply)
	}

	var participant *string
	if (types.MessageInfo{}) != info {
		var sender string
		if msg.FromGroup() {
			sender = fmt.Sprint(info.Sender.User, "@", info.Sender.Server)
		} else {
			sender = fmt.Sprint(info.Chat.User, "@", info.Chat.Server)
		}
		participant = proto.String(sender)
	}

	return &waE2E.ContextInfo{
		StanzaID:      proto.String(msg.InReply),
		Participant:   participant,
		QuotedMessage: quoted,
	}
}

// Default SEND method using WhatsappMessage Interface
func (source *WhatsmeowConnection) Send(msg *whatsapp.WhatsappMessage) (whatsapp.IWhatsappSendResponse, error) {
	logentry := source.GetLogger()
	loglevel := logentry.Level
	logentry = logentry.WithField(LogFields.MessageId, msg.Id)
	logentry.Level = loglevel

	var err error

	// Formatting destination accordingly
	formattedDestination, _ := whatsapp.FormatEndpoint(msg.GetChatId())

	// avoid common issue with incorrect non ascii chat id
	if !isASCII(formattedDestination) {
		err = fmt.Errorf("not an ASCII formatted chat id")
		return msg, err
	}

	// validating jid before remote commands as upload or send
	jid, err := types.ParseJID(formattedDestination)
	if err != nil {
		logentry.Infof("send error on get jid: %s", err)
		return msg, err
	}

	// request message text
	messageText := msg.GetText()

	var newMessage *waE2E.Message
	if !msg.HasAttachment() {
		if IsValidForButtons(messageText) {
			internal := GenerateButtonsMessage(messageText)
			internal.ContextInfo = source.GetContextInfo(*msg)
			newMessage = &waE2E.Message{
				ButtonsMessage: internal,
			}
		} else {
			if msg.Poll != nil {
				newMessage, err = GeneratePollMessage(msg)
				if err != nil {
					return msg, err
				}
			} else {
				internal := &waE2E.ExtendedTextMessage{Text: &messageText}
				internal.ContextInfo = source.GetContextInfo(*msg)
				newMessage = &waE2E.Message{ExtendedTextMessage: internal}
			}
		}
	} else {
		newMessage, err = source.UploadAttachment(*msg)
		if err != nil {
			return msg, err
		}
	}

	// Generating a new unique MessageID
	if len(msg.Id) == 0 {
		msg.Id = source.Client.GenerateMessageID()
	}

	extra := whatsmeow.SendRequestExtra{
		ID: msg.Id,
	}

	// saving cached content for instance of future reply
	if msg.Content == nil {
		msg.Content = newMessage
	}

	resp, err := source.Client.SendMessage(context.Background(), jid, newMessage, extra)
	if err != nil {
		logentry.Errorf("whatsmeow connection send error: %s", err)
		return msg, err
	}

	// updating timestamp
	msg.Timestamp = resp.Timestamp

	if msg.Id != resp.ID {
		logentry.Warnf("send success but msg id differs from response id: %s, type: %v, on: %s", resp.ID, msg.Type, msg.Timestamp)
	} else {
		logentry.Infof("send success, type: %v, on: %s", msg.Type, msg.Timestamp)
	}

	// testing, mark read function
	if source.GetHandlers().ReadUpdate {
		go source.GetHandlers().MarkRead(msg, types.ReceiptTypeRead)
	}

	return msg, err
}

// useful to check if is a member of a group before send a msg.
// fails on recently added groups.
// pending a more efficient code !!!!!!!!!!!!!!
func (source *WhatsmeowConnection) HasChat(chat string) bool {
	jid, err := types.ParseJID(chat)
	if err != nil {
		return false
	}

	info, err := source.Client.Store.ChatSettings.GetChatSettings(context.TODO(), jid)
	if err != nil {
		return false
	}

	return info.Found
}

// func (cli *Client) Upload(ctx context.Context, plaintext []byte, appInfo MediaType) (resp UploadResponse, err error)
func (source *WhatsmeowConnection) UploadAttachment(msg whatsapp.WhatsappMessage) (result *waE2E.Message, err error) {

	content := *msg.Attachment.GetContent()
	if len(content) == 0 {
		err = fmt.Errorf("null or empty content")
		return
	}

	mediaType := GetMediaTypeFromWAMsgType(msg.Type)
	response, err := source.Client.Upload(context.Background(), content, mediaType)
	if err != nil {
		return
	}

	inreplycontext := source.GetInReplyContextInfo(msg)
	result = NewWhatsmeowMessageAttachment(response, msg, mediaType, inreplycontext)
	return
}

func (conn *WhatsmeowConnection) Disconnect() (err error) {
	if conn.Client != nil {
		if conn.Client.IsConnected() {
			conn.Client.Disconnect()
		}
	}
	return
}

//region PAIRING

func (source *WhatsmeowConnection) PairPhone(phone string) (string, error) {

	if !source.Client.IsConnected() {
		err := source.Client.Connect()
		if err != nil {
			log.Errorf("error on connecting for getting whatsapp qrcode: %s", err.Error())
			return "", err
		}
	}

	return source.Client.PairPhone(context.TODO(), phone, true, whatsmeow.PairClientChrome, "Chrome (Linux)")
}

func (conn *WhatsmeowConnection) GetWhatsAppQRCode() string {

	var result string

	// No ID stored, new login
	qrChan, err := conn.Client.GetQRChannel(context.Background())
	if err != nil {
		log.Errorf("error on getting whatsapp qrcode channel: %s", err.Error())
		return ""
	}

	if !conn.Client.IsConnected() {
		err = conn.Client.Connect()
		if err != nil {
			log.Errorf("error on connecting for getting whatsapp qrcode: %s", err.Error())
			return ""
		}
	}

	var wg sync.WaitGroup
	wg.Add(1)

	for evt := range qrChan {
		if evt.Event == "code" {
			result = evt.Code
		}

		wg.Done()

		// ending after the first the loop
		break
	}

	wg.Wait()
	return result
}

//endregion

func TryUpdateChannel(ch chan<- string, value string) (closed bool) {
	defer func() {
		if recover() != nil {
			// the return result can be altered
			// in a defer function call
			closed = false
		}
	}()

	ch <- value // panic if ch is closed
	return true // <=> closed = false; return
}

func (source *WhatsmeowConnection) GetWhatsAppQRChannel(ctx context.Context, out chan<- string) error {
	logger := source.GetLogger()

	// No ID stored, new login
	qrChan, err := source.Client.GetQRChannel(ctx)
	if err != nil {
		logger.Errorf("error on getting whatsapp qrcode channel: %s", err.Error())
		return err
	}

	if !source.Client.IsConnected() {
		err = source.Client.Connect()
		if err != nil {
			logger.Errorf("error on connecting for getting whatsapp qrcode: %s", err.Error())
			return err
		}
	}

	var wg sync.WaitGroup
	wg.Add(1)

	for evt := range qrChan {
		if evt.Event == "code" {
			if !TryUpdateChannel(out, evt.Code) {
				// expected error, means that websocket was closed
				// probably user has gone out page
				return fmt.Errorf("cant write to output")
			}
		} else {
			if evt.Event == "timeout" {
				return errors.New("timeout")
			}
			wg.Done()
			break
		}
	}

	wg.Wait()
	return nil
}

func (source *WhatsmeowConnection) HistorySync(timestamp time.Time) (err error) {
	logentry := source.GetLogger()

	leading := source.GetHandlers().WAHandlers.GetLeading()
	if leading == nil {
		err = fmt.Errorf("no valid msg in cache for retrieve parents")
		return err
	}

	// Convert interface to struct using type assertion
	info, ok := leading.InfoForHistory.(types.MessageInfo)
	if !ok {
		logentry.Error("error converting leading for history")
	}

	logentry.Infof("getting history from: %s", timestamp)
	extra := whatsmeow.SendRequestExtra{Peer: true}

	//info := &types.MessageInfo{ }
	msg := source.Client.BuildHistorySyncRequest(&info, 50)
	response, err := source.Client.SendMessage(context.Background(), source.Client.Store.ID.ToNonAD(), msg, extra)
	if err != nil {
		logentry.Errorf("getting history error: %s", err.Error())
	}

	logentry.Infof("history: %v", response)
	return
}

func (conn *WhatsmeowConnection) UpdateHandler(handlers whatsapp.IWhatsappHandlers) {
	conn.GetHandlers().WAHandlers = handlers
}

func (conn *WhatsmeowConnection) UpdatePairedCallBack(callback func(string)) {
	conn.paired = callback
}

func (conn *WhatsmeowConnection) PairedCallBack(jid types.JID, platform, businessName string) bool {
	if conn.paired != nil {
		go conn.paired(jid.String())
	}
	return true
}

//endregion

/*
<summary>

	Disconnect if connected
	Cleanup Handlers
	Dispose resources
	Does not erase permanent data !

</summary>
*/
func (source *WhatsmeowConnection) Dispose(reason string) {

	source.GetLogger().Infof("disposing connection: %s", reason)

	if source.Handlers != nil {
		go source.Handlers.UnRegister()
		source.Handlers = nil
	}

	if source.Client != nil {
		if source.Client.IsConnected() {
			go source.Client.Disconnect()
		}
		source.Client = nil
	}

	source = nil
}

/*
<summary>

	Erase permanent data + Dispose !

</summary>
*/
func (source *WhatsmeowConnection) Delete() (err error) {
	if source != nil {
		if source.Client != nil {
			if source.Client.IsLoggedIn() {
				err = source.Client.Logout(context.Background())
				if err != nil {
					err = fmt.Errorf("whatsmeow connection, delete logout error: %s", err.Error())
					return
				}
				source.GetLogger().Infof("logged out for delete")
			}

			if source.Client.Store != nil {
				err = source.Client.Store.Delete(context.Background())
				if err != nil {
					// ignoring error about JID, just checked and the delete process was succeed
					if strings.Contains(err.Error(), "device JID must be known before accessing database") {
						err = nil
					} else {
						err = fmt.Errorf("whatsmeow connection, delete store error: %s", err.Error())
						return
					}
				}
			}
		}
	}

	source.Dispose("Delete")
	return
}

func (conn *WhatsmeowConnection) IsInterfaceNil() bool {
	return nil == conn
}

// GetGroupManager returns the group manager instance with lazy initialization
func (conn *WhatsmeowConnection) GetGroupManager() whatsapp.WhatsappGroupManagerInterface {
	if conn.GroupManager == nil {
		conn.GroupManager = NewWhatsmeowGroupManager(conn)
	}
	return conn.GroupManager
}

// GetStatusManager returns the status manager instance with lazy initialization
func (conn *WhatsmeowConnection) GetStatusManager() whatsapp.WhatsappStatusManagerInterface {
	if conn.StatusManager == nil {
		conn.StatusManager = NewWhatsmeowStatusManager(conn)
	}
	return conn.StatusManager
}

// GetContactManager returns the contact manager instance with lazy initialization
func (conn *WhatsmeowConnection) GetContactManager() whatsapp.WhatsappContactManagerInterface {
	if conn.ContactManager == nil {
		conn.ContactManager = NewWhatsmeowContactManager(conn)
	}
	return conn.ContactManager
}

// GetHandlers returns the handlers instance with lazy initialization
func (conn *WhatsmeowConnection) GetHandlers() *WhatsmeowHandlers {
	if conn.Handlers == nil {
		conn.initializeHandlers(nil, WhatsmeowOptions{})
	}
	return conn.Handlers
}

// initializeHandlers creates and configures the handlers with proper options
func (conn *WhatsmeowConnection) initializeHandlers(whatsappOptions *whatsapp.WhatsappOptions, whatsmeowOptions WhatsmeowOptions) error {
	if conn.Handlers != nil {
		return nil // already initialized
	}

	conn.Handlers = NewWhatsmeowHandlers(conn, whatsmeowOptions, whatsappOptions)
	return conn.Handlers.Register()
}

// SendChatPresence updates typing status in a chat
func (conn *WhatsmeowConnection) SendChatPresence(chatId string, presenceType uint) error {
	if conn.Client == nil {
		return fmt.Errorf("client not defined")
	}

	jid, err := types.ParseJID(chatId)
	if err != nil {
		return fmt.Errorf("invalid chat id format: %v", err)
	}

	var state types.ChatPresence
	var media types.ChatPresenceMedia

	// Map our custom presence type to whatsmeow types
	switch whatsapp.WhatsappChatPresenceType(presenceType) {
	case whatsapp.WhatsappChatPresenceTypeText:
		state = types.ChatPresenceComposing // typing
		media = types.ChatPresenceMediaText
	case whatsapp.WhatsappChatPresenceTypeAudio:
		state = types.ChatPresenceComposing // typing
		media = types.ChatPresenceMediaAudio
	default:
		// Default is paused (stop typing)
		state = types.ChatPresencePaused
		media = types.ChatPresenceMediaText
	}

	return conn.Client.SendChatPresence(jid, state, media)
}
