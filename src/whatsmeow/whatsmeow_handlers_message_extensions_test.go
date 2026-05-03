package whatsmeow

import (
	"testing"
	"time"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow/proto/waE2E"
)

func TestExtractExpirationFromMessage_UsesContextInfoExpiration(t *testing.T) {
	expiration := uint32(3600)
	msg := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			ContextInfo: &waE2E.ContextInfo{Expiration: &expiration},
		},
	}

	got := extractExpirationFromMessage(msg)
	if got != expiration {
		t.Fatalf("expected expiration=%d, got %d", expiration, got)
	}
}

func TestHandleEphemeralMessage_ProcessesInnerMessageAndSetsExpiresAt(t *testing.T) {
	expiration := uint32(600)
	baseTs := time.Unix(1710000000, 0).UTC()
	out := &whatsapp.WhatsappMessage{Timestamp: baseTs}

	in := &waE2E.FutureProofMessage{
		Message: &waE2E.Message{
			ExtendedTextMessage: &waE2E.ExtendedTextMessage{
				ContextInfo: &waE2E.ContextInfo{Expiration: &expiration},
			},
		},
	}

	h := minimalHandlers(t)
	entry := log.New().WithField("test", t.Name())
	HandleEphemeralMessage(h, entry, out, in)

	if out.Type != whatsapp.TextMessageType {
		t.Fatalf("expected type=%q, got %q", whatsapp.TextMessageType, out.Type)
	}

	wantExpiresAt := baseTs.Unix() + int64(expiration)
	if out.ExpiresAt != wantExpiresAt {
		t.Fatalf("expected expiresat=%d, got %d", wantExpiresAt, out.ExpiresAt)
	}
}

func TestHandleEphemeralMessage_DoesNotOverwriteExistingExpiresAt(t *testing.T) {
	expiration := uint32(600)
	out := &whatsapp.WhatsappMessage{
		Timestamp: time.Unix(1710000000, 0).UTC(),
		ExpiresAt: 1710009999,
	}

	in := &waE2E.FutureProofMessage{
		Message: &waE2E.Message{
			ExtendedTextMessage: &waE2E.ExtendedTextMessage{
				ContextInfo: &waE2E.ContextInfo{Expiration: &expiration},
			},
		},
	}

	h := minimalHandlers(t)
	entry := log.New().WithField("test", t.Name())
	HandleEphemeralMessage(h, entry, out, in)

	if out.ExpiresAt != 1710009999 {
		t.Fatalf("expected existing expiresat to remain unchanged, got %d", out.ExpiresAt)
	}
}
