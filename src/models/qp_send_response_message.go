package models

type QpSendResponseMessage struct {
	Id      string `json:"id,omitempty"`
	Wid     string `json:"wid,omitempty"`
	ChatId  string `json:"chatId,omitempty"`
	TrackId string `json:"trackId,omitempty"`
}
