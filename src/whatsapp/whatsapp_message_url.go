package whatsapp

type WhatsappMessageUrl struct {
	Reference   string `json:"reference,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

func (source *WhatsappMessageUrl) String() string {
	return source.Reference
}
