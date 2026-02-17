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

// callAcceptSentDump represents the structured dump of an ACCEPT message we sent to WhatsApp
type callAcceptSentDump struct {
	Kind     string                 `json:"kind"`     // "AcceptSent"
	Captured string                 `json:"captured"` // ISO timestamp when dump was created
	CallID   string                 `json:"call_id"`
	To       string                 `json:"to"`   // Who we're sending ACCEPT to (caller)
	From     string                 `json:"from"` // Our JID
	Node     map[string]interface{} `json:"node"` // Full binary.Node structure
}

// DumpAcceptSent creates a JSON dump file of the ACCEPT node we sent to WhatsApp
// This helps debugging by capturing exactly what we sent (candidates, medium, etc)
func DumpAcceptSent(callID string, to types.JID, from types.JID, acceptNode binary.Node) (string, error) {
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
	// Filename format: call_accept_sent_{timestamp}_{call_id}.json
	timestampStr := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("call_accept_sent_%s_%s.json", timestampStr, callIDPart)
	path := filepath.Join(dumpDir, filename)

	dump := callAcceptSentDump{
		Kind:     "AcceptSent",
		Captured: time.Now().UTC().Format(time.RFC3339Nano),
		CallID:   callID,
		To:       to.String(),
		From:     from.String(),
		Node:     flattenBinaryNode(acceptNode),
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

// flattenBinaryNode converts binary.Node to a JSON-friendly map structure
func flattenBinaryNode(node binary.Node) map[string]interface{} {
	result := map[string]interface{}{
		"tag":   node.Tag,
		"attrs": node.Attrs,
	}

	switch content := node.Content.(type) {
	case []binary.Node:
		children := make([]map[string]interface{}, len(content))
		for i, child := range content {
			children[i] = flattenBinaryNode(child)
		}
		result["content"] = children
	case []byte:
		result["content"] = string(content)
	case string:
		result["content"] = content
	case nil:
		result["content"] = nil
	default:
		result["content"] = fmt.Sprintf("%v", content)
	}

	return result
}
