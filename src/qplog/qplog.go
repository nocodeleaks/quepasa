// Package qplog is the project's single logging facade (module
// github.com/nocodeleaks/quepasa/qplog). It is the ONLY package allowed to
// import a concrete logging backend (logrus); every other package logs through
// the qplog.Logger interface, so the backend can be swapped here without
// touching any call site.
//
// The facade covers three call styles so it can replace every logger in the
// project at once, with no method-name collisions:
//
//   - logrus style:  log.Info("msg")          log.Infof("msg %d", x)
//   - waLog style:   log.Sub("module").Infof(...)   (satisfies waLog.Logger)
//   - fluent style:  log.InfoE().Str("k", v).Int("n", x).Msg("text")
//
// The fluent entry points carry an "E" suffix (InfoE/DebugE/...) so they never
// clash with the logrus-style bare level methods. A qplog.Logger directly
// satisfies whatsmeow's waLog.Logger (Infof/Warnf/Errorf/Debugf/Sub).
//
// The package is named qplog (not log) to avoid colliding with the standard
// library log package and the many local identifiers named "log".
package qplog

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/sirupsen/logrus"
)

// Fields is a set of structured log fields (drop-in for logrus.Fields).
type Fields = map[string]any

// Level is a log level name (drop-in for code that used logrus.Level). It is a
// string alias so it interoperates freely with level strings.
type Level = string

// Level constants matching logrus level names.
const (
	TraceLevel = "trace"
	DebugLevel = "debug"
	InfoLevel  = "info"
	WarnLevel  = "warning"
	ErrorLevel = "error"
	FatalLevel = "fatal"
	PanicLevel = "panic"
)

// Logger is the logging entry point used across the whole project.
type Logger interface {
	// logrus style — log immediately with fmt.Sprint/Sprintf semantics.
	Trace(args ...any)
	Debug(args ...any)
	Info(args ...any)
	Warn(args ...any)
	Error(args ...any)
	Fatal(args ...any) // logs then exits the process
	Panic(args ...any) // logs then panics

	Tracef(format string, args ...any)
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
	Panicf(format string, args ...any)

	// Print family (logrus compatibility) — log at Info level.
	Print(args ...any)
	Printf(format string, args ...any)
	Println(args ...any)

	// Warning* are logrus aliases for Warn*.
	Warning(args ...any)
	Warningf(format string, args ...any)
	Warningln(args ...any)

	// Structured context — return a child logger carrying the given context.
	WithField(key string, value any) Logger
	WithFields(fields Fields) Logger
	WithError(err error) Logger
	WithContext(ctx context.Context) Logger

	// Sub returns a child logger scoped to a named module (waLog.Logger).
	Sub(module string) Logger

	// Level reports the logger's current level name; WithLevel returns a child
	// logger limited to the given level (its own backend instance). These
	// replace the logrus.Entry.Level field used for per-component levels.
	Level() Level
	WithLevel(level Level) Logger

	// Writer returns an io.Writer that logs whatever is written to it at the
	// logger's level (drop-in for logrus.Entry.Writer).
	Writer() io.Writer

	// Fluent style — return an Event to attach fields, terminated by Msg/Msgf.
	TraceE() Event
	DebugE() Event
	InfoE() Event
	WarnE() Event
	ErrorE() Event
}

// Event accumulates structured fields and emits one log line on Msg/Msgf.
type Event interface {
	Str(key, val string) Event
	Strs(key string, vals []string) Event
	Int(key string, val int) Event
	Int32(key string, val int32) Event
	Int64(key string, val int64) Event
	Uint(key string, val uint) Event
	Uint8(key string, val uint8) Event
	Uint16(key string, val uint16) Event
	Uint32(key string, val uint32) Event
	Uint64(key string, val uint64) Event
	Float64(key string, val float64) Event
	Bool(key string, val bool) Event
	Err(err error) Event
	Dur(key string, val time.Duration) Event
	Bytes(key string, val []byte) Event
	Msg(msg string)
	Msgf(format string, args ...any)
}

// ---------------------------------------------------------------------------
// constructors and global configuration
// ---------------------------------------------------------------------------

// New returns a Logger backed by the standard logrus logger, so all logs share
// the project's single logging configuration and output.
func New() Logger {
	return &logrusLogger{entry: logrus.NewEntry(logrus.StandardLogger())}
}

// NewWithLogger wraps an existing *logrus.Logger (nil uses the standard logger).
func NewWithLogger(l *logrus.Logger) Logger {
	if l == nil {
		l = logrus.StandardLogger()
	}
	return &logrusLogger{entry: logrus.NewEntry(l)}
}

// NewLeveled returns a Logger on its own backend instance, scoped to module and
// limited to the given level ("trace".."error"). It replaces per-component
// constructors such as waLog.Stdout(module, level, color).
func NewLeveled(module, level string) Logger {
	l := logrus.New()
	l.SetOutput(logrus.StandardLogger().Out)
	l.SetFormatter(logrus.StandardLogger().Formatter)
	if lvl, err := logrus.ParseLevel(level); err == nil {
		l.SetLevel(lvl)
	}
	e := logrus.NewEntry(l)
	if module != "" {
		e = e.WithField("module", module)
	}
	return &logrusLogger{entry: e}
}

// SetLevel sets the standard logger's level ("trace".."panic"). Unknown levels
// are ignored.
func SetLevel(level string) {
	if lvl, err := logrus.ParseLevel(level); err == nil {
		logrus.SetLevel(lvl)
	}
}

// SetOutput sets the standard logger's output writer.
func SetOutput(w io.Writer) { logrus.SetOutput(w) }

// UseTextFormatter / UseJSONFormatter select the standard logger's format.
func UseTextFormatter() { logrus.SetFormatter(&logrus.TextFormatter{}) }
func UseJSONFormatter() { logrus.SetFormatter(&logrus.JSONFormatter{}) }

// MoreVerbose returns whichever of the two levels is more verbose (Trace is the
// most verbose, Panic the least). Replaces numeric comparisons of logrus.Level.
func MoreVerbose(a, b Level) Level {
	la, ea := logrus.ParseLevel(a)
	lb, eb := logrus.ParseLevel(b)
	if ea != nil {
		return b
	}
	if eb != nil {
		return a
	}
	if la >= lb { // logrus: higher level value == more verbose
		return a
	}
	return b
}

// AtLeastVerbose reports whether level is at least as verbose as threshold
// (e.g. AtLeastVerbose(level, DebugLevel) is true for Debug and Trace). Replaces
// numeric >= comparisons of logrus.Level.
func AtLeastVerbose(level, threshold Level) bool {
	lvl, e1 := logrus.ParseLevel(level)
	thr, e2 := logrus.ParseLevel(threshold)
	if e1 != nil || e2 != nil {
		return false
	}
	return lvl >= thr
}

// ParseLevel parses a level string and returns its canonical form, so callers
// never import the backend directly.
func ParseLevel(level string) (string, error) {
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		return "info", err
	}
	return lvl.String(), nil
}

// ---------------------------------------------------------------------------
// package-level convenience logging (drop-in for logrus package-level calls)
// ---------------------------------------------------------------------------

func Trace(args ...any) { logrus.StandardLogger().Trace(args...) }
func Debug(args ...any) { logrus.StandardLogger().Debug(args...) }
func Info(args ...any)  { logrus.StandardLogger().Info(args...) }
func Warn(args ...any)  { logrus.StandardLogger().Warn(args...) }
func Error(args ...any) { logrus.StandardLogger().Error(args...) }
func Fatal(args ...any) { logrus.StandardLogger().Fatal(args...) }
func Panic(args ...any) { logrus.StandardLogger().Panic(args...) }

func Tracef(format string, args ...any) { logrus.StandardLogger().Tracef(format, args...) }
func Debugf(format string, args ...any) { logrus.StandardLogger().Debugf(format, args...) }
func Infof(format string, args ...any)  { logrus.StandardLogger().Infof(format, args...) }
func Warnf(format string, args ...any)  { logrus.StandardLogger().Warnf(format, args...) }
func Errorf(format string, args ...any) { logrus.StandardLogger().Errorf(format, args...) }
func Fatalf(format string, args ...any) { logrus.StandardLogger().Fatalf(format, args...) }
func Panicf(format string, args ...any) { logrus.StandardLogger().Panicf(format, args...) }

func Print(args ...any)                 { logrus.StandardLogger().Print(args...) }
func Printf(format string, args ...any) { logrus.StandardLogger().Printf(format, args...) }
func Println(args ...any)               { logrus.StandardLogger().Println(args...) }

func WithField(key string, value any) Logger        { return New().WithField(key, value) }
func WithFields(fields Fields) Logger               { return New().WithFields(fields) }
func WithError(err error) Logger                    { return New().WithError(err) }
func WithContext(ctx context.Context) Logger        { return New().WithContext(ctx) }

// ---------------------------------------------------------------------------
// logrus-backed implementation
// ---------------------------------------------------------------------------

type logrusLogger struct {
	entry *logrus.Entry
}

func (l *logrusLogger) Trace(args ...any) { l.entry.Trace(args...) }
func (l *logrusLogger) Debug(args ...any) { l.entry.Debug(args...) }
func (l *logrusLogger) Info(args ...any)  { l.entry.Info(args...) }
func (l *logrusLogger) Warn(args ...any)  { l.entry.Warn(args...) }
func (l *logrusLogger) Error(args ...any) { l.entry.Error(args...) }
func (l *logrusLogger) Fatal(args ...any) { l.entry.Fatal(args...) }
func (l *logrusLogger) Panic(args ...any) { l.entry.Panic(args...) }

func (l *logrusLogger) Tracef(format string, args ...any) { l.entry.Tracef(format, args...) }
func (l *logrusLogger) Debugf(format string, args ...any) { l.entry.Debugf(format, args...) }
func (l *logrusLogger) Infof(format string, args ...any)  { l.entry.Infof(format, args...) }
func (l *logrusLogger) Warnf(format string, args ...any)  { l.entry.Warnf(format, args...) }
func (l *logrusLogger) Errorf(format string, args ...any) { l.entry.Errorf(format, args...) }
func (l *logrusLogger) Fatalf(format string, args ...any) { l.entry.Fatalf(format, args...) }
func (l *logrusLogger) Panicf(format string, args ...any) { l.entry.Panicf(format, args...) }

func (l *logrusLogger) Print(args ...any)                 { l.entry.Print(args...) }
func (l *logrusLogger) Printf(format string, args ...any) { l.entry.Printf(format, args...) }
func (l *logrusLogger) Println(args ...any)               { l.entry.Println(args...) }

func (l *logrusLogger) Warning(args ...any)                 { l.entry.Warning(args...) }
func (l *logrusLogger) Warningf(format string, args ...any) { l.entry.Warningf(format, args...) }
func (l *logrusLogger) Warningln(args ...any)               { l.entry.Warningln(args...) }

func (l *logrusLogger) WithField(key string, value any) Logger {
	return &logrusLogger{entry: l.entry.WithField(key, value)}
}
func (l *logrusLogger) WithFields(fields Fields) Logger {
	return &logrusLogger{entry: l.entry.WithFields(logrus.Fields(fields))}
}
func (l *logrusLogger) WithError(err error) Logger {
	return &logrusLogger{entry: l.entry.WithError(err)}
}
func (l *logrusLogger) WithContext(ctx context.Context) Logger {
	return &logrusLogger{entry: l.entry.WithContext(ctx)}
}
func (l *logrusLogger) Sub(module string) Logger {
	return &logrusLogger{entry: l.entry.WithField("module", module)}
}

func (l *logrusLogger) Level() Level {
	return l.entry.Logger.Level.String()
}

func (l *logrusLogger) Writer() io.Writer { return l.entry.Writer() }

func (l *logrusLogger) WithLevel(level Level) Logger {
	nl := logrus.New()
	nl.SetOutput(l.entry.Logger.Out)
	nl.SetFormatter(l.entry.Logger.Formatter)
	if lvl, err := logrus.ParseLevel(level); err == nil {
		nl.SetLevel(lvl)
	}
	ne := logrus.NewEntry(nl)
	for k, v := range l.entry.Data {
		ne = ne.WithField(k, v)
	}
	return &logrusLogger{entry: ne}
}

func (l *logrusLogger) eventAt(level logrus.Level) Event {
	return &logrusEvent{entry: l.entry, level: level, fields: logrus.Fields{}}
}
func (l *logrusLogger) TraceE() Event { return l.eventAt(logrus.TraceLevel) }
func (l *logrusLogger) DebugE() Event { return l.eventAt(logrus.DebugLevel) }
func (l *logrusLogger) InfoE() Event  { return l.eventAt(logrus.InfoLevel) }
func (l *logrusLogger) WarnE() Event  { return l.eventAt(logrus.WarnLevel) }
func (l *logrusLogger) ErrorE() Event { return l.eventAt(logrus.ErrorLevel) }

type logrusEvent struct {
	entry  *logrus.Entry
	level  logrus.Level
	fields logrus.Fields
}

func (e *logrusEvent) Str(k, v string) Event               { e.fields[k] = v; return e }
func (e *logrusEvent) Strs(k string, v []string) Event     { e.fields[k] = v; return e }
func (e *logrusEvent) Int(k string, v int) Event           { e.fields[k] = v; return e }
func (e *logrusEvent) Int32(k string, v int32) Event       { e.fields[k] = v; return e }
func (e *logrusEvent) Int64(k string, v int64) Event       { e.fields[k] = v; return e }
func (e *logrusEvent) Uint(k string, v uint) Event         { e.fields[k] = v; return e }
func (e *logrusEvent) Uint8(k string, v uint8) Event       { e.fields[k] = v; return e }
func (e *logrusEvent) Uint16(k string, v uint16) Event     { e.fields[k] = v; return e }
func (e *logrusEvent) Uint32(k string, v uint32) Event     { e.fields[k] = v; return e }
func (e *logrusEvent) Uint64(k string, v uint64) Event     { e.fields[k] = v; return e }
func (e *logrusEvent) Float64(k string, v float64) Event   { e.fields[k] = v; return e }
func (e *logrusEvent) Bool(k string, v bool) Event         { e.fields[k] = v; return e }
func (e *logrusEvent) Err(err error) Event                 { e.fields[logrus.ErrorKey] = err; return e }
func (e *logrusEvent) Dur(k string, v time.Duration) Event { e.fields[k] = v.String(); return e }
func (e *logrusEvent) Bytes(k string, v []byte) Event      { e.fields[k] = fmt.Sprintf("%x", v); return e }

func (e *logrusEvent) Msg(msg string) {
	e.entry.WithFields(e.fields).Log(e.level, msg)
}
func (e *logrusEvent) Msgf(format string, args ...any) {
	e.entry.WithFields(e.fields).Logf(e.level, format, args...)
}

// ---------------------------------------------------------------------------
// no-op implementation (zero-cost silence)
// ---------------------------------------------------------------------------

type nopLogger struct{}
type nopEvent struct{}

// Nop returns a Logger that discards everything at near-zero cost.
func Nop() Logger { return nopLogger{} }

func (nopLogger) Trace(...any) {}
func (nopLogger) Debug(...any) {}
func (nopLogger) Info(...any)  {}
func (nopLogger) Warn(...any)  {}
func (nopLogger) Error(...any) {}
func (nopLogger) Fatal(...any) {}
func (nopLogger) Panic(...any) {}

func (nopLogger) Tracef(string, ...any) {}
func (nopLogger) Debugf(string, ...any) {}
func (nopLogger) Infof(string, ...any)  {}
func (nopLogger) Warnf(string, ...any)  {}
func (nopLogger) Errorf(string, ...any) {}
func (nopLogger) Fatalf(string, ...any) {}
func (nopLogger) Panicf(string, ...any) {}
func (nopLogger) Print(...any)          {}
func (nopLogger) Printf(string, ...any) {}
func (nopLogger) Println(...any)        {}
func (nopLogger) Warning(...any)        {}
func (nopLogger) Warningf(string, ...any) {}
func (nopLogger) Warningln(...any)      {}

func (nopLogger) WithField(string, any) Logger        { return nopLogger{} }
func (nopLogger) WithFields(Fields) Logger            { return nopLogger{} }
func (nopLogger) WithError(error) Logger              { return nopLogger{} }
func (nopLogger) WithContext(context.Context) Logger  { return nopLogger{} }
func (nopLogger) Sub(string) Logger                   { return nopLogger{} }
func (nopLogger) Level() Level             { return InfoLevel }
func (nopLogger) WithLevel(Level) Logger   { return nopLogger{} }
func (nopLogger) Writer() io.Writer        { return io.Discard }

func (nopLogger) TraceE() Event { return nopEvent{} }
func (nopLogger) DebugE() Event { return nopEvent{} }
func (nopLogger) InfoE() Event  { return nopEvent{} }
func (nopLogger) WarnE() Event  { return nopEvent{} }
func (nopLogger) ErrorE() Event { return nopEvent{} }

func (nopEvent) Str(string, string) Event        { return nopEvent{} }
func (nopEvent) Strs(string, []string) Event     { return nopEvent{} }
func (nopEvent) Int(string, int) Event           { return nopEvent{} }
func (nopEvent) Int32(string, int32) Event       { return nopEvent{} }
func (nopEvent) Int64(string, int64) Event       { return nopEvent{} }
func (nopEvent) Uint(string, uint) Event         { return nopEvent{} }
func (nopEvent) Uint8(string, uint8) Event       { return nopEvent{} }
func (nopEvent) Uint16(string, uint16) Event     { return nopEvent{} }
func (nopEvent) Uint32(string, uint32) Event     { return nopEvent{} }
func (nopEvent) Uint64(string, uint64) Event     { return nopEvent{} }
func (nopEvent) Float64(string, float64) Event   { return nopEvent{} }
func (nopEvent) Bool(string, bool) Event         { return nopEvent{} }
func (nopEvent) Err(error) Event                 { return nopEvent{} }
func (nopEvent) Dur(string, time.Duration) Event { return nopEvent{} }
func (nopEvent) Bytes(string, []byte) Event      { return nopEvent{} }
func (nopEvent) Msg(string)                      {}
func (nopEvent) Msgf(string, ...any)             {}
