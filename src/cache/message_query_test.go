package cache

import (
	"testing"
	"time"

	"github.com/nocodeleaks/quepasa/whatsapp"
)

func entry(id, wid, chat string, ts int64) MessageRecordEntry {
	return MessageRecordEntry{Key: id, Record: MessageRecord{
		Message: &whatsapp.WhatsappMessage{Id: id, Wid: wid, Timestamp: time.Unix(ts, 0),
			Chat: whatsapp.WhatsappChat{Id: chat}},
	}}
}

func TestFilterAndPaginate(t *testing.T) {
	all := []MessageRecordEntry{
		entry("A", "W1", "chatX", 100),
		entry("B", "W1", "chatX", 300),
		entry("C", "W1", "chatY", 200),
		entry("D", "W2", "chatX", 400),
	}
	items, total := FilterAndPaginate(all, MessageQuery{Wid: "W1", ChatID: "chatX", Page: 1, Limit: 1})
	if total != 2 {
		t.Fatalf("total = %d, want 2", total)
	}
	if len(items) != 1 || items[0].Key != "B" {
		t.Fatalf("items = %+v, want [B]", items)
	}
}

func TestFilterAndPaginateKeyPrefix(t *testing.T) {
	// Same wid, different per-server key prefixes: KeyPrefix must isolate.
	mk := func(key, wid string, ts int64) MessageRecordEntry {
		return MessageRecordEntry{Key: key, Record: MessageRecord{
			Message: &whatsapp.WhatsappMessage{Id: key, Wid: wid, Timestamp: time.Unix(ts, 0)}}}
	}
	all := []MessageRecordEntry{
		mk("TOKA:1", "W1", 100),
		mk("TOKB:2", "W1", 200),
	}
	items, total := FilterAndPaginate(all, MessageQuery{Wid: "W1", KeyPrefix: "TOKA", Page: 1, Limit: 10})
	if total != 1 || len(items) != 1 || items[0].Key != "TOKA:1" {
		t.Fatalf("got total=%d items=%+v, want total=1 [TOKA:1]", total, items)
	}
}

func TestFilterAndPaginateTieBreak(t *testing.T) {
	all := []MessageRecordEntry{
		entry("AAA", "W1", "c", 500),
		entry("BBB", "W1", "c", 500), // same ts
	}
	items, total := FilterAndPaginate(all, MessageQuery{Wid: "W1", Page: 1, Limit: 10})
	if total != 2 || len(items) != 2 {
		t.Fatalf("total=%d len=%d, want 2/2", total, len(items))
	}
	if items[0].Key != "BBB" || items[1].Key != "AAA" {
		t.Fatalf("tie order = [%s,%s], want [BBB,AAA] (id desc)", items[0].Key, items[1].Key)
	}
}
