package service

import (
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
)

// DispatchPolicy decides whether a message should be dispatched to a specific target.
// Implementing this interface in the runtime package allows business routing rules
// to live outside the transport layer while keeping DispatchService transport-agnostic.
type DispatchPolicy interface {
	ShouldDispatch(target Target, message *whatsapp.WhatsappMessage, logentry *log.Entry) bool
}

// DefaultDispatchPolicy applies the standard outbound filter rules.
// It honors per-target flags (read receipts, groups, broadcasts, calls) and
// prevents internal message forwarding loops.
type DefaultDispatchPolicy struct{}

func (DefaultDispatchPolicy) ShouldDispatch(target Target, message *whatsapp.WhatsappMessage, logentry *log.Entry) bool {
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
