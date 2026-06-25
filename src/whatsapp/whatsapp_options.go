package whatsapp

// Whatsapp service options, setted on start, so if want to changed then, you have to restart the entire service
type WhatsappOptions struct {

	// should handle groups messages
	Groups WhatsappBoolean `db:"groups" json:"groups,omitempty"`

	// should handle broadcast messages
	Broadcasts WhatsappBoolean `db:"broadcasts" json:"broadcasts,omitempty"`

	// should emit read receipts
	ReadReceipts WhatsappBoolean `db:"readreceipts" json:"readreceipts,omitempty"`

	// should emit delivery receipts
	DeliveryReceipts WhatsappBoolean `db:"deliveryreceipts" json:"deliveryreceipts,omitempty"`

	// should handle calls
	Calls WhatsappBoolean `db:"calls" json:"calls,omitempty"`

	// should send markread requests when receiving messages
	ReadUpdate WhatsappBoolean `db:"readupdate" json:"readupdate,omitempty"`

	// should handle direct (individual) messages (@s.whatsapp.net and @lid); default true
	Direct WhatsappBoolean `db:"direct" json:"direct,omitempty"`

	// VoIPMode controls inbound WhatsApp call handling for this instance:
	// disabled (default reject/relay), exclusive (SIP only, hang up WhatsApp on
	// SIP failure), or additional (SIP as extra device, leave call ringing).
	// Persisted in server metadata, not as a dedicated DB column.
	VoIPMode VoIPMode `json:"voipmode,omitempty"`
}
