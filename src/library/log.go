package library

import (
	"context"
	"reflect"

	log "github.com/sirupsen/logrus"
)

const LogLevelDefault = log.ErrorLevel

type LogInterface interface {
	GetLogger() *log.Entry
	LogWithField(key string, value interface{}) *log.Entry
}

type LogStruct struct {
	LogEntry     *log.Entry `json:"-"` // log entry
	LogInterface `json:"-"`
}

func NewLogEntryWithLevel(level log.Level) *log.Entry {
	logentry := log.WithContext(context.Background())
	logentry.Level = level
	return logentry
}

func NewLogEntry(source any) *log.Entry {
	var typeof string
	if source != nil {
		if stringValue, ok := source.(string); ok {
			typeof = stringValue
		} else {
			typeof = reflect.TypeOf(source).String()
		}
	}

	logentry := log.WithContext(context.Background())
	logentry = logentry.WithField(LogFields.Entry, typeof)
	return logentry
}

func NewLogStruct(level log.Level) LogStruct {
	logentry := NewLogEntryWithLevel(level)
	return LogStruct{LogEntry: logentry}
}

// get default log entry, never nil
func (source *LogStruct) GetLogger() *log.Entry {
	return GetLogger(source)
}

func GetLogger(source *LogStruct) *log.Entry {
	if source != nil && source.LogEntry != nil {
		return source.LogEntry
	}

	logentry := NewLogEntryWithLevel(LogLevelDefault)
	logentry.Infof("generating new log entry for %s, with level: %s", reflect.TypeOf(source), logentry.Level)

	if source != nil {
		source.LogEntry = logentry
	}

	return logentry
}

func (source *LogStruct) LogWithField(key string, value interface{}) *log.Entry {
	logentry := source.GetLogger()
	loglevel := logentry.Level
	logentry = logentry.WithField(key, value)
	logentry.Level = loglevel
	return logentry
}
