package whatsmeow

import (
	"context"
	"fmt"

	"go.mau.fi/whatsmeow/binary"
)

func rawCallAckChild(node *binary.Node) (binary.Node, string, bool) {
	if node == nil || node.Tag != "call" {
		return binary.Node{}, "", false
	}
	children := node.GetChildren()
	if len(children) == 0 {
		return binary.Node{}, "", false
	}

	child := children[0]
	switch child.Tag {
	case "relaylatency", "transport", "terminate", "mute_v2":
	default:
		return binary.Node{}, "", false
	}

	if child.Tag != "relaylatency" {
		return binary.Node{}, child.Tag, true
	}

	attrs := binary.Attrs{}
	if v, ok := child.Attrs["call-creator"]; ok {
		attrs["call-creator"] = v
	}
	if v, ok := child.Attrs["call-id"]; ok {
		attrs["call-id"] = v
	}

	return binary.Node{
		Tag:   child.Tag,
		Attrs: attrs,
	}, child.Tag, true
}

func (source *WhatsmeowHandlers) maybeSendRawCallAck(node *binary.Node) {
	if source == nil || node == nil {
		return
	}
	client := source.Client
	if client == nil {
		return
	}

	child, ackType, ok := rawCallAckChild(node)
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
			"class": "call",
			"id":    id,
			"type":  ackType,
		},
	}
	if ackType == "relaylatency" {
		ack.Attrs["from"] = to
	} else {
		ack.Attrs["to"] = to
	}
	if child.Tag != "" {
		ack.Content = []binary.Node{child}
	}

	logentry := source.GetLogger()
	if err := client.DangerousInternals().SendNode(context.Background(), ack); err != nil {
		logentry.Errorf("[CALL] Raw call ack send failed: type=%s id=%v err=%v", ackType, id, err)
		return
	}
	logentry.Infof("[CALL] Raw call ack sent: type=%s id=%v", ackType, id)
}
