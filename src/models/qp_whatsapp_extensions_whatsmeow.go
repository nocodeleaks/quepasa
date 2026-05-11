package models

import (
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
