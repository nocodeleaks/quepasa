package models

import (
	"encoding/base64"
)

// Mensagem no formato QuePasa
// Utilizada na API do QuePasa para troca com outros sistemas
type QPAttachmentV1 struct {
	Url         string `json:"url,omitempty"`
	B64MediaKey string `json:"b64mediakey,omitempty"`
	Length      uint64 `json:"length,omitempty"`
	MIME        string `json:"mime,omitempty"`
	Base64      string `json:"base64,omitempty"`
	FileName    string `json:"filename,omitempty"`
}

// Traz o MediaKey em []byte apartir de base64
func (source *QPAttachmentV1) MediaKey() ([]byte, error) {
	return base64.StdEncoding.DecodeString(source.B64MediaKey)
}
