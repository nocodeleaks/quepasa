package whatsmeow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nocodeleaks/quepasa/library"
	"go.mau.fi/whatsmeow/types"
)

type callAcceptMetaDump struct {
	Kind        string `json:"kind"`
	Captured    string `json:"captured"`
	CallID      string `json:"call_id"`
	FromRaw     string `json:"from_raw"`
	FromToNonAD string `json:"from_to_non_ad,omitempty"`
	FromIsLID   bool   `json:"from_is_lid"`
	Timestamp   string `json:"timestamp,omitempty"`
	OwnRaw      string `json:"own_raw,omitempty"`
	OwnToNonAD  string `json:"own_to_non_ad,omitempty"`
	OwnIsLID    bool   `json:"own_is_lid,omitempty"`
	RawJSON     string `json:"raw_json,omitempty"`
	Notes       string `json:"notes,omitempty"`
}

// DumpCallAcceptMeta dumps call accepted events that arrive via the BasicCallMeta handler path.
// This is useful because some WhatsApp call state transitions may not emit events.CallAccept.
func DumpCallAcceptMeta(evt types.BasicCallMeta, ownID *types.JID) (string, error) {
	if strings.TrimSpace(evt.CallID) == "" {
		return "", fmt.Errorf("empty CallID")
	}

	dumpDir := strings.TrimSpace(os.Getenv("QP_CALL_DUMP_DIR"))
	if dumpDir == "" {
		dumpDir = filepath.Join(".dist", "call_dumps")
	}
	if err := os.MkdirAll(dumpDir, 0o755); err != nil {
		return "", err
	}

	callIDPart := sanitizeFilenamePart(evt.CallID)
	if callIDPart == "" {
		callIDPart = "unknown"
	}
	timestampStr := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("call_accept_received_meta_%s_%s.json", timestampStr, callIDPart)
	path := filepath.Join(dumpDir, filename)

	fromRaw := fmt.Sprint(evt.From)
	fromNonAD := ""
	if (evt.From != types.JID{}) {
		fromNonAD = fmt.Sprint(evt.From.ToNonAD())
	}
	fromIsLID := strings.Contains(fromRaw, "@lid")

	timestamp := ""
	if !evt.Timestamp.IsZero() {
		timestamp = evt.Timestamp.UTC().Format(time.RFC3339Nano)
	}

	ownRaw := ""
	ownNonAD := ""
	ownIsLID := false
	if ownID != nil {
		ownRaw = fmt.Sprint(*ownID)
		ownNonAD = fmt.Sprint(ownID.ToNonAD())
		ownIsLID = strings.Contains(ownRaw, "@lid")
	}

	dump := callAcceptMetaDump{
		Kind:        "CallAcceptMeta",
		Captured:    time.Now().UTC().Format(time.RFC3339Nano),
		CallID:      evt.CallID,
		FromRaw:     fromRaw,
		FromToNonAD: fromNonAD,
		FromIsLID:   fromIsLID,
		Timestamp:   timestamp,
		OwnRaw:      ownRaw,
		OwnToNonAD:  ownNonAD,
		OwnIsLID:    ownIsLID,
		RawJSON:     library.ToJson(evt),
		Notes:       "Compare from_raw vs from_to_non_ad to see if the event uses @lid or a phone-number JID (s.whatsapp.net).",
	}

	data, err := json.MarshalIndent(dump, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", err
	}
	return path, nil
}
