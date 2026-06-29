package models

import (
	"time"
)

type QpUser struct {
	Username  string    `db:"username" json:"username" validate:"max=255"`
	Password  string    `db:"password" json:"password" validate:"max=255"`
	UI        *string   `db:"ui" json:"ui,omitempty"`
	Timestamp time.Time `db:"timestamp" json:"timestamp,omitempty"`

	// APIKey holds the SHA-256 hash (hex) of the user's personal API key, used to
	// authenticate access to that user's WhatsApp sessions independently of the
	// admin master key. Never the plaintext key. Nil = no key set.
	APIKey *string `db:"apikey" json:"-"`
	// APIKeyRotatedAt records when the API key was last generated/rotated.
	APIKeyRotatedAt *time.Time `db:"apikey_rotated_at" json:"apikey_rotated_at,omitempty"`
}

// QpUserUI holds the persisted UI preferences for a user.
type QpUserUI struct {
	ViewMode string `json:"viewMode,omitempty"`
	PageSize int    `json:"pageSize,omitempty"`
	Theme    string `json:"theme,omitempty"`
	Locale   string `json:"locale,omitempty"`
}
