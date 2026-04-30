package api

import (
	"testing"

	models "github.com/nocodeleaks/quepasa/models"
)

func TestFindConversationLabelStoreReturnsConfiguredStore(t *testing.T) {
	previousService := models.WhatsappService
	defer func() {
		models.WhatsappService = previousService
	}()

	store := &stubConversationLabelStore{}
	models.WhatsappService = &models.QPWhatsappService{
		DB: &models.QpDatabase{
			ConversationLabels: store,
		},
	}

	got, err := findConversationLabelStore()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got != store {
		t.Fatalf("expected configured conversation label store")
	}
}

type stubConversationLabelStore struct{}

func (s *stubConversationLabelStore) FindAllForUser(string, *bool) ([]*models.QpConversationLabel, error) {
	return nil, nil
}

func (s *stubConversationLabelStore) FindByIDForUser(int64, string) (*models.QpConversationLabel, error) {
	return nil, nil
}

func (s *stubConversationLabelStore) Create(label *models.QpConversationLabel) (*models.QpConversationLabel, error) {
	return label, nil
}

func (s *stubConversationLabelStore) Update(*models.QpConversationLabel) error { return nil }

func (s *stubConversationLabelStore) Delete(int64, string) error { return nil }

func (s *stubConversationLabelStore) Assign(string, string, int64, string) (uint, error) {
	return 0, nil
}

func (s *stubConversationLabelStore) Remove(string, string, int64, string) (uint, error) {
	return 0, nil
}

func (s *stubConversationLabelStore) FindConversationLabels(string, string, string) ([]*models.QpConversationLabel, error) {
	return nil, nil
}

func (s *stubConversationLabelStore) FindConversationLabelsMap(string, string, []string) (map[string][]*models.QpConversationLabel, error) {
	return nil, nil
}
