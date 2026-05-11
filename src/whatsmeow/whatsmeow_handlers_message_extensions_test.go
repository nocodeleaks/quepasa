package whatsmeow

import (
	"testing"
	"time"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
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

func TestHandleKnowingMessagesPtvMessageAsVideo(t *testing.T) {
	handler := &WhatsmeowHandlers{}
	out := &whatsapp.WhatsappMessage{
		Id:   "test-ptv-id",
		Chat: whatsapp.WhatsappChat{Id: "12345@s.whatsapp.net"},
	}
	in := &waE2E.Message{
		PtvMessage: &waE2E.VideoMessage{
			Caption:    proto.String("ptv caption"),
			Mimetype:   proto.String("video/mp4"),
			FileLength: proto.Uint64(123),
		},
	}

	HandleKnowingMessages(handler, out, in)

	if out.Type != whatsapp.VideoMessageType {
		t.Fatalf("expected type %v, got %v", whatsapp.VideoMessageType, out.Type)
	}
	if out.Text != "ptv caption" {
		t.Fatalf("expected caption %q, got %q", "ptv caption", out.Text)
	}
	if out.Attachment == nil {
		t.Fatal("expected attachment to be set")
	}
	if out.Attachment.Mimetype != "video/mp4" {
		t.Fatalf("expected mimetype %q, got %q", "video/mp4", out.Attachment.Mimetype)
	}
	if out.Attachment.FileLength != 123 {
		t.Fatalf("expected file length %d, got %d", 123, out.Attachment.FileLength)
	}
	if !out.InVideoNote {
		t.Fatal("expected InVideoNote=true for ptv message")
	}
}

func TestGetDownloadableMessageReturnsPtvMessage(t *testing.T) {
	ptv := &waE2E.VideoMessage{Mimetype: proto.String("video/mp4")}
	msg := &waE2E.Message{PtvMessage: ptv}

	got := GetDownloadableMessage(msg)
	if got == nil {
		t.Fatal("expected downloadable message, got nil")
	}
	video, ok := got.(*waE2E.VideoMessage)
	if !ok {
		t.Fatalf("expected *waE2E.VideoMessage, got %T", got)
	}
	if video != ptv {
		t.Fatal("expected returned message to be the same ptv pointer")
	}
}
