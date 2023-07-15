package whatsmeow

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
	"unicode"

	log "github.com/sirupsen/logrus"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	whatsmeow "go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	types "go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// Must Implement IWhatsappConnection
type WhatsmeowConnection struct {
	Client      *whatsmeow.Client
	Handlers    *WhatsmeowHandlers
	waLogger    waLog.Logger
	log         *log.Entry
	failedToken bool
	paired      func(string)
}

//region IMPLEMENT INTERFACE WHATSAPP CONNECTION

func (conn *WhatsmeowConnection) GetVersion() string { return "multi" }

func (conn *WhatsmeowConnection) GetWid() (wid string, err error) {
	if conn.Client == nil {
		err = fmt.Errorf("client not defined on trying to get wid")
	} else {
		if conn.Client.Store == nil {
			err = fmt.Errorf("device store not defined on trying to get wid")
		} else {
			if conn.Client.Store.ID == nil {
				err = fmt.Errorf("device id not defined on trying to get wid")
			} else {
				wid = conn.Client.Store.ID.User
			}
		}
	}

	return
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

func (conn *WhatsmeowConnection) IsConnected() bool {
	if conn != nil {
		if conn.Client != nil {
			if conn.Client.IsConnected() {
				return true
			}
		}
	}
	return false
}

func (conn *WhatsmeowConnection) GetStatus() whatsapp.WhatsappConnectionState {
	if conn != nil {
		if conn.Client == nil {
			return whatsapp.UnVerified
		} else {
			if conn.Client.IsConnected() {
				if conn.Client.IsLoggedIn() {
					return whatsapp.Ready
				} else {
					return whatsapp.Connected
				}
			} else {
				if conn.failedToken {
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
func (conn *WhatsmeowConnection) Connect() (err error) {
	conn.log.Info("starting whatsmeow connection")

	err = conn.Client.Connect()
	if err != nil {
		conn.failedToken = true
		return
	}

	// waits 2 seconds for loggedin
	// not required
	_ = conn.Client.WaitForConnection(time.Millisecond * 2000)

	conn.failedToken = false
	return
}

// func (cli *Client) Download(msg DownloadableMessage) (data []byte, err error)
func (conn *WhatsmeowConnection) DownloadData(imsg whatsapp.IWhatsappMessage) (data []byte, err error) {
	msg := imsg.GetSource()
	downloadable, ok := msg.(whatsmeow.DownloadableMessage)
	if !ok {
		conn.log.Debug("not downloadable type, trying default message")
		waMsg, ok := msg.(*waProto.Message)
		if !ok {
			attach := imsg.GetAttachment()
			if attach != nil {
				data := attach.GetContent()
				if data != nil {
					return *data, err
				}
			}

			err = fmt.Errorf("parameter msg cannot be converted to an original message")
			return
		}
		return conn.Client.DownloadAny(waMsg)
	}
	return conn.Client.Download(downloadable)
}

func (conn *WhatsmeowConnection) Download(imsg whatsapp.IWhatsappMessage, cache bool) (att *whatsapp.WhatsappAttachment, err error) {
	att = imsg.GetAttachment()
	if att == nil {
		err = fmt.Errorf("message (%s) does not contains attachment info", imsg.GetId())
		return
	}

	if !att.HasContent() || !cache {
		data, err := conn.DownloadData(imsg)
		if err != nil {
			return att, err
		}

		if !cache {
			newAtt := *att
			att = &newAtt
		}

		att.SetContent(&data)
	}

	return
}

func (conn *WhatsmeowConnection) Revoke(msg whatsapp.IWhatsappMessage) error {
	jid, err := types.ParseJID(msg.GetChatId())
	if err != nil {
		conn.log.Infof("revoke error on get jid: %s", err)
		return err
	}

	participantJid, err := types.ParseJID(msg.GetParticipantId())
	if err != nil {
		conn.log.Infof("revoke error on get jid: %s", err)
		return err
	}

	newMessage := conn.Client.BuildRevoke(jid, participantJid, msg.GetId())
	_, err = conn.Client.SendMessage(context.Background(), jid, newMessage)
	if err != nil {
		conn.log.Infof("revoke error: %s", err)
		return err
	}

	return nil
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

// Default SEND method using WhatsappMessage Interface
func (conn *WhatsmeowConnection) Send(msg *whatsapp.WhatsappMessage) (whatsapp.IWhatsappSendResponse, error) {

	var err error
	messageText := msg.GetText()

	var newMessage *waProto.Message
	if !msg.HasAttachment() {
		if IsValidForButtons(messageText) {
			internal := GenerateButtonsMessage(messageText)
			newMessage = &waProto.Message{ButtonsMessage: internal}
		} else {
			internal := &waProto.ExtendedTextMessage{Text: &messageText}
			newMessage = &waProto.Message{ExtendedTextMessage: internal}
		}
	} else {
		newMessage, err = conn.UploadAttachment(*msg)
		if err != nil {
			return msg, err
		}
	}

	// Formatting destination accordly
	formatedDestination, _ := whatsapp.FormatEndpoint(msg.GetChatId())

	// Avoid common issue with incorrect non ascii chat id
	if !isASCII(formatedDestination) {
		err = fmt.Errorf("not an ASCII formated chat id")
		return msg, err
	}

	jid, err := types.ParseJID(formatedDestination)
	if err != nil {
		conn.log.Infof("send error on get jid: %s", err)
		return msg, err
	}

	// Generating a new unique MessageID
	if len(msg.Id) == 0 {
		msg.Id = whatsmeow.GenerateMessageID()
	}

	extra := whatsmeow.SendRequestExtra{
		ID: msg.Id,
	}

	resp, err := conn.Client.SendMessage(context.Background(), jid, newMessage, extra)
	if err != nil {
		conn.log.Infof("send error: %s", err)
		return msg, err
	}
	msg.Timestamp = resp.Timestamp

	conn.log.Infof("send: %s, on: %s", msg.Id, msg.Timestamp)
	return msg, err
}

// func (cli *Client) Upload(ctx context.Context, plaintext []byte, appInfo MediaType) (resp UploadResponse, err error)
func (conn *WhatsmeowConnection) UploadAttachment(msg whatsapp.WhatsappMessage) (result *waProto.Message, err error) {

	content := *msg.Attachment.GetContent()
	if len(content) == 0 {
		err = fmt.Errorf("null or empty content")
		return
	}

	mediaType := GetMediaTypeFromWAMsgType(msg.Type)
	response, err := conn.Client.Upload(context.Background(), content, mediaType)
	if err != nil {
		return
	}

	result = NewWhatsmeowMessageAttachment(response, msg, mediaType)
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

func (conn *WhatsmeowConnection) GetInvite(groupId string) (link string, err error) {
	jid, err := types.ParseJID(groupId)
	if err != nil {
		conn.log.Infof("getting invite error on parse jid: %s", err)
	}

	link, err = conn.Client.GetGroupInviteLink(jid, false)
	return
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
		break
	}

	wg.Wait()
	return result
}

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

func (conn *WhatsmeowConnection) GetWhatsAppQRChannel(ctx context.Context, out chan<- string) (err error) {
	// No ID stored, new login
	qrChan, err := conn.Client.GetQRChannel(ctx)
	if err != nil {
		log.Errorf("error on getting whatsapp qrcode channel: %s", err.Error())
		return
	}

	if !conn.Client.IsConnected() {
		err = conn.Client.Connect()
		if err != nil {
			log.Errorf("error on connecting for getting whatsapp qrcode: %s", err.Error())
			return
		}
	}

	var wg sync.WaitGroup
	wg.Add(1)

	for evt := range qrChan {
		if evt.Event == "code" {
			if !TryUpdateChannel(out, evt.Code) {
				// expected error, means that websocket was closed
				// probably user has gone out page
				err = fmt.Errorf("cant write to output")
				break
			}
		} else {
			if evt.Event == "timeout" {
				err = errors.New("timeout")
			}
			wg.Done()
			break
		}
	}

	wg.Wait()
	return
}

func (conn *WhatsmeowConnection) UpdateLog(entry *log.Entry) {
	conn.log = entry
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
func (conn *WhatsmeowConnection) Dispose(reason string) {
	if conn.log != nil {
		conn.log.Infof("disposing connection: %s", reason)
		conn.log = nil
	}

	if conn.log != nil {
		conn.log = nil
	}

	if conn.Handlers != nil {
		go conn.Handlers.UnRegister()
		conn.Handlers = nil
	}

	if conn.Client != nil {
		if conn.Client.IsConnected() {
			go conn.Client.Disconnect()
		}
		conn.Client = nil
	}

	conn = nil
}

/*
<summary>

	Erase permanent data + Dispose !

</summary>
*/
func (conn *WhatsmeowConnection) Delete() (err error) {
	if conn != nil {
		if conn.Client != nil {
			if conn.Client.IsLoggedIn() {
				err = conn.Client.Logout()
				if err != nil {
					return
				}
				conn.log.Infof("logged out for delete")
			}

			if conn.Client.Store != nil {
				err = conn.Client.Store.Delete()
				if err != nil {
					// ignoring error about JID, just checked and the delete process was succeded
					if strings.Contains(err.Error(), "device JID must be known before accessing database") {
						err = nil
					} else {
						err = fmt.Errorf("error on trying to delete store: %s", err.Error())
						return
					}
				}

				// here the conn.log (*Entry) is already nil
			}
		}
	}

	conn.Dispose("Delete")
	return
}

func (conn *WhatsmeowConnection) IsInterfaceNil() bool {
	return nil == conn
}
