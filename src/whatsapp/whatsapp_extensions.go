package whatsapp

import (
	"fmt"
	"regexp"
	"strings"
)

var AllowedSuffix = map[string]bool{
	"g.us":           true, // Mensagem para um grupo
	"s.whatsapp.net": true, // Mensagem direta a um usuário
}

func PhoneToWid(source string) (destination string) {

	// removing starting + from E164 phones
	destination = strings.TrimLeft(source, "+")

	if !strings.ContainsAny(destination, "@") {
		return destination + "@s.whatsapp.net"
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

			destination = destination + "@g.us"
		} else {
			if IsValidE164(destination) {
				destination = PhoneToWid(destination)
			} else {
				destination = destination + "@g.us"
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

// Returns message type by attachment mime information
func GetMessageType(Mimetype string) WhatsappMessageType {

	// usado pela API para garantir o envio como documento de qualquer anexo
	if strings.Contains(Mimetype, "wa-document") {
		return DocumentMessageType
	}

	// apaga informações após o ;
	// fica somente o mime mesmo
	mimeOnly := strings.Split(Mimetype, ";")
	switch mimeOnly[0] {
	case "image/png", "image/jpeg":
		return ImageMessageType
	case
		"audio/ogg", "application/ogg", "audio/oga", "audio/ogx",
		"audio/x-mpeg-3", "audio/mpeg3", "audio/mpeg",
		"audio/mp4", "audio/wav", "audio/wave", "audio/x-wav":
		return AudioMessageType
	case "video/mp4":
		return VideoMessageType

	// dont needed, just set to near following
	case "text/xml", "application/pdf":
		return DocumentMessageType

	default:
		println("(" + Mimetype + ") mime: " + mimeOnly[0])
		return DocumentMessageType
	}
}
