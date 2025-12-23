package api

import whatsapp "github.com/nocodeleaks/quepasa/whatsapp"

// InfoCreateRequest represents the request body for creating a new bot/server
// Used for pre-configuring a server before QR code scanning
type InfoCreateRequest struct {
	Groups       *whatsapp.WhatsappBoolean `json:"groups,omitempty"`       // should handle groups messages
	Broadcasts   *whatsapp.WhatsappBoolean `json:"broadcasts,omitempty"`   // should handle broadcast messages
	ReadReceipts *whatsapp.WhatsappBoolean `json:"readreceipts,omitempty"` // should emit read receipts
	Calls        *whatsapp.WhatsappBoolean `json:"calls,omitempty"`        // should handle calls
	ReadUpdate   *whatsapp.WhatsappBoolean `json:"readupdate,omitempty"`   // should send markread requests when receiving messages
	Devel        *bool                     `json:"devel,omitempty"`        // enable debug mode (devel)
}
