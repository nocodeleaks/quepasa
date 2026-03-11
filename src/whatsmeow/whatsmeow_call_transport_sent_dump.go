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

type callTransportSentDump struct {
	Kind              string                         `json:"kind"`
	Captured          string                         `json:"captured"`
	CallID            string                         `json:"call_id"`
	To                string                         `json:"to"`
	From              string                         `json:"from"`
	Node              map[string]interface{}         `json:"node"`
	CompactItems      []transportCompactItem         `json:"compact_items,omitempty"`
	CompactCandidates []transportSentCompactCandidate `json:"compact_candidates,omitempty"`
}

// DumpTransportSent salva o transport que enviamos ao peer/WhatsApp
func DumpTransportSent(callID string, to types.JID, from types.JID, transportNode binary.Node) (string, error) {
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
	filename := fmt.Sprintf("call_transport_sent_%s_%s.json", timestampStr, callIDPart)
	path := filepath.Join(dumpDir, filename)

	dump := callTransportSentDump{
		Kind:              "TransportSent",
		Captured:          time.Now().UTC().Format(time.RFC3339Nano),
		CallID:            callID,
		To:                to.String(),
		From:              from.String(),
		Node:              flattenBinaryNode(transportNode),
		CompactItems:      extractCompactTransportItemsFromBinaryNode(transportNode),
		CompactCandidates: extractCompactTransportCandidates(transportNode),
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
