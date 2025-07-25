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
	phone, _ := library.GetPhoneIfValid(source.Id)
	return phone
}

// PopulatePhone attempts to fill the Phone field using available mapping
func (source *WhatsappChat) PopulatePhone(conn interface{}) {
	if len(source.Phone) > 0 {
		return // Already populated
	}

	// Create a LogStruct and use GetLogger for proper logging
	logStruct := library.NewLogStruct(library.LogLevelDefault)
	logentry := logStruct.GetLogger()
	logentry = logentry.WithField(library.LogFields.Entry, "WhatsappChat.PopulatePhone")

	// Try to get phone from different sources
	if strings.Contains(source.Id, "@lid") {
		// For @lid, try to get the corresponding phone number using WhatsmeowConnection interface
		if connection, ok := conn.(interface{ GetPhoneFromLID(string) (string, error) }); ok {
			if phone, err := connection.GetPhoneFromLID(source.Id); err == nil && len(phone) > 0 {
				logentry.WithField("id", source.Id).WithField("phone", phone).Debug("Retrieved phone from LID mapping")

				// Format the phone to E164 if needed
				if formattedPhone, err := library.GetPhoneIfValid(phone); err == nil {
					source.Phone = formattedPhone
					logentry.WithField("id", source.Id).WithField("formatted_phone", formattedPhone).Debug("Phone formatted to E164")
				} else {
					source.Phone = phone
					logentry.WithField("id", source.Id).WithField("phone", phone).WithError(err).Warn("Phone validation failed, using raw phone")
				}
			} else {
				logentry.WithField("id", source.Id).WithError(err).Error("Failed to get phone from LID mapping")
			}
		} else {
			logentry.WithField("id", source.Id).Error("Connection does not support GetPhoneFromLID method")
		}
	} else if strings.Contains(source.Id, "@s.whatsapp.net") {
		// For @s.whatsapp.net, extract phone from ID
		if phone, _ := library.GetPhoneIfValid(source.Id); len(phone) > 0 {
			source.Phone = phone
			logentry.WithField("id", source.Id).WithField("phone", phone).Debug("Extracted phone from s.whatsapp.net ID")
		} else {
			logentry.WithField("id", source.Id).Error("Failed to extract phone from s.whatsapp.net ID")
		}
	} else {
		logentry.WithField("id", source.Id).Debug("No phone extraction method available for this ID format")
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
