package whatsmeow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.mau.fi/whatsmeow/binary"
)

type rawAckDump struct {
	Kind      string                 `json:"kind"`
	Captured  string                 `json:"captured"`
	Class     string                 `json:"class,omitempty"`
	MessageID string                 `json:"message_id,omitempty"`
	CallID    string                 `json:"call_id,omitempty"`
	XML       string                 `json:"xml,omitempty"`
	Node      map[string]interface{} `json:"node,omitempty"`
}

func DumpRawAckNode(node *binary.Node) (string, error) {
	if node == nil {
		return "", fmt.Errorf("nil ack node")
	}

	callID := ""
	if v, ok := node.Attrs["call-id"]; ok {
		callID = fmt.Sprint(v)
	}
	if callID == "" {
		for _, child := range node.GetChildren() {
			if v, ok := child.Attrs["call-id"]; ok && fmt.Sprint(v) != "" {
				callID = fmt.Sprint(v)
				break
			}
		}
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
	classPart := sanitizeFilenamePart(fmt.Sprint(node.Attrs["class"]))
	if classPart == "" {
		classPart = "unknown"
	}
	timestampStr := time.Now().Format("20060102150405.000000000")
	filename := fmt.Sprintf("call_ack_raw_received_%s_%s_%s.json", timestampStr, callIDPart, classPart)
	path := filepath.Join(dumpDir, filename)

	messageID := ""
	if v, ok := node.Attrs["id"]; ok {
		messageID = fmt.Sprint(v)
	}

	dump := rawAckDump{
		Kind:      "RawAckNode",
		Captured:  time.Now().UTC().Format(time.RFC3339Nano),
		Class:     fmt.Sprint(node.Attrs["class"]),
		MessageID: messageID,
		CallID:    callID,
		XML:       node.XMLString(),
		Node:      flattenBinaryNode(*node),
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

func (source *WhatsmeowHandlers) HandleRawAckNode(node *binary.Node) {
	if node == nil {
		return
	}
	logentry := source.GetLogger()
	path, err := DumpRawAckNode(node)
	if err != nil {
		logentry.Errorf("[CALL] Raw ack dump failed: err=%v", err)
		return
	}
	if path != "" {
		logentry.Infof("[CALL] Raw ack dumped: path=%s", path)
	}
}
