package api

import (
	"fmt"
	"testing"

	models "github.com/nocodeleaks/quepasa/models"
)

func TestFindPersistedUserDelegatesToUserStore(t *testing.T) {
	previousService := models.WhatsappService
	defer func() {
		models.WhatsappService = previousService
	}()

	models.WhatsappService = &models.QPWhatsappService{
		DB: &models.QpDatabase{
			Users: &stubAPIUsersData{
				findResult: &models.QpUser{Username: "owner@example.com"},
			},
		},
	}

	user, err := findPersistedUser(" owner@example.com ")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user == nil || user.Username != "owner@example.com" {
		t.Fatalf("expected persisted user lookup result")
	}
}

func TestAuthenticatePersistedUserDelegatesToUserStore(t *testing.T) {
	previousService := models.WhatsappService
	defer func() {
		models.WhatsappService = previousService
	}()

	models.WhatsappService = &models.QPWhatsappService{
		DB: &models.QpDatabase{
			Users: &stubAPIUsersData{
				checkResult: &models.QpUser{Username: "owner@example.com"},
			},
		},
	}

	user, err := authenticatePersistedUser(" owner@example.com ", "secret")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user == nil || user.Username != "owner@example.com" {
		t.Fatalf("expected authenticated user lookup result")
	}
}

func TestUpdatePersistedUserPasswordDelegatesToUserStore(t *testing.T) {
	previousService := models.WhatsappService
	defer func() {
		models.WhatsappService = previousService
	}()

	users := &stubAPIUsersData{existsResult: true}
	models.WhatsappService = &models.QPWhatsappService{
		DB: &models.QpDatabase{Users: users},
	}

	err := updatePersistedUserPassword(" owner@example.com ", "secret")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if users.updatePasswordUsername != "owner@example.com" {
		t.Fatalf("expected trimmed username in password update")
	}
}

type stubAPIUsersData struct {
	findResult             *models.QpUser
	checkResult            *models.QpUser
	existsResult           bool
	updatePasswordUsername string
}

func (s *stubAPIUsersData) Count() (int, error) { return 0, nil }

func (s *stubAPIUsersData) FindAll() ([]*models.QpUser, error) { return nil, nil }

func (s *stubAPIUsersData) Find(username string) (*models.QpUser, error) {
	if s.findResult != nil && s.findResult.Username == username {
		return s.findResult, nil
	}
	return nil, fmt.Errorf("user not found")
}

func (s *stubAPIUsersData) Exists(string) (bool, error) { return s.existsResult, nil }

func (s *stubAPIUsersData) Check(string, string) (*models.QpUser, error) {
	if s.checkResult != nil {
		return s.checkResult, nil
	}
	return nil, fmt.Errorf("user not found")
}

func (s *stubAPIUsersData) Create(string, string) (*models.QpUser, error) { return nil, nil }

func (s *stubAPIUsersData) UpdatePassword(username string, password string) error {
	s.updatePasswordUsername = username
	return nil
}

func (s *stubAPIUsersData) Delete(string) error { return nil }
