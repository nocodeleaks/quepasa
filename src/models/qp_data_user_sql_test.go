package models

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func TestQpDataUserSqlFindAllAndDelete(t *testing.T) {
	db := setupUserSQLTestDB(t)
	users := NewQpDataUserSql(db)

	restore := setUserSQLTestEnv(t, ENV_ACCOUNTSETUP, "true")
	defer restore()

	if _, err := users.Create("beta@example.com", "Password123!"); err != nil {
		t.Fatalf("create beta user: %v", err)
	}
	if _, err := users.Create("alpha@example.com", "Password123!"); err != nil {
		t.Fatalf("create alpha user: %v", err)
	}

	allUsers, err := users.FindAll()
	if err != nil {
		t.Fatalf("find all users: %v", err)
	}

	if len(allUsers) != 2 {
		t.Fatalf("expected 2 users, got %d", len(allUsers))
	}

	if allUsers[0].Username != "alpha@example.com" || allUsers[1].Username != "beta@example.com" {
		t.Fatalf("expected users sorted by username, got %+v", allUsers)
	}

	if err := users.Delete("alpha@example.com"); err != nil {
		t.Fatalf("delete alpha user: %v", err)
	}

	remainingUsers, err := users.FindAll()
	if err != nil {
		t.Fatalf("find all users after delete: %v", err)
	}

	if len(remainingUsers) != 1 || remainingUsers[0].Username != "beta@example.com" {
		t.Fatalf("unexpected users after delete: %+v", remainingUsers)
	}

	if err := users.Delete("missing@example.com"); err == nil {
		t.Fatalf("expected deleting a missing user to fail")
	}
}

func setupUserSQLTestDB(t *testing.T) *sqlx.DB {
	t.Helper()

	db, err := sqlx.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	schema := `
		CREATE TABLE users (
			username TEXT PRIMARY KEY,
			password TEXT NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("create users schema: %v", err)
	}

	return db
}

func setUserSQLTestEnv(t *testing.T, key string, value string) func() {
	t.Helper()

	oldValue, hadValue := os.LookupEnv(key)
	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("set %s: %v", key, err)
	}

	return func() {
		if !hadValue {
			_ = os.Unsetenv(key)
			return
		}
		_ = os.Setenv(key, oldValue)
	}
}
