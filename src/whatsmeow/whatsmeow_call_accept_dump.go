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
	"go.mau.fi/whatsmeow/types/events"
)

type callAcceptReceivedDump struct {
	Kind        string `json:"kind"`
	Captured    string `json:"captured"`
	CallID      string `json:"call_id"`
	FromRaw     string `json:"from_raw"`
	FromToNonAD string `json:"from_to_non_ad,omitempty"`
	FromIsLID   bool   `json:"from_is_lid"`

	CallCreatorRaw     string `json:"call_creator_raw,omitempty"`
	CallCreatorToNonAD string `json:"call_creator_to_non_ad,omitempty"`
	CallCreatorIsLID   bool   `json:"call_creator_is_lid,omitempty"`
	CallCreatorAlt     string `json:"call_creator_alt,omitempty"`

	RemotePlatform string `json:"remote_platform,omitempty"`
	RemoteVersion  string `json:"remote_version,omitempty"`
	Timestamp      string `json:"timestamp,omitempty"`

	OwnRaw     string `json:"own_raw,omitempty"`
	OwnToNonAD string `json:"own_to_non_ad,omitempty"`
	OwnIsLID   bool   `json:"own_is_lid,omitempty"`
	RawJSON    string `json:"raw_json,omitempty"`
	Notes      string `json:"notes,omitempty"`
}

func DumpCallAcceptReceivedEvent(evt *events.CallAccept, ownID *types.JID) (string, error) {
	if evt == nil {
		return "", fmt.Errorf("nil CallAccept")
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
	filename := fmt.Sprintf("call_accept_received_%s_%s.json", timestampStr, callIDPart)
	path := filepath.Join(dumpDir, filename)

	fromRaw := fmt.Sprint(evt.From)
	fromNonAD := ""
	if (evt.From != types.JID{}) {
		fromNonAD = fmt.Sprint(evt.From.ToNonAD())
	}
	fromIsLID := strings.Contains(fromRaw, "@lid")

	callCreatorRaw := ""
	callCreatorNonAD := ""
	callCreatorIsLID := false
	if (evt.CallCreator != types.JID{}) {
		callCreatorRaw = fmt.Sprint(evt.CallCreator)
		callCreatorNonAD = fmt.Sprint(evt.CallCreator.ToNonAD())
		callCreatorIsLID = strings.Contains(callCreatorRaw, "@lid")
	}
	callCreatorAlt := strings.TrimSpace(fmt.Sprint(evt.CallCreatorAlt))

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

	dump := callAcceptReceivedDump{
		Kind:               "CallAcceptReceived",
		Captured:           time.Now().UTC().Format(time.RFC3339Nano),
		CallID:             evt.CallID,
		FromRaw:            fromRaw,
		FromToNonAD:        fromNonAD,
		FromIsLID:          fromIsLID,
		CallCreatorRaw:     callCreatorRaw,
		CallCreatorToNonAD: callCreatorNonAD,
		CallCreatorIsLID:   callCreatorIsLID,
		CallCreatorAlt:     callCreatorAlt,
		RemotePlatform:     strings.TrimSpace(evt.RemotePlatform),
		RemoteVersion:      strings.TrimSpace(evt.RemoteVersion),
		Timestamp:          timestamp,
		OwnRaw:             ownRaw,
		OwnToNonAD:         ownNonAD,
		OwnIsLID:           ownIsLID,
		RawJSON:            library.ToJson(evt),
		Notes:              "Compare from_raw vs from_to_non_ad to see if the event uses @lid or a phone-number JID (s.whatsapp.net).",
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
