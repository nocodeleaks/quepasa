package whatsmeow

import (
	"time"
)

type GroupJoinInfo struct {
	Owner        string    `json:"owner,omitempty"`
	Created      time.Time `json:"time,omitempty"`
	Participants int       `json:"participants,omitempty"`
	Type         string    `json:"type,omitempty"`
	Reason       string    `json:"reason,omitempty"`
}
