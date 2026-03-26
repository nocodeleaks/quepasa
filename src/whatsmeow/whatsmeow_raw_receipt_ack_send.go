package whatsmeow

import (
	"context"
	"fmt"

	"go.mau.fi/whatsmeow/binary"
)

func rawReceiptAckTag(node *binary.Node) (string, bool) {
	if node == nil || node.Tag != "receipt" {
		return "", false
	}
	children := node.GetChildren()
	if len(children) == 0 {
		return "", false
	}
	switch children[0].Tag {
	case "accept":
		return children[0].Tag, true
	default:
		return "", false
	}
}

func (source *WhatsmeowHandlers) maybeSendRawReceiptAck(node *binary.Node) {
	if source == nil || node == nil {
		return
	}
	client := source.Client
	if client == nil {
		return
	}

	receiptTag, ok := rawReceiptAckTag(node)
	if !ok {
		return
	}

	to, ok := node.Attrs["from"]
	if !ok || fmt.Sprint(to) == "" {
		return
	}
	id, ok := node.Attrs["id"]
	if !ok || fmt.Sprint(id) == "" {
		return
	}

	ack := binary.Node{
		Tag: "ack",
		Attrs: binary.Attrs{
			"class": "receipt",
			"id":    id,
			"to":    to,
		},
	}

	logentry := source.GetLogger()
	if err := client.DangerousInternals().SendNode(context.Background(), ack); err != nil {
		logentry.Errorf("[CALL] Raw receipt ack send failed: child=%s id=%v err=%v", receiptTag, id, err)
		return
	}
	logentry.Infof("[CALL] Raw receipt ack sent: child=%s id=%v", receiptTag, id)
}
