package api

import (
	models "github.com/nocodeleaks/quepasa/models"
)

// GetOwnedSessionRecord returns the persisted session record for a token and ensures
// the authenticated user is allowed to access it, using session-oriented naming.
func GetOwnedSessionRecord(user *models.QpUser, token string) (*models.QpServer, error) {
	return GetOwnedServerRecord(user, token)
}

// FindLiveSession returns the in-memory live session instance when present, using
// session-oriented naming.
func FindLiveSession(token string) *models.QpWhatsappSession {
	server := FindLiveServer(token)
	if server == nil {
		return nil
	}
	return server
}

// GetOwnedLiveSession returns the in-memory session instance only after the user
// has been authorized against the persisted session record, using session-oriented naming.
func GetOwnedLiveSession(user *models.QpUser, token string) (*models.QpWhatsappSession, error) {
	server, err := GetOwnedLiveServer(user, token)
	if err != nil {
		return nil, err
	}
	return server, nil
}

// EnsureLiveSessionReady validates that the live session can serve realtime/message operations,
// using session-oriented naming.
func EnsureLiveSessionReady(session *models.QpWhatsappSession) error {
	return EnsureLiveServerReady(session)
}

// CountDispatchingForSession counts dispatching rows from the current live session
// when available, otherwise it falls back to persisted dispatching data, using session-oriented naming.
func CountDispatchingForSession(token string, liveSession *models.QpWhatsappSession) (dispatchCount, webhookCount, rabbitmqCount int) {
	return CountDispatchingForServer(token, liveSession)
}

// BuildSessionSummary creates a stable JSON-friendly session summary for SPA reads,
// using session-oriented naming.
func BuildSessionSummary(dbServer *models.QpServer, liveSession *models.QpWhatsappSession) map[string]interface{} {
	return BuildServerSummary(dbServer, liveSession)
}

// getOwnedReadySessionHelper is the session-oriented version of the authenticated ready-server lookup.
func getOwnedReadySessionHelper(user *models.QpUser, token string) (*models.QpWhatsappSession, error) {
	session, err := GetOwnedLiveSession(user, token)
	if err != nil {
		return nil, err
	}

	if err := EnsureLiveSessionReady(session); err != nil {
		return nil, err
	}

	return session, nil
}
