package models

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func newSpamSectionsTestStore(t *testing.T) (*sqlx.DB, QpDataSpamSectionsInterface) {
	t.Helper()

	db, err := sqlx.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	schema := `
		CREATE TABLE servers (
			token TEXT PRIMARY KEY
		);

		CREATE TABLE spam_sections (
			token TEXT PRIMARY KEY NOT NULL REFERENCES servers(token) ON DELETE CASCADE,
			position INTEGER NOT NULL DEFAULT 0,
			enabled BOOLEAN NOT NULL DEFAULT TRUE,
			label TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`
	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		t.Fatalf("create schema: %v", err)
	}

	for _, token := range []string{"token-1", "token-2", "token-3"} {
		if _, err := db.Exec("INSERT INTO servers (token) VALUES (?)", token); err != nil {
			_ = db.Close()
			t.Fatalf("insert server %s: %v", token, err)
		}
	}

	return db, NewQpDataSpamSectionsSql(db)
}

func TestQpDataSpamSectionsSqlUpsertAssignsNextPosition(t *testing.T) {
	t.Parallel()

	db, store := newSpamSectionsTestStore(t)
	defer db.Close()

	if err := store.Upsert(&QpSpamSection{Token: "token-1", Enabled: true, Label: " primeira "}); err != nil {
		t.Fatalf("upsert token-1: %v", err)
	}
	if err := store.Upsert(&QpSpamSection{Token: "token-2", Enabled: true}); err != nil {
		t.Fatalf("upsert token-2: %v", err)
	}

	first, err := store.Find("token-1")
	if err != nil {
		t.Fatalf("find token-1: %v", err)
	}
	if first.Position != 10 {
		t.Fatalf("expected first position 10, got %d", first.Position)
	}
	if first.Label != "primeira" {
		t.Fatalf("expected trimmed label, got %q", first.Label)
	}

	second, err := store.Find("token-2")
	if err != nil {
		t.Fatalf("find token-2: %v", err)
	}
	if second.Position != 20 {
		t.Fatalf("expected second position 20, got %d", second.Position)
	}
}

func TestQpDataSpamSectionsSqlListAllReturnsConfiguredOrder(t *testing.T) {
	t.Parallel()

	db, store := newSpamSectionsTestStore(t)
	defer db.Close()

	fixtures := []*QpSpamSection{
		{Token: "token-2", Position: 20, Enabled: true},
		{Token: "token-3", Position: 30, Enabled: false},
		{Token: "token-1", Position: 10, Enabled: true},
	}
	for _, fixture := range fixtures {
		if err := store.Upsert(fixture); err != nil {
			t.Fatalf("upsert %s: %v", fixture.Token, err)
		}
	}

	items, err := store.ListAll()
	if err != nil {
		t.Fatalf("list all: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}

	expected := []string{"token-1", "token-2", "token-3"}
	for index, token := range expected {
		if items[index].Token != token {
			t.Fatalf("expected item %d to be %s, got %s", index, token, items[index].Token)
		}
	}
	if items[2].Enabled {
		t.Fatal("expected token-3 to remain disabled")
	}
}

func TestQpDataSpamSectionsSqlUpdatePositionAndDelete(t *testing.T) {
	t.Parallel()

	db, store := newSpamSectionsTestStore(t)
	defer db.Close()

	for _, token := range []string{"token-1", "token-2"} {
		if err := store.Upsert(&QpSpamSection{Token: token, Enabled: true}); err != nil {
			t.Fatalf("upsert %s: %v", token, err)
		}
	}

	if err := store.UpdatePosition("token-2", 5); err != nil {
		t.Fatalf("update position: %v", err)
	}

	items, err := store.ListAll()
	if err != nil {
		t.Fatalf("list after reorder: %v", err)
	}
	if items[0].Token != "token-2" {
		t.Fatalf("expected token-2 first after reorder, got %s", items[0].Token)
	}

	removed, err := store.Delete("token-2")
	if err != nil {
		t.Fatalf("delete token-2: %v", err)
	}
	if !removed {
		t.Fatal("expected token-2 to be removed")
	}

	removedAgain, err := store.Delete("token-2")
	if err != nil {
		t.Fatalf("delete token-2 again: %v", err)
	}
	if removedAgain {
		t.Fatal("expected second delete to report no removal")
	}

	if _, err := store.Find("token-2"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows for deleted token, got %v", err)
	}
}
