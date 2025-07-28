package whatsapp

import (
	"encoding/json"
	"strings"

	library "github.com/nocodeleaks/quepasa/library"
)

type WhatsappChat struct {
	// (Identifier) whatsapp contact id, based on phone number or timestamp
	Id string `json:"id"`

	// (Local Identifier) new whatsapp unique contact id
	LId string `json:"lid,omitempty"`

	// phone number in E164 format
	Phone string `json:"phone,omitempty"`

	Title string `json:"title,omitempty"`
}

func (source *WhatsappChat) GetChatId() string {
	return source.Id
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
	// Return the Phone field if it's already populated
	if len(source.Phone) > 0 {
		return source.Phone
	}

	// Fallback to extracting from ID
	phone, _ := library.GetPhoneIfValid(source.Id)
	return phone
}

// MarshalJSON customizes JSON marshaling to omit lid when it's the same as id
func (source WhatsappChat) MarshalJSON() ([]byte, error) {
	type Alias WhatsappChat
	aux := struct {
		Alias
		Lid string `json:"lid,omitempty"`
	}{
		Alias: Alias(source),
	}

	// Only include lid if it's different from id
	if source.LId != source.Id && len(source.LId) > 0 {
		aux.Lid = source.LId
	}

	return json.Marshal(aux)
}
