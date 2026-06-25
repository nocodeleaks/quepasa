package stun

import qplog "github.com/nocodeleaks/quepasa/qplog"

// pickLog returns the first logger from a variadic logger argument, or a silent
// no-op logger when none was supplied. The stateless encoders/parsers accept a
// trailing variadic logger and resolve it here so callers that pass nothing stay
// silent.
func pickLog(log []qplog.Logger) qplog.Logger {
	if len(log) > 0 {
		return log[0]
	}
	return qplog.Nop()
}
