package models

import (
	"context"
	"encoding/base64"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/skip2/go-qrcode"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

func NewEmptyConnection(callback func(string)) (conn whatsapp.IWhatsappConnection, err error) {
	return NewWhatsmeowEmptyConnection(callback)
}

func NewConnection(options *whatsapp.WhatsappConnectionOptions) (whatsapp.IWhatsappConnection, error) {
	return NewWhatsmeowConnection(options)
}

func TryUpdateHttpChannel(ch chan<- []byte, value []byte) (closed bool) {
	defer func() {
		if recover() != nil {
			closed = false
		}
	}()

	ch <- value
	return true
}

// SignInWithQRCode streams QR code images to the caller while the pairing flow
// waits for the device to be linked.
func SignInWithQRCode(ctx context.Context, pairing QpWhatsappPairing, out chan<- []byte) (err error) {
	con, err := pairing.GetConnection()
	if err != nil {
		return
	}

	logger := con.GetLogger()
	qrChan := make(chan string)
	defer close(qrChan)

	go func() {
		for qrBase64 := range qrChan {
			png, encodeErr := qrcode.Encode(qrBase64, qrcode.Medium, 256)
			if encodeErr != nil {
				logger.Errorf("(qrcode) encode fail, %s", encodeErr.Error())
				return
			}

			encodedPNG := base64.StdEncoding.EncodeToString(png)
			if !TryUpdateHttpChannel(out, []byte(encodedPNG)) {
				logger.Error("(qrcode) cant write to output")
				return
			}
		}
	}()

	logger.Info("(qrcode) getting qrcode channel ...")
	return con.GetWhatsAppQRChannel(ctx, qrChan)
}

func EnsureServerOnCache(currentUserID string, wid string, connection whatsapp.IWhatsappConnection) (err error) {
	server, err := WhatsappService.GetOrCreateServer(currentUserID, wid)
	if err != nil {
		log.Errorf("getting or create server after login : %s", err.Error())
		return
	}

	server.MarkVerified(true)
	go server.UpdateConnection(connection)
	return
}

// GetDownloadPrefixFromToken resolves the internal attachment download prefix
// for a live server token so transport layers can build stable download URLs.
func GetDownloadPrefixFromToken(token string) (path string, err error) {
	server, ok := WhatsappService.Servers[token]
	if !ok {
		err = fmt.Errorf("server not found: %s", token)
		return
	}

	prefix := fmt.Sprintf("/bot/%s/download", server.Token)
	return prefix, err
}

func ToWhatsappMessage(destination string, text string, attach *whatsapp.WhatsappAttachment) (msg *whatsapp.WhatsappMessage, err error) {
	recipient, err := whatsapp.FormatEndpoint(destination)
	if err != nil {
		return
	}

	msg = &whatsapp.WhatsappMessage{}
	msg.FromInternal = true
	msg.FromMe = true
	msg.Type = whatsapp.TextMessageType
	msg.Text = text
	msg.Chat = whatsapp.WhatsappChat{Id: recipient}

	if attach != nil {
		msg.Attachment = attach
		msg.Type = whatsapp.GetMessageType(attach)
	}

	return
}

// ToggleReadReceipts cycles the persisted read-receipt handling mode.
func ToggleReadReceipts(source whatsapp.IWhatsappOptions) error {
	options := source.GetOptions()

	switch options.ReadReceipts {
	case whatsapp.UnSetBooleanType:
		options.ReadReceipts = whatsapp.TrueBooleanType
	case whatsapp.TrueBooleanType:
		options.ReadReceipts = whatsapp.FalseBooleanType
	default:
		options.ReadReceipts = whatsapp.UnSetBooleanType
	}

	reason := fmt.Sprintf("toggle read receipts: %s", options.ReadReceipts)
	return source.Save(reason)
}

// ToggleGroups cycles the persisted group-handling mode.
func ToggleGroups(source whatsapp.IWhatsappOptions) error {
	options := source.GetOptions()

	switch options.Groups {
	case whatsapp.UnSetBooleanType:
		options.Groups = whatsapp.TrueBooleanType
	case whatsapp.TrueBooleanType:
		options.Groups = whatsapp.FalseBooleanType
	default:
		options.Groups = whatsapp.UnSetBooleanType
	}

	reason := fmt.Sprintf("toggle groups: %s", options.Groups)
	return source.Save(reason)
}

// ToggleBroadcasts cycles the persisted broadcast-handling mode.
func ToggleBroadcasts(source whatsapp.IWhatsappOptions) error {
	options := source.GetOptions()

	switch options.Broadcasts {
	case whatsapp.UnSetBooleanType:
		options.Broadcasts = whatsapp.TrueBooleanType
	case whatsapp.TrueBooleanType:
		options.Broadcasts = whatsapp.FalseBooleanType
	default:
		options.Broadcasts = whatsapp.UnSetBooleanType
	}

	reason := fmt.Sprintf("toggle broadcasts: %s", options.Broadcasts)
	return source.Save(reason)
}

// ToggleCalls cycles the persisted call-handling mode.
func ToggleCalls(source whatsapp.IWhatsappOptions) error {
	options := source.GetOptions()

	switch options.Calls {
	case whatsapp.UnSetBooleanType:
		options.Calls = whatsapp.TrueBooleanType
	case whatsapp.TrueBooleanType:
		options.Calls = whatsapp.FalseBooleanType
	default:
		options.Calls = whatsapp.UnSetBooleanType
	}

	reason := fmt.Sprintf("toggle calls: %s", options.Calls)
	return source.Save(reason)
}

// ToggleReadUpdate cycles the persisted mark-read update handling mode.
func ToggleReadUpdate(source whatsapp.IWhatsappOptions) error {
	options := source.GetOptions()

	switch options.ReadUpdate {
	case whatsapp.UnSetBooleanType:
		options.ReadUpdate = whatsapp.TrueBooleanType
	case whatsapp.TrueBooleanType:
		options.ReadUpdate = whatsapp.FalseBooleanType
	default:
		options.ReadUpdate = whatsapp.UnSetBooleanType
	}

	reason := fmt.Sprintf("toggle read update: %s", options.ReadUpdate)
	return source.Save(reason)
}
