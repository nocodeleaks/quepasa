package models

import (
	"strings"
	"testing"
)

// TestGenerateAPIKey checks the key format and that the returned hash matches the
// plaintext, while two generations never collide.
func TestGenerateAPIKey(t *testing.T) {
	plain, hash, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if !strings.HasPrefix(plain, "qp_") {
		t.Fatalf("plaintext key missing qp_ prefix: %q", plain)
	}
	if len(plain) != len("qp_")+64 { // 32 bytes -> 64 hex chars
		t.Fatalf("unexpected key length: %d", len(plain))
	}
	if hash != HashAPIKey(plain) {
		t.Fatal("returned hash does not match HashAPIKey(plaintext)")
	}
	if hash == plain {
		t.Fatal("hash must not equal the plaintext key")
	}

	plain2, _, _ := GenerateAPIKey()
	if plain2 == plain {
		t.Fatal("two generated keys collided")
	}
}

// TestHashAPIKeyRejectsEmpty ensures blank input hashes to "" so callers can deny it.
func TestHashAPIKeyRejectsEmpty(t *testing.T) {
	if HashAPIKey("   ") != "" {
		t.Fatal("whitespace key should hash to empty")
	}
}

// TestUserAPIKeyStoreRoundTrip exercises the SQL layer: set a key hash, then
// resolve the user back by that hash; a wrong hash finds nothing; clearing
// removes it.
func TestUserAPIKeyStoreRoundTrip(t *testing.T) {
	db := setupUserSQLTestDB(t)
	users := NewQpDataUserSql(db)

	restore := setUserSQLTestEnv(t, ENV_ACCOUNTSETUP, "true")
	defer restore()

	if _, err := users.Create("owner@example.com", "Password123!"); err != nil {
		t.Fatalf("create user: %v", err)
	}

	plain, hash, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	if err := users.SetAPIKey("owner@example.com", hash); err != nil {
		t.Fatalf("set api key: %v", err)
	}

	found, err := users.FindByAPIKey(HashAPIKey(plain))
	if err != nil {
		t.Fatalf("find by api key: %v", err)
	}
	if found == nil || found.Username != "owner@example.com" {
		t.Fatalf("api key did not resolve to the owner: %+v", found)
	}
	if found.APIKeyRotatedAt == nil {
		t.Fatal("rotation timestamp not stamped on SetAPIKey")
	}

	if _, err := users.FindByAPIKey(HashAPIKey("qp_wrongkey")); err == nil {
		t.Fatal("a wrong key must not resolve to any user")
	}

	if err := users.ClearAPIKey("owner@example.com"); err != nil {
		t.Fatalf("clear api key: %v", err)
	}
	if _, err := users.FindByAPIKey(hash); err == nil {
		t.Fatal("cleared key must no longer resolve")
	}
}
