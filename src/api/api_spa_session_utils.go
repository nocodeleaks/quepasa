package api

import (
	models "github.com/nocodeleaks/quepasa/models"
)

// GetSPAOwnedSessionRecord returns the persisted session record for a token and ensures
// the authenticated user is allowed to access it, using session-oriented naming.
func GetSPAOwnedSessionRecord(user *models.QpUser, token string) (*models.QpServer, error) {
	return GetSPAOwnedServerRecord(user, token)
}

// FindSPALiveSession returns the in-memory live session instance when present, using
// session-oriented naming.
func FindSPALiveSession(token string) *models.QpWhatsappSession {
	server := FindSPALiveServer(token)
	if server == nil {
		return nil
	}
	return server
}

// GetSPAOwnedLiveSession returns the in-memory session instance only after the user
// has been authorized against the persisted session record, using session-oriented naming.
func GetSPAOwnedLiveSession(user *models.QpUser, token string) (*models.QpWhatsappSession, error) {
	server, err := GetSPAOwnedLiveServer(user, token)
	if err != nil {
		return nil, err
	}
	return server, nil
}

// EnsureSPASessionReady validates that the live session can serve realtime/message operations,
// using session-oriented naming.
func EnsureSPASessionReady(session *models.QpWhatsappSession) error {
	return EnsureSPAServerReady(session)
}

// CountSPADispatchingForSession counts dispatching rows from the current live session
// when available, otherwise it falls back to persisted dispatching data, using session-oriented naming.
func CountSPADispatchingForSession(token string, liveSession *models.QpWhatsappSession) (dispatchCount, webhookCount, rabbitmqCount int) {
	return CountSPADispatchingForServer(token, liveSession)
}

// BuildSPASessionSummary creates a stable JSON-friendly session summary for SPA reads,
// using session-oriented naming.
func BuildSPASessionSummary(dbServer *models.QpServer, liveSession *models.QpWhatsappSession) map[string]interface{} {
	return BuildSPAServerSummary(dbServer, liveSession)
}

// getSPAOwnedReadySessionHelper is the session-oriented version of getSPAOwnedReadyServer.
// Returns token, session, and bool indicating success.
func getSPAOwnedReadySessionHelper(user *models.QpUser, token string) (*models.QpWhatsappSession, error) {
	session, err := GetSPAOwnedLiveSession(user, token)
	if err != nil {
		return nil, err
	}

	if err := EnsureSPASessionReady(session); err != nil {
		return nil, err
	}

	return session, nil
}
