package api

type LeaveGroupRequest struct {
	ChatId string `json:"chatId"` // Required: Group Chat ID to leave
}
