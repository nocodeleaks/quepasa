package models

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strings"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// handle message deliver to dispatching distribution
func PostToDispatchingFromServer(server *QpWhatsappServer, message *whatsapp.WhatsappMessage) (err error) {
	if server == nil {
		err = fmt.Errorf("server nil")
		return err
	}

	// ignoring ssl issues
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	for _, dispatching := range server.QpDataDispatching.Dispatching {

		// updating log
		logentry := dispatching.GetLogger()
		loglevel := logentry.Level
		logentry = logentry.WithField(LogFields.MessageId, message.Id)
		logentry.Level = loglevel

		if message.Id == "readreceipt" && dispatching.IsSetReadReceipts() && !dispatching.ReadReceipts.Boolean() {
			logentry.Debugf("ignoring read receipt message: %s", message.Text)
			continue
		}

		if message.FromGroup() && dispatching.IsSetGroups() && !dispatching.Groups.Boolean() {
			logentry.Debug("ignoring group message")
			continue
		}

		if message.FromDirect() && dispatching.IsSetDirect() && !dispatching.Direct.Boolean() {
			logentry.Debug("ignoring direct message")
			continue
		}

		if message.FromBroadcast() && dispatching.IsSetBroadcasts() && !dispatching.Broadcasts.Boolean() {
			logentry.Debug("ignoring broadcast message")
			continue
		}

		if message.Type == whatsapp.CallMessageType && dispatching.IsSetCalls() && !dispatching.Calls.Boolean() {
			logentry.Debug("ignoring call message")
			continue
		}

		if !message.FromInternal || (dispatching.ForwardInternal && (len(dispatching.TrackId) == 0 || dispatching.TrackId != message.TrackId)) {
			elerr := dispatching.Dispatch(message)
			if elerr != nil {
				logentry.Errorf("error on dispatch: %s", elerr.Error())
			}
		}
	}

	return
}

// region FIND|SEARCH WHATSAPP SERVER
var ErrServerNotFound error = errors.New("the requested whatsapp server was not found")

func GetServerFromID(source string) (server *QpWhatsappServer, err error) {
	server, ok := WhatsappService.Servers[source]
	if !ok {
		err = ErrServerNotFound
		return
	}
	return
}

func GetServerFromBot(source QPBot) (server *QpWhatsappServer, err error) {
	return GetServerFromID(source.Wid)
}

// insecure
func GetServerFirstAvailable() (server *QpWhatsappServer, err error) {
	for _, item := range WhatsappService.Servers {
		if item != nil && item.GetStatus() == whatsapp.Ready {
			server = item
			break
		}
	}

	if server == nil {
		err = ErrServerNotFound
	}

	return
}

func GetServerFromToken(token string) (server *QpWhatsappServer, err error) {
	for _, item := range WhatsappService.Servers {
		if item != nil && strings.EqualFold(item.Token, token) {
			server = item
			break
		}
	}

	if server == nil {
		err = ErrServerNotFound
	}

	return
}

func GetServersForUserID(user string) (servers map[string]*QpWhatsappServer) {
	return WhatsappService.GetServersForUser(user)
}

func GetServersForUser(user *QpUser) (servers map[string]*QpWhatsappServer) {
	return GetServersForUserID(user.Username)
}

// PostToWebhooksModern sends a message to all webhook endpoints using the modern dispatching system
func PostToWebhooksModern(server *QpWhatsappServer, message *whatsapp.WhatsappMessage) (err error) {
	if server == nil {
		err = fmt.Errorf("server nil")
		return err
	}

	// ignoring ssl issues
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	webhookDispatchings := server.GetWebhookDispatchings()

	for _, dispatching := range webhookDispatchings {
		// updating log
		logentry := dispatching.GetLogger()
		loglevel := logentry.Level
		logentry = logentry.WithField(LogFields.MessageId, message.Id)
		logentry.Level = loglevel

		// Check if should ignore based on message type and dispatching options
		if message.Id == "readreceipt" && dispatching.IsSetReadReceipts() && !dispatching.ReadReceipts.Boolean() {
			logentry.Debugf("ignoring read receipt message: %s", message.Text)
			continue
		}

		if message.FromGroup() && dispatching.IsSetGroups() && !dispatching.Groups.Boolean() {
			logentry.Debug("ignoring group message")
			continue
		}

		if message.FromDirect() && dispatching.IsSetDirect() && !dispatching.Direct.Boolean() {
			logentry.Debug("ignoring direct message")
			continue
		}

		if message.FromBroadcast() && dispatching.IsSetBroadcasts() && !dispatching.Broadcasts.Boolean() {
			logentry.Debug("ignoring broadcast message")
			continue
		}

		if message.Type == whatsapp.CallMessageType && dispatching.IsSetCalls() && !dispatching.Calls.Boolean() {
			logentry.Debug("ignoring call message")
			continue
		}

		// Send the message using the dispatching
		err = dispatching.PostWebhook(message)
		if err != nil {
			logentry.Errorf("error posting to webhook: %s", err.Error())
		}
	}

	return err
}

//endregion
