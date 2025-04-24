package models

import "github.com/nocodeleaks/quepasa/whatsapp"

type PollRequest struct {
	whatsapp.WhatsappPoll
	ChatId  string `json:"chat_id"`           // Required: Chat to send the poll to
	TrackId string `json:"trackid,omitempty"` // Optional: For tracking the message
}
