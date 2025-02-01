package whatsapp

type WhatsappMessageUrl struct {
	Reference   string `json:"reference,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`

	// small image representing something in this message, MIME: image/jpeg
	Thumbnail *WhatsappMessageThumbnail `json:"thumbnail,omitempty"`
}

func (source *WhatsappMessageUrl) String() string {
	return source.Reference
}
