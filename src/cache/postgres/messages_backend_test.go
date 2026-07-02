package postgres

import (
	"os"
	"testing"
	"time"

	"github.com/nocodeleaks/quepasa/cache"
	"github.com/nocodeleaks/quepasa/whatsapp"
)

func newTestBackend(t *testing.T) *MessagesBackend {
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN not set")
	}
	b, err := NewFromDSN(dsn, nil, time.Hour)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { b.db.Exec("DROP TABLE IF EXISTS messages"); b.Close() })
	b.db.Exec("DELETE FROM messages")
	return b
}

func rec(id, wid, chat string, ts int64) cache.MessageRecord {
	return cache.MessageRecord{Message: &whatsapp.WhatsappMessage{
		Id: id, Wid: wid, Timestamp: time.Unix(ts, 0), Chat: whatsapp.WhatsappChat{Id: chat}}}
}

func TestSetGetDelete(t *testing.T) {
	b := newTestBackend(t)
	if err := b.Set("A", rec("A", "W1", "c1", 100)); err != nil {
		t.Fatalf("Set: %v", err)
	}
	got, found, err := b.Get("A")
	if err != nil || !found || got.Message.Id != "A" {
		t.Fatalf("Get: got=%+v found=%v err=%v", got, found, err)
	}
	if err := b.Delete("A"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, found, _ := b.Get("A"); found {
		t.Fatalf("still found after delete")
	}
}

func TestForeverIsNullExpiry(t *testing.T) {
	b := newTestBackend(t)
	if err := b.Set("A", rec("A", "W1", "c1", 100)); err != nil { // ExpiresAt zero => forever
		t.Fatalf("Set: %v", err)
	}
	var expNull bool
	b.db.QueryRow("SELECT expires_at IS NULL FROM messages WHERE msgkey='A'").Scan(&expNull)
	if !expNull {
		t.Fatalf("forever record should store expires_at NULL")
	}
}

func TestQueryPaged(t *testing.T) {
	b := newTestBackend(t)
	b.Set("A", rec("A", "W1", "c1", 100))
	b.Set("B", rec("B", "W1", "c1", 300))
	b.Set("C", rec("C", "W1", "c2", 200))
	items, total, err := b.Query(cache.MessageQuery{Wid: "W1", ChatID: "c1", Page: 1, Limit: 1})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if total != 2 || len(items) != 1 || items[0].Key != "B" {
		t.Fatalf("got total=%d items=%+v, want total=2 [B]", total, items)
	}
}
