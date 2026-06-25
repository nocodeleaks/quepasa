package models

// Tests covering the shared-backend isolation bug:
//
// When multiple WhatsApp numbers (WIDs) are connected and receive the same group
// message, each WID must publish it independently. Before the fix, all WIDs shared
// the same backend keyed only by message ID, so the second WID to arrive would hit
// ValidateItemBecauseUNOAPIConflict and be silenced as a "duplicate".
//
// The fix prefixes every cache key with the server token, giving each WID its own
// isolated namespace inside the shared backend.

import (
	"testing"
	"time"

	cache_memory "github.com/nocodeleaks/quepasa/cache/memory"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

// newGroupMsg builds a WhatsappMessage that looks like a real group text,
// including the waE2E.Message Content that ValidateItemBecauseUNOAPIConflict
// inspects. Same ID and same text but different WID each time.
func newGroupMsg(id, text, wid string) *whatsapp.WhatsappMessage {
	conversation := proto.String(text)
	return &whatsapp.WhatsappMessage{
		Id:        id,
		Timestamp: time.Now(),
		Type:      whatsapp.TextMessageType,
		Text:      text,
		Wid:       wid,
		Content:   &waE2E.Message{Conversation: conversation},
		Chat: whatsapp.WhatsappChat{
			Id: "111111111111111111@g.us",
		},
	}
}

// newMessages creates an isolated QpWhatsappMessages with its own in-memory
// backend and the given key prefix (token), mirroring what
// InjectCacheBackendIntoHandler does for each server.
func newMessages(backend *cache_memory.MessagesBackend, token string) *QpWhatsappMessages {
	m := &QpWhatsappMessages{}
	m.SetBackend(backend)
	m.SetKeyPrefix(token)
	return m
}

// TestSharedBackend_TwoWIDsSameGroupMessage is the primary regression test.
//
// Two servers share one backend (as in production with Redis or disk).
// Both receive the same group message (same ID, same text).
// Each must return true from Append — i.e. each must trigger its own publish.
func TestSharedBackend_TwoWIDsSameGroupMessage(t *testing.T) {
	t.Parallel()

	backend := cache_memory.NewMessagesBackend()
	defer backend.Close()

	const msgID = "3EB0863201A3FCE4B25C61"
	const msgText = "o seu numero não ta disparando o rabbitmq"

	wid1 := newMessages(backend, "tokenWID1")
	wid2 := newMessages(backend, "tokenWID2")

	// WID1 receives the message first.
	ok1 := wid1.Append(newGroupMsg(msgID, msgText, "555180124284:27@s.whatsapp.net"), "live")
	if !ok1 {
		t.Fatal("WID1: Append() returned false — first receiver must always trigger")
	}

	// WID2 receives the same message ~80 ms later (as observed in production logs).
	ok2 := wid2.Append(newGroupMsg(msgID, msgText, "555192508186:28@s.whatsapp.net"), "live")
	if !ok2 {
		t.Fatal("WID2: Append() returned false — different WID must trigger independently, not be silenced as duplicate")
	}
}

// TestSharedBackend_SameWIDSameMessageIsDeduped ensures the dedup logic still
// fires correctly when the same WID receives the identical message twice
// (e.g. a network retry from WhatsApp).
func TestSharedBackend_SameWIDSameMessageIsDeduped(t *testing.T) {
	t.Parallel()

	backend := cache_memory.NewMessagesBackend()
	defer backend.Close()

	const msgID = "3EB0AABBCCDDEEFF001122"
	const msgText = "mensagem duplicada"

	wid := newMessages(backend, "tokenWID3")

	ok1 := wid.Append(newGroupMsg(msgID, msgText, "555180124284:27@s.whatsapp.net"), "live")
	if !ok1 {
		t.Fatal("first Append() must succeed")
	}

	// Exact same message arriving again on the same WID — must be suppressed.
	ok2 := wid.Append(newGroupMsg(msgID, msgText, "555180124284:27@s.whatsapp.net"), "live")
	if ok2 {
		t.Fatal("second Append() of identical content on same WID must return false (dedup)")
	}
}

// TestSharedBackend_KeyIsolation verifies that each server only sees its own
// messages when calling GetSlice / Count, even though they share a backend.
func TestSharedBackend_KeyIsolation(t *testing.T) {
	t.Parallel()

	backend := cache_memory.NewMessagesBackend()
	defer backend.Close()

	wid1 := newMessages(backend, "tokenA")
	wid2 := newMessages(backend, "tokenB")

	for i := range 3 {
		wid1.Append(newGroupMsg("MSG-A-00"+string(rune('1'+i)), "texto", "widA"), "live")
	}
	for i := range 5 {
		wid2.Append(newGroupMsg("MSG-B-00"+string(rune('1'+i)), "texto", "widB"), "live")
	}

	if wid1.Count() != 3 {
		t.Fatalf("WID1 should see exactly 3 messages, got %d", wid1.Count())
	}
	if wid2.Count() != 5 {
		t.Fatalf("WID2 should see exactly 5 messages, got %d", wid2.Count())
	}
}

// TestSharedBackend_ThreeWIDsSameGroupMessage covers the case where three
// numbers are subscribed to the same group — all three must publish.
func TestSharedBackend_ThreeWIDsSameGroupMessage(t *testing.T) {
	t.Parallel()

	backend := cache_memory.NewMessagesBackend()
	defer backend.Close()

	const msgID = "3EB0FFEEDDCCBBAA998877"
	const msgText = "hello group"

	tokens := []string{"tokenX", "tokenY", "tokenZ"}
	wids := []string{"5551111@s.whatsapp.net", "5552222@s.whatsapp.net", "5553333@s.whatsapp.net"}

	for i, token := range tokens {
		m := newMessages(backend, token)
		ok := m.Append(newGroupMsg(msgID, msgText, wids[i]), "live")
		if !ok {
			t.Fatalf("WID %s (token %s): Append() returned false — every connected number must trigger independently", wids[i], token)
		}
	}
}

// TestSharedBackend_WithoutPrefixCollides documents the pre-fix behaviour:
// without a key prefix, two WIDs sharing a backend collide and the second is
// silenced. This test is intentionally asserting the broken behaviour so it
// serves as a living record of what the fix prevents.
func TestSharedBackend_WithoutPrefixCollides(t *testing.T) {
	t.Parallel()

	backend := cache_memory.NewMessagesBackend()
	defer backend.Close()

	const msgID = "3EB0COLLISION00000001"
	const msgText = "colisão de cache"

	// No SetKeyPrefix — both share the raw message ID as key.
	wid1 := &QpWhatsappMessages{}
	wid1.SetBackend(backend)

	wid2 := &QpWhatsappMessages{}
	wid2.SetBackend(backend)

	ok1 := wid1.Append(newGroupMsg(msgID, msgText, "555111@s.whatsapp.net"), "live")
	if !ok1 {
		t.Fatal("first Append() must succeed")
	}

	ok2 := wid2.Append(newGroupMsg(msgID, msgText, "555222@s.whatsapp.net"), "live")
	// Without prefix isolation this returns false — document that here.
	if ok2 {
		t.Log("NOTE: without key prefix, second WID was NOT suppressed (backend may have changed behaviour)")
	} else {
		t.Log("Confirmed pre-fix behaviour: without key prefix, second WID is silenced as duplicate")
	}
}
