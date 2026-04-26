package models

import (
	"database/sql"
	"encoding/json"
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
	Wid      sql.NullString `db:"wid" json:"wid" validate:"max=255" swaggertype:"string"`
	Verified bool           `db:"verified" json:"verified"`
	Devel    bool           `db:"devel" json:"devel"`
	Metadata QpMetadata     `db:"metadata" json:"metadata,omitempty"`

	User      sql.NullString `db:"user" json:"user,omitempty" validate:"max=36" swaggertype:"string"`
	Timestamp time.Time      `db:"timestamp" json:"timestamp,omitempty"`
}

func (source QpServer) MarshalJSON() ([]byte, error) {
	type serverJSON struct {
		whatsapp.WhatsappOptions
		Token     string     `json:"token"`
		Wid       string     `json:"wid,omitempty"`
		Verified  bool       `json:"verified"`
		Devel     bool       `json:"devel"`
		Metadata  QpMetadata `json:"metadata,omitempty"`
		User      string     `json:"user,omitempty"`
		Timestamp time.Time  `json:"timestamp,omitempty"`
	}

	return json.Marshal(serverJSON{
		WhatsappOptions: source.WhatsappOptions,
		Token:           source.Token,
		Wid:             source.GetWId(),
		Verified:        source.Verified,
		Devel:           source.Devel,
		Metadata:        source.Metadata,
		User:            source.GetUser(),
		Timestamp:       source.Timestamp,
	})
}

func (source *QpServer) UnmarshalJSON(data []byte) error {
	type serverJSON struct {
		whatsapp.WhatsappOptions
		Token     string     `json:"token"`
		Wid       string     `json:"wid"`
		Verified  bool       `json:"verified"`
		Devel     bool       `json:"devel"`
		Metadata  QpMetadata `json:"metadata,omitempty"`
		User      string     `json:"user"`
		Timestamp time.Time  `json:"timestamp,omitempty"`
	}

	var payload serverJSON
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	source.WhatsappOptions = payload.WhatsappOptions
	source.Token = payload.Token
	source.SetWId(payload.Wid)
	source.Verified = payload.Verified
	source.Devel = payload.Devel
	source.Metadata = payload.Metadata
	source.SetUser(payload.User)
	source.Timestamp = payload.Timestamp
	return nil
}

// custom log entry with fields: wid
func (source *QpServer) GetLogger() *log.Entry {
	if source != nil && source.LogEntry != nil {
		return source.LogEntry
	}

	logentry := library.NewLogEntry(source)
	if source != nil {
		widStr := ""
		if source.Wid.Valid {
			widStr = source.Wid.String
		}
		logentry = logentry.WithField(LogFields.WId, widStr)
		logentry = logentry.WithField(LogFields.Token, source.Token)
		source.LogEntry = logentry
	}

	logentry.Level = source.GetLogLevel()
	logentry.Infof("generating new log entry for %s, with level: %s", reflect.TypeOf(source), logentry.Level)

	return logentry
}

func (source *QpServer) GetWId() string {
	if source == nil || !source.Wid.Valid {
		return ""
	}
	return source.Wid.String
}

// SetWId sets the Wid field
func (source *QpServer) SetWId(wid string) {
	if source == nil {
		return
	}
	if len(wid) == 0 {
		source.Wid = sql.NullString{}
	} else {
		source.Wid = sql.NullString{String: wid, Valid: true}
	}
}

// GetUser returns user as string, handling sql.NullString
func (source *QpServer) GetUser() string {
	if source == nil || !source.User.Valid {
		return ""
	}
	return source.User.String
}

// SetUser sets the User field
func (source *QpServer) SetUser(user string) {
	if source == nil {
		return
	}
	if len(user) == 0 {
		source.User = sql.NullString{}
	} else {
		source.User = sql.NullString{String: user, Valid: true}
	}
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

//#endregion
