package runtime

import (
	"fmt"
	"testing"

	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

func TestStartSessionNilReturnsError(t *testing.T) {
	err := StartSession(nil)
	if err != ErrNilSession {
		t.Fatalf("expected ErrNilSession, got %v", err)
	}
}

func TestStopSessionNilReturnsError(t *testing.T) {
	err := StopSession(nil, "test")
	if err != ErrNilSession {
		t.Fatalf("expected ErrNilSession, got %v", err)
	}
}

func TestRestartSessionNilReturnsError(t *testing.T) {
	err := RestartSession(nil)
	if err != ErrNilSession {
		t.Fatalf("expected ErrNilSession, got %v", err)
	}
}

func TestSendSessionMessageNilReturnsError(t *testing.T) {
	_, err := SendSessionMessage(nil, nil)
	if err != ErrNilSession {
		t.Fatalf("expected ErrNilSession, got %v", err)
	}
}

func TestSaveSessionNilReturnsError(t *testing.T) {
	err := SaveSession(nil, "test")
	if err != ErrNilSession {
		t.Fatalf("expected ErrNilSession, got %v", err)
	}
}

func TestRestartSessionAsyncNilIsNoop(t *testing.T) {
	RestartSessionAsync(nil)
}

func TestToggleSessionDebugNilReturnsError(t *testing.T) {
	_, err := ToggleSessionDebug(nil)
	if err != ErrNilSession {
		t.Fatalf("expected ErrNilSession, got %v", err)
	}
}

func TestToggleSessionOptionNilReturnsError(t *testing.T) {
	_, err := ToggleSessionOption(nil, "groups")
	if err != ErrNilSession {
		t.Fatalf("expected ErrNilSession, got %v", err)
	}
}

func TestApplySessionConfigurationPatchNilReturnsError(t *testing.T) {
	_, err := ApplySessionConfigurationPatch(nil, &SessionConfigurationPatch{})
	if err != ErrNilSession {
		t.Fatalf("expected ErrNilSession, got %v", err)
	}
}

func TestCreateSessionRecordNilInfoReturnsError(t *testing.T) {
	_, err := CreateSessionRecord(nil, "test")
	if err != ErrNilSessionInfo {
		t.Fatalf("expected ErrNilSessionInfo, got %v", err)
	}
}

func TestLoadSessionRecordNilInfoReturnsError(t *testing.T) {
	_, err := LoadSessionRecord(nil)
	if err != ErrNilSessionInfo {
		t.Fatalf("expected ErrNilSessionInfo, got %v", err)
	}
}

func TestDeleteSessionRecordNilReturnsError(t *testing.T) {
	err := DeleteSessionRecord(nil, "test")
	if err != ErrNilSession {
		t.Fatalf("expected ErrNilSession, got %v", err)
	}
}

func TestApplySessionUserNilReturnsError(t *testing.T) {
	_, err := ApplySessionUser(nil, "user@example.com")
	if err != ErrNilSession {
		t.Fatalf("expected ErrNilSession, got %v", err)
	}
}

func TestBuildSessionRecordAppliesPatch(t *testing.T) {
	groups := whatsapp.TrueBooleanType
	readUpdate := whatsapp.FalseBooleanType
	devel := true

	info := BuildSessionRecord("token-123", "user@example.com", &SessionConfigurationPatch{
		Groups:     &groups,
		ReadUpdate: &readUpdate,
		Devel:      &devel,
	})

	if info == nil {
		t.Fatalf("expected session record")
	}

	if info.Token != "token-123" {
		t.Fatalf("expected token to be preserved")
	}

	if info.GetUser() != "user@example.com" {
		t.Fatalf("expected user to be applied")
	}

	if info.Groups != groups {
		t.Fatalf("expected groups to be applied")
	}

	if info.ReadUpdate != readUpdate {
		t.Fatalf("expected readupdate to be applied")
	}

	if info.Devel != devel {
		t.Fatalf("expected devel to be applied")
	}
}

func TestFindLiveSessionByTokenMatchesExistingSession(t *testing.T) {
	previousService := models.WhatsappService
	defer func() {
		models.WhatsappService = previousService
	}()

	session := &models.QpWhatsappSession{QpServer: &models.QpServer{Token: "Token-123"}}
	models.WhatsappService = &models.QPWhatsappService{
		Servers: map[string]*models.QpWhatsappServer{
			"token-123": session,
		},
	}

	found, ok := FindLiveSessionByToken("token-123")
	if !ok {
		t.Fatalf("expected live session lookup to succeed")
	}

	if found != session {
		t.Fatalf("expected the same live session instance")
	}
}

func TestGetLiveSessionByTokenMissingReturnsNotFound(t *testing.T) {
	previousService := models.WhatsappService
	defer func() {
		models.WhatsappService = previousService
	}()

	models.WhatsappService = &models.QPWhatsappService{Servers: map[string]*models.QpWhatsappServer{}}

	_, err := GetLiveSessionByToken("missing")
	if err != models.ErrSessionNotFound {
		t.Fatalf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestListLiveSessionsForUserFiltersOwner(t *testing.T) {
	previousService := models.WhatsappService
	defer func() {
		models.WhatsappService = previousService
	}()

	owned := &models.QpWhatsappSession{QpServer: &models.QpServer{Token: "owned-token"}}
	owned.SetUser("owner@example.com")
	other := &models.QpWhatsappSession{QpServer: &models.QpServer{Token: "other-token"}}
	other.SetUser("other@example.com")

	models.WhatsappService = &models.QPWhatsappService{
		Servers: map[string]*models.QpWhatsappServer{
			owned.Token: owned,
			other.Token: other,
		},
	}

	sessions := ListLiveSessionsForUser("owner@example.com")
	if len(sessions) != 1 {
		t.Fatalf("expected one owned session, got %d", len(sessions))
	}

	if sessions[0] != owned {
		t.Fatalf("expected the owned session instance")
	}
}

func TestGetSessionDownloadPrefixUsesLiveSessionToken(t *testing.T) {
	previousService := models.WhatsappService
	defer func() {
		models.WhatsappService = previousService
	}()

	session := &models.QpWhatsappSession{QpServer: &models.QpServer{Token: "token-123"}}
	models.WhatsappService = &models.QPWhatsappService{
		Servers: map[string]*models.QpWhatsappServer{
			session.Token: session,
		},
	}

	prefix, err := GetSessionDownloadPrefix("token-123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if prefix != "/bot/token-123/download" {
		t.Fatalf("unexpected download prefix: %s", prefix)
	}
}

func TestListPersistedSessionRecordsNilServiceReturnsError(t *testing.T) {
	previousService := models.WhatsappService
	defer func() {
		models.WhatsappService = previousService
	}()

	models.WhatsappService = nil

	_, err := ListPersistedSessionRecords()
	if err == nil {
		t.Fatalf("expected error for missing service")
	}
}

func TestFindPersistedUserNilServiceReturnsError(t *testing.T) {
	previousService := models.WhatsappService
	defer func() {
		models.WhatsappService = previousService
	}()

	models.WhatsappService = nil

	_, err := FindPersistedUser("owner@example.com")
	if err == nil {
		t.Fatalf("expected error for missing user service")
	}
}

func TestUpdatePersistedUserPasswordDelegatesToUserStore(t *testing.T) {
	previousService := models.WhatsappService
	defer func() {
		models.WhatsappService = previousService
	}()

	users := &stubRuntimeUsersData{existsResult: true}
	models.WhatsappService = &models.QPWhatsappService{
		DB: &models.QpDatabase{Users: users},
	}

	err := UpdatePersistedUserPassword(" owner@example.com ", "secret")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if users.updatePasswordUsername != "owner@example.com" {
		t.Fatalf("expected trimmed username in password update")
	}
}

func TestCountPersistedUsersDelegatesToUserStore(t *testing.T) {
	previousService := models.WhatsappService
	defer func() {
		models.WhatsappService = previousService
	}()

	users := &stubRuntimeUsersData{countResult: 3}
	models.WhatsappService = &models.QPWhatsappService{
		DB: &models.QpDatabase{Users: users},
	}

	count, err := CountPersistedUsers()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if count != 3 {
		t.Fatalf("expected count 3, got %d", count)
	}
}

func TestListPersistedUsersDelegatesToUserStore(t *testing.T) {
	previousService := models.WhatsappService
	defer func() {
		models.WhatsappService = previousService
	}()

	users := &stubRuntimeUsersData{findAllResult: []*models.QpUser{{Username: "owner@example.com"}}}
	models.WhatsappService = &models.QPWhatsappService{
		DB: &models.QpDatabase{Users: users},
	}

	list, err := ListPersistedUsers()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(list) != 1 || list[0].Username != "owner@example.com" {
		t.Fatalf("expected persisted users list to be returned")
	}
}

func TestCreatePersistedUserDelegatesToUserStore(t *testing.T) {
	previousService := models.WhatsappService
	defer func() {
		models.WhatsappService = previousService
	}()

	users := &stubRuntimeUsersData{createResult: &models.QpUser{Username: "owner@example.com"}}
	models.WhatsappService = &models.QPWhatsappService{
		DB: &models.QpDatabase{Users: users},
	}

	created, err := CreatePersistedUser(" owner@example.com ", "secret")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if created == nil || created.Username != "owner@example.com" {
		t.Fatalf("expected created user result")
	}

	if users.createUsername != "owner@example.com" {
		t.Fatalf("expected trimmed username in create")
	}
}

func TestDeletePersistedUserDelegatesToUserStore(t *testing.T) {
	previousService := models.WhatsappService
	defer func() {
		models.WhatsappService = previousService
	}()

	users := &stubRuntimeUsersData{}
	models.WhatsappService = &models.QPWhatsappService{
		DB: &models.QpDatabase{Users: users},
	}

	err := DeletePersistedUser(" owner@example.com ")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if users.deleteUsername != "owner@example.com" {
		t.Fatalf("expected trimmed username in delete")
	}
}

func TestGetConversationLabelStoreNilServiceReturnsError(t *testing.T) {
	previousService := models.WhatsappService
	defer func() {
		models.WhatsappService = previousService
	}()

	models.WhatsappService = nil

	_, err := GetConversationLabelStore()
	if err == nil {
		t.Fatalf("expected error for missing conversation label service")
	}
}

func TestGetFirstReadySessionMissingReturnsNotFound(t *testing.T) {
	previousService := models.WhatsappService
	defer func() {
		models.WhatsappService = previousService
	}()

	models.WhatsappService = &models.QPWhatsappService{
		Servers: map[string]*models.QpWhatsappServer{},
	}

	_, err := GetFirstReadySession()
	if err != models.ErrSessionNotFound {
		t.Fatalf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestDiagnoseOrphanedSessionsNilServiceReturnsError(t *testing.T) {
	previousService := models.WhatsappService
	defer func() {
		models.WhatsappService = previousService
	}()

	models.WhatsappService = nil

	_, err := DiagnoseOrphanedSessions()
	if err != ErrSessionServiceUnavailable {
		t.Fatalf("expected ErrSessionServiceUnavailable, got %v", err)
	}
}

type stubRuntimeUsersData struct {
	countResult            int
	findAllResult          []*models.QpUser
	existsResult           bool
	createResult           *models.QpUser
	createUsername         string
	deleteUsername         string
	updatePasswordUsername string
}

func (s *stubRuntimeUsersData) Count() (int, error) { return s.countResult, nil }

func (s *stubRuntimeUsersData) FindAll() ([]*models.QpUser, error) { return s.findAllResult, nil }

func (s *stubRuntimeUsersData) Find(string) (*models.QpUser, error) {
	return nil, fmt.Errorf("user not found")
}

func (s *stubRuntimeUsersData) Exists(string) (bool, error) { return s.existsResult, nil }

func (s *stubRuntimeUsersData) Check(string, string) (*models.QpUser, error) {
	return nil, fmt.Errorf("user not found")
}

func (s *stubRuntimeUsersData) Create(username string, password string) (*models.QpUser, error) {
	s.createUsername = username
	if s.createResult != nil {
		return s.createResult, nil
	}
	return nil, nil
}

func (s *stubRuntimeUsersData) UpdatePassword(username string, password string) error {
	s.updatePasswordUsername = username
	return nil
}

func (s *stubRuntimeUsersData) Delete(username string) error {
	s.deleteUsername = username
	return nil
}

func TestRestoreOrphanedSessionsNilServiceReturnsError(t *testing.T) {
	previousService := models.WhatsappService
	defer func() {
		models.WhatsappService = previousService
	}()

	models.WhatsappService = nil

	_, err := RestoreOrphanedSessions()
	if err != ErrSessionServiceUnavailable {
		t.Fatalf("expected ErrSessionServiceUnavailable, got %v", err)
	}
}

func TestRestoreSessionManuallyNilServiceReturnsError(t *testing.T) {
	previousService := models.WhatsappService
	defer func() {
		models.WhatsappService = previousService
	}()

	models.WhatsappService = nil

	err := RestoreSessionManually("token", "jid")
	if err != ErrSessionServiceUnavailable {
		t.Fatalf("expected ErrSessionServiceUnavailable, got %v", err)
	}
}

func TestApplySessionConfigurationPatchUpdatesFields(t *testing.T) {
	groups := whatsapp.TrueBooleanType
	broadcasts := whatsapp.FalseBooleanType
	devel := true

	session := &models.QpWhatsappSession{QpServer: &models.QpServer{}}
	update, err := ApplySessionConfigurationPatch(session, &SessionConfigurationPatch{
		Groups:     &groups,
		Broadcasts: &broadcasts,
		Devel:      &devel,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if session.Groups != groups {
		t.Fatalf("expected groups to be updated")
	}

	if session.Broadcasts != broadcasts {
		t.Fatalf("expected broadcasts to be updated")
	}

	if session.Devel != devel {
		t.Fatalf("expected devel to be updated")
	}

	if update == "" {
		t.Fatalf("expected non-empty update summary")
	}
}
