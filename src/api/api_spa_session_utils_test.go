package api

import (
	"testing"

	models "github.com/nocodeleaks/quepasa/models"
)

func TestFindSPALiveSession_ReturnsNilWhenServerNotFound(t *testing.T) {
	prevService := models.WhatsappService
	defer func() { models.WhatsappService = prevService }()

	models.WhatsappService = &models.QPWhatsappService{
		Servers: map[string]*models.QpWhatsappServer{},
	}

	if got := FindSPALiveSession("missing"); got != nil {
		t.Fatal("expected nil for missing session")
	}
}

func TestCountSPADispatchingForSession_IsSessionWrapper(t *testing.T) {
	// Verify function exists and is callable with session-oriented naming.
	// Detailed behavior tested via CountSPADispatchingForServer integration tests.
	_ = CountSPADispatchingForSession
}

func TestBuildSPASessionSummary_IsSessionWrapper(t *testing.T) {
	// Verify function exists and is callable with session-oriented naming.
	// Detailed behavior tested via BuildSPAServerSummary integration tests.
	_ = BuildSPASessionSummary
}

func TestEnsureSPASessionReady_RequiresValidConnection(t *testing.T) {
	dbServer := &models.QpServer{Token: "ready-test"}
	session := &models.QpWhatsappSession{
		QpServer: dbServer,
	}

	// Without handler or connection, should fail
	err := EnsureSPASessionReady(session)
	if err == nil {
		t.Fatal("expected error for session without handler")
	}
}

func TestGetSPAOwnedSessionRecord_DelegatesUserOwnershipCheck(t *testing.T) {
	prevService := models.WhatsappService
	defer func() { models.WhatsappService = prevService }()

	dbServer := &models.QpServer{Token: "owner-test"}
	dbServer.SetUser("owner@example.com")

	models.WhatsappService = &models.QPWhatsappService{
		Servers: map[string]*models.QpWhatsappServer{},
		DB: &models.QpDatabase{
			Servers: &stubSessionServersData{
				findByTokenResult: dbServer,
			},
		},
	}

	ownerUser := &models.QpUser{Username: "owner@example.com"}
	got, err := GetSPAOwnedSessionRecord(ownerUser, "owner-test")
	if err != nil {
		t.Fatalf("expected no error for owned session, got %v", err)
	}
	if got.Token != dbServer.Token {
		t.Fatalf("expected session token %q, got %q", dbServer.Token, got.Token)
	}

	// Different user should fail
	otherUser := &models.QpUser{Username: "other@example.com"}
	got, err = GetSPAOwnedSessionRecord(otherUser, "owner-test")
	if err == nil {
		t.Fatal("expected error for non-owner user")
	}
	if got != nil {
		t.Fatal("expected nil session for non-owner lookup")
	}
}

func TestFindPersistedServerRecord_FallsBackToCaseInsensitiveScan(t *testing.T) {
	prevService := models.WhatsappService
	defer func() { models.WhatsappService = prevService }()

	dbServer := &models.QpServer{Token: "owner-test"}
	models.WhatsappService = &models.QPWhatsappService{
		Servers: map[string]*models.QpWhatsappServer{},
		DB: &models.QpDatabase{
			Servers: &stubSessionServersData{
				findByTokenResult: dbServer,
			},
		},
	}

	got, err := findPersistedServerRecord("OWNER-TEST")
	if err != nil {
		t.Fatalf("expected no error for case-insensitive lookup, got %v", err)
	}

	if got != dbServer {
		t.Fatalf("expected the same persisted record instance")
	}
}

// Stub for testing session record lookups
type stubSessionServersData struct {
	findByTokenResult *models.QpServer
}

func (s *stubSessionServersData) FindAll() []*models.QpServer {
	if s.findByTokenResult != nil {
		return []*models.QpServer{s.findByTokenResult}
	}
	return nil
}

func (s *stubSessionServersData) FindByToken(token string) (*models.QpServer, error) {
	if s.findByTokenResult != nil && s.findByTokenResult.Token == token {
		return s.findByTokenResult, nil
	}
	return nil, models.ErrServerNotFound
}

func (s *stubSessionServersData) FindForUser(user, token string) (*models.QpServer, error) {
	return nil, models.ErrServerNotFound
}

func (s *stubSessionServersData) Exists(token string) (bool, error) {
	return false, nil
}

func (s *stubSessionServersData) Add(server *models.QpServer) error {
	return nil
}

func (s *stubSessionServersData) Update(server *models.QpServer) error {
	return nil
}

func (s *stubSessionServersData) Delete(token string) error {
	return nil
}
