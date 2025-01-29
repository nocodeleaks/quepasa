package models

import (
	"context"
	"fmt"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
)

type QpWhatsappServerWebhook struct {
	*QpWebhook

	server *QpWhatsappServer
}

// get default log entry, never nil
func (source *QpWhatsappServerWebhook) GetLogger() *log.Entry {
	var logentry *log.Entry
	if source != nil && source.server != nil {
		logentry = source.server.GetLogger()
		if source.QpWebhook != nil {
			logentry = source.QpWebhook.GetLogger()
		}
	} else {
		logentry = log.New().WithContext(context.Background())
	}

	return logentry
}

//#region IMPLEMENTING WHATSAPP OPTIONS INTERFACE

func (source *QpWhatsappServerWebhook) GetOptions() *whatsapp.WhatsappOptions {
	if source == nil {
		return nil
	}

	return &source.WhatsappOptions
}

//#endregion

func (source *QpWhatsappServerWebhook) Save(reason string) (err error) {

	if source == nil {
		err = fmt.Errorf("nil webhook source")
		return err
	}

	if source.server == nil {
		err = fmt.Errorf("nil server")
		return err
	}

	if source.QpWebhook == nil {
		err = fmt.Errorf("nil source webhook")
		return err
	}

	logentry := source.GetLogger()
	logentry.Debugf("saving webhook info, reason: %s, content: %+v", reason, source)

	affected, err := source.server.WebhookAddOrUpdate(source.QpWebhook)
	if err == nil {
		logentry.Infof("saved webhook with %v affected rows", affected)
	}

	return err
}

func (source *QpWhatsappServerWebhook) ToggleForwardInternal() (handle bool, err error) {
	source.ForwardInternal = !source.ForwardInternal

	reason := fmt.Sprintf("toggle forward internal: %v", source.ForwardInternal)
	return source.ForwardInternal, source.Save(reason)
}
