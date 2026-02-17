package whatsmeow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.mau.fi/whatsmeow/types/events"
)

type callTransportDump struct {
	Kind      string             `json:"kind"`
	Captured  string             `json:"captured"`
	CallID    string             `json:"call_id"`
	From      string             `json:"from"`
	Timestamp string             `json:"timestamp,omitempty"`
	Attrs     map[string]string  `json:"attrs,omitempty"`
	Data      *TransportDataNode `json:"data"`
}

func DumpCallTransportEvent(evt *events.CallTransport, normalized *WhatsmeowCallTransport) (string, error) {
	if evt == nil {
		return "", fmt.Errorf("nil CallTransport")
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
	filename := fmt.Sprintf("call_transport_received_%s_%s.json", timestampStr, callIDPart)
	path := filepath.Join(dumpDir, filename)

	timestamp := ""
	if !evt.Timestamp.IsZero() {
		timestamp = evt.Timestamp.UTC().Format(time.RFC3339Nano)
	}

	data := &TransportDataNode{Tag: "transport", Attrs: map[string]string{}, Content: nil}
	if normalized != nil {
		data = normalized.GetData()
	}

	attrs := map[string]string{}
	for k, v := range data.Attrs {
		attrs[k] = v
	}

	dump := callTransportDump{
		Kind:      "CallTransport",
		Captured:  time.Now().UTC().Format(time.RFC3339Nano),
		CallID:    evt.CallID,
		From:      fmt.Sprint(evt.From),
		Timestamp: timestamp,
		Attrs:     attrs,
		Data:      data,
	}

	bytes, err := json.MarshalIndent(dump, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, bytes, 0o644); err != nil {
		return "", err
	}

	return path, nil
}

func sanitizeFilenamePart(in string) string {
	s := strings.TrimSpace(in)
	if s == "" {
		return ""
	}
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	s = replacer.Replace(s)
	if len(s) > 64 {
		s = s[:64]
	}
	return s
}
