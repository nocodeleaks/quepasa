package whatsmeow

import (
	"encoding/json"
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
	"strings"

	"go.mau.fi/whatsmeow/binary"
)

// CompactEndpoint is the 6-byte endpoint shape now confirmed in Desktop call stanzas:
// IPv4 (4 bytes) + port (2 bytes, big-endian).
type CompactEndpoint struct {
	RawHex   string `json:"raw_hex,omitempty"`
	IP       string `json:"ip,omitempty"`
	Port     int    `json:"port,omitempty"`
	Endpoint string `json:"endpoint,omitempty"`
}

func decodeCompactEndpoint6Bytes(raw []byte) *CompactEndpoint {
	if len(raw) != 6 {
		return nil
	}
	ip := net.IPv4(raw[0], raw[1], raw[2], raw[3]).String()
	port := int(raw[4])<<8 | int(raw[5])
	return &CompactEndpoint{
		RawHex:   strings.ToLower(hex.EncodeToString(raw)),
		IP:       ip,
		Port:     port,
		Endpoint: fmt.Sprintf("%s:%d", ip, port),
	}
}

func decodeCompactEndpoint6String(s string) *CompactEndpoint {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	if len(s) == 12 && isHexString(s) {
		if raw, err := hex.DecodeString(s); err == nil {
			return decodeCompactEndpoint6Bytes(raw)
		}
	}
	if len(s) == 6 {
		return decodeCompactEndpoint6Bytes([]byte(s))
	}
	return nil
}

func encodeCompactEndpoint6(ip string, port int) *CompactEndpoint {
	parsed := net.ParseIP(strings.TrimSpace(ip))
	if parsed == nil {
		return nil
	}
	ip4 := parsed.To4()
	if ip4 == nil || port < 0 || port > 65535 {
		return nil
	}
	raw := []byte{ip4[0], ip4[1], ip4[2], ip4[3], byte(port >> 8), byte(port)}
	return decodeCompactEndpoint6Bytes(raw)
}

func isHexString(s string) bool {
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}

type transportCompactItem struct {
	Tag      string           `json:"tag"`
	Priority string           `json:"priority,omitempty"`
	Compact  *CompactEndpoint `json:"compact,omitempty"`
}

func extractCompactTransportItems(data *TransportDataNode) []transportCompactItem {
	if data == nil {
		return nil
	}
	var out []transportCompactItem
	var walk func(nodes []RawNode)
	walk = func(nodes []RawNode) {
		for _, node := range nodes {
			tag := strings.ToLower(strings.TrimSpace(node.Tag))
			switch tag {
			case "te", "rte":
				var s string
				if err := jsonUnmarshalString(node.Content, &s); err == nil {
					if ep := decodeCompactEndpoint6String(s); ep != nil {
						out = append(out, transportCompactItem{
							Tag:      tag,
							Priority: strings.TrimSpace(node.Attrs["priority"]),
							Compact:  ep,
						})
					}
				}
			}
			var children []RawNode
			if err := jsonUnmarshalRawNodes(node.Content, &children); err == nil && len(children) > 0 {
				walk(children)
			}
		}
	}
	walk(data.Content)
	return out
}

type transportSentCompactCandidate struct {
	ID       string           `json:"id,omitempty"`
	Type     string           `json:"type,omitempty"`
	IP       string           `json:"ip,omitempty"`
	Port     string           `json:"port,omitempty"`
	Priority string           `json:"priority,omitempty"`
	RelAddr  string           `json:"rel_addr,omitempty"`
	RelPort  string           `json:"rel_port,omitempty"`
	Compact  *CompactEndpoint `json:"compact,omitempty"`
}

func extractCompactTransportItemsFromBinaryNode(node binary.Node) []transportCompactItem {
	var out []transportCompactItem
	var walk func(nodes []binary.Node)
	walk = func(nodes []binary.Node) {
		for _, n := range nodes {
			tag := strings.ToLower(strings.TrimSpace(n.Tag))
			switch tag {
			case "te", "rte":
				switch content := n.Content.(type) {
				case []byte:
					if ep := decodeCompactEndpoint6Bytes(content); ep != nil {
						out = append(out, transportCompactItem{
							Tag:      tag,
							Priority: strings.TrimSpace(fmt.Sprint(n.Attrs["priority"])),
							Compact:  ep,
						})
					}
				case string:
					if ep := decodeCompactEndpoint6String(content); ep != nil {
						out = append(out, transportCompactItem{
							Tag:      tag,
							Priority: strings.TrimSpace(fmt.Sprint(n.Attrs["priority"])),
							Compact:  ep,
						})
					}
				}
			}
			if children, ok := n.Content.([]binary.Node); ok && len(children) > 0 {
				walk(children)
			}
		}
	}
	if children, ok := node.Content.([]binary.Node); ok && len(children) > 0 {
		walk(children)
	}
	return out
}

func extractCompactTransportCandidates(node binary.Node) []transportSentCompactCandidate {
	var out []transportSentCompactCandidate
	var walk func(nodes []binary.Node)
	walk = func(nodes []binary.Node) {
		for _, n := range nodes {
			if strings.EqualFold(strings.TrimSpace(n.Tag), "candidate") {
				ip := fmt.Sprint(n.Attrs["ip"])
				portRaw := fmt.Sprint(n.Attrs["port"])
				port, _ := strconv.Atoi(strings.TrimSpace(portRaw))
				out = append(out, transportSentCompactCandidate{
					ID:       fmt.Sprint(n.Attrs["id"]),
					Type:     fmt.Sprint(n.Attrs["type"]),
					IP:       ip,
					Port:     portRaw,
					Priority: fmt.Sprint(n.Attrs["priority"]),
					RelAddr:  fmt.Sprint(n.Attrs["rel-addr"]),
					RelPort:  fmt.Sprint(n.Attrs["rel-port"]),
					Compact:  encodeCompactEndpoint6(ip, port),
				})
			}
			if children, ok := n.Content.([]binary.Node); ok && len(children) > 0 {
				walk(children)
			}
		}
	}
	if children, ok := node.Content.([]binary.Node); ok && len(children) > 0 {
		walk(children)
	}
	return out
}

func jsonUnmarshalString(raw []byte, out *string) error {
	return json.Unmarshal(raw, out)
}

func jsonUnmarshalRawNodes(raw []byte, out *[]RawNode) error {
	return json.Unmarshal(raw, out)
}
