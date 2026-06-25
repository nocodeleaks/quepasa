package api

import (
	"net/http"
	"strings"

	environment "github.com/nocodeleaks/quepasa/environment"
	library "github.com/nocodeleaks/quepasa/library"
	models "github.com/nocodeleaks/quepasa/models"
)

// masterKeyHeader points to the shared library constant so the string literal
// is defined in exactly one place across all modules.
const masterKeyHeader = library.HeaderMasterKey

// isMasterKeyEnabled reports whether a MASTERKEY has been configured in the environment.
func isMasterKeyEnabled() bool {
	return strings.TrimSpace(models.ENV.MasterKey()) != ""
}

func isMasterKeyConfigured(masterKey string) bool {
	return strings.TrimSpace(masterKey) != ""
}

// isMasterKeyRequest validates the X-Master-Key header against the configured master key.
// Returns false when the master key is not configured or the header does not match.
func isMasterKeyRequest(r *http.Request) bool {
	configured := strings.TrimSpace(models.ENV.MasterKey())
	if configured == "" {
		return false
	}
	candidate := strings.TrimSpace(r.Header.Get(masterKeyHeader))
	return candidate != "" && candidate == configured
}

// isRelaxedSessions reports whether session creation is open to any authenticated user.
// Controlled by RELAXED_SESSIONS env var; defaults to true (open).
// When false, session creation requires a master key in addition to a valid user.
func isRelaxedSessions() bool {
	return environment.Settings.API.RelaxedSessions
}

func buildMasterKeyStatusResponse(masterKey string) map[string]interface{} {
	return map[string]interface{}{
		"configured": isMasterKeyConfigured(masterKey),
	}
}
