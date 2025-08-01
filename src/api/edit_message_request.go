package api

type EditMessageRequest struct {
	MessageId string `json:"messageId"` // Required: Message ID to edit
	Content   string `json:"content"`   // Required: New content for the message
}
