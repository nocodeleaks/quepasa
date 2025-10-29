package models

import (
	library "github.com/nocodeleaks/quepasa/library"
	log "github.com/sirupsen/logrus"
)

// quepasa build version, if ends with .0 means stable versions.
// version 3.YY.MMDD.HHMM
const QpVersion = "3.25.1029.1216"

const QpLogLevel = log.InfoLevel

// copying log fields names
var LogFields = library.LogFields
