package models

type PollRequest struct {
	ChatId        string   `json:"chat_id"`                  // Required: Chat to send the poll to
	Question      string   `json:"question"`                 // Required: Poll question/title
	Options       []string `json:"options"`                  // Required: Array of poll options
	MaxSelections int      `json:"max_selections,omitempty"` // Optional: Maximum number of options a user can select (default: 1)
	TrackId       string   `json:"track_id,omitempty"`       // Optional: For tracking the message
}
