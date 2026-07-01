package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"sync"

	"github.com/jmoiron/sqlx"
)

// GlobalMessageConfig is the instance-wide override tier (nil field = inherit env).
type GlobalMessageConfig struct {
	StoreRetentionDays *int    `json:"store_retention_days,omitempty"`
	DispatchTypes      *string `json:"dispatch_types,omitempty"`
}

// LoadGlobalConfig reads the single app_settings row (id=1). Missing row = empty config.
func LoadGlobalConfig(db *sqlx.DB) (GlobalMessageConfig, error) {
	var cfg GlobalMessageConfig
	var raw string
	err := db.Get(&raw, "SELECT config FROM app_settings WHERE id = 1")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return cfg, nil
		}
		return cfg, err
	}
	if raw == "" {
		return cfg, nil
	}
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return GlobalMessageConfig{}, err
	}
	return cfg, nil
}

// SaveGlobalConfig upserts the single app_settings row (id=1).
func SaveGlobalConfig(db *sqlx.DB, cfg GlobalMessageConfig) error {
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	_, err = db.Exec(`INSERT INTO app_settings (id, config, updated_at) VALUES (1, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(id) DO UPDATE SET config = excluded.config, updated_at = CURRENT_TIMESTAMP`, string(data))
	return err
}

// --- process cache (no-restart reload) ---
var (
	globalCfg    GlobalMessageConfig
	globalMu     sync.RWMutex
	globalLoaded bool
)

// GetGlobalMessageConfig returns the cached global config, lazily loading it once.
func GetGlobalMessageConfig() GlobalMessageConfig {
	globalMu.RLock()
	if globalLoaded {
		defer globalMu.RUnlock()
		return globalCfg
	}
	globalMu.RUnlock()

	globalMu.Lock()
	defer globalMu.Unlock()
	if !globalLoaded {
		if cfg, err := LoadGlobalConfig(GetDB()); err == nil {
			globalCfg = cfg
		}
		globalLoaded = true
	}
	return globalCfg
}

// SetGlobalMessageConfig persists + refreshes the cache (runtime, no restart).
func SetGlobalMessageConfig(cfg GlobalMessageConfig) error {
	if err := SaveGlobalConfig(GetDB(), cfg); err != nil {
		return err
	}
	globalMu.Lock()
	globalCfg = cfg
	globalLoaded = true
	globalMu.Unlock()
	return nil
}

// SetGlobalMessageConfigForTest overrides the in-memory global config (tests only).
func SetGlobalMessageConfigForTest(cfg GlobalMessageConfig) {
	globalMu.Lock()
	globalCfg = cfg
	globalLoaded = true
	globalMu.Unlock()
}
