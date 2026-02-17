package sipproxy

import (
	"strconv"
	"strings"
)

// parseSDPAudioEndpoint extracts the remote RTP endpoint from an SDP body.
// It returns the first global connection address (c=) and the first audio media port (m=audio).
// This is best-effort and intentionally simple.
func parseSDPAudioEndpoint(sdp []byte) (ip string, port int) {
	if len(sdp) == 0 {
		return "", 0
	}

	lines := strings.Split(string(sdp), "\n")
	for _, raw := range lines {
		line := strings.TrimSpace(strings.TrimSuffix(raw, "\r"))
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "c=") {
			// c=IN IP4 1.2.3.4
			parts := strings.Fields(strings.TrimPrefix(line, "c="))
			if len(parts) >= 3 {
				ip = strings.TrimSpace(parts[2])
			}
			continue
		}

		if strings.HasPrefix(line, "m=audio") {
			// m=audio 10250 RTP/AVP 0 8 101
			parts := strings.Fields(strings.TrimPrefix(line, "m="))
			if len(parts) >= 2 {
				if p, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
					port = p
				}
			}
			// We can stop early once we have both.
			if ip != "" && port > 0 {
				return ip, port
			}
		}
	}

	return ip, port
}
