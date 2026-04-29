package models

import (
	dispatchservice "github.com/nocodeleaks/quepasa/dispatch/service"
	"github.com/nocodeleaks/quepasa/library"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// OutboundDispatchingSubscriber is the default outbound subscriber attached to
// models.DispatchingHandler. It bridges cached inbound/outbound events to the
// transport dispatch service using server dispatching configuration.
type OutboundDispatchingSubscriber struct {
	library.LogStruct // logging
	server            *QpWhatsappServer
}

// NewOutboundDispatchingSubscriber builds a dispatching subscriber bound to a server.
func NewOutboundDispatchingSubscriber(server *QpWhatsappServer) *OutboundDispatchingSubscriber {
	out := &OutboundDispatchingSubscriber{server: server}
	if server != nil {
		out.LogEntry = server.GetLogger()
	}
	return out
}

func (source *OutboundDispatchingSubscriber) HandleDispatching(payload *whatsapp.WhatsappMessage) {
	if !source.HasDispatching() {
		return
	}

	// updating log
	logentry := source.GetLogger()
	loglevel := logentry.Level
	logentry = logentry.WithField(LogFields.MessageId, payload.Id)
	logentry.Level = loglevel

	err := DispatchOutboundFromServer(source.server, payload)
	if err != nil {
		logentry.Errorf("error on handle dispatching distributions: %s", err.Error())
	}
}

func (source *OutboundDispatchingSubscriber) isDispatchingSubscriber() {}

// DispatchOutboundFromServer sends a message to all dispatching targets of a server.
func DispatchOutboundFromServer(server *QpWhatsappServer, message *whatsapp.WhatsappMessage) error {
	if server == nil {
		return nil
	}

	return DispatchOutboundToTargets(server, server.QpDataDispatching.Dispatching, message)
}

// DispatchOutboundToTargets sends a message to an explicit dispatching snapshot.
// This is used by normal flow and delete/redispatch flows that require stable targets.
func DispatchOutboundToTargets(server *QpWhatsappServer, dispatchings []*QpDispatching, message *whatsapp.WhatsappMessage) error {
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

	err := dispatchservice.GetInstance().DispatchOutbound(request)
	if server != nil {
		if syncErr := server.QpDataDispatching.DispatchingSyncHealth(dispatchings); syncErr != nil {
			if err != nil {
				return err
			}
			return syncErr
		}
	}

	return err
}

func (source *OutboundDispatchingSubscriber) HasDispatching() bool {
	if source.server != nil {
		return len(source.server.QpDataDispatching.Dispatching) > 0
	}
	return false
}

func (source *OutboundDispatchingSubscriber) HasWebhook() bool {
	if source.server != nil {
		webhooks := source.server.GetWebhooks()
		return len(webhooks) > 0
	}
	return false
}
