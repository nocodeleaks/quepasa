package models

// Response basic defaults
// Model response of a Invite Request
type QpInviteResponse struct {
	QpResponse
	Url string `json:"url,omitempty"` // invite public link
}
