package models

import (
	"fmt"
	"strings"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	signalr "github.com/philippseith/signalr"
)

type QpSignalRHub struct {
	signalr.Hub
	tokens map[string]string
	proxy  map[string]signalr.ClientProxy
}

var SignalRHub = &QpSignalRHub{
	tokens: map[string]string{},
	proxy:  map[string]signalr.ClientProxy{},
}

// not used
func SignalRHubFactory() signalr.HubInterface {
	return SignalRHub
}

func (source *QpSignalRHub) IsInterfaceNil() bool {
	return source == nil
}

func (source *QpSignalRHub) OnConnected(ConnectionId string) {
	info, _ := source.Logger()
	info.Log("connection", ConnectionId, "status", "connected")

	source.proxy[ConnectionId] = source.Clients().Caller()
}

func (source *QpSignalRHub) OnDisconnected(ConnectionId string) {
	info, _ := source.Logger()
	info.Log("connection", ConnectionId, "status", "disconnected")

	delete(source.tokens, ConnectionId)
	delete(source.proxy, ConnectionId)
}

func (source *QpSignalRHub) TrySend(ConnectionId string, target string, args ...interface{}) {
	if source == nil {
		return
	}

	proxy := source.proxy[ConnectionId]
	if proxy != nil {
		proxy.Send(target, args...)
	}
}

func (source *QpSignalRHub) GetToken() string {
	ConnectionId := source.ConnectionID()
	token := source.tokens[ConnectionId]

	message := fmt.Sprintf("connection id: %s, token: %s", ConnectionId, token)
	source.Clients().Caller().Send(ConnectionId, "system", message)
	return token
}

func (source *QpSignalRHub) Token(token string) {
	ConnectionId := source.ConnectionID()
	source.tokens[ConnectionId] = token

	info, _ := source.Logger()
	info.Log("connection", ConnectionId, "token", token)
}

func (source *QpSignalRHub) GetActiveConnections(token string) (active []string) {
	if source != nil {
		masterkey := ENV.MasterKey()
		for ConnectionId, _token := range source.tokens {
			if strings.EqualFold(masterkey, _token) || _token == token {
				active = append(active, ConnectionId)
			}
		}
	}

	return
}

func (source *QpSignalRHub) Dispatch(token string, payload *whatsapp.WhatsappMessage) {
	for _, ConnectionId := range source.GetActiveConnections(token) {
		source.TrySend(ConnectionId, "message", payload)
	}
}

func (source *QpSignalRHub) HasActiveConnections(token string) bool {
	connections := source.GetActiveConnections(token)
	return len(connections) > 0
}
