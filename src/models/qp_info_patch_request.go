package models

import whatsapp "github.com/nocodeleaks/quepasa/whatsapp"

// Information Request Body
type QpInfoPatchRequest struct {
	Groups       *whatsapp.WhatsappBoolean `db:"groups" json:"groups,omitempty"`             // should handle groups messages
	Broadcasts   *whatsapp.WhatsappBoolean `db:"broadcasts" json:"broadcasts,omitempty"`     // should handle broadcast messages
	ReadReceipts *whatsapp.WhatsappBoolean `db:"readreceipts" json:"readreceipts,omitempty"` // should emit read receipts
	Calls        *whatsapp.WhatsappBoolean `db:"calls" json:"calls,omitempty"`               // should handle calls
	Username     *string                   `json:"username,omitempty" validate:"max=255"`
}
