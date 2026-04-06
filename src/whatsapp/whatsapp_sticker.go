package whatsapp

// WhatsappSticker represents a sticker to be sent.
// Content can be any image or video format — it will be converted to WebP automatically.
// Provide either Url (remote) or Content (base64 data URI).
type WhatsappSticker struct {
	Url     string `json:"url,omitempty"`     // Remote URL to download the image/video
	Content string `json:"content,omitempty"` // Base64 data URI (e.g. data:image/png;base64,...)
}
