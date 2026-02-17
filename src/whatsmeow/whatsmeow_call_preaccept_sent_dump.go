package whatsmeow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.mau.fi/whatsmeow/binary"
	"go.mau.fi/whatsmeow/types"
)

type callPreacceptSentDump struct {
	Kind     string                 `json:"kind"`
	Captured string                 `json:"captured"`
	CallID   string                 `json:"call_id"`
	To       string                 `json:"to"`
	From     string                 `json:"from"`
	Node     map[string]interface{} `json:"node"`
}

// DumpPreacceptSent salva o preaccept que enviamos para o disco (quando ativado)
func DumpPreacceptSent(callID string, to types.JID, from types.JID, preNode binary.Node) (string, error) {
	if callID == "" {
		return "", fmt.Errorf("empty callID")
	}
	dumpDir := strings.TrimSpace(os.Getenv("QP_CALL_DUMP_DIR"))
	if dumpDir == "" {
		dumpDir = filepath.Join(".dist", "call_dumps")
	}
	if err := os.MkdirAll(dumpDir, 0o755); err != nil {
		return "", err
	}

	callIDPart := sanitizeFilenamePart(callID)
	if callIDPart == "" {
		callIDPart = "unknown"
	}
	timestampStr := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("call_preaccept_sent_%s_%s.json", timestampStr, callIDPart)
	path := filepath.Join(dumpDir, filename)

	dump := callPreacceptSentDump{
		Kind:     "PreacceptSent",
		Captured: time.Now().UTC().Format(time.RFC3339Nano),
		CallID:   callID,
		To:       to.String(),
		From:     from.String(),
		Node:     flattenBinaryNode(preNode),
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
