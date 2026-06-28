package models

import (
	"github.com/nocodeleaks/quepasa/ports"
	"github.com/nocodeleaks/quepasa/whatsapp"
)

// NewWhatsmeowEmptyConnection creates an empty unpaired connection via the injected driver.
// Breaking models -> whatsmeow import per ADR-0003 and PLAN P1.1.
func NewWhatsmeowEmptyConnection(callback func(string)) (conn whatsapp.IWhatsappConnection, err error) {
	if ports.GlobalWhatsappDriverFactory == nil {
		panic("GlobalWhatsappDriverFactory not injected — call ports.SetWhatsappDriver() in main.go")
	}

	conn, err = ports.GlobalWhatsappDriverFactory.CreateEmptyConnection()
	if err != nil {
		return
	}

	conn.UpdatePairedCallBack(callback)
	return
}

// NewWhatsmeowConnection creates a connection from options via the injected driver.
func NewWhatsmeowConnection(options *whatsapp.WhatsappConnectionOptions) (whatsapp.IWhatsappConnection, error) {
	if ports.GlobalWhatsappDriverFactory == nil {
		panic("GlobalWhatsappDriverFactory not injected — call ports.SetWhatsappDriver() in main.go")
	}

	return ports.GlobalWhatsappDriverFactory.CreateConnection(options)
}
