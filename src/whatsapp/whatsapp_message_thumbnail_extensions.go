package whatsapp

import (
	"encoding/base64"
	"fmt"
	"strings"
)

// base64 contents common starts with these signatures
var ContentSignatures = map[string]string{
	"/9j/":        "image/jpg",
	"R0lGODdh":    "image/gif",
	"R0lGODlh":    "image/gif",
	"iVBORw0KGgo": "image/png",
	"JVBERi0":     "application/pdf",
}

func NewWhatsappMessageThumbnail(bytes []byte) *WhatsappMessageThumbnail {
	data := base64.StdEncoding.EncodeToString(bytes)
	return NewWhatsappMessageThumbnailFromBase64(data)
}

func NewWhatsappMessageThumbnailFromBase64(data string) *WhatsappMessageThumbnail {
	thumbnail := &WhatsappMessageThumbnail{}
	thumbnail.Data = data
	thumbnail.Mime = GetThumbnailMime(thumbnail.Data)
	thumbnail.UrlPrefix = GetThumbnailUrlPrefix(thumbnail.Mime)
	return thumbnail
}

func GetThumbnailUrlPrefix(mime string) string {
	return fmt.Sprintf("data:%s;base64,", mime)
}

func GetThumbnailMime(data string) string {
	for k, v := range ContentSignatures {
		if strings.HasPrefix(data, k) {
			return v
		}
	}
	return "application/octet-stream"
}
