package api

import whatsapp "github.com/nocodeleaks/quepasa/whatsapp"

// InfoPatchRequest represents the request body for partial session updates.
type InfoPatchRequest struct {
	Groups       *whatsapp.WhatsappBoolean `db:"groups" json:"groups,omitempty"`
	Broadcasts   *whatsapp.WhatsappBoolean `db:"broadcasts" json:"broadcasts,omitempty"`
	ReadReceipts     *whatsapp.WhatsappBoolean `db:"readreceipts" json:"readreceipts,omitempty"`
	DeliveryReceipts *whatsapp.WhatsappBoolean `db:"deliveryreceipts" json:"deliveryreceipts,omitempty"`
	Calls            *whatsapp.WhatsappBoolean `db:"calls" json:"calls,omitempty"`
	ReadUpdate   *whatsapp.WhatsappBoolean `db:"readupdate" json:"readupdate,omitempty"`
	Direct       *whatsapp.WhatsappBoolean `db:"direct" json:"direct,omitempty"`
	HistorySyncDays *uint32                `db:"historysyncdays" json:"historysyncdays,omitempty"`
	Username     *string                   `json:"username,omitempty" validate:"max=255"`
	Devel        *bool                     `json:"devel,omitempty"`
	ContextId    *string                   `json:"contextid,omitempty" validate:"max=100"`
}