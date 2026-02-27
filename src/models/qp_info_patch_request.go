package models

import whatsapp "github.com/nocodeleaks/quepasa/whatsapp"

// Information Request Body
type QpInfoPatchRequest struct {
	Groups       *whatsapp.WhatsappBoolean `db:"groups" json:"groups,omitempty"`             // should handle groups messages
	Direct       *whatsapp.WhatsappBoolean `db:"direct" json:"direct,omitempty"`             // should handle direct messages
	Individuals  *whatsapp.WhatsappBoolean `db:"-" json:"individuals,omitempty"`             // deprecated alias for direct
	Broadcasts   *whatsapp.WhatsappBoolean `db:"broadcasts" json:"broadcasts,omitempty"`     // should handle broadcast messages
	ReadReceipts *whatsapp.WhatsappBoolean `db:"readreceipts" json:"readreceipts,omitempty"` // should emit read receipts
	Calls        *whatsapp.WhatsappBoolean `db:"calls" json:"calls,omitempty"`               // should handle calls
	ReadUpdate   *whatsapp.WhatsappBoolean `db:"readupdate" json:"readupdate,omitempty"`     // should send markread requests when receiving messages
	Username     *string                   `json:"username,omitempty" validate:"max=255"`
	Devel        *bool                     `json:"devel,omitempty"` // enable debug mode (devel)
}

func (source *QpInfoPatchRequest) GetDirect() *whatsapp.WhatsappBoolean {
	if source == nil {
		return nil
	}
	if source.Direct != nil {
		return source.Direct
	}
	return source.Individuals
}
