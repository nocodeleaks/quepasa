package whatsmeow

import (
	library "github.com/nocodeleaks/quepasa/library"
	"github.com/sirupsen/logrus"
	types "go.mau.fi/whatsmeow/types"
)

const WhatsmeowLogLevel = logrus.WarnLevel // default log level for whatsmeow
const WhatsmeowClientLogLevel = "INFO"     // default log level for whatsmeow client
const WhatsmeowDBLogLevel = "WARN"         // default log level for whatsmeow database

// default service presence state
const WhatsmeowPresenceDefault = types.PresenceUnavailable

// copying log fields names
var LogFields = library.LogFields
