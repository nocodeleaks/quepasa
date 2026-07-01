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

// ResolveMessageSettings resolves the effective settings for a server:
// per-caixa override (server value beats env), falling back to env when NULL.
func ResolveMessageSettings(server *QpWhatsappServer) ResolvedMessageSettings {
	m := environment.Settings.Messages
	retention := m.RetentionDays
	types := m.DispatchTypes
	if server != nil {
		cfg := server.QpServer
		if cfg != nil && cfg.StoreRetentionDays.Valid {
			retention = int(cfg.StoreRetentionDays.Int64)
		}
		if cfg != nil && cfg.DispatchTypes.Valid && cfg.DispatchTypes.String != "" {
			types = environment.ParseDispatchTypes(cfg.DispatchTypes.String)
		}
	}
	return ResolvedMessageSettings{RetentionDays: retention, DispatchTypes: types}
}
