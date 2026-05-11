package models

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeServerStore struct {
	deletedToken string
	deleteErr    error
}

func (f *fakeServerStore) FindAll() []*QpServer {
	return nil
}

func (f *fakeServerStore) FindByToken(string) (*QpServer, error) {
	return nil, nil
}

func (f *fakeServerStore) FindForUser(string, string) (*QpServer, error) {
	return nil, nil
}

func (f *fakeServerStore) Exists(string) (bool, error) {
	return false, nil
}

func (f *fakeServerStore) Add(*QpServer) error {
	return nil
}

func (f *fakeServerStore) Update(*QpServer) error {
	return nil
}

func (f *fakeServerStore) Delete(token string) error {
	f.deletedToken = token
	return f.deleteErr
}

func TestDeleteDispatchesDeletedWebhookWithStoppingState(t *testing.T) {
	type receivedPayload struct {
		Info map[string]interface{} `json:"info"`
	}

	var payload receivedPayload
	var receivedWid string
	var decodeErr error

	webhookServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedWid = r.Header.Get("X-QUEPASA-WID")
		decodeErr = json.NewDecoder(r.Body).Decode(&payload)

		w.WriteHeader(http.StatusOK)
	}))
	defer webhookServer.Close()

	store := &fakeServerStore{}
	server := &QpWhatsappServer{
		QpServer: &QpServer{
			Token:    "delete-token",
			Verified: true,
		},
		QpDataDispatching: QpDataDispatching{
			Dispatching: []*QpDispatching{{
				ConnectionString: webhookServer.URL,
				Type:             DispatchingTypeWebhook,
			}},
		},
		db: store,
	}
	server.QpServer.SetWId("5511999999999@s.whatsapp.net")

	if err := server.Delete("api"); err != nil {
		t.Fatalf("expected delete to succeed, got: %v", err)
	}

	if store.deletedToken != "delete-token" {
		t.Fatalf("expected token delete-token to be deleted, got %s", store.deletedToken)
	}

	if receivedWid != "5511999999999@s.whatsapp.net" {
		t.Fatalf("expected X-QUEPASA-WID header to be synchronized, got %s", receivedWid)
	}

	if decodeErr != nil {
		t.Fatalf("failed to decode webhook payload: %v", decodeErr)
	}

	state, ok := payload.Info["state"].(string)
	if !ok || state != "Stopping" {
		t.Fatalf("expected deleted event state to be Stopping, got %#v", payload.Info["state"])
	}

	previousState, ok := payload.Info["previous_state"].(string)
	if !ok || previousState != "UnPrepared" {
		t.Fatalf("expected previous_state to be UnPrepared, got %#v", payload.Info["previous_state"])
	}

	event, ok := payload.Info["event"].(string)
	if !ok || event != "deleted" {
		t.Fatalf("expected event deleted, got %#v", payload.Info["event"])
	}

	if len(server.QpDataDispatching.Dispatching) != 0 {
		t.Fatalf("expected dispatching memory cache to be cleared, got %d item(s)", len(server.QpDataDispatching.Dispatching))
	}
}

func TestDeleteRestoresStateAndDispatchingsWhenDatabaseDeleteFails(t *testing.T) {
	store := &fakeServerStore{
		deleteErr: errors.New("delete failed"),
	}

	dispatching := &QpDispatching{
		ConnectionString: "https://example.com/webhook",
		Type:             DispatchingTypeWebhook,
	}

	server := &QpWhatsappServer{
		QpServer: &QpServer{
			Token:    "delete-token",
			Verified: true,
		},
		QpDataDispatching: QpDataDispatching{
			Dispatching: []*QpDispatching{dispatching},
		},
		db: store,
	}

	err := server.Delete("api")
	if err == nil {
		t.Fatal("expected delete to fail")
	}

	if server.Intent.IsDeleteRequested() {
		t.Fatal("expected Intent to be restored to no delete after failure")
	}

	if server.Intent.IsStopRequested() {
		t.Fatal("expected Intent to be restored to no stop after failure")
	}

	if len(server.QpDataDispatching.Dispatching) != 1 {
		t.Fatalf("expected dispatchings to be restored after failure, got %d item(s)", len(server.QpDataDispatching.Dispatching))
	}

	if server.QpDataDispatching.Dispatching[0] != dispatching {
		t.Fatal("expected original dispatching to be restored after failure")
	}
}
