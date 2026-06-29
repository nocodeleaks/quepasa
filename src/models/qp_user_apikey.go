package models

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// apiKeyPrefix tags the plaintext personal API key so it is recognizable in
// logs/configs and distinguishable from session tokens.
const apiKeyPrefix = "qp_"

// GenerateAPIKey creates a new personal API key. It returns the plaintext key
// (shown to the user exactly once) and the SHA-256 hash that is persisted.
//
// The key carries 256 bits of entropy, so an unsalted SHA-256 is an appropriate
// store: it allows O(1) lookup on authentication while keeping the plaintext out
// of the database (a leaked dump does not expose usable keys).
func GenerateAPIKey() (plaintext string, hash string, err error) {
	buf := make([]byte, 32)
	if _, err = rand.Read(buf); err != nil {
		return "", "", err
	}
	plaintext = apiKeyPrefix + hex.EncodeToString(buf)
	return plaintext, HashAPIKey(plaintext), nil
}

// HashAPIKey returns the hex SHA-256 of a plaintext API key, trimming whitespace.
// Returns "" for an empty input so callers can reject it.
func HashAPIKey(plaintext string) string {
	plaintext = strings.TrimSpace(plaintext)
	if plaintext == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(sum[:])
}
