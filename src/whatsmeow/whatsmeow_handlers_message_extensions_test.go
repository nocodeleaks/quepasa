package whatsmeow

import (
	"testing"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

func TestHandleKnowingMessagesPtvMessageAsVideo(t *testing.T) {
	handler := &WhatsmeowHandlers{}
	out := &whatsapp.WhatsappMessage{
		Id:   "test-message-id",
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
		t.Fatal("expected invideonote=true for ptv message")
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
		t.Fatal("expected returned message to be the same ptv message pointer")
	}
}
