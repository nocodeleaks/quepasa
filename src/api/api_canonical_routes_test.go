package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi/v5"
	environment "github.com/nocodeleaks/quepasa/environment"
	library "github.com/nocodeleaks/quepasa/library"
	models "github.com/nocodeleaks/quepasa/models"
)

func TestCanonicalAuthConfigIsPublicButSessionRemainsProtected(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	router := newCanonicalTestRouter()

	loginReq := httptest.NewRequest(http.MethodGet, "/api/auth/config", nil)
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)

	if loginRec.Code != http.StatusOK {
		t.Fatalf("expected /api/auth/config to return 200, got %d", loginRec.Code)
	}

	var loginConfig map[string]any
	if err := json.Unmarshal(loginRec.Body.Bytes(), &loginConfig); err != nil {
		t.Fatalf("decode /api/auth/config response: %v", err)
	}

	if _, ok := loginConfig["version"]; !ok {
		t.Fatalf("expected auth config to expose version, got %v", loginConfig)
	}

	sessionReq := httptest.NewRequest(http.MethodGet, "/api/auth/session", nil)
	sessionRec := httptest.NewRecorder()
	router.ServeHTTP(sessionRec, sessionReq)

	if sessionRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected /api/auth/session to require auth and return 401, got %d", sessionRec.Code)
	}
}

func TestCanonicalUsersLifecycleAndEnvironment(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	restore := setCanonicalAccountSetupEnv(t, "true")
	defer restore()
	CreateTestUser(t, "owner@example.com", "Password123!")
	CreateTestUser(t, "other@example.com", "Password123!")

	router := newCanonicalTestRouter()

	usersReq := newCanonicalAuthRequest(t, http.MethodGet, "/api/users", nil, "owner@example.com")
	usersReq.Header.Set(library.HeaderMasterKey, strings.TrimSpace(models.ENV.MasterKey()))
	usersRec := httptest.NewRecorder()
	router.ServeHTTP(usersRec, usersReq)

	if usersRec.Code != http.StatusOK {
		t.Fatalf("expected /api/users to return 200, got %d", usersRec.Code)
	}

	var usersPayload struct {
		Users []map[string]any `json:"users"`
	}
	if err := json.Unmarshal(usersRec.Body.Bytes(), &usersPayload); err != nil {
		t.Fatalf("decode /api/users response: %v", err)
	}

	if len(usersPayload.Users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(usersPayload.Users))
	}

	foundSelf := false
	for _, user := range usersPayload.Users {
		if user["username"] == "owner@example.com" && user["is_self"] == true {
			foundSelf = true
		}
	}
	if !foundSelf {
		t.Fatalf("expected authenticated user to be marked as self: %+v", usersPayload.Users)
	}

	envReq := newCanonicalAuthRequest(t, http.MethodGet, "/api/system/environment", nil, "owner@example.com")
	envRec := httptest.NewRecorder()
	router.ServeHTTP(envRec, envReq)

	if envRec.Code != http.StatusOK {
		t.Fatalf("expected /api/system/environment to return 200, got %d", envRec.Code)
	}

	var envPayload map[string]any
	if err := json.Unmarshal(envRec.Body.Bytes(), &envPayload); err != nil {
		t.Fatalf("decode /api/system/environment response: %v", err)
	}

	if _, hasSettings := envPayload["settings"]; !hasSettings {
		if _, hasPreview := envPayload["preview"]; !hasPreview {
			t.Fatalf("expected /api/system/environment to include settings or preview, got %v", envPayload)
		}
	}

	deleteSelfReq := newCanonicalAuthRequest(t, http.MethodDelete, "/api/users", []byte(`{"username":"owner@example.com"}`), "owner@example.com")
	deleteSelfReq.Header.Set(library.HeaderMasterKey, strings.TrimSpace(models.ENV.MasterKey()))
	deleteSelfRec := httptest.NewRecorder()
	router.ServeHTTP(deleteSelfRec, deleteSelfReq)

	if deleteSelfRec.Code != http.StatusBadRequest {
		t.Fatalf("expected deleting self to return 400, got %d", deleteSelfRec.Code)
	}

	deleteOtherReq := newCanonicalAuthRequest(t, http.MethodDelete, "/api/users", []byte(`{"username":"other@example.com"}`), "owner@example.com")
	deleteOtherReq.Header.Set(library.HeaderMasterKey, strings.TrimSpace(models.ENV.MasterKey()))
	deleteOtherRec := httptest.NewRecorder()
	router.ServeHTTP(deleteOtherRec, deleteOtherReq)

	if deleteOtherRec.Code != http.StatusOK {
		t.Fatalf("expected deleting another user to return 200, got %d", deleteOtherRec.Code)
	}

	body := []byte(`{"email":"created@example.com","password":"CorrectHorseBatteryStaple!2026"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(body))
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusOK {
		t.Fatalf("expected public /api/users create to return 200, got %d with body %s", createRec.Code, createRec.Body.String())
	}

	finalUsersReq := newCanonicalAuthRequest(t, http.MethodGet, "/api/users", nil, "owner@example.com")
	finalUsersReq.Header.Set(library.HeaderMasterKey, strings.TrimSpace(models.ENV.MasterKey()))
	finalUsersRec := httptest.NewRecorder()
	router.ServeHTTP(finalUsersRec, finalUsersReq)

	if finalUsersRec.Code != http.StatusOK {
		t.Fatalf("expected final /api/users to return 200, got %d", finalUsersRec.Code)
	}

	var finalUsersPayload struct {
		Users []map[string]any `json:"users"`
	}
	if err := json.Unmarshal(finalUsersRec.Body.Bytes(), &finalUsersPayload); err != nil {
		t.Fatalf("decode final /api/users response: %v", err)
	}

	if len(finalUsersPayload.Users) != 2 {
		t.Fatalf("expected 2 users after delete+create cycle, got %d", len(finalUsersPayload.Users))
	}
}

func TestCanonicalSessionCreateThenGetReturnsCreatedServer(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	CreateTestUser(t, "owner@example.com", "Password123!")

	router := newCanonicalTestRouter()

	createReq := newCanonicalAuthRequest(t, http.MethodPost, "/api/sessions", []byte(`{}`), "owner@example.com")
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected create session to return 201, got %d with body %s", createRec.Code, createRec.Body.String())
	}

	var createPayload struct {
		Server struct {
			Token string `json:"token"`
			User  string `json:"user"`
		} `json:"server"`
	}
	if err := json.Unmarshal(createRec.Body.Bytes(), &createPayload); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	if createPayload.Server.Token == "" {
		t.Fatalf("expected create response to contain server token, got %s", createRec.Body.String())
	}

	if createPayload.Server.User != "owner@example.com" {
		t.Fatalf("expected created server user to be owner@example.com, got %q", createPayload.Server.User)
	}

	infoReq := newCanonicalAuthRequest(t, http.MethodPost, "/api/sessions/get", []byte(`{"token":"`+createPayload.Server.Token+`"}`), "owner@example.com")
	infoRec := httptest.NewRecorder()
	router.ServeHTTP(infoRec, infoReq)

	if infoRec.Code != http.StatusOK {
		t.Fatalf("expected session get for created server to return 200, got %d with body %s", infoRec.Code, infoRec.Body.String())
	}

	var infoPayload struct {
		Server struct {
			Token string `json:"token"`
			User  string `json:"user"`
		} `json:"server"`
	}
	if err := json.Unmarshal(infoRec.Body.Bytes(), &infoPayload); err != nil {
		t.Fatalf("decode get response: %v", err)
	}

	if infoPayload.Server.Token != createPayload.Server.Token {
		t.Fatalf("expected get token %q, got %q", createPayload.Server.Token, infoPayload.Server.Token)
	}

	if infoPayload.Server.User != "owner@example.com" {
		t.Fatalf("expected get user to be owner@example.com, got %q", infoPayload.Server.User)
	}
}

func TestCanonicalSessionCreateAcceptsXQuePasaTokenHeader(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	CreateTestUser(t, "owner@example.com", "Password123!")

	restoreRelaxed := setRelaxedSessionsForTest(t, true)
	defer restoreRelaxed()

	router := newCanonicalTestRouter()

	createReq := newCanonicalAuthRequest(t, http.MethodPost, "/api/sessions", []byte(`{}`), "owner@example.com")
	createReq.Header.Set(library.HeaderToken, "custom-session-token-001")
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected create session to return 201 with X-QUEPASA-TOKEN, got %d with body %s", createRec.Code, createRec.Body.String())
	}

	var createPayload struct {
		Server struct {
			Token string `json:"token"`
		} `json:"server"`
	}
	if err := json.Unmarshal(createRec.Body.Bytes(), &createPayload); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	if createPayload.Server.Token != "custom-session-token-001" {
		t.Fatalf("expected created session token to match X-QUEPASA-TOKEN, got %q", createPayload.Server.Token)
	}
}

func TestCanonicalSessionCreateRequiresMasterKeyWhenNotRelaxed(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	CreateTestUser(t, "owner@example.com", "Password123!")

	restoreRelaxed := setRelaxedSessionsForTest(t, false)
	defer restoreRelaxed()

	router := newCanonicalTestRouter()

	createReq := newCanonicalAuthRequest(t, http.MethodPost, "/api/sessions", []byte(`{}`), "owner@example.com")
	createReq.Header.Set(library.HeaderToken, "custom-session-token-002")
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusForbidden {
		t.Fatalf("expected create session to return 403 when RELAXED_SESSIONS=false and no master key, got %d with body %s", createRec.Code, createRec.Body.String())
	}
}

func TestCanonicalSessionCreateMasterKeyOnlyStillRequiresAuthIdentity(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	cleanupMasterKey := SetupTestMasterKey(t, "master-key-123")
	defer cleanupMasterKey()

	router := newCanonicalTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/api/sessions", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(library.HeaderMasterKey, "master-key-123")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected create session to return 401 without JWT/X-QUEPASA-TOKEN, got %d with body %s", rec.Code, rec.Body.String())
	}
}

func TestCanonicalScopedSessionTokenAuthListsOnlyScopedSession(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	CreateTestUser(t, "owner@example.com", "Password123!")

	restoreRelaxed := setRelaxedSessionsForTest(t, true)
	defer restoreRelaxed()

	router := newCanonicalTestRouter()

	createA := newCanonicalAuthRequest(t, http.MethodPost, "/api/sessions", []byte(`{}`), "owner@example.com")
	createA.Header.Set(library.HeaderToken, "session-a")
	createARec := httptest.NewRecorder()
	router.ServeHTTP(createARec, createA)
	if createARec.Code != http.StatusCreated {
		t.Fatalf("expected session-a create 201, got %d with body %s", createARec.Code, createARec.Body.String())
	}

	createB := newCanonicalAuthRequest(t, http.MethodPost, "/api/sessions", []byte(`{}`), "owner@example.com")
	createB.Header.Set(library.HeaderToken, "session-b")
	createBRec := httptest.NewRecorder()
	router.ServeHTTP(createBRec, createB)
	if createBRec.Code != http.StatusCreated {
		t.Fatalf("expected session-b create 201, got %d with body %s", createBRec.Code, createBRec.Body.String())
	}

	req := newCanonicalScopedTokenAuthRequest(http.MethodGet, "/api/sessions", nil, "session-a")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected scoped token auth list to return 200, got %d with body %s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Servers []struct {
			Token string `json:"token"`
		} `json:"servers"`
		Total int `json:"total"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode sessions list response: %v", err)
	}

	if payload.Total != 1 || len(payload.Servers) != 1 {
		t.Fatalf("expected exactly one scoped session, got total=%d len=%d", payload.Total, len(payload.Servers))
	}

	if payload.Servers[0].Token != "session-a" {
		t.Fatalf("expected scoped session token session-a, got %q", payload.Servers[0].Token)
	}
}

func TestCanonicalScopedSessionTokenAuthForcesAuthenticatedSessionToken(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	CreateTestUser(t, "owner@example.com", "Password123!")

	restoreRelaxed := setRelaxedSessionsForTest(t, true)
	defer restoreRelaxed()

	router := newCanonicalTestRouter()

	createA := newCanonicalAuthRequest(t, http.MethodPost, "/api/sessions", []byte(`{}`), "owner@example.com")
	createA.Header.Set(library.HeaderToken, "session-a")
	createARec := httptest.NewRecorder()
	router.ServeHTTP(createARec, createA)
	if createARec.Code != http.StatusCreated {
		t.Fatalf("expected session-a create 201, got %d with body %s", createARec.Code, createARec.Body.String())
	}

	createB := newCanonicalAuthRequest(t, http.MethodPost, "/api/sessions", []byte(`{}`), "owner@example.com")
	createB.Header.Set(library.HeaderToken, "session-b")
	createBRec := httptest.NewRecorder()
	router.ServeHTTP(createBRec, createB)
	if createBRec.Code != http.StatusCreated {
		t.Fatalf("expected session-b create 201, got %d with body %s", createBRec.Code, createBRec.Body.String())
	}

	req := newCanonicalScopedTokenAuthRequest(http.MethodPost, "/api/sessions/get", []byte(`{"token":"session-b"}`), "session-a")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected scoped token auth to resolve to authenticated session, got %d with body %s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Server struct {
			Token string `json:"token"`
		} `json:"server"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode session get response: %v", err)
	}

	if payload.Server.Token != "session-a" {
		t.Fatalf("expected scoped auth to force token session-a, got %q", payload.Server.Token)
	}
}

// TestCanonicalHealthAliasRoutes validates that /api/health is an alias for /api/system/health
func TestCanonicalHealthAliasRoutes(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	router := newCanonicalTestRouter()

	// Test GET /api/system/health
	sysReq := httptest.NewRequest(http.MethodGet, "/api/system/health", nil)
	sysRec := httptest.NewRecorder()
	router.ServeHTTP(sysRec, sysReq)

	if sysRec.Code != http.StatusOK {
		t.Fatalf("expected /api/system/health to return 200, got %d", sysRec.Code)
	}

	var sysPayload map[string]interface{}
	if err := json.Unmarshal(sysRec.Body.Bytes(), &sysPayload); err != nil {
		t.Fatalf("decode /api/system/health response: %v", err)
	}

	// Test GET /api/health (alias)
	aliasReq := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	aliasRec := httptest.NewRecorder()
	router.ServeHTTP(aliasRec, aliasReq)

	if aliasRec.Code != http.StatusOK {
		t.Fatalf("expected /api/health to return 200, got %d", aliasRec.Code)
	}

	var aliasPayload map[string]interface{}
	if err := json.Unmarshal(aliasRec.Body.Bytes(), &aliasPayload); err != nil {
		t.Fatalf("decode /api/health response: %v", err)
	}

	if sysPayload["status"] != aliasPayload["status"] || sysPayload["success"] != aliasPayload["success"] {
		t.Fatalf("expected /api/health and /api/system/health to expose the same health state\n/api/system/health: %+v\n/api/health: %+v", sysPayload, aliasPayload)
	}

	// Test HEAD /api/system/health
	sysHeadReq := httptest.NewRequest(http.MethodHead, "/api/system/health", nil)
	sysHeadRec := httptest.NewRecorder()
	router.ServeHTTP(sysHeadRec, sysHeadReq)

	if sysHeadRec.Code != http.StatusOK {
		t.Fatalf("expected HEAD /api/system/health to return 200, got %d", sysHeadRec.Code)
	}

	// Test HEAD /api/health (alias)
	aliasHeadReq := httptest.NewRequest(http.MethodHead, "/api/health", nil)
	aliasHeadRec := httptest.NewRecorder()
	router.ServeHTTP(aliasHeadRec, aliasHeadReq)

	if aliasHeadRec.Code != http.StatusOK {
		t.Fatalf("expected HEAD /api/health to return 200, got %d", aliasHeadRec.Code)
	}

	t.Log("Both /api/health and /api/system/health routes working correctly as aliases")
}

func newCanonicalTestRouter() chi.Router {
	router := chi.NewRouter()
	router.Route("/api", func(r chi.Router) {
		r.Group(func(router chi.Router) {
			RegisterAPIV5Controllers(router, true)
		})
	})
	return router
}

func newCanonicalAuthRequest(t *testing.T, method string, target string, body []byte, username string) *http.Request {
	t.Helper()

	req := httptest.NewRequest(method, target, bytes.NewReader(body))
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}

	_, tokenString, err := GetAuthenticatedTokenAuth().Encode(jwt.MapClaims{"user_id": username})
	if err != nil {
		t.Fatalf("encode authenticated canonical token: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+tokenString)
	return req
}

func newCanonicalScopedTokenAuthRequest(method string, target string, body []byte, scopedToken string) *http.Request {
	req := httptest.NewRequest(method, target, bytes.NewReader(body))
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set(library.HeaderToken, scopedToken)
	return req
}

func setRelaxedSessionsForTest(t *testing.T, value bool) func() {
	t.Helper()

	oldValue := environment.Settings.API.RelaxedSessions
	environment.Settings.API.RelaxedSessions = value

	return func() {
		environment.Settings.API.RelaxedSessions = oldValue
	}
}

func setCanonicalAccountSetupEnv(t *testing.T, value string) func() {
	t.Helper()

	oldValue, hadValue := os.LookupEnv(models.ENV_ACCOUNTSETUP)
	if err := os.Setenv(models.ENV_ACCOUNTSETUP, value); err != nil {
		t.Fatalf("set %s: %v", models.ENV_ACCOUNTSETUP, err)
	}

	return func() {
		if !hadValue {
			_ = os.Unsetenv(models.ENV_ACCOUNTSETUP)
			return
		}
		_ = os.Setenv(models.ENV_ACCOUNTSETUP, oldValue)
	}
}
