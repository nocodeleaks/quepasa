package models

import (
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// Mensagem no formato QuePasa
// Utilizada na API do QuePasa para troca com outros sistemas
type QPAttachment struct {
	whatsapp.WhatsappAttachment

	// Public URL to direct download without encryption
	DirectPath string `json:"url,omitempty"`
}
