package whatsapp

// Whatsapp service options, setted on start, so if want to changed then, you have to restart the entire service
type WhatsappOptions struct {

	// should handle groups messages
	Groups WhatsappBoolean `db:"groups" json:"groups,omitempty"`

	// should handle broadcast messages
	Broadcasts WhatsappBoolean `db:"broadcasts" json:"broadcasts,omitempty"`

	// should emit read receipts
	ReadReceipts WhatsappBoolean `db:"readreceipts" json:"readreceipts,omitempty"`

	// should handle calls
	Calls WhatsappBoolean `db:"calls" json:"calls,omitempty"`
}
