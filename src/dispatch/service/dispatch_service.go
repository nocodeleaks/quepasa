package service

import (
	"crypto/tls"
	"net/http"
	"sync"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
)

// DispatchService coordinates outbound delivery to external integrations.
// It does not handle internal persistence/caching concerns.

// Target represents one configured dispatch destination (webhook, rabbitmq, etc.).
type Target interface {
	GetDispatchType() string
	GetWid() string
	SetWid(wid string)
	GetLogger() *log.Entry

	IsSetReadReceipts() bool
	GetReadReceipts() bool
	IsSetGroups() bool
	GetGroups() bool
	IsSetBroadcasts() bool
	GetBroadcasts() bool
	IsSetCalls() bool
	GetCalls() bool

	IsFromInternalForwardEnabled() bool
	GetTrackId() string

	Dispatch(message *whatsapp.WhatsappMessage) error
}

// HandlerSubscriber represents one downstream dispatch hook that receives
// messages validated by the dispatch orchestration flow.
type HandlerSubscriber interface {
	HandleDispatching(*whatsapp.WhatsappMessage)
}

// HandlerFlowRequest encapsulates all callbacks required for dispatch flow
// orchestration without coupling this module to higher-level packages.
type HandlerFlowRequest struct {
	Payload *whatsapp.WhatsappMessage

	Validate  func(payload *whatsapp.WhatsappMessage) string
	OnInvalid func(reason string, payload *whatsapp.WhatsappMessage)

	MarkEventTimestamp   func()
	MarkMessageTimestamp func()

	SetMessageWid    func(payload *whatsapp.WhatsappMessage)
	PublishRealtime  func(payload *whatsapp.WhatsappMessage)
	HandlerCallbacks []HandlerSubscriber
}

// OutboundRequest encapsulates external-delivery context for webhook/rabbitmq.
type OutboundRequest struct {
	ServerWid string
	Message   *whatsapp.WhatsappMessage
	Targets   []Target

	Enrich func(message *whatsapp.WhatsappMessage) *whatsapp.WhatsappMessage
}

type DispatchService struct{}

var dispatchServiceOnce sync.Once
var dispatchServiceInstance *DispatchService

func GetInstance() *DispatchService {
	dispatchServiceOnce.Do(func() {
		dispatchServiceInstance = &DispatchService{}
	})

	return dispatchServiceInstance
}

// DispatchHandlerFlow centralizes trigger orchestration for outbound dispatch.
// It validates payload, updates event/message timestamps through callbacks,
// publishes realtime notifications, and fan-outs to registered handlers.
func (service *DispatchService) DispatchHandlerFlow(request *HandlerFlowRequest) {
	if service == nil || request == nil || request.Payload == nil {
		return
	}

	if request.Validate != nil {
		reason := request.Validate(request.Payload)
		if reason != "" {
			if request.OnInvalid != nil {
				request.OnInvalid(reason, request.Payload)
			}
			return
		}
	}

	isEvent := request.Payload.Type == whatsapp.UnhandledMessageType ||
		request.Payload.Type == whatsapp.SystemMessageType ||
		request.Payload.Id == "readreceipt"

	if isEvent {
		if request.MarkEventTimestamp != nil {
			request.MarkEventTimestamp()
		}
	} else {
		if request.MarkMessageTimestamp != nil {
			request.MarkMessageTimestamp()
		}
	}

	if request.SetMessageWid != nil {
		request.SetMessageWid(request.Payload)
	}

	if request.PublishRealtime != nil {
		request.PublishRealtime(request.Payload)
	}

	for _, handler := range request.HandlerCallbacks {
		if handler == nil {
			continue
		}
		go handler.HandleDispatching(request.Payload)
	}
}

// DispatchOutbound applies optional enrichment and then dispatches a message
// to all configured external targets (webhook/rabbitmq).
func (service *DispatchService) DispatchOutbound(request *OutboundRequest) error {
	if service == nil || request == nil || request.Message == nil {
		return nil
	}

	message := request.Message
	if request.Enrich != nil {
		message = request.Enrich(message)
	}

	return service.DispatchTargets(request.ServerWid, message, request.Targets)
}

// DispatchTargets sends one message to every configured external target,
// applying destination-level filters before delivery.
func (service *DispatchService) DispatchTargets(serverWid string, message *whatsapp.WhatsappMessage, targets []Target) error {
	if service == nil || message == nil {
		return nil
	}

	// Keep legacy behavior for compatibility with self-signed webhook endpoints.
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	for _, target := range targets {
		if target == nil {
			continue
		}

		if target.GetWid() != serverWid {
			target.SetWid(serverWid)
		}

		logentry := target.GetLogger()
		loglevel := logentry.Level
		logentry = logentry.WithField("messageid", message.Id)
		logentry.Level = loglevel

		if !service.shouldDispatchToTarget(target, message, logentry) {
			continue
		}

		err := target.Dispatch(message)
		if err != nil {
			logentry.Errorf("error on dispatch: %s", err.Error())
		}
	}

	return nil
}

// shouldDispatchToTarget centralizes outbound filter rules shared by all
// external dispatch transports.
func (service *DispatchService) shouldDispatchToTarget(target Target, message *whatsapp.WhatsappMessage, logentry *log.Entry) bool {
	if message.Id == "readreceipt" && target.IsSetReadReceipts() && !target.GetReadReceipts() {
		logentry.Debugf("ignoring read receipt message: %s", message.Text)
		return false
	}

	if message.FromGroup() && target.IsSetGroups() && !target.GetGroups() {
		logentry.Debug("ignoring group message")
		return false
	}

	if message.FromBroadcast() && target.IsSetBroadcasts() && !target.GetBroadcasts() {
		logentry.Debug("ignoring broadcast message")
		return false
	}

	if message.Type == whatsapp.CallMessageType && target.IsSetCalls() && !target.GetCalls() {
		logentry.Debug("ignoring call message")
		return false
	}

	if message.FromInternal && (!target.IsFromInternalForwardEnabled() || (target.GetTrackId() != "" && target.GetTrackId() == message.TrackId)) {
		return false
	}

	return true
}
