package whatsapp

import (
	"strings"

	library "github.com/nocodeleaks/quepasa/library"
)

type WhatsappChat struct {
	// whatsapp contact id, based on phone number or timestamp
	Id string `json:"id"`

	// new whatsapp unique contact id
	Lid string `json:"lid,omitempty"`

	Title string `json:"title,omitempty"`
}

var WASYSTEMCHAT = WhatsappChat{Id: "system", Title: "Internal System Message"}

func (source *WhatsappChat) FormatContact() {
	// removing session id
	if strings.Contains(source.Id, ":") {
		prefix := strings.Split(source.Id, ":")[0]
		suffix := strings.Split(source.Id, "@")[1]
		source.Id = prefix + "@" + suffix
	}
}

// get phone number if exists
func (source *WhatsappChat) GetPhone() string {
	phone, _ := library.ExtractPhoneIfValid(source.Id)
	return phone
}
