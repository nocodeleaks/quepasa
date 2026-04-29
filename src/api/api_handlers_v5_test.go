package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestWithCanonicalParamsInjectsBodyAndHeaderValues(t *testing.T) {
	router := chi.NewRouter()
	router.With(withCanonicalParams(canonicalTokenParam, canonicalMessageIDParam)).Post("/messages/get", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(chi.URLParam(r, "token") + "|" + chi.URLParam(r, "messageid")))
	})

	req := httptest.NewRequest(http.MethodPost, "/messages/get", strings.NewReader(`{"messageId":"msg-123"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-QUEPASA-TOKEN", "token-abc")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}

	body, _ := io.ReadAll(res.Body)
	if got := string(body); got != "token-abc|msg-123" {
		t.Fatalf("expected adapted params token-abc|msg-123, got %s", got)
	}
}

func TestBuildMasterKeyStatusResponseNeverReturnsSecret(t *testing.T) {
	payload := buildMasterKeyStatusResponse("super-secret-master-key")

	if payload["configured"] != true {
		t.Fatalf("expected configured=true, got %v", payload["configured"])
	}

	if _, exists := payload["status"]; exists {
		t.Fatal("status text should not be present in the payload")
	}

	if _, exists := payload["masterKey"]; exists {
		t.Fatal("master key secret should never be present in the payload")
	}
}
