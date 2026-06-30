package voip

import (
	"strings"
	"testing"

	types "go.mau.fi/whatsmeow/types"
)

func TestSIPHeadersIncludeCallerMetadata(t *testing.T) {
	mgr := &VoipManager{sectionID: "section-123"}
	meta := callSIPMetadata{
		CallID:    "CALL-123",
		FromPhone: "5511999999999",
		ToPhone:   "5521967609095",
		Peer:      types.NewJID("5511999999999", types.DefaultUserServer),
		CallerInfo: CallerInfo{
			JID:          "5511999999999@s.whatsapp.net",
			LID:          "111222333@lid",
			Phone:        "5511999999999",
			PhoneE164:    "+5511999999999",
			Title:        "Cliente Principal",
			FullName:     "Cliente Principal",
			BusinessName: "Empresa Cliente",
			PushName:     "Cliente",
		},
	}

	headers := mgr.sipHeaders(meta)

	assertHeader(t, headers, "X-QuePasa-SectionId", "section-123")
	assertHeader(t, headers, "X-QuePasa-Token", "section-123")
	assertHeader(t, headers, "X-QuePasa-CallId", "CALL-123")
	assertHeader(t, headers, "X-QuePasa-Direction", "inbound-whatsapp")
	assertHeader(t, headers, "X-QuePasa-Account-Phone", "5521967609095")
	assertHeader(t, headers, "X-QuePasa-Caller-Phone", "5511999999999")
	assertHeader(t, headers, "X-QuePasa-Caller-E164", "+5511999999999")
	assertHeader(t, headers, "X-QuePasa-Caller-JID", "5511999999999@s.whatsapp.net")
	assertHeader(t, headers, "X-QuePasa-Caller-LID", "111222333@lid")
	assertHeader(t, headers, "X-QuePasa-Caller-Title", "Cliente Principal")
	assertHeader(t, headers, "X-QuePasa-Caller-FullName", "Cliente Principal")
	assertHeader(t, headers, "X-QuePasa-Caller-BusinessName", "Empresa Cliente")
	assertHeader(t, headers, "X-QuePasa-Caller-PushName", "Cliente")
}

func TestSIPHeadersSanitizeValues(t *testing.T) {
	mgr := &VoipManager{sectionID: "section-123\r\nInjected: bad"}
	meta := callSIPMetadata{
		CallID:    "CALL-123",
		FromPhone: "5511999999999",
		Peer:      types.NewJID("5511999999999", types.DefaultUserServer),
		CallerInfo: CallerInfo{
			Title: "Nome\r\nX-Bad: 1\tTeste",
		},
	}

	headers := mgr.sipHeaders(meta)

	assertHeader(t, headers, "X-QuePasa-SectionId", "section-123 Injected: bad")
	assertHeader(t, headers, "X-QuePasa-Caller-Title", "Nome X-Bad: 1 Teste")
	for name, value := range headers {
		if strings.ContainsAny(value, "\r\n\t") {
			t.Fatalf("header %s contains unsafe whitespace: %q", name, value)
		}
	}
}

func assertHeader(t *testing.T, headers map[string]string, name, want string) {
	t.Helper()
	if got := headers[name]; got != want {
		t.Fatalf("%s = %q, want %q", name, got, want)
	}
}
