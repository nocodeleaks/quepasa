package api

import "strings"

func isMasterKeyConfigured(masterKey string) bool {
	return strings.TrimSpace(masterKey) != ""
}

func buildMasterKeyStatusResponse(masterKey string) map[string]interface{} {
	return map[string]interface{}{
		"configured": isMasterKeyConfigured(masterKey),
	}
}
