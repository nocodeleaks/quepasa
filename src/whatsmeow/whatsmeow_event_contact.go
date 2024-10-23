package whatsmeow

import (
	"time"
)

type ContactInfo struct {
	JID                      string    `json:"jid"`
	FullName                 string    `json:"full_name,omitempty"`
	FirstName                string    `json:"first_name,omitempty"`
	LidJID                   string    `json:"lid_jid,omitempty"`
	SaveOnPrimaryAddressbook bool      `json:"save_on_primary_addressbook,omitempty"`
	FromFullSync             bool      `json:"from_full_sync"`
	Timestamp                time.Time `json:"timestamp"`
}
