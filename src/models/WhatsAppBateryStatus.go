package models

import "time"

type WhatsAppBateryStatus struct {
	Timestamp  time.Time `json:"timestamp"`
	Plugged    bool      `json:"plugged"`
	Powersave  bool      `json:"powersave"`
	Percentage int       `json:"percentage"`
}
