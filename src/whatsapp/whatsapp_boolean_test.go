package whatsapp

import "testing"

func TestWhatsappBooleanBooleanInvalidValueFallsBackToFalse(t *testing.T) {
	if got := WhatsappBoolean(99).Boolean(); got {
		t.Fatalf("expected invalid value to fall back to false, got true")
	}
}

func TestWhatsappBooleanBooleanUnsetFallsBackToFalse(t *testing.T) {
	if got := UnSetBooleanType.Boolean(); got {
		t.Fatalf("expected unset value to fall back to false, got true")
	}
}
