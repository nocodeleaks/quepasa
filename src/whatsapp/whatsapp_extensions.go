package whatsapp

import (
	"fmt"
	"regexp"
	"strings"

	library "github.com/nocodeleaks/quepasa/library"
	log "github.com/sirupsen/logrus"
)

// Helper function to convert phone numbers or partial WIDs to full WIDs
func PhonesToWids(sources []string) []string {
	result := make([]string, len(sources))

	for i, item := range sources {
		result[i] = PhoneToWid(item)
	}

	return result
}

// PhoneToWid converts a phone number to a WhatsApp WID
func PhoneToWid(source string) (destination string) {

	// removing starting + from E164 phones
	destination = strings.TrimLeft(source, "+")

	if !strings.ContainsAny(destination, "@") {
		return destination + WHATSAPP_SERVERDOMAIN_USER_SUFFIX
	}
	return
}

// Formata um texto qualquer em formato de destino válido para o sistema do whatsapp
func FormatEndpoint(source string) (destination string, err error) {

	// removing whitespaces
	destination = strings.Replace(source, " ", "", -1)

	if len(destination) < 8 {
		if len(destination) == 0 {
			err = fmt.Errorf("empty chatid recipient")
			return
		}

		err = fmt.Errorf("invalid chatid length")
		return
	}

	// if have a + as prefix, is a phone number
	if strings.HasPrefix(destination, "+") {
		destination = PhoneToWid(destination)
		return
	}

	if strings.ContainsAny(destination, "@") {
		splited := strings.Split(destination, "@")
		if !AllowedSuffix[splited[1]] {
			err = fmt.Errorf("invalid suffix @ recipient %s", destination)
			return
		}
	} else {
		if strings.Contains(destination, "-") {
			splited := strings.Split(destination, "-")
			if !IsValidE164(splited[0]) {
				err = fmt.Errorf("contains - but its not a valid group: %s", source)
				return
			}

			destination = destination + WHATSAPP_SERVERDOMAIN_GROUP_SUFFIX
		} else {
			if IsValidE164(destination) {
				destination = PhoneToWid(destination)
			} else {
				destination = destination + WHATSAPP_SERVERDOMAIN_GROUP_SUFFIX
			}
		}
	}

	return
}

var RegexValidE164Test = string(`\d`)

func IsValidE164(phone string) bool {
	regex, err := regexp.Compile(RegexValidE164Test)
	if err != nil {
		panic("invalid regex on IsValidE164 :: " + RegexValidE164Test)
	}
	matches := regex.FindAllString(phone, -1)
	if len(matches) >= 9 && len(matches) <= 15 {
		return true
	}
	return false
}

func GetPhoneIfValid(source string) (phone string, err error) {

	// Removing whitespace
	response := strings.TrimSpace(source)

	// Validating minimum length for a phone number
	response = strings.TrimLeft(response, "+")

	// If it has @s.whatsapp.net, remove it
	if strings.HasSuffix(response, WHATSAPP_SERVERDOMAIN_USER_SUFFIX) {
		response = strings.Split(response, "@")[0]

		if strings.Contains(response, ":") {
			// removing session id
			response = strings.Split(response, ":")[0]
		}
	}

	// Regex to match E164 format
	return library.GetPhoneIfValid(response)
}

/*
<summary>

	Checks if the given ID is a valid group ID
	It returns true if the ID is valid, false otherwise.

</summary>
<remarks>

	It checks if the ID ends with "@g.us"
	It checks the prefix (before "@") max 20 caracters.

<remarks>
*/
func IsValidGroupId(id string) bool {
	return strings.HasSuffix(id, WHATSAPP_SERVERDOMAIN_GROUP_SUFFIX) && len(id) <= 29
}

func GetMessageType(attach *WhatsappAttachment) WhatsappMessageType {
	if attach == nil {
		return TextMessageType
	}

	if attach.ptt {
		return AudioMessageType
	}

	if strings.HasPrefix(attach.FileName, InvalidFilePrefix) {
		return DocumentMessageType
	}

	return GetMessageTypeFromMIME(attach.Mimetype)
}

// Returns message type by attachment mime information
func GetMessageTypeFromMIME(mime string) WhatsappMessageType {

	// should force to send as document ?
	if strings.Contains(mime, "wa-document") {
		return DocumentMessageType
	}

	// switch for basic mime type, ignoring suffix
	mimeOnly := strings.Split(mime, ";")[0]

	for _, item := range WhatsappMIMEAudio {
		if item == mimeOnly {
			return AudioMessageType
		}
	}

	for _, item := range WhatsappMIMEVideo {
		if item == mimeOnly {
			return VideoMessageType
		}
	}

	for _, item := range WhatsappMIMEImage {
		if item == mimeOnly {
			return ImageMessageType
		}
	}

	for _, item := range WhatsappMIMEDocument {
		if item == mimeOnly {
			return DocumentMessageType
		}
	}

	log.Debugf("whatsapp extensions default, full mime: (%s) extract mime: %s", mime, mimeOnly)
	return DocumentMessageType
}
