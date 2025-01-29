package whatsapp

type WhatsappEndpoint struct {
	ID        string `json:"id,omitempty"`
	UserName  string `json:"username,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Title     string `json:"title,omitempty"`
}
