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

type rawReceiptDump struct {
	Kind      string                 `json:"kind"`
	Captured  string                 `json:"captured"`
	MessageID string                 `json:"message_id,omitempty"`
	CallID    string                 `json:"call_id,omitempty"`
	ChildTags []string               `json:"child_tags,omitempty"`
	XML       string                 `json:"xml,omitempty"`
	Node      map[string]interface{} `json:"node,omitempty"`
}

func shouldDumpRawReceipt(node *binary.Node) (bool, string, []string) {
	if node == nil {
		return false, "", nil
	}
	children := node.GetChildren()
	if len(children) == 0 {
		return false, "", nil
	}
	childTags := make([]string, 0, len(children))
	callID := ""
	for _, child := range children {
		childTags = append(childTags, child.Tag)
		if v, ok := child.Attrs["call-id"]; ok && fmt.Sprint(v) != "" {
			callID = fmt.Sprint(v)
		}
	}
	return true, callID, childTags
}

func DumpRawReceiptNode(node *binary.Node) (string, error) {
	if node == nil {
		return "", fmt.Errorf("nil receipt node")
	}
	ok, callID, childTags := shouldDumpRawReceipt(node)
	if !ok {
		return "", nil
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
	childPart := "unknown"
	if len(childTags) > 0 {
		childPart = sanitizeFilenamePart(childTags[0])
		if childPart == "" {
			childPart = "unknown"
		}
	}
	filename := fmt.Sprintf("call_receipt_raw_received_%s_%s_%s.json", timestampStr, callIDPart, childPart)
	path := filepath.Join(dumpDir, filename)

	messageID := ""
	if v, ok := node.Attrs["id"]; ok {
		messageID = fmt.Sprint(v)
	}

	dump := rawReceiptDump{
		Kind:      "RawReceiptNode",
		Captured:  time.Now().UTC().Format(time.RFC3339Nano),
		MessageID: messageID,
		CallID:    callID,
		ChildTags: childTags,
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

func (source *WhatsmeowHandlers) HandleRawReceiptNode(node *binary.Node) {
	if node == nil {
		return
	}
	logentry := source.GetLogger()
	ok, callID, childTags := shouldDumpRawReceipt(node)
	if !ok {
		return
	}
	logentry.Infof("[CALL] Raw receipt node: callID=%s childTags=%v", callID, childTags)
	source.maybeSendRawReceiptAck(node)
	path, err := DumpRawReceiptNode(node)
	if err != nil {
		logentry.Errorf("[CALL] Raw receipt dump failed: err=%v", err)
		return
	}
	if path != "" {
		logentry.Infof("[CALL] Raw receipt dumped: path=%s", path)
	}
}
