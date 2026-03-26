package models

import (
	"context"
	"fmt"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
)

type QpWhatsappServerDispatching struct {
	*QpDispatching

	server *QpWhatsappServer
}

// get default log entry, never nil
func (source *QpWhatsappServerDispatching) GetLogger() *log.Entry {
	var logentry *log.Entry
	if source != nil && source.server != nil {
		logentry = source.server.GetLogger()
		if source.QpDispatching != nil {
			logentry = source.QpDispatching.GetLogger()
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

// Save validates the wrapper state before persisting the dispatching settings.
//
// The dispatching instance keeps a cached copy of the server WID that is later
// reused as the X-QUEPASA-WID header during webhook delivery. Synchronizing the
// current server WID here prevents saving a configuration that would keep using
// an outdated or empty header value after the server identity becomes available.
func (source *QpWhatsappServerDispatching) Save(reason string) (err error) {

	if source == nil {
		err = fmt.Errorf("nil dispatching source")
		return err
	}

	if source.server == nil {
		err = fmt.Errorf("nil server")
		return err
	}

	if source.QpDispatching == nil {
		err = fmt.Errorf("nil source configuration")
		return err
	}

	source.QpDispatching.Wid = source.server.GetWId()

	logentry := source.GetLogger()
	logentry.Debugf("saving configuration info, reason: %s, content: %+v", reason, source)

	affected, err := source.server.DispatchingAddOrUpdate(source.QpDispatching)
	if err == nil {
		logentry.Infof("saved configuration as dispatching with %v affected rows, type: %s", affected, source.QpDispatching.Type)
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

// NewFromDispatching creates a QpWhatsappServerDispatching from a QpDispatching
func NewQpWhatsappServerDispatchingFromDispatching(dispatching *QpDispatching, server *QpWhatsappServer) *QpWhatsappServerDispatching {
	if dispatching == nil {
		return nil
	}

	result := &QpWhatsappServerDispatching{
		QpDispatching: dispatching,
		server:        server,
	}

	return result
}
