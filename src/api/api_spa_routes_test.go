package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi/v5"
	models "github.com/nocodeleaks/quepasa/models"
)

func TestSPALoginConfigIsPublicButSessionRemainsProtected(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	router := newSPATestRouter()

	loginReq := httptest.NewRequest(http.MethodGet, "/spa/login/config", nil)
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)

	if loginRec.Code != http.StatusOK {
		t.Fatalf("expected /spa/login/config to return 200, got %d", loginRec.Code)
	}

	var loginConfig map[string]any
	if err := json.Unmarshal(loginRec.Body.Bytes(), &loginConfig); err != nil {
		t.Fatalf("decode /spa/login/config response: %v", err)
	}

	if _, ok := loginConfig["version"]; !ok {
		t.Fatalf("expected login config to expose version, got %v", loginConfig)
	}

	sessionReq := httptest.NewRequest(http.MethodGet, "/spa/session", nil)
	sessionRec := httptest.NewRecorder()
	router.ServeHTTP(sessionRec, sessionReq)

	if sessionRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected /spa/session to require auth and return 401, got %d", sessionRec.Code)
	}
}

func TestSPAUsersLifecycleAndEnvironment(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	restore := setSPAAccountSetupEnv(t, "true")
	defer restore()

	CreateTestUser(t, "owner@example.com", "Password123!")
	CreateTestUser(t, "other@example.com", "Password123!")

	router := newSPATestRouter()

	usersReq := newSPAAuthRequest(t, http.MethodGet, "/spa/users", nil, "owner@example.com")
	usersRec := httptest.NewRecorder()
	router.ServeHTTP(usersRec, usersReq)

	if usersRec.Code != http.StatusOK {
		t.Fatalf("expected /spa/users to return 200, got %d", usersRec.Code)
	}

	var usersPayload struct {
		Users []map[string]any `json:"users"`
	}
	if err := json.Unmarshal(usersRec.Body.Bytes(), &usersPayload); err != nil {
		t.Fatalf("decode /spa/users response: %v", err)
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

	envReq := newSPAAuthRequest(t, http.MethodGet, "/spa/environment", nil, "owner@example.com")
	envRec := httptest.NewRecorder()
	router.ServeHTTP(envRec, envReq)

	if envRec.Code != http.StatusOK {
		t.Fatalf("expected /spa/environment to return 200, got %d", envRec.Code)
	}

	var envPayload map[string]any
	if err := json.Unmarshal(envRec.Body.Bytes(), &envPayload); err != nil {
		t.Fatalf("decode /spa/environment response: %v", err)
	}

	if _, ok := envPayload["settings"]; !ok {
		t.Fatalf("expected /spa/environment to include settings, got %v", envPayload)
	}

	deleteSelfReq := newSPAAuthRequest(t, http.MethodDelete, "/spa/user/owner@example.com", nil, "owner@example.com")
	deleteSelfRec := httptest.NewRecorder()
	router.ServeHTTP(deleteSelfRec, deleteSelfReq)

	if deleteSelfRec.Code != http.StatusBadRequest {
		t.Fatalf("expected deleting self to return 400, got %d", deleteSelfRec.Code)
	}

	deleteOtherReq := newSPAAuthRequest(t, http.MethodDelete, "/spa/user/other@example.com", nil, "owner@example.com")
	deleteOtherRec := httptest.NewRecorder()
	router.ServeHTTP(deleteOtherRec, deleteOtherReq)

	if deleteOtherRec.Code != http.StatusOK {
		t.Fatalf("expected deleting another user to return 200, got %d", deleteOtherRec.Code)
	}

	body := []byte(`{"email":"created@example.com","password":"CorrectHorseBatteryStaple!2026"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/spa/users", bytes.NewReader(body))
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusOK {
		t.Fatalf("expected public /spa/users create to return 200, got %d with body %s", createRec.Code, createRec.Body.String())
	}

	finalUsersReq := newSPAAuthRequest(t, http.MethodGet, "/spa/users", nil, "owner@example.com")
	finalUsersRec := httptest.NewRecorder()
	router.ServeHTTP(finalUsersRec, finalUsersReq)

	if finalUsersRec.Code != http.StatusOK {
		t.Fatalf("expected final /spa/users to return 200, got %d", finalUsersRec.Code)
	}

	var finalUsersPayload struct {
		Users []map[string]any `json:"users"`
	}
	if err := json.Unmarshal(finalUsersRec.Body.Bytes(), &finalUsersPayload); err != nil {
		t.Fatalf("decode final /spa/users response: %v", err)
	}

	if len(finalUsersPayload.Users) != 2 {
		t.Fatalf("expected 2 users after delete+create cycle, got %d", len(finalUsersPayload.Users))
	}
}

func TestSPAServerCreateThenInfoReturnsCreatedServer(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	CreateTestUser(t, "owner@example.com", "Password123!")

	router := newSPATestRouter()

	createReq := newSPAAuthRequest(t, http.MethodPost, "/spa/server/create", []byte(`{}`), "owner@example.com")
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected create server to return 201, got %d with body %s", createRec.Code, createRec.Body.String())
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

	infoReq := newSPAAuthRequest(t, http.MethodGet, "/spa/server/"+createPayload.Server.Token+"/info", nil, "owner@example.com")
	infoRec := httptest.NewRecorder()
	router.ServeHTTP(infoRec, infoReq)

	if infoRec.Code != http.StatusOK {
		t.Fatalf("expected info for created server to return 200, got %d with body %s", infoRec.Code, infoRec.Body.String())
	}

	var infoPayload struct {
		Server struct {
			Token string `json:"token"`
			User  string `json:"user"`
		} `json:"server"`
	}
	if err := json.Unmarshal(infoRec.Body.Bytes(), &infoPayload); err != nil {
		t.Fatalf("decode info response: %v", err)
	}

	if infoPayload.Server.Token != createPayload.Server.Token {
		t.Fatalf("expected info token %q, got %q", createPayload.Server.Token, infoPayload.Server.Token)
	}

	if infoPayload.Server.User != "owner@example.com" {
		t.Fatalf("expected info user to be owner@example.com, got %q", infoPayload.Server.User)
	}
}

func newSPATestRouter() chi.Router {
	router := chi.NewRouter()
	router.Route("/spa", func(r chi.Router) {
		r.Group(RegisterSPAPublicControllers)
		r.Group(RegisterSPAControllers)
	})
	return router
}

func newSPAAuthRequest(t *testing.T, method string, target string, body []byte, username string) *http.Request {
	t.Helper()

	req := httptest.NewRequest(method, target, bytes.NewReader(body))
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}

	_, tokenString, err := GetSPATokenAuth().Encode(jwt.MapClaims{"user_id": username})
	if err != nil {
		t.Fatalf("encode spa auth token: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+tokenString)
	return req
}

func setSPAAccountSetupEnv(t *testing.T, value string) func() {
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
