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

type callRelayLatencyDump struct {
	Kind      string          `json:"kind"`
	Captured  string          `json:"captured"`
	CallID    string          `json:"call_id"`
	From      string          `json:"from"`
	Endpoints []RelayEndpoint `json:"endpoints,omitempty"`
}

func DumpCallRelayLatencyEvent(evt *events.CallRelayLatency, endpoints []RelayEndpoint) (string, error) {
	if evt == nil {
		return "", fmt.Errorf("nil CallRelayLatency")
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
	filename := fmt.Sprintf("call_relaylatency_%s_%s.json", timestampStr, callIDPart)
	path := filepath.Join(dumpDir, filename)

	dump := callRelayLatencyDump{
		Kind:      "CallRelayLatency",
		Captured:  time.Now().UTC().Format(time.RFC3339Nano),
		CallID:    evt.CallID,
		From:      fmt.Sprint(evt.From),
		Endpoints: endpoints,
	}
	for i := range dump.Endpoints {
		if dump.Endpoints[i].CompactHex == "" {
			if ep := encodeCompactEndpoint6(dump.Endpoints[i].IP, dump.Endpoints[i].Port); ep != nil {
				dump.Endpoints[i].CompactHex = ep.RawHex
			}
		}
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
