package cable

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

type testFrame struct {
	Type    string          `json:"type"`
	ID      string          `json:"id,omitempty"`
	Command string          `json:"command,omitempty"`
	Event   string          `json:"event,omitempty"`
	Topic   string          `json:"topic,omitempty"`
	OK      bool            `json:"ok,omitempty"`
	Error   *ProtocolError  `json:"error,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func TestCableRejectsUnauthenticatedWebsocket(t *testing.T) {
	router := chi.NewRouter()
	Configure(router)

	server := httptest.NewServer(router)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/cable"
	conn, response, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if conn != nil {
		_ = conn.Close()
	}
	if response != nil {
		defer response.Body.Close()
	}

	if err == nil {
		t.Fatal("expected unauthenticated websocket dial to fail")
	}

	if response == nil || response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected handshake to fail with 401, got response %#v and err %v", response, err)
	}
}

func TestCableSessionSubscribeAndServerMessageEvent(t *testing.T) {
	db := newCableTestDatabase(t)
	setupCableTestService(t, db)
	defer cleanupCableTestService(t, db)

	user, err := models.WhatsappService.DB.Users.Create("owner@example.com", "Password123!")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	serverRecord := &models.QpServer{
		Token: "server-token-1",
		User:  user.Username,
		Wid:   "5511999999999@s.whatsapp.net",
	}
	if err := models.WhatsappService.DB.Servers.Add(serverRecord); err != nil {
		t.Fatalf("add server record: %v", err)
	}

	liveServer, err := models.WhatsappService.AppendNewServer(serverRecord)
	if err != nil {
		t.Fatalf("append live server: %v", err)
	}

	originalHub := CableHub
	CableHub = NewHub()
	defer func() {
		CableHub = originalHub
	}()

	router := chi.NewRouter()
	Configure(router)

	httpServer := httptest.NewServer(router)
	defer httpServer.Close()

	conn := dialCableTestWebsocket(t, httpServer.URL, user.Username)
	defer conn.Close()

	sessionReady := readCableTestFrame(t, conn)
	if sessionReady.Type != "event" || sessionReady.Event != "session.ready" {
		t.Fatalf("expected first frame to be session.ready event, got %+v", sessionReady)
	}

	subscribeCommand := map[string]any{
		"id":      "cmd-subscribe",
		"command": "subscribe",
		"data": map[string]any{
			"token": serverRecord.Token,
		},
	}
	if err := conn.WriteJSON(subscribeCommand); err != nil {
		t.Fatalf("write subscribe command: %v", err)
	}

	subscribeResponse := readCableTestFrame(t, conn)
	if subscribeResponse.Type != "response" || subscribeResponse.Command != "subscribe" || !subscribeResponse.OK {
		t.Fatalf("expected successful subscribe response, got %+v", subscribeResponse)
	}

	var subscribePayload SubscriptionResponsePayload
	if err := json.Unmarshal(subscribeResponse.Data, &subscribePayload); err != nil {
		t.Fatalf("decode subscribe response: %v", err)
	}
	if len(subscribePayload.Subscriptions) != 1 || subscribePayload.Subscriptions[0] != "server:"+serverRecord.Token {
		t.Fatalf("unexpected subscriptions payload: %+v", subscribePayload)
	}

	CableHub.PublishServerMessage(liveServer, &whatsapp.WhatsappMessage{
		Id:        "msg-1",
		Timestamp: time.Now().UTC(),
		Type:      whatsapp.TextMessageType,
		Chat: whatsapp.WhatsappChat{
			Id:    "5511999999999@s.whatsapp.net",
			Phone: "5511999999999",
			Title: "Contato Teste",
		},
		Text: "hello from cable test",
	})

	messageEvent := readCableTestFrame(t, conn)
	if messageEvent.Type != "event" || messageEvent.Event != "server.message" {
		t.Fatalf("expected server.message event, got %+v", messageEvent)
	}
	if messageEvent.Topic != "server:"+serverRecord.Token {
		t.Fatalf("expected event topic server:%s, got %q", serverRecord.Token, messageEvent.Topic)
	}

	var messagePayload struct {
		Token   string                 `json:"token"`
		Message map[string]interface{} `json:"message"`
	}
	if err := json.Unmarshal(messageEvent.Data, &messagePayload); err != nil {
		t.Fatalf("decode server.message payload: %v", err)
	}
	if messagePayload.Token != serverRecord.Token {
		t.Fatalf("expected payload token %s, got %s", serverRecord.Token, messagePayload.Token)
	}
	if messagePayload.Message == nil || messagePayload.Message["id"] != "msg-1" {
		t.Fatalf("expected published message payload, got %+v", messagePayload.Message)
	}
}

func dialCableTestWebsocket(t *testing.T, baseURL string, username string) *websocket.Conn {
	t.Helper()

	wsURL := "ws" + strings.TrimPrefix(baseURL, "http") + "/cable"
	header := http.Header{}
	header.Set("Authorization", "Bearer "+encodeCableTestToken(t, username))

	conn, response, err := websocket.DefaultDialer.Dial(wsURL, header)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		t.Fatalf("dial cable websocket: %v (response=%#v)", err, response)
	}

	return conn
}

func encodeCableTestToken(t *testing.T, username string) string {
	t.Helper()

	_, tokenString, err := cableTokenAuth.Encode(jwt.MapClaims{"user_id": username})
	if err != nil {
		t.Fatalf("encode cable token: %v", err)
	}

	return tokenString
}

func readCableTestFrame(t *testing.T, conn *websocket.Conn) testFrame {
	t.Helper()

	if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}

	_, payload, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read websocket message: %v", err)
	}

	var frame testFrame
	if err := json.Unmarshal(payload, &frame); err != nil {
		t.Fatalf("decode websocket frame: %v; payload=%s", err, string(payload))
	}

	return frame
}

func newCableTestDatabase(t *testing.T) *sqlx.DB {
	t.Helper()

	db, err := sqlx.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite database: %v", err)
	}

	schema := `
		CREATE TABLE IF NOT EXISTS users (
			username TEXT PRIMARY KEY,
			password TEXT NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS servers (
			token TEXT PRIMARY KEY,
			wid TEXT UNIQUE,
			user TEXT NOT NULL,
			verified BOOLEAN DEFAULT 0,
			devel BOOLEAN DEFAULT 0,
			metadata TEXT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			groups INTEGER DEFAULT 1,
			broadcasts INTEGER DEFAULT 1,
			readreceipts INTEGER DEFAULT 1,
			calls INTEGER DEFAULT 1,
			readupdate INTEGER DEFAULT 1,
			FOREIGN KEY (user) REFERENCES users(username)
		);
		CREATE TABLE IF NOT EXISTS dispatching (
			context TEXT NOT NULL,
			connection_string TEXT NOT NULL,
			type TEXT NOT NULL DEFAULT 'webhook',
			forward_internal BOOLEAN DEFAULT 0,
			track_id TEXT,
			extra TEXT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			groups INTEGER DEFAULT 1,
			broadcasts INTEGER DEFAULT 1,
			readreceipts INTEGER DEFAULT 1,
			calls INTEGER DEFAULT 1,
			PRIMARY KEY (context, connection_string)
		);
	`
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		t.Fatalf("create cable test schema: %v", err)
	}

	return db
}

func setupCableTestService(t *testing.T, db *sqlx.DB) {
	t.Helper()

	models.WhatsappService = &models.QPWhatsappService{
		Servers: make(map[string]*models.QpWhatsappServer),
		DB: &models.QpDatabase{
			Connection:  db,
			Users:       models.NewQpDataUserSql(db),
			Servers:     models.NewQpDataServerSql(db),
			Dispatching: models.NewQpDataServerDispatchingSql(db),
		},
	}
}

func cleanupCableTestService(t *testing.T, db *sqlx.DB) {
	t.Helper()

	if db != nil {
		_ = db.Close()
	}
	models.WhatsappService = nil
}
