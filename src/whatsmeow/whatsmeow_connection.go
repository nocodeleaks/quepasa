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

	"github.com/nocodeleaks/quepasa/library"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	whatsmeow "go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	types "go.mau.fi/whatsmeow/types"
)

// Must Implement IWhatsappConnection
type WhatsmeowConnection struct {
	library.LogStruct // logging
	Client            *whatsmeow.Client
	Handlers          *WhatsmeowHandlers

	failedToken  bool
	paired       func(string)
	IsConnecting bool `json:"isconnecting"` // used to avoid multiple connection attempts
}

//#region IMPLEMENT WHATSAPP CONNECTION OPTIONS INTERFACE

func (conn *WhatsmeowConnection) GetWid() string {
	if conn != nil {
		wid, err := conn.GetWidInternal()
		if err != nil {
			return wid
		}
	}

	return ""
}

// get default log entry, never nil
func (source *WhatsmeowConnection) GetLogger() *log.Entry {
	if source != nil && source.LogEntry != nil {
		return source.LogEntry
	}

	logentry := library.NewLogEntry(source)
	if source != nil {
		wid, _ := source.GetWidInternal()
		if len(wid) > 0 {
			logentry = logentry.WithField(LogFields.WId, wid)
		}
		source.LogEntry = logentry
	}

	logentry.Level = log.ErrorLevel
	logentry.Infof("generating new log entry for %s, with level: %s", reflect.TypeOf(source), logentry.Level)
	return logentry
}

func (conn *WhatsmeowConnection) SetReconnect(value bool) {
	if conn != nil {
		if conn.Client != nil {
			conn.Client.EnableAutoReconnect = value
		}
	}
}

func (conn *WhatsmeowConnection) GetReconnect() bool {
	if conn != nil {
		if conn.Client != nil {
			return conn.Client.EnableAutoReconnect
		}
	}

	return false
}

//#endregion

//region IMPLEMENT INTERFACE WHATSAPP CONNECTION

func (conn *WhatsmeowConnection) GetVersion() string { return "multi" }

func (conn *WhatsmeowConnection) GetWidInternal() (string, error) {
	if conn.Client == nil {
		err := fmt.Errorf("client not defined on trying to get wid")
		return "", err
	}

	if conn.Client.Store == nil {
		err := fmt.Errorf("device store not defined on trying to get wid")
		return "", err
	}

	if conn.Client.Store.ID == nil {
		err := fmt.Errorf("device id not defined on trying to get wid")
		return "", err
	}

	wid := conn.Client.Store.ID.User
	return wid, nil
}

func (conn *WhatsmeowConnection) IsValid() bool {
	if conn != nil {
		if conn.Client != nil {
			if conn.Client.IsConnected() {
				if conn.Client.IsLoggedIn() {
					return true
				}
			}
		}
	}
	return false
}

func (source *WhatsmeowConnection) IsConnected() bool {
	if source != nil {

		// manual checks for avoid thread locking
		if source.IsConnecting {
			return false
		}

		if source.Client != nil {
			if source.Client.IsConnected() {
				return true
			}
		}
	}
	return false
}

func (source *WhatsmeowConnection) GetStatus() whatsapp.WhatsappConnectionState {
	if source != nil {
		if source.Client == nil {
			return whatsapp.UnVerified
		} else {

			// manual checks for avoid thread locking
			if source.IsConnecting {
				return whatsapp.Connecting
			}

			// this is connected method locks the socket thread, so, if its in connecting state, it will be blocked here
			if source.Client.IsConnected() {
				if source.Client.IsLoggedIn() {
					return whatsapp.Ready
				} else {
					return whatsapp.Connected
				}
			} else {
				if source.failedToken {
					return whatsapp.Failed
				} else {
					return whatsapp.Disconnected
				}
			}
		}
	} else {
		return whatsapp.UnPrepared
	}
}

// returns a valid chat title from local memory store
func (conn *WhatsmeowConnection) GetChatTitle(wid string) string {
	jid, err := types.ParseJID(wid)
	if err == nil {
		return GetChatTitle(conn.Client, jid)
	}

	return ""
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

func (source *WhatsmeowConnection) GetContacts() (chats []whatsapp.WhatsappChat, err error) {
	if source.Client == nil {
		err = errors.New("invalid client")
		return chats, err
	}

	if source.Client.Store == nil {
		err = errors.New("invalid store")
		return chats, err
	}

	contacts, err := source.Client.Store.Contacts.GetAllContacts(context.TODO())
	if err != nil {
		return chats, err
	}

	// Map to track contacts by phone number
	contactMap := make(map[string]whatsapp.WhatsappChat)

	for jid, info := range contacts {
		title := info.FullName
		if len(title) == 0 {
			title = info.BusinessName
			if len(title) == 0 {
				title = info.PushName
			}
		}

		var phoneNumber string
		var lid string
		var phoneE164 string

		if strings.Contains(jid.String(), "@lid") {
			// For @lid contacts, get the corresponding phone number
			pnJID, err := source.Client.Store.LIDs.GetPNForLID(context.TODO(), jid)
			if err == nil && !pnJID.IsEmpty() {
				phoneNumber = pnJID.User
				lid = jid.String()
				// Format phone to E164
				if phone, err := library.ExtractPhoneIfValid(phoneNumber); err == nil {
					phoneE164 = phone
				}
			} else {
				// If no mapping found, use the LID as unique identifier
				phoneNumber = jid.String()
				lid = ""
			}
		} else {
			// For regular @s.whatsapp.net contacts
			phoneNumber = jid.User
			// Format phone to E164
			if phone, err := library.ExtractPhoneIfValid(phoneNumber); err == nil {
				phoneE164 = phone
			}

			// Try to get corresponding LID
			lidJID, err := source.Client.Store.LIDs.GetLIDForPN(context.TODO(), jid)
			if err == nil && !lidJID.IsEmpty() {
				lid = lidJID.String()
			} else {
				lid = ""
			}
		}

		// Check if contact with this phone number already exists
		existingContact, exists := contactMap[phoneNumber]

		if !exists {
			// First contact with this phone number
			contactMap[phoneNumber] = whatsapp.WhatsappChat{
				Id:    jid.String(),
				Lid:   lid,
				Title: title,
				Phone: phoneE164,
			}
		} else {
			// Contact already exists, merge information
			var finalId, finalLid, finalPhone string

			if strings.Contains(jid.String(), "@lid") {
				// Current is @lid, keep existing as Id and use current as Lid
				finalId = existingContact.Id
				finalLid = jid.String()
				finalPhone = existingContact.Phone
				if len(finalPhone) == 0 && len(phoneE164) > 0 {
					finalPhone = phoneE164
				}
			} else {
				// Current is @s.whatsapp.net, use as Id and keep existing Lid
				finalId = jid.String()
				finalLid = existingContact.Lid
				if len(finalLid) == 0 && len(lid) > 0 {
					finalLid = lid
				}
				finalPhone = phoneE164
				if len(finalPhone) == 0 && len(existingContact.Phone) > 0 {
					finalPhone = existingContact.Phone
				}
			}

			// Keep the best available title
			finalTitle := title
			if len(finalTitle) == 0 && len(existingContact.Title) > 0 {
				finalTitle = existingContact.Title
			}

			contactMap[phoneNumber] = whatsapp.WhatsappChat{
				Id:    finalId,
				Lid:   finalLid,
				Title: finalTitle,
				Phone: finalPhone,
			}
		}
	}

	// Convert map to slice
	for _, contact := range contactMap {
		chats = append(chats, contact)
	}

	return chats, nil
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
	if !ok {
		attach := imsg.GetAttachment()
		if attach != nil {
			data := attach.GetContent()
			if data != nil {
				logentry.Trace("no waMsg, found attachment, returning content")
				return *data, err
			}
		}

		err = fmt.Errorf("parameter msg cannot be converted to an original message")
		return
	}

	downloadable = GetDownloadableMessage(waMsg)
	if downloadable != nil {
		logentry.Trace("waMsg implements DownloadableMessage, using Client.Download()")
		return source.Client.Download(context.Background(), downloadable)
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

// func (cli *Client) BuildEdit(chat types.JID, id types.MessageID, newContent *waE2E.Message) *waE2E.Message {
func (source *WhatsmeowConnection) Edit(msg whatsapp.IWhatsappMessage, newContent string) error {
	logentry := source.GetLogger()

	jid, err := types.ParseJID(msg.GetChatId())
	if err != nil {
		logentry.Infof("edit message error on get jid: %s", err)
		return err
	}

	// Create a new message with the edited content
	editedMessage := &waE2E.Message{
		Conversation: proto.String(newContent),
	}

	// Build the edit message using the new content
	newMessage := source.Client.BuildEdit(jid, msg.GetId(), editedMessage)
	_, err = source.Client.SendMessage(context.Background(), jid, newMessage)
	if err != nil {
		logentry.Infof("edit message error: %s", err)
		return err
	}

	return nil
}

func (conn *WhatsmeowConnection) IsOnWhatsApp(phones ...string) (registered []string, err error) {
	results, err := conn.Client.IsOnWhatsApp(phones)
	if err != nil {
		return
	}

	for _, result := range results {
		if result.IsIn {
			registered = append(registered, result.JID.String())
		}
	}

	return
}

func (conn *WhatsmeowConnection) GetProfilePicture(wid string, knowingId string) (picture *whatsapp.WhatsappProfilePicture, err error) {
	jid, err := types.ParseJID(wid)
	if err != nil {
		return
	}

	params := &whatsmeow.GetProfilePictureParams{}
	params.ExistingID = knowingId
	params.Preview = false

	pictureInfo, err := conn.Client.GetProfilePictureInfo(jid, params)
	if err != nil {
		return
	}

	if pictureInfo != nil {
		picture = &whatsapp.WhatsappProfilePicture{
			Id:   pictureInfo.ID,
			Type: pictureInfo.Type,
			Url:  pictureInfo.URL,
		}
	}
	return
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

	cached, _ := source.Handlers.WAHandlers.GetById(msg.InReply)
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
	if source.Handlers.ReadUpdate {
		go source.Handlers.MarkRead(msg, types.ReceiptTypeRead)
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

func (source *WhatsmeowConnection) GetInvite(groupId string) (link string, err error) {
	jid, err := types.ParseJID(groupId)
	if err != nil {
		source.GetLogger().Infof("getting invite error on parse jid: %s", err)
	}

	link, err = source.Client.GetGroupInviteLink(jid, false)
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

	leading := source.Handlers.WAHandlers.GetLeading()
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
	conn.Handlers.WAHandlers = handlers
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
				err = source.Client.Logout(context.TODO())
				if err != nil {
					err = fmt.Errorf("whatsmeow connection, delete logout error: %s", err.Error())
					return
				}
				source.GetLogger().Infof("logged out for delete")
			}

			if source.Client.Store != nil {
				err = source.Client.Store.Delete(context.TODO())
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
func (conn *WhatsmeowConnection) GetJoinedGroups() ([]interface{}, error) {
	if conn.Client == nil {
		return nil, fmt.Errorf("client not defined")
	}

	// Get the group info slice
	groupInfos, err := conn.Client.GetJoinedGroups()
	if err != nil {
		return nil, err
	}

	// Iterate over groupInfos and set the DisplayName for each participant
	for _, groupInfo := range groupInfos {
		if groupInfo.Participants != nil {
			for i, participant := range groupInfo.Participants {
				// Get the contact info from the store
				contact, err := conn.Client.Store.Contacts.GetContact(context.TODO(), participant.JID)
				if err != nil {
					// If no contact info is found, fallback to JID user part
					groupInfo.Participants[i].DisplayName = participant.JID.User
				} else {
					// Set the DisplayName field to the contact's full name or push name
					if len(contact.FullName) > 0 {
						groupInfo.Participants[i].DisplayName = contact.FullName
					} else if len(contact.PushName) > 0 {
						groupInfo.Participants[i].DisplayName = contact.PushName
					} else {
						groupInfo.Participants[i].DisplayName = "" // Fallback to JID user part
					}
				}
			}
		} else {

			// If Participants is nil, initialize it to an empty slice
			groupInfo.Participants = []types.GroupParticipant{}
			// You might want to log this or handle it differently
			conn.GetLogger().Warnf("Group %s has nil Participants, initializing to empty slice", groupInfo.JID.String())
		}
	}

	groups := make([]interface{}, len(groupInfos))
	for i, group := range groupInfos {
		groups[i] = group
	}

	return groups, nil
}

func (conn *WhatsmeowConnection) GetGroupInfo(groupId string) (interface{}, error) {
	if conn.Client == nil {
		return nil, fmt.Errorf("client not defined")
	}

	jid, err := types.ParseJID(groupId)
	if err != nil {
		return nil, err
	}

	groupInfo, err := conn.Client.GetGroupInfo(jid)
	if err != nil {
		return nil, err
	}

	// Fill contact names for participants
	if groupInfo.Participants != nil {
		for i, participant := range groupInfo.Participants {
			// Get the contact info from the store
			contact, err := conn.Client.Store.Contacts.GetContact(context.TODO(), participant.JID)
			if err != nil {
				// If no contact info is found, fallback to JID user part
				groupInfo.Participants[i].DisplayName = participant.JID.User
			} else {
				// Set the DisplayName field to the contact's full name or push name
				if len(contact.FullName) > 0 {
					groupInfo.Participants[i].DisplayName = contact.FullName
				} else if len(contact.PushName) > 0 {
					groupInfo.Participants[i].DisplayName = contact.PushName
				} else {
					groupInfo.Participants[i].DisplayName = "" // Fallback to JID user part
				}
			}
		}
	} else {
		// If Participants is nil, initialize it to an empty slice
		groupInfo.Participants = []types.GroupParticipant{}
		conn.GetLogger().Warnf("Group %s has nil Participants, initializing to empty slice", groupInfo.JID.String())
	}

	return groupInfo, nil
}

func (conn *WhatsmeowConnection) CreateGroup(name string, participants []string) (interface{}, error) {
	if conn.Client == nil {
		return nil, fmt.Errorf("client not defined")
	}

	// Convert participants to JID format
	var participantsJID []types.JID
	for _, participant := range participants {
		// Check if it's already in JID format
		if strings.Contains(participant, "@") {
			jid, err := types.ParseJID(participant)
			if err != nil {
				return nil, fmt.Errorf("invalid JID format for participant %s: %v", participant, err)
			}
			participantsJID = append(participantsJID, jid)
		} else {
			// Assume it's a phone number and convert to JID
			jid := types.JID{
				User:   participant,
				Server: "s.whatsapp.net", // Use the standard WhatsApp server
			}
			participantsJID = append(participantsJID, jid)
		}
	}

	// Create the request struct
	groupConfig := whatsmeow.ReqCreateGroup{
		Name:         name,
		Participants: participantsJID,
	}

	// Call the existing method with the constructed request
	return conn.Client.CreateGroup(groupConfig)
}

func (conn *WhatsmeowConnection) LeaveGroup(groupID string) error {
	if conn.Client == nil {
		return fmt.Errorf("client not defined")
	}

	// Parse the group ID to JID format
	jid, err := types.ParseJID(groupID)
	if err != nil {
		return fmt.Errorf("invalid group JID format: %v", err)
	}

	// Leave the group using whatsmeow client
	err = conn.Client.LeaveGroup(jid)
	if err != nil {
		return fmt.Errorf("failed to leave group: %v", err)
	}

	return nil
}

func (conn *WhatsmeowConnection) UpdateGroupSubject(groupID string, name string) (interface{}, error) {
	if conn.Client == nil {
		return nil, fmt.Errorf("client not defined")
	}

	// Parse the group ID to JID format
	jid, err := types.ParseJID(groupID)
	if err != nil {
		return nil, fmt.Errorf("invalid group JID format: %v", err)
	}

	// Update the group subject
	err = conn.Client.SetGroupName(jid, name)
	if err != nil {
		return nil, fmt.Errorf("failed to update group subject: %v", err)
	}

	// Return the updated group info
	return conn.Client.GetGroupInfo(jid)
}

func (conn *WhatsmeowConnection) UpdateGroupPhoto(groupID string, imageData []byte) (string, error) {
	if conn.Client == nil {
		return "", fmt.Errorf("client not defined")
	}

	// Parse the group ID to JID format
	jid, err := types.ParseJID(groupID)
	if err != nil {
		return "", fmt.Errorf("invalid group JID format: %v", err)
	}

	// Update the group photo
	pictureID, err := conn.Client.SetGroupPhoto(jid, imageData)
	if err != nil {
		return "", fmt.Errorf("failed to update group photo: %v", err)
	}

	return pictureID, nil
}

func (conn *WhatsmeowConnection) UpdateGroupTopic(groupID string, topic string) (interface{}, error) {
	if conn.Client == nil {
		return nil, fmt.Errorf("client not defined")
	}

	// Parse the group ID to JID format
	jid, err := types.ParseJID(groupID)
	if err != nil {
		return nil, fmt.Errorf("invalid group JID format: %v", err)
	}

	// Update the group topic (description)
	// SetGroupTopic requires: jid, previousID, newID, topic
	// Let the whatsmeow library handle previousID and newID automatically by passing empty strings
	err = conn.Client.SetGroupTopic(jid, "", "", topic)
	if err != nil {
		return nil, fmt.Errorf("failed to update group topic: %v", err)
	}

	// Return the updated group info
	return conn.Client.GetGroupInfo(jid)
}
func (conn *WhatsmeowConnection) UpdateGroupParticipants(groupJID string, participants []string, action string) ([]interface{}, error) {
	if conn.Client == nil {
		return nil, fmt.Errorf("client not defined")
	}

	// Parse the group JID
	jid, err := types.ParseJID(groupJID)
	if err != nil {
		return nil, fmt.Errorf("invalid group JID format: %v", err)
	}

	// Convert participant strings to JIDs
	participantJIDs := make([]types.JID, len(participants))
	for i, participant := range participants {
		participantJIDs[i], err = types.ParseJID(participant)
		if err != nil {
			return nil, fmt.Errorf("invalid participant JID format for %s: %v", participant, err)
		}
	}

	// Map the action string to the ParticipantChange type
	var participantAction whatsmeow.ParticipantChange
	switch action {
	case "add":
		participantAction = whatsmeow.ParticipantChangeAdd
	case "remove":
		participantAction = whatsmeow.ParticipantChangeRemove
	case "promote":
		participantAction = whatsmeow.ParticipantChangePromote
	case "demote":
		participantAction = whatsmeow.ParticipantChangeDemote
	default:
		return nil, fmt.Errorf("invalid action %s", action)
	}

	// Call the whatsmeow method
	result, err := conn.Client.UpdateGroupParticipants(jid, participantJIDs, participantAction)
	if err != nil {
		return nil, fmt.Errorf("failed to update group participants: %v", err)
	}

	// Convert to interface array for the generic return type
	interfaceResults := make([]interface{}, len(result))
	for i, r := range result {
		interfaceResults[i] = r
	}

	return interfaceResults, nil
}

func (conn *WhatsmeowConnection) GetGroupJoinRequests(groupJID string) ([]interface{}, error) {
	if conn.Client == nil {
		return nil, fmt.Errorf("client not defined")
	}

	// Parse the group JID
	jid, err := types.ParseJID(groupJID)
	if err != nil {
		return nil, fmt.Errorf("invalid group JID format: %v", err)
	}

	// Call the whatsmeow method
	requests, err := conn.Client.GetGroupRequestParticipants(jid)
	if err != nil {
		return nil, fmt.Errorf("failed to get group join requests: %v", err)
	}

	// Convert to interface array for the generic return type
	interfaceResults := make([]interface{}, len(requests))
	for i, r := range requests {
		interfaceResults[i] = r
	}

	return interfaceResults, nil
}

func (conn *WhatsmeowConnection) HandleGroupJoinRequests(groupJID string, participants []string, action string) ([]interface{}, error) {
	if conn.Client == nil {
		return nil, fmt.Errorf("client not defined")
	}

	// Parse the group JID
	jid, err := types.ParseJID(groupJID)
	if err != nil {
		return nil, fmt.Errorf("invalid group JID format: %v", err)
	}

	// Convert participant strings to JIDs
	participantJIDs := make([]types.JID, len(participants))
	for i, participant := range participants {
		participantJIDs[i], err = types.ParseJID(participant)
		if err != nil {
			return nil, fmt.Errorf("invalid participant JID format for %s: %v", participant, err)
		}
	}

	// Map the action string to the ParticipantRequestChange type
	var requestAction whatsmeow.ParticipantRequestChange
	switch action {
	case "approve":
		requestAction = whatsmeow.ParticipantChangeApprove
	case "reject":
		requestAction = whatsmeow.ParticipantChangeReject
	default:
		return nil, fmt.Errorf("invalid action %s", action)
	}

	// Call the correct WhatsApp method which returns participant results
	result, err := conn.Client.UpdateGroupRequestParticipants(jid, participantJIDs, requestAction)
	if err != nil {
		return nil, fmt.Errorf("failed to handle group join requests: %v", err)
	}

	// Convert the typed results to interface array
	interfaceResults := make([]interface{}, len(result))
	for i, r := range result {
		interfaceResults[i] = r
	}

	return interfaceResults, nil
}

func (conn *WhatsmeowConnection) CreateGroupExtended(title string, participants []string) (interface{}, error) {
	if conn.Client == nil {
		return nil, fmt.Errorf("client not defined")
	}

	// Convert participants to JIDs
	participantJIDs := make([]types.JID, len(participants))
	for i, participant := range participants {
		jid, err := types.ParseJID(participant)
		if err != nil {
			return nil, fmt.Errorf("invalid participant JID: %v", err)
		}
		participantJIDs[i] = jid
	}

	// Create request structure
	req := whatsmeow.ReqCreateGroup{
		Name:         title,
		Participants: participantJIDs,
	}

	// Call the WhatsApp method
	return conn.Client.CreateGroup(req)
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

// GetLIDFromPhone returns the @lid for a given phone number
func (conn *WhatsmeowConnection) GetLIDFromPhone(phone string) (string, error) {
	if conn.Client == nil {
		return "", fmt.Errorf("client not defined")
	}

	if conn.Client.Store == nil {
		return "", fmt.Errorf("store not defined")
	}

	// Parse the phone number to JID format
	phoneJID := types.JID{
		User:   phone,
		Server: "s.whatsapp.net",
	}

	// try to get the LID from local store
	lidJID, err := conn.Client.Store.LIDs.GetLIDForPN(context.TODO(), phoneJID)
	if err == nil && !lidJID.IsEmpty() {
		conn.GetLogger().Debugf("LID found in local store for phone %s: %s", phone, lidJID.String())
		return lidJID.String(), nil
	}
	return "", fmt.Errorf("no LID found for phone %s", phone)
}

// GetPhoneFromLID returns the phone number for a given @lid
func (conn *WhatsmeowConnection) GetPhoneFromLID(lid string) (string, error) {
	if conn.Client == nil {
		return "", fmt.Errorf("client not defined")
	}

	if conn.Client.Store == nil {
		return "", fmt.Errorf("store not defined")
	}

	// Parse the LID to JID format
	lidJID, err := types.ParseJID(lid)
	if err != nil {
		return "", fmt.Errorf("invalid LID format: %v", err)
	}

	// Get the corresponding phone number from local store
	phoneJID, err := conn.Client.Store.LIDs.GetPNForLID(context.TODO(), lidJID)
	if err != nil {
		return "", fmt.Errorf("failed to get phone for LID %s: %v", lid, err)
	}

	if phoneJID.IsEmpty() {
		return "", fmt.Errorf("no phone found for LID %s", lid)
	}

	conn.GetLogger().Debugf("Phone found in local store for LID %s: %s", lid, phoneJID.User)
	return phoneJID.User, nil
}

// UserInfoResponse represents the structured response for user information
type UserInfoResponse struct {
	JID          string              `json:"jid"`
	LID          string              `json:"lid,omitempty"`
	Phone        string              `json:"phone,omitempty"`
	PhoneE164    string              `json:"phoneE164,omitempty"`
	Status       string              `json:"status,omitempty"`
	PictureID    string              `json:"pictureId,omitempty"`
	Devices      []types.JID         `json:"devices,omitempty"`
	VerifiedName *types.VerifiedName `json:"verifiedName,omitempty"`
	DisplayName  string              `json:"displayName,omitempty"`
	FullName     string              `json:"fullName,omitempty"`
	BusinessName string              `json:"businessName,omitempty"`
	PushName     string              `json:"pushName,omitempty"`
}

// GetUserInfo retrieves comprehensive user information for given JIDs
func (conn *WhatsmeowConnection) GetUserInfo(jids []string) ([]interface{}, error) {
	if conn.Client == nil {
		return nil, fmt.Errorf("client not defined")
	}

	if conn.Client.Store == nil {
		return nil, fmt.Errorf("store not defined")
	}

	// Convert string JIDs to types.JID
	var parsedJIDs []types.JID
	for _, jidStr := range jids {
		// Check if it's a phone number (no @ symbol) and validate E164 format
		if !strings.Contains(jidStr, "@") {
			// This is a phone number, validate and format to E164
			validPhone, err := library.ExtractPhoneIfValid(jidStr)
			if err != nil {
				return nil, fmt.Errorf("invalid phone number format for %s: %v (must be E164 format starting with +)", jidStr, err)
			}

			// Remove the + from E164 format for JID creation
			phoneNumber := strings.TrimPrefix(validPhone, "+")
			jid := types.JID{
				User:   phoneNumber,
				Server: "s.whatsapp.net",
			}
			parsedJIDs = append(parsedJIDs, jid)
		} else {
			// This is already a JID, parse normally
			jid, err := types.ParseJID(jidStr)
			if err != nil {
				return nil, fmt.Errorf("invalid JID format for %s: %v", jidStr, err)
			}
			parsedJIDs = append(parsedJIDs, jid)
		}
	}

	// Get user info from WhatsApp - this returns a map[types.JID]types.UserInfo
	userInfoMap, err := conn.Client.GetUserInfo(parsedJIDs)
	logentry := conn.GetLogger()
	logentry.Debugf("GetUserInfo for JIDs: %v, result: %v", parsedJIDs, userInfoMap)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %v", err)
	}

	// Convert map to interface array for generic return type
	result := make([]interface{}, 0, len(userInfoMap))
	for jid, info := range userInfoMap {
		// Get contact info from local store - try both JID and corresponding phone/LID
		contactInfo, contactErr := conn.Client.Store.Contacts.GetContact(context.TODO(), jid)

		// Get LID/Phone mapping information
		var lid, phoneNumber string
		var phoneJID types.JID

		if strings.Contains(jid.String(), "@lid") {
			// This is a LID, try to get corresponding phone
			lid = jid.String()
			pnJID, err := conn.Client.Store.LIDs.GetPNForLID(context.TODO(), jid)
			if err == nil && !pnJID.IsEmpty() {
				phoneNumber = pnJID.User
				phoneJID = pnJID

				// If we didn't get contact info from LID, try with phone JID
				if contactErr != nil {
					contactInfo, contactErr = conn.Client.Store.Contacts.GetContact(context.TODO(), phoneJID)
				}
			}
		} else {
			// This is a phone number JID, try to get corresponding LID
			phoneNumber = jid.User
			lidJID, err := conn.Client.Store.LIDs.GetLIDForPN(context.TODO(), jid)
			if err == nil && !lidJID.IsEmpty() {
				lid = lidJID.String()

				// If we didn't get contact info from phone JID, try with LID
				if contactErr != nil {
					contactInfo, contactErr = conn.Client.Store.Contacts.GetContact(context.TODO(), lidJID)
				}
			}
		}

		// Format phone to E164 if available
		var phoneE164 string
		if phoneNumber != "" {
			if phone, err := library.ExtractPhoneIfValid(phoneNumber); err == nil {
				phoneE164 = phone
			}
		}

		// Determine the best display name
		var displayName string
		if contactErr == nil {
			if contactInfo.FullName != "" {
				displayName = contactInfo.FullName
			} else if contactInfo.BusinessName != "" {
				displayName = contactInfo.BusinessName
			} else if contactInfo.PushName != "" {
				displayName = contactInfo.PushName
			}
		}

		// If no local contact name, use verified name from user info
		if displayName == "" && info.VerifiedName != nil {
			displayName = info.VerifiedName.Details.GetVerifiedName()
		}

		// Check if we have meaningful contact information
		hasContactInfo := contactErr == nil && (contactInfo.FullName != "" || contactInfo.BusinessName != "" || contactInfo.PushName != "")
		hasVerifiedName := info.VerifiedName != nil && info.VerifiedName.Details.GetVerifiedName() != ""
		hasStatus := info.Status != ""
		hasPictureID := info.PictureID != ""
		hasDevices := len(info.Devices) > 0
		hasLID := lid != ""

		// Only include contacts that have meaningful information beyond just phone/JID
		if !hasContactInfo && !hasVerifiedName && !hasStatus && !hasPictureID && !hasDevices && !hasLID {
			logentry.Debugf("Skipping contact %s - no meaningful information possible non whatsapp number", jid.String())
			continue
		}

		// Create a comprehensive response with omitempty support
		userInfoResponse := UserInfoResponse{
			JID:          jid.String(),
			LID:          lid,
			Phone:        phoneNumber,
			PhoneE164:    phoneE164,
			Status:       info.Status,
			PictureID:    info.PictureID,
			Devices:      info.Devices,
			VerifiedName: info.VerifiedName,
			DisplayName:  displayName,
		}

		// Add contact-specific information if available
		if contactErr == nil {
			userInfoResponse.FullName = contactInfo.FullName
			userInfoResponse.BusinessName = contactInfo.BusinessName
			userInfoResponse.PushName = contactInfo.PushName
		}

		result = append(result, userInfoResponse)
	}

	return result, nil
}
