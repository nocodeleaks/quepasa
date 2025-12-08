package models

import (
	"time"

	library "github.com/nocodeleaks/quepasa/library"
	log "github.com/sirupsen/logrus"
)

// quepasa build version, if ends with .0 means stable versions.
// version 3.YY.MMDD.HHMM
const QpVersion = "3.25.1208.1515"

const QpLogLevel = log.InfoLevel

// copying log fields names
var LogFields = library.LogFields

// ApplicationStartTime stores when the application was started
var ApplicationStartTime time.Time

func init() {
	ApplicationStartTime = time.Now()
}
