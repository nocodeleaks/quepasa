package models

import (
	"time"
)

type QpUser struct {
	Username  string    `db:"username" json:"username" validate:"max=255"`
	Password  string    `db:"password" json:"password" validate:"max=255"`
	Timestamp time.Time `db:"timestamp" json:"timestamp,omitempty"`
}
