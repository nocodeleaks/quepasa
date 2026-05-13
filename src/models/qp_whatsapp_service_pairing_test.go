package models

import (
	"context"
	"sync"
	"testing"
	"time"

	library "github.com/nocodeleaks/quepasa/library"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
)

type pairingTestServersData struct{}

func (pairingTestServersData) FindAll() []*QpServer                          { return nil }
func (pairingTestServersData) FindByToken(string) (*QpServer, error)         { return nil, nil }
func (pairingTestServersData) FindForUser(string, string) (*QpServer, error) { return nil, nil }
func (pairingTestServersData) Exists(string) (bool, error)                   { return false, nil }
func (pairingTestServersData) Add(*QpServer) error                           { return nil }
func (pairingTestServersData) Update(*QpServer) error                        { return nil }
func (pairingTestServersData) Delete(string) error                           { return nil }

type pairingTestDispatchingData struct{}

func (pairingTestDispatchingData) Find(string, string) (*QpServerDispatching, error) { return nil, nil }
func (pairingTestDispatchingData) FindAll(string) ([]*QpServerDispatching, error)    { return nil, nil }
func (pairingTestDispatchingData) All() ([]*QpServerDispatching, error)              { return nil, nil }
func (pairingTestDispatchingData) Add(*QpServerDispatching) error                    { return nil }
func (pairingTestDispatchingData) Update(*QpServerDispatching) error                 { return nil }
func (pairingTestDispatchingData) UpdateContext(*QpServerDispatching, string) error  { return nil }
func (pairingTestDispatchingData) Remove(string, string) error                       { return nil }
func (pairingTestDispatchingData) Clear(string) error                                { return nil }
func (pairingTestDispatchingData) DispatchingAddOrUpdate(string, *QpDispatching) (uint, error) {
	return 0, nil
}
func (pairingTestDispatchingData) DispatchingUpdateHealth(string, *QpDispatching) error { return nil }
func (pairingTestDispatchingData) DispatchingRemove(string, string) (uint, error)       { return 0, nil }
func (pairingTestDispatchingData) DispatchingClear(string) error                        { return nil }
func (pairingTestDispatchingData) GetWebhooks() []*QpWebhook                            { return nil }
func (pairingTestDispatchingData) GetRabbitMQConfigs() []*QpRabbitMQConfig              { return nil }

type pairingTestConnection struct {
	logEntry               *log.Entry
	updatedHandler         whatsapp.IWhatsappHandlers
	updatedPairedCallback  func(string)
	disposed               bool
}

func newPairingTestConnection(t *testing.T) *pairingTestConnection {
	t.Helper()
	return &pairingTestConnection{
		logEntry: log.New().WithField("test", t.Name()),
	}
}

func (c *pairingTestConnection) GetChatTitle(string) string { return "" }
func (c *pairingTestConnection) Connect() error             { return nil }
func (c *pairingTestConnection) Disconnect() error          { return nil }
func (c *pairingTestConnection) GetWhatsAppQRChannel(_ context.Context, _ chan<- string) error {
	return nil
}
func (c *pairingTestConnection) GetWhatsAppQRCode() string { return "" }
func (c *pairingTestConnection) UpdateHandler(h whatsapp.IWhatsappHandlers) {
	c.updatedHandler = h
}
func (c *pairingTestConnection) UpdatePairedCallBack(cb func(string)) { c.updatedPairedCallback = cb }
func (c *pairingTestConnection) DownloadData(whatsapp.IWhatsappMessage) ([]byte, error) {
	return nil, nil
}
func (c *pairingTestConnection) Download(whatsapp.IWhatsappMessage, bool) (*whatsapp.WhatsappAttachment, error) {
	return nil, nil
}
func (c *pairingTestConnection) Revoke(whatsapp.IWhatsappMessage) error { return nil }
func (c *pairingTestConnection) Edit(whatsapp.IWhatsappMessage, string) error { return nil }
func (c *pairingTestConnection) MarkRead(whatsapp.IWhatsappMessage) error     { return nil }
func (c *pairingTestConnection) Send(*whatsapp.WhatsappMessage) (whatsapp.IWhatsappSendResponse, error) {
	return nil, nil
}
func (c *pairingTestConnection) SendReaction(string, string, bool, string) error { return nil }
func (c *pairingTestConnection) PublishStatus(string, *whatsapp.WhatsappAttachment) (string, error) {
	return "", nil
}
func (c *pairingTestConnection) HasChat(string) bool         { return false }
func (c *pairingTestConnection) GetLogger() *log.Entry       { return c.logEntry }
func (c *pairingTestConnection) Dispose(string)              { c.disposed = true }
func (c *pairingTestConnection) Delete() error               { return nil }
func (c *pairingTestConnection) IsInterfaceNil() bool        { return c == nil }
func (c *pairingTestConnection) HistorySync(time.Time) error { return nil }
func (c *pairingTestConnection) PairPhone(string) (string, error) {
	return "", nil
}
func (c *pairingTestConnection) SendChatPresence(string, uint) error { return nil }
func (c *pairingTestConnection) GetStatusManager() whatsapp.WhatsappStatusManagerInterface {
	return nil
}
func (c *pairingTestConnection) GetContactManager() whatsapp.WhatsappContactManagerInterface {
	return nil
}
func (c *pairingTestConnection) GetResume() *whatsapp.WhatsappConnectionStatus { return nil }
func (c *pairingTestConnection) GetOptions() *whatsapp.WhatsappOptions         { return &whatsapp.WhatsappOptions{} }
func (c *pairingTestConnection) SetOptions(*whatsapp.WhatsappOptions)          {}
func (c *pairingTestConnection) GetWId() string                                { return "" }
func (c *pairingTestConnection) SetWId(string)                                 {}
func (c *pairingTestConnection) GetReconnect() bool                            { return false }
func (c *pairingTestConnection) SetReconnect(bool)                             {}
func (c *pairingTestConnection) GetLogLevel() log.Level                        { return log.InfoLevel }
func (c *pairingTestConnection) SetLogLevel(log.Level)                         {}

type pairingTestHandlers struct{}

func (pairingTestHandlers) Message(*whatsapp.WhatsappMessage, string) {}
func (pairingTestHandlers) MessageStatusUpdate(string, whatsapp.WhatsappMessageStatus) bool {
	return false
}
func (pairingTestHandlers) Receipt(*whatsapp.WhatsappMessage)                   {}
func (pairingTestHandlers) LoggedOut(string)                                    {}
func (pairingTestHandlers) GetLeading() *whatsapp.WhatsappMessage               { return nil }
func (pairingTestHandlers) GetById(string) (*whatsapp.WhatsappMessage, error)   { return nil, nil }
func (pairingTestHandlers) OnConnected()                                        {}
func (pairingTestHandlers) OnDisconnected(string, string)                       {}
func (pairingTestHandlers) IsInterfaceNil() bool                                { return false }

func newPairingTestService() *QPWhatsappService {
	return &QPWhatsappService{
		Servers: make(map[string]*QpWhatsappServer),
		DB: &QpDatabase{
			Servers:     pairingTestServersData{},
			Dispatching: pairingTestDispatchingData{},
		},
		LogStruct: library.LogStruct{LogEntry: log.New().WithField("component", "pairing-test-service")},
		initlock:   &sync.Mutex{},
		appendlock: &sync.Mutex{},
	}
}

func TestAppendPaired_ReusesPairingConnectionAndWiresHandler(t *testing.T) {
	service := newPairingTestService()
	conn := newPairingTestConnection(t)

	paired := &QpWhatsappPairing{
		Token:    "pair-token",
		Wid:      "554791234567@s.whatsapp.net",
		Username: "tester@example.com",
		conn:     conn,
	}

	server, err := service.AppendPaired(paired)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if server == nil {
		t.Fatal("expected server instance")
	}
	if server.GetConnection() != conn {
		t.Fatal("expected paired connection to be reused by server")
	}
	if conn.updatedHandler != server.Handler {
		t.Fatal("expected pairing connection handler to be replaced by live server handler")
	}
	if !server.Verified {
		t.Fatal("expected paired server to be marked verified")
	}
	if got := server.QpServer.GetUser(); got != paired.Username {
		t.Fatalf("expected username %q, got %q", paired.Username, got)
	}
}

func TestAppendPaired_UpdatesExistingCachedServerConnection(t *testing.T) {
	service := newPairingTestService()

	existing := &QpWhatsappServer{
		QpServer:        &QpServer{Token: "existing-token"},
		Handler:         pairingTestHandlers{},
		syncConnection:  &sync.Mutex{},
		syncMessages:    &sync.Mutex{},
		Timestamps:      QpTimestamps{Start: time.Now().UTC()},
		Intent:          SessionIntentNone,
		db:              service.DB.Servers,
		LogStruct:       library.LogStruct{LogEntry: log.New().WithField("test", t.Name())},
	}
	existing.QpServer.SetUser("owner@example.com")
	oldConn := newPairingTestConnection(t)
	existing.connection = oldConn
	service.Servers["existing-token"] = existing

	newConn := newPairingTestConnection(t)
	paired := &QpWhatsappPairing{
		Token:    "existing-token",
		Wid:      "554792345678@s.whatsapp.net",
		Username: "owner@example.com",
		conn:     newConn,
	}

	server, err := service.AppendPaired(paired)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if server != existing {
		t.Fatal("expected existing cached server to be reused")
	}
	if !oldConn.disposed {
		t.Fatal("expected previous connection to be disposed when replacing pairing connection")
	}
	if server.GetConnection() != newConn {
		t.Fatal("expected new pairing connection to replace old server connection")
	}
	if newConn.updatedHandler != server.Handler {
		t.Fatal("expected new connection to receive live server handler")
	}
	if got := server.GetWId(); got != paired.Wid {
		t.Fatalf("expected wid %q, got %q", paired.Wid, got)
	}
}
