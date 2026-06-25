package whatsapp

type WhatsappPoll struct {
	Question   string   `json:"question"`             // Required: Poll question/title
	Options    []string `json:"options"`              // Required: Array of poll options
	Selections uint     `json:"selections,omitempty"` // Optional: Maximum number of options a user can select (default: 1)
}
