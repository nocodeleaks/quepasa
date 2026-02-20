package models

import (
	"time"

	library "github.com/nocodeleaks/quepasa/library"
	log "github.com/sirupsen/logrus"
)

// quepasa build version format has 4 sections only: 3.YY.MMDD.HHMM
// stable versions are identified when HHMM last digit is 0
const QpVersion = "3.26.0220.0005"

const QpLogLevel = log.InfoLevel

// copying log fields names
var LogFields = library.LogFields

// ApplicationStartTime stores when the application was started
var ApplicationStartTime time.Time

func init() {
	ApplicationStartTime = time.Now()
}
