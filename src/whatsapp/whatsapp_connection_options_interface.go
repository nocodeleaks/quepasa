package whatsapp

import log "github.com/sirupsen/logrus"

type IWhatsappConnectionOptions interface {
	GetWid() string

	// should auto reconnect, false for qrcode scanner
	SetReconnect(bool)

	GetReconnect() bool

	GetLogger() *log.Entry
}
