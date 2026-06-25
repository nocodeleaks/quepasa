package models

import "time"

type QpWebhookInterface interface {
	GetUrl() string
	GetFailure() *time.Time
}
