package models

import "github.com/nocodeleaks/quepasa/whatsapp"

type QpContactsResponse struct {
	QpResponse
	Total    int                     `json:"total"`
	Contacts []whatsapp.WhatsappChat `json:"contacts,omitempty"`
}
