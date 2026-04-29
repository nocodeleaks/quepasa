package runtime

import (
	"strconv"
	"time"

	events "github.com/nocodeleaks/quepasa/events"
	"github.com/nocodeleaks/quepasa/library"
	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// DispatchingHandler implements the business logic for dispatching outbound messages
// from a WhatsApp server to external targets (webhooks, RabbitMQ, etc.).
//
// ARCHITECTURAL NOTE:
// DispatchingHandler is RUNTIME LOGIC (business rules), not TRANSPORT LOGIC.
//
//   - DISPATCH (transport layer): HOW to send data (webhooks, rabbitmq, realtime)
//     Located: src/dispatch/service/
//     Concern: Technical implementation of outbound transport mechanisms
//     Contracts: Target, OutboundRequest, Transport interfaces
//
//   - RUNTIME (business layer): WHEN and WHY to dispatch, which rules apply
//     Located: src/runtime/
//     Concern: Business decisions (server config, enrichment, filtering, triggers)
//     Contracts: Server state, message routing, domain events
//
// DispatchingHandler knows about:
// - server.QpDataDispatching (business config: which targets to use)
// - CloneAndEnrichMessageForServer (domain enrichment rule)
// - message flow pipeline integration (when to trigger dispatch)
//
// DispatchingHandler does NOT know about:
// - How webhooks are sent (that's dispatch.service.SendWebhook)
// - How RabbitMQ topics are determined (that's dispatch.service.PublishRabbitMQ)
// - Connection details or transport mechanisms
//
// This separation ensures models/domain logic stays independent of delivery
// mechanism changes, and transport concerns stay isolated in the dispatch module.
type DispatchingHandler struct {
	library.LogStruct // logging
	server            *models.QpWhatsappServer
}

// HandleDispatching processes an inbound WhatsApp message and dispatches it
// to configured outbound targets if dispatching is enabled on the server.
func (source *DispatchingHandler) HandleDispatching(payload *whatsapp.WhatsappMessage) {
	if !source.HasDispatching() {
		return
	}

	// updating log
	logentry := source.GetLogger()
	loglevel := logentry.Level
	logentry = logentry.WithField(models.LogFields.MessageId, payload.Id)
	logentry.Level = loglevel

	startedAt := time.Now()
	err := models.DispatchOutboundFromServer(source.server, payload)
	targetCount := 0
	if source.server != nil {
		targetCount = len(source.server.QpDataDispatching.Dispatching)
	}
	if err != nil {
		publishRuntimeDispatchEvent("error", time.Since(startedAt), targetCount)
		logentry.Errorf("error on handle dispatching distributions: %s", err.Error())
		return
	}

	publishRuntimeDispatchEvent("success", time.Since(startedAt), targetCount)
}

func publishRuntimeDispatchEvent(status string, duration time.Duration, targetCount int) {
	events.Publish(events.Event{
		Name:     "runtime.dispatch.outbound",
		Source:   "runtime.dispatching_handler",
		Status:   status,
		Duration: duration,
		Attributes: map[string]string{
			"target_count": strconv.Itoa(targetCount),
		},
	})
}

// HasDispatching returns true if the server has at least one dispatching target configured.
func (source *DispatchingHandler) HasDispatching() bool {
	if source.server != nil {
		return len(source.server.QpDataDispatching.Dispatching) > 0
	}
	return false
}

// HasWebhook returns true if the server has at least one webhook configured.
func (source *DispatchingHandler) HasWebhook() bool {
	if source.server != nil {
		webhooks := source.server.GetWebhooks()
		return len(webhooks) > 0
	}
	return false
}
