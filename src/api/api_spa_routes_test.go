package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestSPALoginConfigIsPublicButSessionRemainsProtected(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	router := chi.NewRouter()
	router.Route("/spa", func(r chi.Router) {
		r.Group(RegisterSPAPublicControllers)
		r.Group(RegisterSPAControllers)
	})

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
