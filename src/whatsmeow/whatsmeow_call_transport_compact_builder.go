package whatsmeow

import (
	"encoding/hex"
	"os"
	"strconv"
	"strings"

	"go.mau.fi/whatsmeow/binary"
)

const compactTransportRTEDesktopPortDelta = -10256

func buildCompactTransportNodes(localIP string, localPort int, publicIP string, publicPort int, includeSrflx bool) []binary.Node {
	nodes := make([]binary.Node, 0, 3)

	if ep := encodeCompactEndpoint6(localIP, localPort); ep != nil {
		raw, err := hex.DecodeString(ep.RawHex)
		if err == nil && len(raw) == 6 {
		nodes = append(nodes, binary.Node{
			Tag: "te",
			Attrs: binary.Attrs{
				"priority": "96",
			},
				Content: raw,
		})
		}
	}

	if includeSrflx {
		if ep := encodeCompactEndpoint6(publicIP, publicPort); ep != nil {
			raw, err := hex.DecodeString(ep.RawHex)
			if err == nil && len(raw) == 6 {
			nodes = append(nodes, binary.Node{
				Tag: "te",
				Attrs: binary.Attrs{
					"priority": "32",
				},
					Content: raw,
			})
			}
		}
	}

	if ep := getCompactTransportRTE(publicIP, publicPort, includeSrflx); ep != nil {
		raw, err := hex.DecodeString(ep.RawHex)
		if err == nil && len(raw) == 6 {
			nodes = append(nodes, binary.Node{
				Tag:     "rte",
				Content: raw,
			})
		}
	}

	return nodes
}

func appendCompactTransportNodes(netContent []binary.Node, localIP string, localPort int, publicIP string, publicPort int, includeSrflx bool) []binary.Node {
	compact := buildCompactTransportNodes(localIP, localPort, publicIP, publicPort, includeSrflx)
	if len(compact) == 0 {
		return netContent
	}
	out := make([]binary.Node, 0, len(netContent)+len(compact))
	out = append(out, compact...)
	out = append(out, netContent...)
	return out
}

func shouldIncludeCompactTransportNodes() bool {
	v := strings.TrimSpace(os.Getenv("QP_CALL_TRANSPORT_INCLUDE_COMPACT_TE"))
	if v == "" {
		v = "1"
	}
	v = strings.TrimSpace(strings.ToLower(v))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

func shouldIncludeCompactTransportRTE() bool {
	v := strings.TrimSpace(os.Getenv("QP_CALL_TRANSPORT_INCLUDE_COMPACT_RTE"))
	if v == "" {
		v = "1"
	}
	v = strings.TrimSpace(strings.ToLower(v))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

func getCompactTransportRTE(publicIP string, publicPort int, includeSrflx bool) *CompactEndpoint {
	if rawHex := strings.TrimSpace(os.Getenv("QP_CALL_TRANSPORT_COMPACT_RTE_HEX")); rawHex != "" {
		return decodeCompactEndpoint6String(rawHex)
	}
	if publicIP == "" {
		return nil
	}
	if port := getCompactTransportRTEPort("QP_CALL_TRANSPORT_COMPACT_RTE_PORT"); port > 0 {
		return encodeCompactEndpoint6(publicIP, port)
	}
	if delta := getCompactTransportRTEPort("QP_CALL_TRANSPORT_COMPACT_RTE_PORT_DELTA"); delta != 0 && publicPort > 0 {
		port := publicPort + delta
		if port > 0 && port <= 65535 {
			return encodeCompactEndpoint6(publicIP, port)
		}
	}
	if includeSrflx && shouldUseDesktopRTEPortDelta() && publicPort > 0 {
		port := publicPort + compactTransportRTEDesktopPortDelta
		if port > 0 && port <= 65535 {
			return encodeCompactEndpoint6(publicIP, port)
		}
	}
	if includeSrflx && shouldIncludeCompactTransportRTE() && publicPort > 0 {
		return encodeCompactEndpoint6(publicIP, publicPort)
	}
	return nil
}

func getCompactTransportRTEPort(envKey string) int {
	raw := strings.TrimSpace(os.Getenv(envKey))
	if raw == "" {
		return 0
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0
	}
	return value
}

func shouldUseDesktopRTEPortDelta() bool {
	v := strings.TrimSpace(os.Getenv("QP_CALL_TRANSPORT_COMPACT_RTE_DESKTOP_DELTA"))
	if v == "" {
		v = "1"
	}
	v = strings.TrimSpace(strings.ToLower(v))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}
