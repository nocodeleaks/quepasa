package whatsmeow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nocodeleaks/quepasa/library"
	"go.mau.fi/whatsmeow/binary"
	"go.mau.fi/whatsmeow/types/events"
)

type callUnknownDump struct {
	Kind     string                 `json:"kind"`
	Captured string                 `json:"captured"`
	Tag      string                 `json:"tag,omitempty"`
	Subtag   string                 `json:"subtag,omitempty"`
	CallID   string                 `json:"call_id,omitempty"`
	XML      string                 `json:"xml,omitempty"`
	Node     map[string]interface{} `json:"node,omitempty"`
	RawJSON  string                 `json:"raw_json,omitempty"`
}

func DumpCallUnknownEvent(evt *events.UnknownCallEvent) (string, error) {
	if evt == nil || evt.Node == nil {
		return "", fmt.Errorf("nil UnknownCallEvent")
	}

	dumpDir := strings.TrimSpace(os.Getenv("QP_CALL_DUMP_DIR"))
	if dumpDir == "" {
		dumpDir = filepath.Join(".dist", "call_dumps")
	}
	if err := os.MkdirAll(dumpDir, 0o755); err != nil {
		return "", err
	}

	callID := ""
	subtag := ""
	if children, ok := evt.Node.Content.([]binary.Node); ok && len(children) > 0 {
		subtag = children[0].Tag
		if v, ok := children[0].Attrs["call-id"]; ok {
			callID = fmt.Sprint(v)
		}
	}
	callIDPart := sanitizeFilenamePart(callID)
	if callIDPart == "" {
		callIDPart = "unknown"
	}
	timestampStr := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("call_unknown_received_%s_%s.json", timestampStr, callIDPart)
	path := filepath.Join(dumpDir, filename)

	dump := callUnknownDump{
		Kind:     "UnknownCallEvent",
		Captured: time.Now().UTC().Format(time.RFC3339Nano),
		Tag:      evt.Node.Tag,
		Subtag:   subtag,
		CallID:   callID,
		XML:      evt.Node.XMLString(),
		Node:     flattenBinaryNode(*evt.Node),
		RawJSON:  library.ToJson(evt),
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
