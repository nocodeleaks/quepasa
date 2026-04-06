package api

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	media "github.com/nocodeleaks/quepasa/media"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// ResolveStickerAttachment downloads or decodes the sticker content,
// converts it to WebP using FFmpeg, and returns a ready-to-upload attachment.
func ResolveStickerAttachment(sticker *whatsapp.WhatsappSticker) (*whatsapp.WhatsappAttachment, error) {
	var rawData []byte
	var sourceMime string

	switch {
	case len(sticker.Url) > 0:
		resp, err := http.Get(sticker.Url)
		if err != nil {
			return nil, fmt.Errorf("failed to download sticker from URL: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status %d downloading sticker", resp.StatusCode)
		}

		rawData, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read sticker URL response: %w", err)
		}
		sourceMime = resp.Header.Get("Content-Type")

	case len(sticker.Content) > 0:
		content := sticker.Content

		// Parse data URI: data:<mime>;base64,<data>
		if strings.HasPrefix(content, "data:") {
			parts := strings.SplitN(content, ",", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid data URI format")
			}
			header := parts[0] // "data:image/png;base64"
			if idx := strings.Index(header, ":"); idx >= 0 {
				mimeAndEnc := header[idx+1:]
				sourceMime = strings.Split(mimeAndEnc, ";")[0]
			}
			content = parts[1]
		}

		var err error
		rawData, err = base64.StdEncoding.DecodeString(content)
		if err != nil {
			return nil, fmt.Errorf("failed to decode sticker base64: %w", err)
		}

	default:
		return nil, fmt.Errorf("sticker requires either 'url' or 'content'")
	}

	if len(rawData) == 0 {
		return nil, fmt.Errorf("sticker content is empty")
	}

	// If already static WebP, skip conversion
	if sourceMime == "image/webp" {
		attach := &whatsapp.WhatsappAttachment{
			Mimetype:   "image/webp",
			FileLength: uint64(len(rawData)),
		}
		attach.SetContent(&rawData)
		return attach, nil
	}

	// Determine if source is video (animated sticker)
	isVideo := isVideoMime(sourceMime)

	webpData, webpMime, err := media.ConvertToWebP(rawData, isVideo)
	if err != nil {
		return nil, fmt.Errorf("WebP conversion failed: %w", err)
	}

	attach := &whatsapp.WhatsappAttachment{
		Mimetype:   webpMime,
		FileLength: uint64(len(webpData)),
	}
	attach.SetContent(&webpData)
	return attach, nil
}

// isVideoMime returns true for MIME types that should produce animated stickers
func isVideoMime(mime string) bool {
	mime = strings.Split(mime, ";")[0]
	switch mime {
	case "video/mp4", "video/webm", "video/mpeg", "video/3gpp",
		"video/quicktime", "image/gif", "video/webp":
		return true
	}
	return false
}
