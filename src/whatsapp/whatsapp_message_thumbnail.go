package whatsapp

// small image representing something in this message, MIME: image/jpeg
type WhatsappMessageThumbnail struct {

	// base64 data
	Data string `json:"data,omitempty"`

	// content mime type
	Mime string `json:"mime,omitempty"`

	// trick for '<img src=' urls prefix
	UrlPrefix string `json:"urlprefix,omitempty"`
}

func (source *WhatsappMessageThumbnail) GetThumbnailAsUrl() string {
	if len(source.UrlPrefix) == 0 {
		mime := source.GetThumbnailMime()
		source.UrlPrefix = GetThumbnailUrlPrefix(mime)
	}
	return source.UrlPrefix + source.Data
}

func (source *WhatsappMessageThumbnail) GetThumbnailMime() string {
	if len(source.Mime) == 0 {
		source.Mime = GetThumbnailMime(source.Data)
	}
	return source.Mime
}
