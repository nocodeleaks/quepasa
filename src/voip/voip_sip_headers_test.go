package voip

import (
	"strings"
	"testing"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	types "go.mau.fi/whatsmeow/types"
)

const (
	testAccountPhone = "5500000000000"
	testCallerPhone  = "5511000000000"
)

func TestSIPHeadersIncludeCallerMetadata(t *testing.T) {
	mgr := &VoipManager{sectionID: testAccountPhone + ":2", sessionToken: "token-123"}
	meta := callSIPMetadata{
		CallID:    "CALL-123",
		FromPhone: testCallerPhone,
		ToPhone:   testAccountPhone,
		Peer:      types.NewJID(testCallerPhone, types.DefaultUserServer),
		CallerInfo: CallerInfo{
			JID:          testCallerPhone + "@s.whatsapp.net",
			LID:          "111222333@lid",
			Phone:        testCallerPhone,
			PhoneE164:    "+" + testCallerPhone,
			Title:        "Cliente Principal",
			FullName:     "Cliente Principal",
			BusinessName: "Empresa Cliente",
			PushName:     "Cliente",
		},
	}

	headers := mgr.sipHeaders(meta)

	assertHeader(t, headers, "X-QuePasa-SessionId", testAccountPhone+":2")
	assertHeader(t, headers, "X-QuePasa-CallId", "CALL-123")
	assertHeader(t, headers, "X-QuePasa-Direction", "inbound-whatsapp")
	assertHeader(t, headers, "X-QuePasa-Account-Phone", testAccountPhone)
	assertHeader(t, headers, "X-QuePasa-Caller-Phone", testCallerPhone)
	assertHeader(t, headers, "X-QuePasa-Caller-E164", "+"+testCallerPhone)
	assertHeader(t, headers, "X-QuePasa-Caller-JID", testCallerPhone+"@s.whatsapp.net")
	assertHeader(t, headers, "X-QuePasa-Caller-LID", "111222333@lid")
	assertHeader(t, headers, "X-QuePasa-Caller-Title", "Cliente Principal")
	assertHeader(t, headers, "X-QuePasa-Caller-FullName", "Cliente Principal")
	assertHeader(t, headers, "X-QuePasa-Caller-BusinessName", "Empresa Cliente")
	assertHeader(t, headers, "X-QuePasa-Caller-PushName", "Cliente")
}

func TestSIPHeadersSanitizeValues(t *testing.T) {
	mgr := &VoipManager{sectionID: "5545343444095:2\r\nInjected: bad", sessionToken: "token-123\r\nInjected: bad"}
	meta := callSIPMetadata{
		CallID:    "CALL-123",
		FromPhone: "5511999999999",
		Peer:      types.NewJID("5511999999999", types.DefaultUserServer),
		CallerInfo: CallerInfo{
			Title: "Nome\r\nX-Bad: 1\tTeste",
		},
	}

	headers := mgr.sipHeaders(meta)

	assertHeader(t, headers, "X-QuePasa-SessionId", "5545343444095:2 Injected: bad")
	assertHeader(t, headers, "X-QuePasa-Caller-Title", "Nome X-Bad: 1 Teste")
	for name, value := range headers {
		if strings.ContainsAny(value, "\r\n\t") {
			t.Fatalf("header %s contains unsafe whitespace: %q", name, value)
		}
	}
}

func TestSIPHeadersResolveSectionInfoAtInviteTime(t *testing.T) {
	mgr := &VoipManager{
		callerInfoResolver: fakeSectionInfoResolver{
			sectionID:    "5545343444095:18",
			sessionToken: "token-live",
		},
	}
	meta := callSIPMetadata{
		CallID:    "CALL-123",
		FromPhone: "5511999999999",
		ToPhone:   "5545343444095",
		Peer:      types.NewJID("5511999999999", types.DefaultUserServer),
	}

	headers := mgr.sipHeaders(meta)

	assertHeader(t, headers, "X-QuePasa-SessionId", "5545343444095:18")
}

func TestSIPHeadersFallbackToAccountPhoneWhenSectionIsUnavailable(t *testing.T) {
	mgr := &VoipManager{}
	meta := callSIPMetadata{
		CallID:    "CALL-123",
		FromPhone: testCallerPhone,
		ToPhone:   testAccountPhone,
		Peer:      types.NewJID(testCallerPhone, types.DefaultUserServer),
	}

	headers := mgr.sipHeaders(meta)

	assertHeader(t, headers, "X-QuePasa-SessionId", testAccountPhone)
}

func TestVoipManagerShouldHandleCallsHonorsLiveCallMode(t *testing.T) {
	mgr := &VoipManager{
		mode: whatsapp.VoIPModeExclusive,
	}
	if !mgr.ShouldHandleCalls() {
		t.Fatal("exclusive mode should handle calls")
	}

	mgr.modeResolver = func() (whatsapp.VoIPMode, whatsapp.WhatsappBoolean) {
		return whatsapp.VoIPModeDisabled, whatsapp.TrueBooleanType
	}
	if mgr.ShouldHandleCalls() {
		t.Fatal("ignore mode must not forward calls to SIP")
	}

	mgr.modeResolver = func() (whatsapp.VoIPMode, whatsapp.WhatsappBoolean) {
		return whatsapp.VoIPModeDisabled, whatsapp.FalseBooleanType
	}
	if mgr.ShouldHandleCalls() {
		t.Fatal("deny mode must not forward calls to SIP")
	}

	mgr.modeResolver = func() (whatsapp.VoIPMode, whatsapp.WhatsappBoolean) {
		return whatsapp.VoIPModeExclusive, whatsapp.UnSetBooleanType
	}
	if !mgr.ShouldHandleCalls() {
		t.Fatal("forward mode should reactivate SIP forwarding without reconnecting")
	}
}

type fakeSectionInfoResolver struct {
	sectionID    string
	sessionToken string
}

func (resolver fakeSectionInfoResolver) ResolveVoIPCallerInfo(peer types.JID) CallerInfo {
	return CallerInfo{JID: peer.String(), Phone: peer.User}
}

func (resolver fakeSectionInfoResolver) GetVoIPSectionID() string {
	return resolver.sectionID
}

func (resolver fakeSectionInfoResolver) GetSessionToken() string {
	return resolver.sessionToken
}

func assertHeader(t *testing.T, headers map[string]string, name, want string) {
	t.Helper()
	if got := headers[name]; got != want {
		t.Fatalf("%s = %q, want %q", name, got, want)
	}
}
