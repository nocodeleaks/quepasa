package models

import (
	"strings"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

func IsValidExtensionFor(request string, content string) bool {
	switch {
	case
		request == ".jpg" && content == ".jpeg", // used for correct old windows 3 characters extensions
		request == ".jpeg" && content == ".jpg", // inverse is even true
		request == ".csv" && content == ".txt",
		request == ".json" && content == ".txt",
		request == ".sql" && content == ".txt",
		request == ".oga" && content == ".webm",
		request == ".oga" && content == ".ogx",
		request == ".opus" && content == ".ogx",
		request == ".ovpn" && content == ".txt",
		request == ".svg" && content == ".xml":
		return true
	}

	return request == content
}

func IsCompatibleWithPTT(mime string) bool {
	// switch for basic mime type, ignoring suffix
	mimeOnly := strings.Split(mime, ";")[0]

	for _, item := range whatsapp.WhatsappMIMEAudioPTTCompatible {
		if item == mimeOnly {
			return true
		}
	}

	return false
}
