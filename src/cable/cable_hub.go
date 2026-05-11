package cable

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	dispatchservice "github.com/nocodeleaks/quepasa/dispatch/service"
	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 1024 * 1024
	sendQueueSize  = 128
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := strings.TrimSpace(r.Header.Get("Origin"))
		if origin == "" {
			return true
		}

		return strings.Contains(origin, r.Host)
	},
}

// CommandHandler is the signature implemented by websocket commands.
type CommandHandler func(client *Client, command ClientCommand) (interface{}, error)

// Hub keeps the full realtime state for websocket cable connections.
//
// Indices are intentionally redundant:
// - clients: direct lookup by connection id
// - userClients: fan-out to every live connection owned by the same user
// - serverClients: server-token subscriptions used for targeted message streams
type Hub struct {
	mu            sync.RWMutex
	clients       map[string]*Client
	userClients   map[string]map[string]*Client
	serverClients map[string]map[string]*Client
	commands      map[string]CommandHandler
}

// Client represents one websocket connection, not one user.
//
// A single user can open many tabs/devices simultaneously, and each connection
// has its own send queue and subscription set.
type Client struct {
	id            string
	user          *models.QpUser
	conn          *websocket.Conn
	hub           *Hub
	send          chan []byte
	closeOnce     sync.Once
	closed        atomic.Bool // Track if client has been closed to prevent sends on closed channel
	subscriptions map[string]struct{}
}

// NewHub creates a ready-to-use cable hub with the default command set.
func NewHub() *Hub {
	hub := &Hub{
		clients:       map[string]*Client{},
		userClients:   map[string]map[string]*Client{},
		serverClients: map[string]map[string]*Client{},
		commands:      map[string]CommandHandler{},
	}

	hub.registerDefaultCommands()
	return hub
}

// ServeWS upgrades the request and starts the client read/write pumps.
func (hub *Hub) ServeWS(w http.ResponseWriter, r *http.Request, user *models.QpUser) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.WithError(err).Warn("cable upgrade failed")
		return
	}

	client := &Client{
		id:            uuid.NewString(),
		user:          user,
		conn:          conn,
		hub:           hub,
		send:          make(chan []byte, sendQueueSize),
		subscriptions: map[string]struct{}{},
	}

	hub.addClient(client)
	client.sendEvent("session.ready", "", SessionReadyPayload{
		ConnectionID:  client.id,
		User:          user.Username,
		Subscriptions: []string{},
		Commands: []string{
			"ping",
			"subscribe",
			"unsubscribe",
			"server.enable",
			"server.disable",
			"message.send",
			"message.edit",
			"message.revoke",
			"chat.archive",
			"chat.presence",
		},
	})

	go client.writePump()
	client.readPump()
}

func (hub *Hub) addClient(client *Client) {
	hub.mu.Lock()
	defer hub.mu.Unlock()

	hub.clients[client.id] = client

	userClients := hub.userClients[client.user.Username]
	if userClients == nil {
		userClients = map[string]*Client{}
		hub.userClients[client.user.Username] = userClients
	}
	userClients[client.id] = client
}

func (hub *Hub) removeClient(client *Client) {
	if client == nil {
		return
	}

	client.closeOnce.Do(func() {
		// Mark client as closed BEFORE closing the channel to prevent concurrent sends
		client.closed.Store(true)

		hub.mu.Lock()
		defer hub.mu.Unlock()

		delete(hub.clients, client.id)

		if userClients := hub.userClients[client.user.Username]; userClients != nil {
			delete(userClients, client.id)
			if len(userClients) == 0 {
				delete(hub.userClients, client.user.Username)
			}
		}

		for token := range client.subscriptions {
			if subscribers := hub.serverClients[token]; subscribers != nil {
				delete(subscribers, client.id)
				if len(subscribers) == 0 {
					delete(hub.serverClients, token)
				}
			}
		}

		close(client.send)
		_ = client.conn.Close()
	})
}

// SubscribeServer attaches the client to a server token topic after ownership has
// already been validated by the caller.
func (hub *Hub) SubscribeServer(client *Client, token string) {
	token = normalizeToken(token)
	if token == "" {
		return
	}

	hub.mu.Lock()
	defer hub.mu.Unlock()

	subscribers := hub.serverClients[token]
	if subscribers == nil {
		subscribers = map[string]*Client{}
		hub.serverClients[token] = subscribers
	}

	subscribers[client.id] = client
	client.subscriptions[token] = struct{}{}
}

// UnsubscribeServer removes a client from a server topic.
func (hub *Hub) UnsubscribeServer(client *Client, token string) {
	token = normalizeToken(token)
	if token == "" {
		return
	}

	hub.mu.Lock()
	defer hub.mu.Unlock()

	delete(client.subscriptions, token)

	if subscribers := hub.serverClients[token]; subscribers != nil {
		delete(subscribers, client.id)
		if len(subscribers) == 0 {
			delete(hub.serverClients, token)
		}
	}
}

func (hub *Hub) getSubscriptions(client *Client) []string {
	hub.mu.RLock()
	defer hub.mu.RUnlock()

	items := make([]string, 0, len(client.subscriptions))
	for token := range client.subscriptions {
		items = append(items, serverTopic(token))
	}
	return items
}

func (hub *Hub) queueFrame(client *Client, frame ServerFrame) {
	// Check if client has been closed to prevent sending on closed channel
	if client.closed.Load() {
		return
	}

	frame.Timestamp = time.Now().UTC()
	payload, err := json.Marshal(frame)
	if err != nil {
		log.WithError(err).Warn("failed to marshal cable frame")
		return
	}

	select {
	case client.send <- payload:
	default:
		log.WithField("client_id", client.id).Warn("closing slow cable client")
		hub.removeClient(client)
	}
}

func (hub *Hub) sendResponse(client *Client, command ClientCommand, data interface{}, err error) {
	frame := ServerFrame{
		Type:    "response",
		ID:      command.ID,
		Command: command.Command,
		OK:      err == nil,
		Data:    data,
	}

	if err != nil {
		frame.Error = &ProtocolError{
			Code:    "command_error",
			Message: err.Error(),
		}
	}

	hub.queueFrame(client, frame)
}

func (hub *Hub) sendEventToClient(client *Client, event string, topic string, data interface{}) {
	hub.queueFrame(client, ServerFrame{
		Type:  "event",
		Event: event,
		Topic: topic,
		Data:  data,
	})
}

func (hub *Hub) sendEventToUser(username string, event string, topic string, data interface{}) {
	hub.mu.RLock()
	recipients := make([]*Client, 0)
	for _, client := range hub.userClients[username] {
		recipients = append(recipients, client)
	}
	hub.mu.RUnlock()

	for _, client := range recipients {
		hub.sendEventToClient(client, event, topic, data)
	}
}

func (hub *Hub) sendEventToServer(token string, event string, data interface{}) {
	token = normalizeToken(token)

	hub.mu.RLock()
	recipients := make([]*Client, 0)
	for _, client := range hub.serverClients[token] {
		recipients = append(recipients, client)
	}
	hub.mu.RUnlock()

	for _, client := range recipients {
		hub.sendEventToClient(client, event, serverTopic(token), data)
	}
}

// PublishMessage implements dispatchservice.RealtimePublisher.
func (hub *Hub) PublishMessage(payload interface{}) {
	event, ok := payload.(*dispatchservice.RealtimeServerMessage)
	if !ok || event == nil || event.Message == nil {
		return
	}

	hub.sendEventToServer(event.Token, "server.message", ServerMessageEventPayload{
		Token:   event.Token,
		User:    event.User,
		WID:     event.WID,
		State:   event.State,
		Message: event.Message,
	})
}

// PublishLifecycle implements dispatchservice.RealtimePublisher.
func (hub *Hub) PublishLifecycle(payload interface{}) {
	event, ok := payload.(*dispatchservice.RealtimeLifecycleEvent)
	if !ok || event == nil {
		return
	}

	eventName := "server." + event.Kind
	if event.User != "" {
		hub.sendEventToUser(event.User, eventName, "", event)
	}
	if event.Token != "" {
		hub.sendEventToServer(event.Token, eventName, event)
	}
}

// PublishServerMessage keeps compatibility with existing tests/helpers.
func (hub *Hub) PublishServerMessage(server *models.QpWhatsappServer, payload *whatsapp.WhatsappMessage) {
	if server == nil || payload == nil {
		return
	}

	enriched := models.CloneAndEnrichMessageForServer(server, payload)
	hub.PublishMessage(&dispatchservice.RealtimeServerMessage{
		Token:   server.Token,
		User:    server.GetUser(),
		WID:     server.GetWId(),
		State:   server.GetState().String(),
		Message: enriched,
	})
}

// PublishServerLifecycle keeps compatibility with existing tests/helpers.
func (hub *Hub) PublishServerLifecycle(event *models.RealtimeLifecycleEvent) {
	if event == nil {
		return
	}

	hub.PublishLifecycle((*dispatchservice.RealtimeLifecycleEvent)(event))
}

func (client *Client) readPump() {
	defer client.hub.removeClient(client)

	client.conn.SetReadLimit(maxMessageSize)
	_ = client.conn.SetReadDeadline(time.Now().Add(pongWait))
	client.conn.SetPongHandler(func(string) error {
		return client.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, payload, err := client.conn.ReadMessage()
		if err != nil {
			return
		}

		var command ClientCommand
		if err := json.Unmarshal(payload, &command); err != nil {
			client.hub.queueFrame(client, ServerFrame{
				Type: "response",
				OK:   false,
				Error: &ProtocolError{
					Code:    "invalid_json",
					Message: err.Error(),
				},
			})
			continue
		}

		client.hub.handleCommand(client, command)
	}
}

func (client *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.hub.removeClient(client)
	}()

	for {
		select {
		case payload, ok := <-client.send:
			_ = client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.conn.WriteMessage(websocket.TextMessage, payload); err != nil {
				return
			}

		case <-ticker.C:
			_ = client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (client *Client) sendEvent(event string, topic string, data interface{}) {
	client.hub.sendEventToClient(client, event, topic, data)
}

func normalizeToken(token string) string {
	return strings.ToLower(strings.TrimSpace(token))
}
