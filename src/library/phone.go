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

// RemoveDigit9IfElegible strips the extra 9th digit for BR mobile numbers (DDD > 30).
func RemoveDigit9IfElegible(source string) (response string, err error) {
	if len(source) == 14 {

		// mobile phones with 9 digit (subscriber must start with 5-9)
		r, _ := regexp.Compile(`\+55([4-9][1-9]|[3-9][1-9])9[5-9]\d{7}$`)
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

// AddDigit9IfEligible inserts the extra 9th digit for legacy BR mobile numbers (DDD > 30).
func AddDigit9IfEligible(source string) (response string, err error) {
	if len(source) == 13 {
		r, _ := regexp.Compile(`\+55([4-9][1-9]|[3-9][1-9])[5-9]\d{7}$`)
		if r.MatchString(source) {
			prefix := source[0:5]
			response = prefix + "9" + source[5:13]
		} else {
			err = fmt.Errorf("not eligible match")
		}
	} else {
		err = fmt.Errorf("not eligible number of digits")
	}

	return
}

// AddDigit9BRAllDDDs inserts the extra 9th digit for any BR DDD (mobile only).
func AddDigit9BRAllDDDs(source string) (response string, err error) {
	if len(source) == 13 {
		r, _ := regexp.Compile(`^\+55\d{2}[5-9]\d{7}$`)
		if r.MatchString(source) {
			prefix := source[0:5]
			response = prefix + "9" + source[5:13]
		} else {
			err = fmt.Errorf("not eligible match")
		}
	} else {
		err = fmt.Errorf("not eligible number of digits")
	}

	return
}

// RemoveDigit9BRAllDDDs strips the extra 9th digit for any BR DDD (mobile only).
func RemoveDigit9BRAllDDDs(source string) (response string, err error) {
	if len(source) == 14 {
		r, _ := regexp.Compile(`^\+55\d{2}9[5-9]\d{7}$`)
		if r.MatchString(source) {
			prefix := source[0:5]
			response = prefix + source[6:14]
		} else {
			err = fmt.Errorf("not eligible match")
		}
	} else {
		err = fmt.Errorf("not eligible number of digits")
	}

	return
}
