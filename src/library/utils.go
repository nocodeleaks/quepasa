package library

import (
	"encoding/base64"
	"encoding/json"
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

// Removes line breaks and leading/trailing spaces from a string
func NormalizeForTitle(source string) string {

	response := source

	response = strings.ReplaceAll(response, "\r\n", " ")
	response = strings.ReplaceAll(response, "\r", " ")
	response = strings.ReplaceAll(response, "\n", " ")

	// removing leading and trailing spaces
	response = strings.TrimSpace(response)

	return response
}

func ToJson(in interface{}) string {
	bytes, err := json.Marshal(in)
	if err == nil {
		return string(bytes)
	}
	return ""
}

// DecodeBase64 decodes a base64 encoded string to bytes
func DecodeBase64(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}
