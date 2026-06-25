package whatsapp

import log "github.com/nocodeleaks/quepasa/qplog"

type IWhatsappConnectionOptions interface {
	GetLogger() log.Logger
}
