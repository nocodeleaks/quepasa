package models

type QpReceiveResponseV2 struct {
	QpResponse
	Messages []QpMessageV2 `json:"messages,omitempty"`
	Bot      QpServerV2    `json:"bot,omitempty"`
}
