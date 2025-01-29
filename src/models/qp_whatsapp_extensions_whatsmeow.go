package models

import (
	"strings"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	whatsmeow "github.com/nocodeleaks/quepasa/whatsmeow"
)

func NewWhatsmeowEmptyConnection(callback func(string)) (conn whatsapp.IWhatsappConnection, err error) {
	conn, err = whatsmeow.WhatsmeowService.CreateEmptyConnection()
	if err != nil {
		return
	}

	conn.UpdatePairedCallBack(callback)
	return
}

func NewWhatsmeowConnection(options *whatsapp.WhatsappConnectionOptions) (whatsapp.IWhatsappConnection, error) {
	return whatsmeow.WhatsmeowService.CreateConnection(options)
}

func ToQpMessageV2(source whatsapp.WhatsappMessage, server *QpWhatsappServer) (message QpMessageV2) {
	message.ID = source.Id
	message.Timestamp = uint64(source.Timestamp.Unix())
	message.Text = source.Text
	message.FromMe = source.FromMe

	message.Controller = QPEndpointV2{}
	if !strings.Contains(server.Wid, "@") {
		message.Controller.ID = server.GetNumber() + "@c.us"
	} else {
		message.Controller.ID = server.GetNumber()
	}

	message.ReplyTo = ChatToQPEndPointV2(source.Chat)
	message.Chat = ChatToQPChatV2(source.Chat)

	if source.Participant != nil {
		message.Participant = ChatToQPEndPointV2(*source.Participant)
	}

	if source.HasAttachment() {
		message.Attachment = ToQPAttachmentV1(source.Attachment, message.ID, server.Token)
	}

	if len(source.InReply) > 0 {
		message.Text = "*(IN REPLY) " + message.Text
	}

	if source.ForwardingScore > 0 {
		message.Text = "*(FORWARDED) " + message.Text
	}

	return
}

func ToQPMessageV1(source whatsapp.WhatsappMessage, wid string) (message QPMessageV1) {
	message.ID = source.Id
	message.Timestamp = uint64(source.Timestamp.Unix())
	message.Text = source.Text
	message.FromMe = source.FromMe

	message.Controller = QPEndpointV1{}
	if !strings.Contains(wid, "@") {
		message.Controller.ID = wid + "@c.us"
	} else {
		message.Controller.ID = wid
	}

	message.ReplyTo = ChatToQPEndPointV1(source.Chat)

	if source.Participant != nil {
		message.Participant = ChatToQPEndPointV1(*source.Participant)
	}

	if source.HasAttachment() {
		message.Attachment = ToQPAttachmentV1(source.Attachment, message.ID, wid)
	}

	return
}
