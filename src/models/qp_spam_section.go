package models

import "time"

// QpSpamSection stores the ordered set of WhatsApp sections allowed for /spam.
type QpSpamSection struct {
	Token     string    `db:"token" json:"token"`
	Position  int       `db:"position" json:"position"`
	Enabled   bool      `db:"enabled" json:"enabled"`
	Label     string    `db:"label" json:"label,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"createdAt,omitempty"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt,omitempty"`
}
