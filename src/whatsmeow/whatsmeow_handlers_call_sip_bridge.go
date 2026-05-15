package whatsmeow

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/emiago/sipgo/sip"
	sipproxy "github.com/nocodeleaks/quepasa/sipproxy"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
	waBinary "go.mau.fi/whatsmeow/binary"
	types "go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

type sipCallBridgeContext struct {
	handler  *WhatsmeowHandlers
	callFrom types.JID
}

var sipBridgeState = struct {
	mu                 sync.RWMutex
	callbacksBound     bool
	activeCallContexts map[string]sipCallBridgeContext
}{
	activeCallContexts: make(map[string]sipCallBridgeContext),
}

func addSIPCallBridgeContext(callID string, handler *WhatsmeowHandlers, callFrom types.JID) {
	if len(callID) == 0 || handler == nil {
		return
	}

	sipBridgeState.mu.Lock()
	defer sipBridgeState.mu.Unlock()
	sipBridgeState.activeCallContexts[callID] = sipCallBridgeContext{handler: handler, callFrom: callFrom}
}

func popSIPCallBridgeContext(callID string) (sipCallBridgeContext, bool) {
	sipBridgeState.mu.Lock()
	defer sipBridgeState.mu.Unlock()

	ctx, ok := sipBridgeState.activeCallContexts[callID]
	if ok {
		delete(sipBridgeState.activeCallContexts, callID)
	}

	return ctx, ok
}

func bindSIPBridgeCallbacks(logentry *log.Entry) bool {
	if sipproxy.SIPProxy == nil {
		return false
	}

	sipBridgeState.mu.Lock()
	defer sipBridgeState.mu.Unlock()

	if sipBridgeState.callbacksBound {
		return true
	}

	sipproxy.SIPProxy.SetCallAcceptedHandler(func(callID, fromPhone, toPhone string, _ *sip.Response) {
		callCtx, ok := popSIPCallBridgeContext(callID)
		if !ok || callCtx.handler == nil {
			return
		}

		handlerLog := callCtx.handler.GetLogger().WithField(LogFields.MessageId, callID)
		err := callCtx.handler.acceptIncomingCall(callCtx.callFrom)
		if err != nil {
			handlerLog.Warnf("SIP accepted but WhatsApp accept failed: %v", err)
			return
		}

		handlerLog.Infof("SIP accepted and WhatsApp call acknowledged")
	})

	sipproxy.SIPProxy.SetCallRejectedHandler(func(callID, fromPhone, toPhone string, _ *sip.Response) {
		callCtx, ok := popSIPCallBridgeContext(callID)
		if !ok || callCtx.handler == nil || callCtx.handler.Client == nil {
			return
		}

		handlerLog := callCtx.handler.GetLogger().WithField(LogFields.MessageId, callID)
		err := callCtx.handler.Client.RejectCall(context.Background(), callCtx.callFrom, callID)
		if err != nil {
			handlerLog.Warnf("SIP rejected but WhatsApp reject failed: %v", err)
			return
		}

		handlerLog.Infof("SIP rejected and WhatsApp call rejected")
	})

	sipBridgeState.callbacksBound = true
	if logentry != nil {
		logentry.Infof("SIP bridge callbacks configured")
	}

	return true
}

func (source *WhatsmeowHandlers) rejectIncomingCall(evt types.BasicCallMeta) {
	if source == nil || source.Client == nil {
		return
	}

	logentry := source.GetLogger().WithField(LogFields.MessageId, evt.CallID)
	err := source.Client.RejectCall(context.Background(), evt.From, evt.CallID)
	if err != nil {
		logentry.Errorf("error on rejecting call: %s", err.Error())
		return
	}

	logentry.Infof("rejecting incoming call from: %s", evt.From)
}

func (source *WhatsmeowHandlers) acceptIncomingCall(from types.JID) error {
	if source == nil {
		return fmt.Errorf("nil source handler")
	}

	if source.Client == nil || source.Client.Store == nil || source.Client.Store.ID.IsEmpty() {
		return fmt.Errorf("whatsapp client not ready to accept call")
	}

	node := waBinary.Node{
		Tag: "ack",
		Attrs: waBinary.Attrs{
			"id":    source.Client.GenerateMessageID(),
			"to":    from,
			"class": "receipt",
			"from":  source.Client.Store.ID.String(),
		},
	}

	return source.Client.DangerousInternals().SendNode(context.Background(), node)
}

func (source *WhatsmeowHandlers) forwardIncomingCallToSIP(evt types.BasicCallMeta) {
	if source == nil || source.Client == nil {
		return
	}

	logentry := source.GetLogger().WithField(LogFields.MessageId, evt.CallID)
	if sipproxy.SIPProxy == nil {
		logentry.Warn("SIP proxy is disabled or not initialized, skipping SIP bridge for incoming call")
		return
	}

	if !bindSIPBridgeCallbacks(logentry) {
		logentry.Warn("failed to bind SIP bridge callbacks, skipping SIP bridge for incoming call")
		return
	}

	identifiers := ResolveCallIdentifiers(evt, logentry)
	callerJID := identifiers.ChatJID
	if callerJID.IsEmpty() {
		callerJID = evt.From
	}

	fromPhone := resolveSIPBridgeCallerPhone(callerJID, source.GetContactManager(), logentry)
	toPhone := source.Client.Store.ID.User
	if len(fromPhone) == 0 || len(toPhone) == 0 {
		logentry.Warnf("missing call phones for SIP bridge (from=%q, to=%q), rejecting incoming call", fromPhone, toPhone)
		source.rejectIncomingCall(evt)
		return
	}

	addSIPCallBridgeContext(evt.CallID, source, evt.From)

	callManager := sipproxy.NewSIPProxyCallAnswerManager(sipproxy.SIPProxy)
	err := callManager.AnswerCallWithReceiver(fromPhone, toPhone, evt.CallID)
	if err != nil {
		_, _ = popSIPCallBridgeContext(evt.CallID)
		logentry.Errorf("failed forwarding call to SIP proxy: %v", err)
		source.rejectIncomingCall(evt)
		return
	}

	logentry.Infof("incoming call forwarded to SIP proxy (%s -> %s)", fromPhone, toPhone)
}

func resolveSIPBridgeCallerPhone(callerJID types.JID, contactManager whatsapp.WhatsappContactManagerInterface, logentry *log.Entry) string {
	if callerJID.IsEmpty() || len(callerJID.User) == 0 {
		return ""
	}

	callerID := callerJID.ToNonAD().String()
	if !strings.HasSuffix(callerID, "@lid") {
		return callerJID.User
	}

	if contactManager == nil {
		if logentry != nil {
			logentry.Warnf("caller is LID but contact manager is unavailable, using opaque caller %s", callerID)
		}
		return callerJID.User
	}

	phone, err := contactManager.GetPhoneFromContactId(callerID)
	if err != nil || len(phone) == 0 {
		if logentry != nil {
			logentry.Warnf("failed to resolve LID caller %s, using opaque caller ID", callerID)
		}
		return callerJID.User
	}

	normalizedPhone, err := whatsapp.GetPhoneIfValid(phone)
	if err != nil || len(normalizedPhone) == 0 {
		if logentry != nil {
			logentry.Warnf("resolved LID caller phone is invalid (%s), using opaque caller ID", phone)
		}
		return callerJID.User
	}

	if logentry != nil {
		logentry.Infof("resolved LID caller to phone for SIP bridge")
	}

	return normalizedPhone
}

func (source *WhatsmeowHandlers) HandleCallTerminate(evt events.CallTerminate) {
	if source == nil {
		return
	}

	logentry := source.GetLogger().WithField(LogFields.MessageId, evt.CallID)
	logentry.Infof("call terminate received (reason=%s)", evt.Reason)

	_, _ = popSIPCallBridgeContext(evt.CallID)

	if sipproxy.SIPProxy == nil {
		return
	}

	if err := sipproxy.SIPProxy.CancelCall(evt.CallID); err != nil {
		logentry.Warnf("failed to cancel SIP call on terminate event: %v", err)
		return
	}

	logentry.Infof("SIP call canceled due to terminate event")
}
