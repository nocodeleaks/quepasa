package api

import models "github.com/nocodeleaks/quepasa/models"

// SendResponseMessage is the nested payload returned after a successful send.
type SendResponseMessage struct {
	Id      string `json:"id,omitempty"`
	Wid     string `json:"wid,omitempty"`
	ChatId  string `json:"chatId,omitempty"`
	TrackId string `json:"trackId,omitempty"`
}

// SendResponse is the API transport shape for send endpoints.
type SendResponse struct {
	models.QpResponse
	Message *SendResponseMessage `json:"message,omitempty"`
}

// ParseSuccess fills the send response with the standard QuePasa success message.
func (source *SendResponse) ParseSuccess(message *SendResponseMessage) {
	source.QpResponse.ParseSuccess("sended with success")
	source.Message = message
}
