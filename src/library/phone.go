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

// RemoveDigit9IfElegible removes the extra 9th digit from Brazilian mobile numbers.
//
// Brazilian mobile phones with DDDs > 30 were migrated from 8-digit to 9-digit numbers
// (adding a leading "9" after the DDD). Some WhatsApp accounts were registered before
// the migration and still use the 8-digit format. This function strips the extra digit
// so both variants can be tried during send resolution.
//
// Eligible format: +55 <DDD> 9 <8 digits>  (total 14 chars including +)
// Mobile-only rule: original 8-digit number must start with 5-9.
// DDD eligibility: first digit 3-9, second digit 1-9 (i.e. DDD > 30, e.g. 31, 41, 47, 51, 61, 71)
//
// TEMPORARY WORKAROUND: WhatsApp has not standardized behavior for accounts registered
// under either variant. This function is part of a dual-variant resolution strategy
// until WhatsApp enforces a single canonical phone format on their platform.
func RemoveDigit9IfElegible(source string) (response string, err error) {
	if len(source) == 14 {

		// mobile phones with 9 digit (original first subscriber digit must be 5-9)
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

// AddDigit9IfEligible adds the extra 9th digit to a Brazilian mobile number that is missing it.
//
// This is the symmetric counterpart of RemoveDigit9IfElegible. Given a legacy 8-digit
// Brazilian mobile number, it produces the modern 9-digit variant so both forms can be
// persisted as valid lookup keys in the LID store.
//
// Eligible format: +55 <DDD> <8 digits>  (total 13 chars including +)
// Mobile-only rule: first subscriber digit must be 5-9.
// First subscriber digit must NOT be 9 (otherwise it already has the extra digit).
// DDD eligibility: first digit 3-9, second digit 1-9 (DDDs > 30, e.g. 31, 41, 47, 51, 61, 71).
//
// NOTE: For LID store augmentation (where an extra mapping is harmless), prefer
// AddDigit9BRAllDDDs which applies to all Brazilian DDDs including 11-29 (SP/RJ area).
//
// TEMPORARY WORKAROUND: WhatsApp has not standardized behavior for accounts registered
// under either variant. This function is part of a dual-variant resolution strategy
// until WhatsApp enforces a single canonical phone format on their platform.
func AddDigit9IfEligible(source string) (response string, err error) {
	if len(source) == 13 {

		// legacy 8-digit mobile phones: +55 <DDD 2 digits> <8 digits>
		// first subscriber digit must be 5-9
		r, _ := regexp.Compile(`\+55([4-9][1-9]|[3-9][1-9])[5-9]\d{7}$`)
		if r.MatchString(source) {
			prefix := source[0:5] // "+55" + DDD (2 digits)
			response = prefix + "9" + source[5:13]
		} else {
			err = fmt.Errorf("not eligible match")
		}
	} else {
		err = fmt.Errorf("not eligible number of digits")
	}

	return
}

// AddDigit9BRAllDDDs adds the extra 9th digit to a Brazilian mobile number for any DDD.
//
// Unlike AddDigit9IfEligible, this function applies to all Brazilian DDDs including
// 11-29 (São Paulo, Rio de Janeiro metropolitan areas). It is intended exclusively for
// LID store augmentation, where persisting an extra mapping that turns out to be unused
// is harmless. Do NOT use this for the phone-only send path — incorrect digit insertion
// for non-eligible DDDs could produce a wrong destination.
//
// Eligible format: +55 <any 2-digit DDD> <8 digits starting with 5-9> (total 13 chars)
//
// TEMPORARY WORKAROUND: Remove once WhatsApp enforces a single canonical phone format.
func AddDigit9BRAllDDDs(source string) (response string, err error) {
	if len(source) == 13 {
		r, _ := regexp.Compile(`^\+55\d{2}[5-9]\d{7}$`)
		if r.MatchString(source) {
			prefix := source[0:5] // "+55" + DDD (2 digits)
			response = prefix + "9" + source[5:13]
		} else {
			err = fmt.Errorf("not eligible match")
		}
	} else {
		err = fmt.Errorf("not eligible number of digits")
	}

	return
}

// RemoveDigit9BRAllDDDs removes the extra 9th digit from a Brazilian mobile number for any DDD.
//
// Unlike RemoveDigit9IfElegible, this function applies to all Brazilian DDDs including
// 11-29 (São Paulo, Rio de Janeiro metropolitan areas). It is intended exclusively for
// LID store augmentation. Do NOT use this for the phone-only send path.
//
// Eligible format: +55 <any 2-digit DDD> 9 <8 digits> (total 14 chars)
// Mobile-only rule: the first digit after the inserted 9 must be 5-9.
//
// TEMPORARY WORKAROUND: Remove once WhatsApp enforces a single canonical phone format.
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
