package whatsmeow

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
)

// StartFakeRTP starts a goroutine that sends synthetic RTP-like packets to a target IP/port.
// It's used only for META-only debugging after successful call acceptance.
// Packets: 12-byte RTP header + 20 bytes dummy payload (zeroed or random).
// Sequence increments by 1; timestamp increments by 960 per packet (20ms @48k) for realism.
func StartFakeRTP(callID string, targetIP string, targetPort int, stop <-chan struct{}, logger *log.Entry) {
	if targetPort <= 0 {
		logger.Warnf("FAKE-RTP: invalid targetPort=%d, aborting", targetPort)
		return
	}
	if targetIP == "" {
		targetIP = "127.0.0.1"
	}
	addr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(targetIP, fmt.Sprintf("%d", targetPort)))
	if err != nil {
		logger.Errorf("FAKE-RTP resolve error: %v", err)
		return
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		logger.Errorf("FAKE-RTP dial error: %v", err)
		return
	}
	logger.Infof("FAKE-RTP: started → %s (CallID=%s)", addr.String(), callID)

	go func() {
		defer conn.Close()
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		var seq uint16 = uint16(rng.Intn(65535))
		var ts uint32 = rng.Uint32()
		ssrc := rng.Uint32()
		ticker := time.NewTicker(20 * time.Millisecond) // 20ms frame interval
		defer ticker.Stop()
		packetCount := 0
		for {
			select {
			case <-stop:
				logger.Infof("FAKE-RTP: stopping (CallID=%s, sent=%d packets)", callID, packetCount)
				return
			case <-ticker.C:
				buf := make([]byte, 12+20)
				// RTP header
				buf[0] = 0x80       // V=2,P=0,X=0,CC=0
				buf[1] = 111 & 0x7F // M=0, PT=111 (Opus dynamic)
				binary.BigEndian.PutUint16(buf[2:4], seq)
				binary.BigEndian.PutUint32(buf[4:8], ts)
				binary.BigEndian.PutUint32(buf[8:12], ssrc)
				// Payload (20 bytes) - random noise
				rng.Read(buf[12:])
				if _, err := conn.Write(buf); err != nil {
					logger.Errorf("FAKE-RTP write error: %v", err)
					continue
				}
				packetCount++
				seq++
				ts += 960 // 20ms @ 48kHz
				if packetCount <= 5 || packetCount%100 == 0 {
					logger.Infof("FAKE-RTP: sent packet #%d seq=%d ts=%d", packetCount, seq-1, ts-960)
				}
			}
		}
	}()
}
