package rtp

import qplog "github.com/nocodeleaks/quepasa/qplog"

// Option configures optional, non-behavioral aspects (currently the diagnostic logger).
type Option func(*config)

type config struct {
	log qplog.Logger
}

func resolveConfig(opts []Option) config {
	c := config{log: qplog.Nop()}
	for _, opt := range opts {
		opt(&c)
	}
	return c
}

// WithLogger sets the zerolog logger for debug/trace diagnostics; default is silent.
func WithLogger(l qplog.Logger) Option {
	return func(c *config) { c.log = l }
}

// pickLog returns the first supplied logger, or a silent Nop logger when none is given.
func pickLog(log []qplog.Logger) qplog.Logger {
	if len(log) > 0 {
		return log[0]
	}
	return qplog.Nop()
}
