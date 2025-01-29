package models

type QPSendResult struct {
	Source    string `json:"source"`
	Recipient string `json:"recipient"`
	MessageId string `json:"messageId"`
}
