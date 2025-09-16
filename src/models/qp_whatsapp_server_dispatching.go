package models

import (
	"context"
	"fmt"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
)

type QpWhatsappServerDispatching struct {
	*QpWebhook

	server *QpWhatsappServer
}

// get default log entry, never nil
func (source *QpWhatsappServerDispatching) GetLogger() *log.Entry {
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

func (source *QpWhatsappServerDispatching) GetOptions() *whatsapp.WhatsappOptions {
	if source == nil {
		return nil
	}

	return &source.WhatsappOptions
}

//#endregion

func (source *QpWhatsappServerDispatching) Save(reason string) (err error) {

	if source == nil {
		err = fmt.Errorf("nil dispatching source")
		return err
	}

	if source.server == nil {
		err = fmt.Errorf("nil server")
		return err
	}

	if source.QpWebhook == nil {
		err = fmt.Errorf("nil source configuration")
		return err
	}

	logentry := source.GetLogger()
	logentry.Debugf("saving configuration info, reason: %s, content: %+v", reason, source)

	// Convert to dispatching format and save via new system
	dispatching := source.QpWebhook.ToDispatching()
	affected, err := source.server.DispatchingAddOrUpdate(dispatching)
	if err == nil {
		logentry.Infof("saved configuration as dispatching with %v affected rows", affected)
	}

	return err
}

func (source *QpWhatsappServerDispatching) ToggleForwardInternal() (handle bool, err error) {
	source.ForwardInternal = !source.ForwardInternal

	reason := fmt.Sprintf("toggle forward internal: %v", source.ForwardInternal)
	return source.ForwardInternal, source.Save(reason)
}

// SetServer sets the server reference for this dispatching configuration
func (source *QpWhatsappServerDispatching) SetServer(server *QpWhatsappServer) {
	source.server = server
}
