package models

import (
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
)

// QpWhatsappSessionDispatching is the preferred session-oriented wrapper around
// a dispatching configuration. It delegates to the legacy server wrapper while
// the codebase migrates incrementally.
type QpWhatsappSessionDispatching QpWhatsappServerDispatching

func (source *QpWhatsappSessionDispatching) legacy() *QpWhatsappServerDispatching {
	return (*QpWhatsappServerDispatching)(source)
}

func (source *QpWhatsappSessionDispatching) GetLogger() *log.Entry {
	return source.legacy().GetLogger()
}

func (source *QpWhatsappSessionDispatching) GetOptions() *whatsapp.WhatsappOptions {
	return source.legacy().GetOptions()
}

func (source *QpWhatsappSessionDispatching) Save(reason string) error {
	return source.legacy().Save(reason)
}

func (source *QpWhatsappSessionDispatching) ToggleForwardInternal() (bool, error) {
	return source.legacy().ToggleForwardInternal()
}

func (source *QpWhatsappSessionDispatching) SetSession(session *QpWhatsappSession) {
	source.legacy().SetServer(session)
}

func NewQpWhatsappSessionDispatchingFromDispatching(dispatching *QpDispatching, session *QpWhatsappSession) *QpWhatsappSessionDispatching {
	legacy := NewQpWhatsappServerDispatchingFromDispatching(dispatching, session)
	if legacy == nil {
		return nil
	}
	return (*QpWhatsappSessionDispatching)(legacy)
}
