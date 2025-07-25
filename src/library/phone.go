package library

import (
	"fmt"
	"regexp"
	"strings"
)

// Usado também para identificar o número do bot
// Meramente visual
func GetPhoneByWId(wid string) string {

	// removing whitespaces
	out := strings.Replace(wid, " ", "", -1)
	if strings.Contains(out, "@") {
		// capturando tudo antes do @
		splited := strings.Split(out, "@")
		out = splited[0]

		if strings.Contains(out, ".") {
			// capturando tudo antes do "."
			splited = strings.Split(out, ".")
			out = splited[0]

			return out
		}
	}

	re, err := regexp.Compile(`\d*`)
	if err == nil {
		matches := re.FindAllString(out, -1)
		if len(matches) > 0 {
			out = matches[0]
		}
	}
	return out
}

// ExtractPhoneFromWhatsappID extrai o telefone de um ID do WhatsApp do tipo "@s.whatsapp.net".
// Retorna o telefone em formato E164 se possível, ou string vazia se não conseguir.
func ExtractPhoneFromWhatsappID(id string) string {
	if strings.Contains(id, "@s.whatsapp.net") {
		phone, _ := GetPhoneIfValid(id)
		return phone
	}
	return ""
}

// GetPhoneIfValid tenta extrair e validar um número de telefone no formato E164 a partir de uma string.
// Retorna o número de telefone formatado ou um erro se não for válido.
func GetPhoneIfValid(source string) (phone string, err error) {

	// Validating minimum length for a phone number
	// This is a basic check; you might want to adjust it based on your requirements.
	if len(source) < 7 {
		err = fmt.Errorf("not a valid E164 phone number")
		return
	}

	r, _ := regexp.Compile(`^[1-9]\d{6,14}$`)
	if r.MatchString(source) {
		phone = "+" + source
	} else {
		err = fmt.Errorf("not a valid phone number")
	}
	return
}

// Whatsapp issue on understanding mobile phones with ddd bigger than 30, only mobile
func RemoveDigit9IfElegible(source string) (response string, err error) {
	if len(source) == 14 {

		// mobile phones with 9 digit
		r, _ := regexp.Compile(`\+55([4-9][1-9]|[3-9][1-9])9\d{8}$`)
		if r.MatchString(source) {
			prefix := source[0:5]
			response = prefix + source[6:14]
		} else {
			err = fmt.Errorf("not elegible match")
		}
	} else {
		err = fmt.Errorf("not elegible number of digits")
	}

	return
}
