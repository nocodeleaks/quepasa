package models

import (
	"crypto/tls"
	"errors"
	"net/http"
	"strings"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// Encaminha msg ao WebHook específicado
func PostToWebHookFromServer(server *QpWhatsappServer, message *whatsapp.WhatsappMessage) (err error) {
	wid := server.GetWid()

	// Ignorando certificado ao realizar o post
	// Não cabe a nós a segurança do cliente
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	for _, element := range server.Webhooks {
		if !message.FromInternal || (element.ForwardInternal && (len(element.TrackId) == 0 || element.TrackId != message.TrackId)) {
			element.Post(wid, message)
		}
	}

	return
}

//region FIND|SEARCH WHATSAPP SERVER
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
	return GetServerFromID(source.WId)
}

func GetServerFromToken(token string) (server *QpWhatsappServer, err error) {
	for _, item := range WhatsappService.Servers {
		if item != nil && strings.ToLower(item.Token) == strings.ToLower(token) {
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
