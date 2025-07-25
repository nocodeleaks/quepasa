package whatsapp

import log "github.com/sirupsen/logrus"

type IWhatsappConnectionOptions interface {
	GetLogger() *log.Entry
}
