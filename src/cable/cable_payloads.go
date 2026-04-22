package cable

import whatsapp "github.com/nocodeleaks/quepasa/whatsapp"

// SessionReadyPayload is the first event delivered after the websocket session
// becomes authenticated and registered in the hub.
type SessionReadyPayload struct {
	ConnectionID  string   `json:"connectionId"`
	User          string   `json:"user"`
	Subscriptions []string `json:"subscriptions"`
	Commands      []string `json:"commands"`
}

// PingResponsePayload returns the current session snapshot for health checks and
// client-side reconnect validation.
type PingResponsePayload struct {
	ConnectionID  string   `json:"connectionId"`
	User          string   `json:"user"`
	Subscriptions []string `json:"subscriptions"`
}

// SubscriptionResponsePayload is used by subscribe/unsubscribe commands to
// confirm the effective topic set after normalization and authorization.
type SubscriptionResponsePayload struct {
	Subscriptions []string `json:"subscriptions"`
	Removed       []string `json:"removed,omitempty"`
}

// ServerStatePayload summarizes the server identity and current lifecycle state
// after state-changing commands such as enable/disable.
type ServerStatePayload struct {
	Token    string `json:"token"`
	User     string `json:"user"`
	WID      string `json:"wid"`
	State    string `json:"state"`
	Verified bool   `json:"verified"`
}

// SendMessageResponsePayload is the command response emitted after a successful
// `message.send` request.
type SendMessageResponsePayload struct {
	ID      string `json:"id"`
	ChatID  string `json:"chatId"`
	TrackID string `json:"trackId"`
	WID     string `json:"wid"`
	Token   string `json:"token"`
}

// ServerMessageEventPayload is the push event delivered to subscribers whenever
// a server emits a message through the realtime publisher bridge.
type ServerMessageEventPayload struct {
	Token   string                    `json:"token"`
	User    string                    `json:"user"`
	WID     string                    `json:"wid"`
	State   string                    `json:"state"`
	Message *whatsapp.WhatsappMessage `json:"message"`
}
