package models

import (
	dispatchservice "github.com/nocodeleaks/quepasa/dispatch/service"
	"github.com/nocodeleaks/quepasa/library"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// DEPRECATED: QPDispatchingHandler has been moved to the runtime module.
//
// See: src/runtime/dispatching_handler.go
//
// This implementation remains temporarily for backward compatibility but will be removed
// in a future refactoring. New code should use runtime.DispatchingHandler instead.
//
// The separation follows the architectural pattern:
// - runtime module: WHEN/WHY to dispatch (business logic)
// - dispatch module: HOW to send (transport implementation)
type QPDispatchingHandler struct {
	library.LogStruct // logging
	server            *QpWhatsappServer
}

func (source *QPDispatchingHandler) HandleDispatching(payload *whatsapp.WhatsappMessage) {
	if !source.HasDispatching() {
		return
	}

	// updating log
	logentry := source.GetLogger()
	loglevel := logentry.Level
	logentry = logentry.WithField(LogFields.MessageId, payload.Id)
	logentry.Level = loglevel

	err := dispatchOutboundFromServer(source.server, payload)
	if err != nil {
		logentry.Errorf("error on handle dispatching distributions: %s", err.Error())
	}
}

func dispatchOutboundFromServer(server *QpWhatsappServer, message *whatsapp.WhatsappMessage) error {
	if server == nil {
		return nil
	}

	return dispatchOutboundToTargets(server, server.QpDataDispatching.Dispatching, message)
}

func dispatchOutboundToTargets(server *QpWhatsappServer, dispatchings []*QpDispatching, message *whatsapp.WhatsappMessage) error {
	if message == nil {
		return nil
	}

	serverWid := ""
	if server != nil {
		serverWid = server.GetWId()
	}

	targets := make([]dispatchservice.Target, 0, len(dispatchings))
	for _, dispatching := range dispatchings {
		if dispatching == nil {
			continue
		}
		targets = append(targets, dispatching)
	}

	request := &dispatchservice.OutboundRequest{
		ServerWid: serverWid,
		Message:   message,
		Targets:   targets,
		Enrich: func(payload *whatsapp.WhatsappMessage) *whatsapp.WhatsappMessage {
			if server == nil {
				return payload
			}
			return CloneAndEnrichMessageForServer(server, payload)
		},
	}

	return dispatchservice.GetInstance().DispatchOutbound(request)
}

func (source *QPDispatchingHandler) HasDispatching() bool {
	if source.server != nil {
		return len(source.server.QpDataDispatching.Dispatching) > 0
	}
	return false
}

func (source *QPDispatchingHandler) HasWebhook() bool {
	if source.server != nil {
		webhooks := source.server.GetWebhooks()
		return len(webhooks) > 0
	}
	return false
}
