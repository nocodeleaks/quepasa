package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nocodeleaks/quepasa/models"
)

func TestSpamAdminStatusReportsMissingMasterKey(t *testing.T) {
	restore := SetupTestMasterKey(t, "")
	defer restore()

	req := httptest.NewRequest(http.MethodGet, "/api/spam/status", nil)
	rec := httptest.NewRecorder()

	SpamAdminStatusController(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var response struct {
		Configured bool `json:"configured"`
		Unlocked   bool `json:"unlocked"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Configured {
		t.Fatal("expected configured=false without master key")
	}
	if response.Unlocked {
		t.Fatal("expected unlocked=false without master key")
	}
}

func TestSpamAdminSectionsRejectsMissingMasterHeader(t *testing.T) {
	restore := SetupTestMasterKey(t, "spam-master-key")
	defer restore()

	req := httptest.NewRequest(http.MethodGet, "/api/spam/sections", nil)
	rec := httptest.NewRecorder()

	SpamAdminSectionsListController(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSpamAdminSectionsSearchReturnsServerRows(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	restore := SetupTestMasterKey(t, "spam-master-key")
	defer restore()

	if _, err := testDB.Exec(`
		CREATE TABLE IF NOT EXISTS spam_sections (
			token TEXT PRIMARY KEY NOT NULL REFERENCES servers(token) ON DELETE CASCADE,
			position INTEGER NOT NULL DEFAULT 0,
			enabled BOOLEAN NOT NULL DEFAULT 1,
			label TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`); err != nil {
		t.Fatalf("ensure spam_sections table: %v", err)
	}
	models.WhatsappService.DB.SpamSections = models.NewQpDataSpamSectionsSql(testDB)

	if _, err := testDB.Exec(
		`INSERT INTO users (username, password) VALUES (?, ?)`,
		"owner@example.com",
		"hash",
	); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	server := &models.QpServer{Token: "spam-token-1", Verified: true}
	server.SetWId("5511999999999@s.whatsapp.net")
	server.SetUser("owner@example.com")
	server.SetContextId("context-1")
	if err := models.WhatsappService.DB.Servers.Add(server); err != nil {
		t.Fatalf("add server: %v", err)
	}

	body, _ := json.Marshal(spamSearchRequest{Search: "5511999999999", Limit: 10})
	req := httptest.NewRequest(http.MethodPost, "/api/spam/sections/search", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-QUEPASA-MASTERKEY", "spam-master-key")
	rec := httptest.NewRecorder()

	SpamAdminSectionsSearchController(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var response struct {
		Items []spamSectionView `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(response.Items) != 1 {
		t.Fatalf("expected one search result, got %d", len(response.Items))
	}
	if response.Items[0].Token != "spam-token-1" {
		t.Fatalf("expected token spam-token-1, got %q", response.Items[0].Token)
	}
	if response.Items[0].InSpam {
		t.Fatal("expected search result to start outside spam queue")
	}
}
