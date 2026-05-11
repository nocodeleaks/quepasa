package models

import "testing"

type stubSessionServersData struct{}

func (stubSessionServersData) FindAll() []*QpServer                          { return nil }
func (stubSessionServersData) FindByToken(string) (*QpServer, error)         { return nil, nil }
func (stubSessionServersData) FindForUser(string, string) (*QpServer, error) { return nil, nil }
func (stubSessionServersData) Exists(string) (bool, error)                   { return false, nil }
func (stubSessionServersData) Add(*QpServer) error                           { return nil }
func (stubSessionServersData) Update(*QpServer) error                        { return nil }
func (stubSessionServersData) Delete(string) error                           { return nil }

type stubSessionDispatchingData struct{}

func (stubSessionDispatchingData) Find(string, string) (*QpServerDispatching, error) { return nil, nil }
func (stubSessionDispatchingData) FindAll(string) ([]*QpServerDispatching, error)    { return nil, nil }
func (stubSessionDispatchingData) All() ([]*QpServerDispatching, error)              { return nil, nil }
func (stubSessionDispatchingData) Add(*QpServerDispatching) error                    { return nil }
func (stubSessionDispatchingData) Update(*QpServerDispatching) error                 { return nil }
func (stubSessionDispatchingData) UpdateContext(*QpServerDispatching, string) error  { return nil }
func (stubSessionDispatchingData) Remove(string, string) error                       { return nil }
func (stubSessionDispatchingData) Clear(string) error                                { return nil }
func (stubSessionDispatchingData) DispatchingAddOrUpdate(string, *QpDispatching) (uint, error) {
	return 0, nil
}
func (stubSessionDispatchingData) DispatchingUpdateHealth(string, *QpDispatching) error { return nil }
func (stubSessionDispatchingData) DispatchingRemove(string, string) (uint, error)       { return 0, nil }
func (stubSessionDispatchingData) DispatchingClear(string) error                        { return nil }
func (stubSessionDispatchingData) GetWebhooks() []*QpWebhook                            { return nil }
func (stubSessionDispatchingData) GetRabbitMQConfigs() []*QpRabbitMQConfig              { return nil }

func TestQpWhatsappSessionAliasPreservesMethods(t *testing.T) {
	serverData := &QpServer{Token: "session-token"}
	serverData.SetWId("5511999999999@s.whatsapp.net")
	session := &QpWhatsappSession{QpServer: serverData}

	if got := session.GetToken(); got != "session-token" {
		t.Fatalf("expected session token %q, got %q", "session-token", got)
	}
	if got := session.ID(); got != "5511999999999@s.whatsapp.net" {
		t.Fatalf("expected session ID %q, got %q", "5511999999999@s.whatsapp.net", got)
	}
}

func TestSessionServiceWrappersDelegateToServerImplementations(t *testing.T) {
	service := &QPWhatsappService{
		DB: &QpDatabase{
			Servers:     stubSessionServersData{},
			Dispatching: stubSessionDispatchingData{},
		},
	}
	info := &QpServer{Token: "wrapper-token"}

	session, err := service.NewQpWhatsappSession(info)
	if err != nil {
		t.Fatalf("expected no error creating session, got %v", err)
	}
	if session == nil {
		t.Fatal("expected session instance")
	}
	if session.Token != info.Token {
		t.Fatalf("expected token %q, got %q", info.Token, session.Token)
	}
	if session.syncConnection == nil {
		t.Fatal("expected syncConnection to be initialized")
	}
	if session.syncMessages == nil {
		t.Fatal("expected syncMessages to be initialized")
	}
}

func TestSessionLookupWrappersReuseServerLookups(t *testing.T) {
	prevService := WhatsappService
	defer func() { WhatsappService = prevService }()

	session := &QpWhatsappSession{QpServer: &QpServer{Token: "lookup-token"}}
	WhatsappService = &QPWhatsappService{
		Servers: map[string]*QpWhatsappServer{
			session.Token: session,
		},
	}

	got, err := GetSessionFromToken("lookup-token")
	if err != nil {
		t.Fatalf("expected no error looking up session, got %v", err)
	}
	if got != session {
		t.Fatal("expected session lookup to return cached session instance")
	}

	userSessions := WhatsappService.GetSessionsForUser("")
	if len(userSessions) != 1 {
		t.Fatalf("expected one session in wrapper map, got %d", len(userSessions))
	}
	if userSessions[session.Token] != session {
		t.Fatal("expected session map to expose the cached session instance")
	}
}
