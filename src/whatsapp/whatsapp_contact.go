package whatsapp

type WhatsappContact struct {
	Phone string `json:"phone"`           // Required: Contact phone number
	Name  string `json:"name"`            // Required: Contact display name
	Vcard string `json:"vcard,omitempty"` // Optional: Full vCard string
}
