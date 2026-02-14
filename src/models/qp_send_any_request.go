package models

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
)

/*
<summary>

	Request to send any type of message
	1º Attachment Url
	2º Attachment Content Base64
	3º Text Message

</summary>
*/
type QpSendAnyRequest struct {
	QpSendRequest

	// Url for download content
	Url string `json:"url,omitempty"`

	// BASE64 embed content
	Content string `json:"content,omitempty"`

	// Preview Options (default: true)
	// Controls both thumbnail generation for media (image/video/PDF) and link preview for URLs in text
	// Set to false to disable thumbnail generation and link preview
	Preview *bool `json:"preview,omitempty"`

	// Custom title for link preview (overrides fetched title)
	PreviewTitle string `json:"preview_title,omitempty"`

	// Custom description for link preview (overrides fetched description)
	PreviewDesc string `json:"preview_desc,omitempty"`

	// Custom thumbnail URL for link preview (overrides fetched image)
	PreviewThumb string `json:"preview_thumb,omitempty"`
}

// ShouldGeneratePreview returns true if preview/thumbnail should be generated (default: true)
func (source *QpSendAnyRequest) ShouldGeneratePreview() bool {
	if source.Preview == nil {
		return true // Default: generate preview
	}
	return *source.Preview
}

// From BASE64 content
func (source *QpSendAnyRequest) GenerateEmbedContent() (err error) {
	content := source.Content

	// Check if content is a data URI (e.g., "data:image/png;base64,<base64data>")
	if strings.HasPrefix(content, "data:") {
		// Parse data URI
		parts := strings.SplitN(content, ",", 2)
		if len(parts) != 2 {
			err = fmt.Errorf("invalid data URI format")
			return
		}

		// Extract MIME type from data URI
		header := parts[0]
		if strings.HasPrefix(header, "data:") && strings.Contains(header, ";base64") {
			mimePart := header[5:]                      // Remove "data:"
			mimeType := strings.Split(mimePart, ";")[0] // Get MIME before ";base64"
			if len(source.Mimetype) == 0 {
				source.Mimetype = mimeType
			}
		}

		// Use the base64 part
		content = parts[1]
	}

	decoded, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return
	}

	source.QpSendRequest.Content = decoded

	// Set the correct file length for decoded content
	source.FileLength = uint64(len(decoded))

	return
}

// From Url content
func (source *QpSendAnyRequest) GenerateUrlContent() (err error) {
	resp, err := http.Get(source.Url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = fmt.Errorf("error on generate url content, unexpected status code: %v", resp.StatusCode)

		logentry := source.GetLogger()
		logentry.Error(err)
		return
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	source.QpSendRequest.Content = content

	if resp.ContentLength > -1 {
		source.FileLength = uint64(resp.ContentLength)
	}

	if len(source.Mimetype) == 0 {
		source.Mimetype = resp.Header.Get("Content-Type")
	}

	// setting filename if empty
	if len(source.FileName) == 0 {
		source.FileName = path.Base(source.Url)

		if len(source.FileName) > 0 {

			// unescaping filename from url, on error, just warn ... dont panic
			filename, unescapeErr := url.QueryUnescape(source.FileName)
			if unescapeErr != nil {
				logentry := source.GetLogger()
				logentry.Warnf("fail to unescape from url, filename: %s, err: %s", source.FileName, unescapeErr)
			} else {
				source.FileName = filename
			}
		}
	}

	return
}
