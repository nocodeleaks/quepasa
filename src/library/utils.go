package library

import (
	"fmt"
	"mime"
	"net/http"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// Validate email string
func IsValidEMail(s string) bool {
	var rx = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	if len(s) < 255 && rx.MatchString(s) {
		return true
	}

	return false
}

// Returns a string representation of source type interface
func GetTypeString(myvar interface{}) string {
	if t := reflect.TypeOf(myvar); t.Kind() == reflect.Ptr {
		return "*" + t.Elem().Name()
	} else {
		return t.Name()
	}
}

func GetMimeTypeFromContent(content []byte) string {
	mimeType := http.DetectContentType(content)
	if len(mimeType) > 0 && mimeType != "application/octet-stream" && mimeType != "application/zip" {
		log.Tracef("utils - detected mime type from content: %s", mimeType)
		return mimeType
	}
	return ""
}

// used by send attachment via url
// 2024/04/10 changed priority by extension, then content.
func GetMimeTypeFromContentAndExtension(content []byte, filename string) string {

	mimeType := http.DetectContentType(content)
	if len(mimeType) > 0 && mimeType != "application/octet-stream" && mimeType != "application/zip" {
		log.Tracef("utils - detected mime type from content: %s", mimeType)
		return mimeType
	}

	if len(filename) > 0 {
		extension := filepath.Ext(filename)

		// checking for static manual mime maps
		for key, value := range MIMEs {
			if value == extension {
				log.Tracef("utils - detected mime type from static maps: %s", key)
				return key
			}
		}

		newMimeType := mime.TypeByExtension(extension)
		log.Tracef("utils - detected mime type from extension: %s, filename: %s, mime: %s", extension, filename, newMimeType)

		if len(newMimeType) > 0 {
			return newMimeType
		}
	}

	return mimeType
}

// Get the first discovered extension from a given mime type (with dot = {.ext})
func TryGetExtensionFromMimeType(mimeType string) (exten string, success bool) {

	// lowering
	normalized := strings.ToLower(mimeType)

	// removing everything after ;
	if strings.Contains(normalized, ";") {
		normalized = strings.Split(normalized, ";")[0]
	}

	// removing white spaces
	normalized = strings.TrimSpace(normalized)

	// looking for at static mappings
	if exten, success = MIMEs[normalized]; success {
		return exten, true
	}

	extensions, err := mime.ExtensionsByType(normalized)
	if err != nil {
		log.Errorf("error getting internal mime for: %s, %s", mimeType, err.Error())
		return exten, false
	}

	if len(extensions) > 0 {
		exten = extensions[0]
		if !strings.HasPrefix(exten, ".") {
			exten = "." + exten
		}
		return exten, true
	} else {
		return exten, false
	}
}

func GenerateFileNameFromMimeType(mimeType string) string {

	const layout = "20060201150405"
	t := time.Now().UTC()
	filename := "file-" + t.Format(layout)

	// get file extension from mime type
	extension, _ := TryGetExtensionFromMimeType(mimeType)
	if len(extension) > 0 {
		filename = filename + extension
	}

	return filename
}

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

func ExtractPhoneIfValid(source string) (phone string, err error) {
	response := strings.TrimLeft(source, "+")
	if strings.HasSuffix(response, "@s.whatsapp.net") {
		response = strings.Split(response, "@")[0]
	}

	r, _ := regexp.Compile(`^[1-9]\d{6,14}$`)
	if r.MatchString(response) {
		phone = "+" + response
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
