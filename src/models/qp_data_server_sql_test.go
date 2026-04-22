package models

import (
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

func TestQpDataServerSqlFindByTokenAndUserReturnServerRows(t *testing.T) {
	t.Parallel()

	db, err := sqlx.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	schema := `
	CREATE TABLE servers (
		token TEXT PRIMARY KEY,
		wid TEXT,
		verified BOOLEAN,
		devel BOOLEAN,
		groups INTEGER,
		broadcasts INTEGER,
		readreceipts INTEGER,
		calls INTEGER,
		readupdate INTEGER,
		user TEXT,
		timestamp DATETIME
	);`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("create schema: %v", err)
	}

	if _, err := db.Exec(
		`INSERT INTO servers (token, wid, verified, devel, groups, broadcasts, readreceipts, calls, readupdate, user)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"token-1",
		"5511999999999@s.whatsapp.net",
		true,
		false,
		int(whatsapp.TrueBooleanType),
		int(whatsapp.FalseBooleanType),
		int(whatsapp.UnSetBooleanType),
		int(whatsapp.TrueBooleanType),
		int(whatsapp.FalseBooleanType),
		"tester@example.com",
	); err != nil {
		t.Fatalf("insert server: %v", err)
	}

	store := QpDataServerSql{db: db}

	byToken, err := store.FindByToken("token-1")
	if err != nil {
		t.Fatalf("FindByToken returned error: %v", err)
	}
	if byToken == nil {
		t.Fatal("FindByToken returned nil server")
	}
	if byToken.Token != "token-1" {
		t.Fatalf("expected token token-1, got %q", byToken.Token)
	}
	if byToken.User != "tester@example.com" {
		t.Fatalf("expected user tester@example.com, got %q", byToken.User)
	}

	byUser, err := store.FindForUser("token-1", "tester@example.com")
	if err != nil {
		t.Fatalf("FindForUser returned error: %v", err)
	}
	if byUser == nil {
		t.Fatal("FindForUser returned nil server")
	}
	if byUser.Wid != "5511999999999@s.whatsapp.net" {
		t.Fatalf("expected wid to be loaded, got %q", byUser.Wid)
	}
}
