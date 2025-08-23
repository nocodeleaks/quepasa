package controllers

type EditMessageRequest struct {
	MessageId string `json:"messageId"` // Required: Message ID to edit
	Content   string `json:"content"`   // New content for the message
}
