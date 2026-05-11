package models

import (
	"errors"
	"fmt"
	"strings"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// handle message deliver to individual dispatching distribution
func PostToDispatchingFromServer(server *QpWhatsappServer, message *whatsapp.WhatsappMessage) (err error) {
	return DispatchOutboundFromServer(server, message)
}

// PostToDispatchings delivers a message to the provided dispatching targets.
// It is used by the normal server flow and by deletion flows that need to use
// a preserved snapshot instead of the server's live dispatching slice.
func PostToDispatchings(server *QpWhatsappServer, dispatchings []*QpDispatching, message *whatsapp.WhatsappMessage) (err error) {
	return DispatchOutboundToTargets(server, dispatchings, message)
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
	return GetServerFromID(source.GetWId())
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

	return DispatchOutboundToTargets(server, server.GetWebhookDispatchings(), message)
}

//endregion
