package models

import (
	"time"
)

type QpUser struct {
	Username  string    `db:"username" json:"username" validate:"max=255"`
	Password  string    `db:"password" json:"password" validate:"max=255"`
	UI        *string   `db:"ui" json:"ui,omitempty"`
	Timestamp time.Time `db:"timestamp" json:"timestamp,omitempty"`
}

// QpUserUI holds the persisted UI preferences for a user.
type QpUserUI struct {
	ViewMode string `json:"viewMode,omitempty"`
	PageSize int    `json:"pageSize,omitempty"`
	Theme    string `json:"theme,omitempty"`
	Locale   string `json:"locale,omitempty"`
}
