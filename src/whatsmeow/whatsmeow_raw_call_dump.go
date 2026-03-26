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

type rawCallDump struct {
	Kind     string                 `json:"kind"`
	Captured string                 `json:"captured"`
	Subtag   string                 `json:"subtag,omitempty"`
	CallID   string                 `json:"call_id,omitempty"`
	XML      string                 `json:"xml,omitempty"`
	Node     map[string]interface{} `json:"node,omitempty"`
}

func DumpRawCallNode(node *binary.Node) (string, error) {
	if node == nil {
		return "", fmt.Errorf("nil call node")
	}
	children := node.GetChildren()
	if len(children) == 0 {
		return "", nil
	}

	subtag := children[0].Tag
	callID := ""
	if v, ok := children[0].Attrs["call-id"]; ok {
		callID = fmt.Sprint(v)
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
	timestampStr := time.Now().Format("20060102150405.000000000")
	subtagPart := sanitizeFilenamePart(subtag)
	if subtagPart == "" {
		subtagPart = "unknown"
	}
	filename := fmt.Sprintf("call_raw_received_%s_%s_%s.json", timestampStr, callIDPart, subtagPart)
	path := filepath.Join(dumpDir, filename)

	dump := rawCallDump{
		Kind:     "RawCallNode",
		Captured: time.Now().UTC().Format(time.RFC3339Nano),
		Subtag:   subtag,
		CallID:   callID,
		XML:      node.XMLString(),
		Node:     flattenBinaryNode(*node),
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

func (source *WhatsmeowHandlers) HandleRawCallNode(node *binary.Node) {
	if node == nil {
		return
	}
	logentry := source.GetLogger()
	path, err := DumpRawCallNode(node)
	if err != nil {
		logentry.Errorf("[CALL] Raw call dump failed: err=%v", err)
		return
	}
	if path != "" {
		logentry.Infof("[CALL] Raw call dumped: path=%s", path)
	}
	source.maybeSendRawCallAck(node)
}
