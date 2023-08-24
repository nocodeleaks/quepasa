package library

import (
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

func GetMimeTypeFromContent(content []byte, filename string) string {
	mimeType := http.DetectContentType(content)
	if mimeType == "application/octet-stream" && len(filename) > 0 {
		extension := filepath.Ext(filename)
		newMimeType := mime.TypeByExtension(extension)
		if len(newMimeType) > 0 {
			mimeType = newMimeType
		}
	}
	return mimeType
}

func GenerateFileNameFromMimeType(mimeType string) string {

	const layout = "20060201150405"
	t := time.Now().UTC()
	filename := "file-" + t.Format(layout)

	// get file extension from mime type
	extension, _ := mime.ExtensionsByType(mimeType)
	if len(extension) > 0 {
		filename = filename + extension[0]
	}

	return filename
}

// Get the first discovered extension from a given mime type (with dot = {.ext})
func TryGetExtensionFromMimeType(mimeType string) (exten string, success bool) {
	if exten, success = MIMEs[mimeType]; success {
		return exten, true
	}

	extensions, err := mime.ExtensionsByType(mimeType)
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

func RemoveDigit9(source string) string {
	response := source

	// if is direct message, not group
	if strings.HasSuffix(source, "@s.whatsapp.net") {
		phonenumber := source[0:13]
		if len(phonenumber) == 13 {

			// mobile phones with 9 digit
			r, _ := regexp.Compile("55([4-9]\\d|[3-9][1-9])9\\d\\d\\d\\d\\d\\d\\d\\d$")
			if r.MatchString(phonenumber) {

				prefix := phonenumber[0:4]
				response = prefix + phonenumber[5:13] + "@s.whatsapp.net"
			}
		}
	}

	return response
}
