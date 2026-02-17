package sipproxy

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
)

// sendRTPProbe sends a short burst of minimal RTP packets from an already-bound local RTP socket
// to the remote RTP endpoint negotiated via SDP (usually from the 200 OK answer).
//
// Purpose: help Asterisk Strict RTP / NAT learning and confirm routing.
func sendRTPProbe(logger *log.Entry, conn *net.UDPConn, remoteIP string, remotePort int, callID string) {
	if logger == nil {
		logger = log.WithField("module", "sipproxy")
	}
	if conn == nil {
		return
	}
	if remoteIP == "" || remotePort <= 0 {
		return
	}

	raddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", remoteIP, remotePort))
	if err != nil {
		logger.Warnf("🎵 [RTP-PROBE] Resolve remote failed: %v (CallID=%s)", err, callID)
		return
	}

	// Minimal RTP header (12 bytes)
	// V=2, P=0, X=0, CC=0
	// M=0, PT=0 (PCMU)
	var seq uint16 = uint16(time.Now().UnixNano())
	ts := uint32(time.Now().UnixNano() / int64(time.Millisecond))
	ssrc := uint32(time.Now().UnixNano())

	pkt := make([]byte, 12)
	pkt[0] = 0x80
	pkt[1] = 0x00

	logger.Infof("🎵🧪 [RTP-PROBE] Sending probe burst to %s (CallID=%s)", raddr.String(), callID)

	for i := 0; i < 10; i++ {
		binary.BigEndian.PutUint16(pkt[2:4], seq)
		binary.BigEndian.PutUint32(pkt[4:8], ts)
		binary.BigEndian.PutUint32(pkt[8:12], ssrc)

		if _, err := conn.WriteToUDP(pkt, raddr); err != nil {
			logger.Warnf("🎵🧪 [RTP-PROBE] Write failed: %v (CallID=%s)", err, callID)
			return
		}

		seq++
		ts += 160 // 20ms at 8kHz
		time.Sleep(20 * time.Millisecond)
	}

	logger.Infof("🎵🧪 [RTP-PROBE] Probe burst sent (CallID=%s)", callID)
}
