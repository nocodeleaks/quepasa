package models

import (
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func newAppSettingsDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.MustExec("CREATE TABLE app_settings (id INTEGER PRIMARY KEY, config TEXT NOT NULL DEFAULT '{}', updated_at TIMESTAMP)")
	t.Cleanup(func() { db.Close() })
	return db
}

func TestGlobalConfigRoundTrip(t *testing.T) {
	db := newAppSettingsDB(t)
	got, err := LoadGlobalConfig(db)
	if err != nil {
		t.Fatalf("load empty: %v", err)
	}
	if got.StoreRetentionDays != nil || got.DispatchTypes != nil {
		t.Fatalf("empty config should have nil fields, got %+v", got)
	}
	n := 15
	types := "text,image"
	if err := SaveGlobalConfig(db, GlobalMessageConfig{StoreRetentionDays: &n, DispatchTypes: &types}); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err = LoadGlobalConfig(db)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if got.StoreRetentionDays == nil || *got.StoreRetentionDays != 15 || got.DispatchTypes == nil || *got.DispatchTypes != "text,image" {
		t.Fatalf("round-trip mismatch: %+v", got)
	}
	if err := SaveGlobalConfig(db, GlobalMessageConfig{}); err != nil {
		t.Fatalf("save empty: %v", err)
	}
	var count int
	db.Get(&count, "SELECT count(*) FROM app_settings")
	if count != 1 {
		t.Fatalf("expected exactly 1 row, got %d", count)
	}
}
