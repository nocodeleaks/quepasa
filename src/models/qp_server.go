package models

import (
	"reflect"
	"time"

	"github.com/nocodeleaks/quepasa/library"
	"github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
)

/*
<summary>

	Database representation for whatsapp controller service

</summary>
*/
type QpServer struct {
	library.LogStruct // logging

	// Optional whatsapp options
	// ------------------------
	whatsapp.WhatsappOptions

	// Public token
	Token string `db:"token" json:"token" validate:"max=100"`

	// Whatsapp session id
	Wid      string `db:"wid" json:"wid" validate:"max=255"`
	Verified bool   `db:"verified" json:"verified"`
	Devel    bool   `db:"devel" json:"devel"`

	User      string    `db:"user" json:"user,omitempty" validate:"max=36"`
	Timestamp time.Time `db:"timestamp" json:"timestamp,omitempty"`
}

// custom log entry with fields: wid
func (source *QpServer) GetLogger() *log.Entry {
	if source != nil && source.LogEntry != nil {
		return source.LogEntry
	}

	logentry := library.NewLogEntry(source)
	if source != nil {
		logentry = logentry.WithField(LogFields.WId, source.Wid)
		logentry = logentry.WithField(LogFields.Token, source.Token)
		source.LogEntry = logentry
	}

	logentry.Level = source.GetLogLevel()
	logentry.Infof("generating new log entry for %s, with level: %s", reflect.TypeOf(source), logentry.Level)

	return logentry
}

func (source *QpServer) GetWId() string {
	return source.Wid
}

// used for view
func (source *QpServer) GetLogLevel() log.Level {
	if source.Devel {
		loglevel := ENV.LogLevelFromLogrus(log.DebugLevel)
		if loglevel < log.DebugLevel {
			return log.DebugLevel
		}
		return loglevel
	} else {
		return log.InfoLevel
	}
}

//#region VIEW TRICKS

// used for view
func (source QpServer) IsSetCalls() bool {
	return source.Calls != whatsapp.UnSetBooleanType
}

// used for view
func (source QpServer) IsSetReadUpdate() bool {
	return source.ReadUpdate != whatsapp.UnSetBooleanType
}

// used for view
func (source QpServer) GetReadUpdate() bool {
	return source.ReadUpdate.Boolean()
}

// used for view
func (source QpServer) GetCalls() bool {
	return source.Calls.Boolean()
}

// used for view
func (source QpServer) IsSetReadReceipts() bool {
	return source.ReadReceipts != whatsapp.UnSetBooleanType
}

// used for view
func (source QpServer) GetReadReceipts() bool {
	return source.ReadReceipts.Boolean()
}

// used for view
func (source QpServer) IsSetBroadcasts() bool {
	return source.Broadcasts != whatsapp.UnSetBooleanType
}

// used for view
func (source QpServer) GetBroadcasts() bool {
	return source.Broadcasts.Boolean()
}

// used for view
func (source QpServer) IsSetGroups() bool {
	return source.Groups != whatsapp.UnSetBooleanType
}

// used for view
func (source QpServer) GetGroups() bool {
	return source.Groups.Boolean()
}

// used for view
func (source QpServer) IsSetIndividuals() bool {
	return source.Individuals != whatsapp.UnSetBooleanType
}

// used for view
func (source QpServer) GetIndividuals() bool {
	return source.Individuals.Boolean()
}

//#endregion
