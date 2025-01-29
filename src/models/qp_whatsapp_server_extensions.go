package models

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strings"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// handle message deliver to individual webhook distribution
func PostToWebHookFromServer(server *QpWhatsappServer, message *whatsapp.WhatsappMessage) (err error) {
	if server == nil {
		err = fmt.Errorf("server nil")
		return err
	}

	// ignoring ssl issues
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	for _, element := range server.Webhooks {

		// updating log
		logentry := element.GetLogger()
		loglevel := logentry.Level
		logentry = logentry.WithField(LogFields.MessageId, message.Id)
		logentry.Level = loglevel

		if message.Id == "readreceipt" && element.IsSetReadReceipts() && !element.ReadReceipts.Boolean() {
			logentry.Debugf("ignoring read receipt message: %s", message.Text)
			continue
		}

		if message.FromGroup() && element.IsSetGroups() && !element.Groups.Boolean() {
			logentry.Debug("ignoring group message")
			continue
		}

		if message.FromBroadcast() && element.IsSetBroadcasts() && !element.Broadcasts.Boolean() {
			logentry.Debug("ignoring broadcast message")
			continue
		}

		if message.Type == whatsapp.CallMessageType && element.IsSetCalls() && !element.Calls.Boolean() {
			logentry.Debug("ignoring call message")
			continue
		}

		if !message.FromInternal || (element.ForwardInternal && (len(element.TrackId) == 0 || element.TrackId != message.TrackId)) {
			elerr := element.Post(message)
			if elerr != nil {
				logentry.Errorf("error on post webhook: %s", elerr.Error())
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

//endregion
