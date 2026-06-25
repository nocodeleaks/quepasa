package models

import "time"

// QpTimestamps holds timestamp information for server activity
type QpTimestamps struct {
	Message *time.Time `json:"message,omitempty"` // Last received message timestamp
	Event   *time.Time `json:"event,omitempty"`   // Last received event timestamp
	Start   time.Time  `json:"start"`             // Server start timestamp
	Update  time.Time  `json:"update"`            // Last database update timestamp
}
