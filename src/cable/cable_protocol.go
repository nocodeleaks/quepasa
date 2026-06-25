package cable

import (
	"encoding/json"
	"time"
)

// ClientCommand is the inbound frame format used by the browser/app to talk to
// the websocket cable transport.
//
// The command envelope is intentionally compact:
// - id: client-generated correlation identifier
// - command: stable command name such as "subscribe" or "message.send"
// - data: command-specific payload
type ClientCommand struct {
	Type    string          `json:"type,omitempty"`
	ID      string          `json:"id,omitempty"`
	Command string          `json:"command"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// ProtocolError is returned in response frames when a command cannot be handled.
type ProtocolError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ServerFrame is the single outbound frame format used both for command
// responses and for unsolicited events.
//
// Clients should branch on:
// - type == "response": answer for a previously sent command
// - type == "event": broadcast pushed by the backend
type ServerFrame struct {
	Type      string         `json:"type"`
	ID        string         `json:"id,omitempty"`
	Command   string         `json:"command,omitempty"`
	Event     string         `json:"event,omitempty"`
	Topic     string         `json:"topic,omitempty"`
	OK        bool           `json:"ok,omitempty"`
	Error     *ProtocolError `json:"error,omitempty"`
	Data      interface{}    `json:"data,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

// serverTopic is the canonical subscription topic for a WhatsApp server token.
func serverTopic(token string) string {
	return "server:" + token
}
