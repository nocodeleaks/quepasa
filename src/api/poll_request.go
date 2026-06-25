package api

import "github.com/nocodeleaks/quepasa/whatsapp"

// PollRequest represents the request body for poll send operations.
type PollRequest struct {
	whatsapp.WhatsappPoll
	ChatId  string `json:"chat_id"`
	TrackId string `json:"trackid,omitempty"`
}
