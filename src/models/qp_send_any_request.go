package models

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"path"
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
}

// From BASE64 content
func (source *QpSendAnyRequest) GenerateEmbedContent() (err error) {
	decoded, err := base64.StdEncoding.DecodeString(source.Content)
	if err != nil {
		return
	}

	source.QpSendRequest.Content = decoded
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
	}

	return
}
