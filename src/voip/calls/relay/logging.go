package relay

import qplog "github.com/nocodeleaks/quepasa/qplog"

// Option configures optional, non-behavioral aspects of the relay channel —
// currently the diagnostic logger. The zero configuration logs nothing.
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

// WithLogger sets the zerolog logger for debug/trace diagnostics. The library never
// configures logging itself; without this option the channel is silent at zero cost.
// Pass the logger from a context, e.g. WithLogger(*zerolog.Ctx(ctx)).
func WithLogger(l qplog.Logger) Option {
	return func(c *config) { c.log = l }
}

func pickLog(log []qplog.Logger) qplog.Logger {
	if len(log) > 0 {
		return log[0]
	}
	return qplog.Nop()
}
