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

type callTerminateDump struct {
	Kind     string `json:"kind"`
	Captured string `json:"captured"`
	CallID   string `json:"call_id"`
	From     string `json:"from"`
	Reason   string `json:"reason"`
	RawJSON  string `json:"raw_json,omitempty"`
}

func DumpCallTerminateEvent(evt *events.CallTerminate) (string, error) {
	if evt == nil {
		return "", fmt.Errorf("nil CallTerminate")
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
	filename := fmt.Sprintf("call_terminate_received_%s_%s.json", timestampStr, callIDPart)
	path := filepath.Join(dumpDir, filename)

	reason := ""
	reason = strings.TrimSpace(evt.Reason)

	dump := callTerminateDump{
		Kind:     "CallTerminate",
		Captured: time.Now().UTC().Format(time.RFC3339Nano),
		CallID:   evt.CallID,
		From:     fmt.Sprint(evt.From),
		Reason:   reason,
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

// DumpCallTerminateMeta dumps termination events that arrive via the BasicCallMeta handler path.
// Some terminations do not emit events.CallTerminate, so this is needed to capture "call failed" reasons.
func DumpCallTerminateMeta(evt types.BasicCallMeta, reason interface{}) (string, error) {
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
	filename := fmt.Sprintf("call_terminate_received_%s_%s.json", timestampStr, callIDPart)
	path := filepath.Join(dumpDir, filename)

	reasonStr := strings.TrimSpace(fmt.Sprintf("%v", reason))

	// Note: BasicCallMeta does not include the same fields as events.CallTerminate.
	// We still persist raw JSON for offline inspection.
	dump := callTerminateDump{
		Kind:     "CallTerminate",
		Captured: time.Now().UTC().Format(time.RFC3339Nano),
		CallID:   evt.CallID,
		From:     fmt.Sprint(evt.From),
		Reason:   reasonStr,
		RawJSON:  library.ToJson(map[string]interface{}{"meta": evt, "reason": reason}),
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
