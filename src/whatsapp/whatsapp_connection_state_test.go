package whatsapp

import "testing"

func TestWhatsappConnectionStateStringOutOfRangeFallsBackToUnknown(t *testing.T) {
	t.Parallel()

	state := WhatsappConnectionState(999)
	if got := state.String(); got != "Unknown" {
		t.Fatalf("expected Unknown for out-of-range state, got %q", got)
	}
}
