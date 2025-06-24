package whatsapp

import (
	"encoding/json"
	"strings"

	library "github.com/nocodeleaks/quepasa/library"
)

type WhatsappChat struct {
	// whatsapp contact id, based on phone number or timestamp
	Id string `json:"id"`

	// new whatsapp unique contact id
	Lid string `json:"lid,omitempty"`

	// phone number in E164 format
	Phone string `json:"phone,omitempty"`

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
	// Return the Phone field if it's already populated
	if len(source.Phone) > 0 {
		return source.Phone
	}

	// Fallback to extracting from ID
	phone, _ := library.ExtractPhoneIfValid(source.Id)
	return phone
}

// PopulatePhone attempts to fill the Phone field using available mapping
func (source *WhatsappChat) PopulatePhone(conn interface{}) {
	if len(source.Phone) > 0 {
		return // Already populated
	}

	// Try to get phone from different sources
	if strings.Contains(source.Id, "@lid") {
		// For @lid, try to get the corresponding phone number
		if connection, ok := conn.(interface{ GetPhoneFromLID(string) (string, error) }); ok {
			if phone, err := connection.GetPhoneFromLID(source.Id); err == nil && len(phone) > 0 {
				// Format the phone to E164 if needed
				if formattedPhone, err := library.ExtractPhoneIfValid(phone); err == nil {
					source.Phone = formattedPhone
				} else {
					source.Phone = phone
				}
			}
		}
	} else if strings.Contains(source.Id, "@s.whatsapp.net") {
		// For @s.whatsapp.net, extract phone from ID
		if phone, _ := library.ExtractPhoneIfValid(source.Id); len(phone) > 0 {
			source.Phone = phone
		}
	}
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
	if source.Lid != source.Id && len(source.Lid) > 0 {
		aux.Lid = source.Lid
	}

	return json.Marshal(aux)
}
