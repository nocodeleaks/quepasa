package models

import (
	library "github.com/nocodeleaks/quepasa/library"
	log "github.com/sirupsen/logrus"
)

// quepasa build version, if ends with .0 means stable versions.
const QpVersion = "3.25.404.1549"

const QpLogLevel = log.InfoLevel

// copying log fields names
var LogFields = library.LogFields
