package api

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	apiModels "github.com/nocodeleaks/quepasa/api/models"
	media "github.com/nocodeleaks/quepasa/media"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
)

// ResolveStickerAttachment downloads or decodes the sticker content and
// converts it to WebP format suitable for WhatsApp sticker messages.
// Returns a fully populated WhatsappAttachment ready for sending.
func ResolveStickerAttachment(sticker *apiModels.WhatsappSticker) (*whatsapp.WhatsappAttachment, error) {
	if sticker == nil {
		return nil, fmt.Errorf("sticker is nil")
	}

	var rawData []byte
	var inputMime string

	switch {
	case len(sticker.Url) > 0:
		// Download from remote URL
		data, mime, err := downloadStickerFromURL(sticker.Url)
		if err != nil {
			return nil, fmt.Errorf("error downloading sticker from URL: %w", err)
		}
		rawData = data
		inputMime = mime

	case len(sticker.Content) > 0:
		// Decode base64 or data URI
		data, mime, err := decodeStickerContent(sticker.Content)
		if err != nil {
			return nil, fmt.Errorf("error decoding sticker content: %w", err)
		}
		rawData = data
		inputMime = mime

	default:
		return nil, fmt.Errorf("sticker has no url or content")
	}

	if len(rawData) == 0 {
		return nil, fmt.Errorf("sticker resolved to empty content")
	}

	// If already WebP, skip conversion
	mimeOnly := strings.Split(inputMime, ";")[0]
	var webpData []byte
	var outputMime string

	if mimeOnly == "image/webp" || mimeOnly == "video/webp" {
		log.Infof("[ResolveStickerAttachment] content is already WebP (%s), skipping conversion", inputMime)
		webpData = rawData
		outputMime = mimeOnly
	} else {
		// Convert to WebP via FFmpeg
		converted, mime, err := media.ConvertToWebP(rawData, inputMime)
		if err != nil {
			return nil, fmt.Errorf("error converting sticker to WebP: %w", err)
		}
		webpData = converted
		outputMime = mime
	}

	attach := &whatsapp.WhatsappAttachment{
		Mimetype:   outputMime,
		FileLength: uint64(len(webpData)),
		FileName:   "sticker.webp",
	}
	attach.SetContent(&webpData)

	log.Infof("[ResolveStickerAttachment] sticker resolved: mime=%s, size=%d bytes", outputMime, len(webpData))
	return attach, nil
}

// downloadStickerFromURL downloads content from a URL and returns raw bytes and MIME type.
func downloadStickerFromURL(url string) ([]byte, string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, "", fmt.Errorf("error fetching URL %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("unexpected status code %d fetching sticker URL", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("error reading sticker response body: %w", err)
	}

	mime := resp.Header.Get("Content-Type")
	if len(mime) == 0 {
		mime = "application/octet-stream"
	}

	return data, mime, nil
}

// decodeStickerContent decodes a plain base64 string or a data URI into raw bytes and MIME type.
func decodeStickerContent(content string) ([]byte, string, error) {
	var mime string
	var b64 string

	if strings.HasPrefix(content, "data:") {
		parts := strings.SplitN(content, ",", 2)
		if len(parts) != 2 {
			return nil, "", fmt.Errorf("invalid data URI format for sticker content")
		}

		header := parts[0] // e.g. "data:image/png;base64"
		b64 = parts[1]

		// Extract MIME type from header
		headerInner := strings.TrimPrefix(header, "data:")
		if idx := strings.LastIndex(headerInner, ";base64"); idx >= 0 {
			mime = strings.TrimSpace(headerInner[:idx])
		} else {
			mime = strings.TrimSpace(headerInner)
		}
	} else {
		b64 = content
		mime = "application/octet-stream"
	}

	decoded, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, "", fmt.Errorf("error decoding base64 sticker content: %w", err)
	}

	return decoded, mime, nil
}
