package models

import (
	"time"

	"github.com/nocodeleaks/quepasa/environment"
)

const storeRetentionNone = -1

// ResolvedMessageSettings is the effective per-message config after cascade
// resolution. M2: env only; M3/M4 layer per-server and global overrides into
// ResolveMessageSettings.
type ResolvedMessageSettings struct {
	RetentionDays int
	DispatchTypes map[string]bool
}

func (r ResolvedMessageSettings) Store() bool { return r.RetentionDays != storeRetentionNone }

// ExpiryFor returns the record ExpiresAt: zero (never-expire → NULL in postgres)
// for forever, now+N days otherwise. Not meaningful when !Store().
func (r ResolvedMessageSettings) ExpiryFor() time.Time {
	if r.RetentionDays <= 0 {
		return time.Time{}
	}
	return time.Now().Add(time.Duration(r.RetentionDays) * 24 * time.Hour)
}

func (r ResolvedMessageSettings) DispatchAllowed(msgType string) bool {
	if len(r.DispatchTypes) == 0 {
		return true
	}
	return r.DispatchTypes[msgType]
}

// ResolveMessageSettings resolves the effective settings for a server.
// M2: env only (server unused now; M3 adds the per-caixa override).
func ResolveMessageSettings(server *QpWhatsappServer) ResolvedMessageSettings {
	m := environment.Settings.Messages
	return ResolvedMessageSettings{
		RetentionDays: m.RetentionDays,
		DispatchTypes: m.DispatchTypes,
	}
}
