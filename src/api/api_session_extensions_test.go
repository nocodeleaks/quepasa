package api

import (
	"net/http/httptest"
	"testing"

	models "github.com/nocodeleaks/quepasa/models"
)

func TestGetSessionReturnsSessionByToken(t *testing.T) {
	prevService := models.WhatsappService
	defer func() { models.WhatsappService = prevService }()

	session := &models.QpWhatsappSession{QpServer: &models.QpServer{Token: "session-token"}}
	models.WhatsappService = &models.QPWhatsappService{
		Servers: map[string]*models.QpWhatsappServer{
			session.Token: session,
		},
	}

	req := httptest.NewRequest("GET", "/api/v5/sessions/get?token=session-token", nil)
	got, err := GetSession(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got != session {
		t.Fatal("expected GetSession to return the cached session instance")
	}
}

func TestGetSessionRespondOnErrorReturnsNotFoundWhenMissing(t *testing.T) {
	prevService := models.WhatsappService
	defer func() { models.WhatsappService = prevService }()

	models.WhatsappService = &models.QPWhatsappService{Servers: map[string]*models.QpWhatsappServer{}}

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v5/sessions/get?token=missing", nil)
	got, err := GetSessionRespondOnError(recorder, req)
	if err == nil {
		t.Fatal("expected lookup error for missing session")
	}
	if got != nil {
		t.Fatal("expected nil session for missing lookup")
	}
	if recorder.Code != 404 {
		t.Fatalf("expected 404 status, got %d", recorder.Code)
	}
}
