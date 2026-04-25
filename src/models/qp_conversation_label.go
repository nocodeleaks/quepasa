package models

import (
	"strings"
	"time"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

type QpConversationLabel struct {
	ID        int64     `db:"id" json:"id"`
	User      string    `db:"user" json:"user,omitempty" validate:"max=255"`
	Name      string    `db:"name" json:"name" validate:"max=100"`
	Color     string    `db:"color" json:"color,omitempty" validate:"max=32"`
	Active    bool      `db:"active" json:"active"`
	Timestamp time.Time `db:"timestamp" json:"timestamp,omitempty"`
}

func (source *QpConversationLabel) Normalize() {
	if source == nil {
		return
	}

	source.User = strings.TrimSpace(source.User)
	source.Name = strings.TrimSpace(source.Name)
	source.Color = strings.TrimSpace(source.Color)
}

func (source *QpConversationLabel) ToWhatsappLabel() whatsapp.WhatsappChatLabel {
	if source == nil {
		return whatsapp.WhatsappChatLabel{}
	}

	return whatsapp.WhatsappChatLabel{
		ID:     source.ID,
		Name:   source.Name,
		Color:  source.Color,
		Active: source.Active,
	}
}
