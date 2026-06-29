package whatsmeow

import (
	"strings"
	"testing"

	"go.mau.fi/whatsmeow/types"
)

// Characterization tests for the pure (connection-free) whatsmeow helpers:
// identity normalization and the interactive "buttons:" text protocol. These
// are the driver's translation seams and a known source of LID-vs-phone bugs;
// freezing their behaviour guards future refactors (PLAN P4.1).

// TestExtractContactNamePriority locks the documented name-resolution order:
// FullName > BusinessName > PushName > FirstName, and "" when not found.
func TestExtractContactNamePriority(t *testing.T) {
	full := types.ContactInfo{Found: true, FullName: "Full", BusinessName: "Biz", PushName: "Push", FirstName: "First"}
	if got := ExtractContactName(full); got != "Full" {
		t.Fatalf("FullName should win: got %q", got)
	}

	biz := types.ContactInfo{Found: true, BusinessName: "Biz", PushName: "Push", FirstName: "First"}
	if got := ExtractContactName(biz); got != "Biz" {
		t.Fatalf("BusinessName should win when no FullName: got %q", got)
	}

	push := types.ContactInfo{Found: true, PushName: "Push", FirstName: "First"}
	if got := ExtractContactName(push); got != "Push" {
		t.Fatalf("PushName should win when no Full/Business: got %q", got)
	}

	first := types.ContactInfo{Found: true, FirstName: "First"}
	if got := ExtractContactName(first); got != "First" {
		t.Fatalf("FirstName is the last resort: got %q", got)
	}

	if got := ExtractContactName(types.ContactInfo{Found: false, FullName: "Ignored"}); got != "" {
		t.Fatalf("not-found contact must resolve to empty, got %q", got)
	}
}

// TestCleanJIDStripsDeviceAndAgent verifies that CleanJID reduces a JID to its
// User+Server identity, dropping the device/session suffix that otherwise causes
// the same contact to be treated as several distinct chats.
func TestCleanJIDStripsDeviceAndAgent(t *testing.T) {
	in := types.JID{User: "5511999999999", Server: types.DefaultUserServer, Device: 7}

	out := CleanJID(in)

	if out.User != in.User {
		t.Fatalf("User changed: got %q, want %q", out.User, in.User)
	}
	if out.Server != in.Server {
		t.Fatalf("Server changed: got %q, want %q", out.Server, in.Server)
	}
	if out.Device != 0 {
		t.Fatalf("Device suffix not stripped: got %d", out.Device)
	}
}

// TestIsValidForButtons locks the trigger contract: plain text is not a buttons
// message; the "$buttons:[...]" syntax is.
func TestIsValidForButtons(t *testing.T) {
	if IsValidForButtons("just a normal message") {
		t.Fatal("plain text wrongly detected as buttons message")
	}
	if IsValidForButtons("the word buttons: alone is not enough") {
		t.Fatal("bare 'buttons:' without the $/# marker must not match")
	}
	if !IsValidForButtons("Pick one $buttons:[(1) Yes,(2) No] Thanks") {
		t.Fatal("valid $buttons syntax not detected")
	}
}

// TestConvertButtonsToText freezes the rendering of the buttons protocol into a
// plain-text fallback: content first, each button as "*id)* display", footer last.
func TestConvertButtonsToText(t *testing.T) {
	out := ConvertButtonsToText("Pick one $buttons:[(1) Yes,(2) No] Thanks")

	for _, want := range []string{"Pick one", "*1)* Yes", "*2)* No", "Thanks"} {
		if !strings.Contains(out, want) {
			t.Fatalf("rendered buttons text missing %q\n full output:\n%s", want, out)
		}
	}

	// Button order must be preserved.
	if strings.Index(out, "*1)* Yes") > strings.Index(out, "*2)* No") {
		t.Fatalf("button order not preserved:\n%s", out)
	}
}
