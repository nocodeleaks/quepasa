package library

import (
	"reflect"

	qplog "github.com/nocodeleaks/quepasa/qplog"
)

const LogLevelDefault = qplog.ErrorLevel

type LogInterface interface {
	GetLogger() qplog.Logger
	LogWithField(key string, value interface{}) qplog.Logger
}

type LogStruct struct {
	LogEntry     qplog.Logger `json:"-"` // log entry
	LogInterface `json:"-"`
}

func NewLogEntryWithLevel(level qplog.Level) qplog.Logger {
	return qplog.New().WithLevel(level)
}

// the parameter source is used just to identify the log entry, it can be a string or any other type
func NewLogEntry(source any) qplog.Logger {
	var typeof string
	if source != nil {
		if stringValue, ok := source.(string); ok {
			typeof = stringValue
		} else {
			typeof = reflect.TypeOf(source).String()
		}
	}

	return qplog.New().WithField(LogFields.Entry, typeof)
}

func NewLogStruct(level qplog.Level) LogStruct {
	logentry := NewLogEntryWithLevel(level)
	return LogStruct{LogEntry: logentry}
}

// get default log entry, never nil
func (source *LogStruct) GetLogger() qplog.Logger {
	return GetLogger(source)
}

func GetLogger(source *LogStruct) qplog.Logger {
	if source != nil && source.LogEntry != nil {
		return source.LogEntry
	}

	logentry := NewLogEntryWithLevel(LogLevelDefault)
	logentry.Infof("generating new log entry for %s, with level: %s", reflect.TypeOf(source), logentry.Level())

	if source != nil {
		source.LogEntry = logentry
	}

	return logentry
}

/*
<summary>

	LogWithField adds a field to the logger while preserving its level. With the
	qplog facade the level lives on the logger instance, so WithField alone keeps
	it; no explicit save/restore is needed.

</summary>
*/
func (source *LogStruct) LogWithField(key string, value interface{}) qplog.Logger {
	return source.GetLogger().WithField(key, value)
}
