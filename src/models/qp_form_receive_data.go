package models

import (
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// Parameters to be accessed/passed on Views (receive.tmpl)
type QPFormReceiveData struct {
	PageTitle           string
	ErrorMessage        string
	Number              string
	Token               string
	DownloadPrefix      string
	FormAccountEndpoint string
	Messages            []whatsapp.WhatsappMessage
}

func (source QPFormReceiveData) Count() int {
	return len(source.Messages)
}
