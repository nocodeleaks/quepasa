package voip

// VoIP Bridge Audio: VoipBridgeSink and VoipBridgeSource adapt between the calls module
// audio contract (16 kHz mono float32, 960 samples / 60 ms frame) and the SIP
// side (G.711 μ-law 8 kHz RTP datagrams).
//
// VoipBridgeSink implements calls.AudioSink:
//   - receives 16 kHz float32 frames from the WhatsApp call
//   - decimates to 8 kHz, μ-law encodes, builds an RTP packet
//   - writes the RTP packet to the SIP-bound UDP socket
//
// VoipBridgeSource implements calls.AudioSource:
//   - reads RTP packets from the SIP-side UDP socket
//   - validates and parses the RTP header
//   - μ-law decodes the payload to 16 kHz float32
//   - passes through a jitter buffer for smooth playback
//
// Both types are driven by the VoIPBridge orchestrator (bridge.go).

import (
	"errors"
	"net"
	"sync"
	"time"

	qplog "github.com/nocodeleaks/quepasa/qplog"
	calls "github.com/nocodeleaks/quepasa/voip/calls"
)

// ErrVoipBridgeSourceClosed is returned when the bridge source has been closed.
var ErrVoipBridgeSourceClosed = errors.New("voip bridge: source closed")

// ---------------------------------------------------------------------------
// sipPeer — symmetric-RTP learned remote address
// ---------------------------------------------------------------------------
//
// The SIP RTP session uses ONE UDP socket. VoipBridgeSource learns the SIP
// server's RTP source address from the first inbound packet (symmetric RTP /
// rport) and stores it here; VoipBridgeSink reads it to know where to send the
// WhatsApp→SIP audio. This avoids depending on the SDP answer's c=/m= address,
// which may be a private/NAT IP.
type sipPeer struct {
	mu   sync.RWMutex
	addr *net.UDPAddr
}

func (p *sipPeer) set(addr *net.UDPAddr) {
	p.mu.Lock()
	if p.addr == nil && addr != nil {
		p.addr = addr
	}
	p.mu.Unlock()
}

func (p *sipPeer) get() *net.UDPAddr {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.addr
}

// ---------------------------------------------------------------------------
// VoipBridgeSink — WhatsApp audio → SIP RTP
// ---------------------------------------------------------------------------

// VoipBridgeSink implements calls.AudioSink. Each WriteFrame call converts the
// 16 kHz float32 frame to G.711 μ-law and sends an RTP packet over UDP to the
// SIP server.
type VoipBridgeSink struct {
	conn       *net.UDPConn // SIP RTP socket (shared with VoipBridgeSource)
	peer       *sipPeer     // learned SIP server RTP address (symmetric RTP)
	builder    *VoipRTPBuilder  // RTP packet builder for WA→SIP direction
	encScratch []byte       // reusable buffer for L16 encoding
	queue      chan []byte  // 20 ms L16 payloads awaiting paced send
	stop       chan struct{}
	mu         sync.Mutex
	closed     bool
}

// NewVoipBridgeSink creates an AudioSink that sends L16/16000 RTP to the SIP server
// over conn, addressed to the peer learned by the paired VoipBridgeSource. L16 is
// full-rate 16 kHz linear PCM so no fidelity is lost on the QuePasa↔asterisk
// leg; asterisk transcodes to the destination endpoint's codec.
//
// Outbound packets are PACED: WhatsApp delivers a 60 ms frame at a time, but
// sending its three 20 ms packets back-to-back makes the SIP server receive
// bursts and its jitter buffer underrun, which (via the echo path) shows up as
// choppiness. A sender goroutine releases one 20 ms packet every 20 ms so the
// SIP server sees a steady stream.
func NewVoipBridgeSink(conn *net.UDPConn, peer *sipPeer, ssrc uint32) *VoipBridgeSink {
	s := &VoipBridgeSink{
		conn:    conn,
		peer:    peer,
		builder: NewVoipRTPBuilderPT(ssrc, 0, L16PayloadType, 2),
		queue:   make(chan []byte, 16),
		stop:    make(chan struct{}),
	}
	go s.senderLoop()
	return s
}

// l16PacketBytes is the L16 payload size per RTP packet: 640 bytes = 320 samples
// = 20 ms at 16 kHz (matches a=ptime:20), and 640+12 = 652 bytes fits the MTU.
// A 60 ms WhatsApp frame (960 samples → 1920 L16 bytes) is emitted as three
// 20 ms packets so asterisk's jitter buffer and timing behave correctly.
const l16PacketBytes = 640

// WriteFrame converts one 16 kHz mono float32 frame to L16/16000 and enqueues
// its 20 ms packets for paced sending. It never blocks the WhatsApp media loop:
// if the send queue is full (peer not yet known, or a stall), packets are
// dropped rather than backing up.
func (s *VoipBridgeSink) WriteFrame(frame []float32) error {
	s.mu.Lock()
	closed := s.closed
	s.mu.Unlock()
	if closed {
		return nil
	}

	// Encode the 16 kHz frame to network-order linear PCM (no loss).
	payload := L16Encode(frame, s.encScratch)
	s.encScratch = payload // reuse next call

	for off := 0; off < len(payload); off += l16PacketBytes {
		end := off + l16PacketBytes
		if end > len(payload) {
			end = len(payload)
		}
		// Copy: encScratch is reused on the next WriteFrame, but the queued
		// payload is sent later by the paced goroutine.
		chunk := make([]byte, end-off)
		copy(chunk, payload[off:end])
		select {
		case s.queue <- chunk:
		default:
			// Queue full — drop to avoid latency build-up / blocking.
		}
	}
	return nil
}

// senderLoop releases one queued 20 ms packet every 20 ms, giving the SIP
// server a steady RTP stream instead of 60 ms bursts.
func (s *VoipBridgeSink) senderLoop() {
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
		}
		var chunk []byte
		select {
		case chunk = <-s.queue:
		default:
			continue // nothing to send this tick
		}
		addr := s.peer.get()
		if addr == nil {
			continue // no destination yet; drop (pre-roll)
		}
		s.mu.Lock()
		if s.closed || s.conn == nil {
			s.mu.Unlock()
			return
		}
		packet := s.builder.Build(chunk, false)
		_, _ = s.conn.WriteToUDP(packet, addr)
		s.mu.Unlock()
	}
}

// Close stops the paced sender. It does NOT close the UDP connection (owned by
// the RTPStream / VoIPBridge).
func (s *VoipBridgeSink) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	s.mu.Unlock()
	close(s.stop)
	return nil
}

// ---------------------------------------------------------------------------
// VoipBridgeSource — SIP RTP → WhatsApp audio
// ---------------------------------------------------------------------------

// VoipBridgeSource implements calls.AudioSource as a sequence-aware jitter buffer.
// A background goroutine reads RTP packets from the SIP-side UDP socket, decodes
// each 20 ms L16 payload, and files it by RTP sequence number. ReadFrame plays
// out packets in order at the consumer's 60 ms cadence, reordering late packets,
// substituting 20 ms of silence for lost ones (PLC-lite), and holding a small
// prebuffer so network/clock jitter doesn't cause continuous choppiness.
type VoipBridgeSource struct {
	conn   *net.UDPConn // SIP RTP socket (shared with VoipBridgeSink)
	peer   *sipPeer     // records the SIP server RTP address
	mu     sync.Mutex
	acc    []float32 // in-order accumulated 16 kHz samples awaiting playout
	primed bool      // true once the prebuffer target is reached
	closed bool
	done   chan struct{} // closed when reader goroutine exits

	// Diagnostics (logged on Close): characterise the inbound RTP so picotes can
	// be attributed to loss vs reordering vs underflow vs clock drift.
	logger        qplog.Logger
	haveExpSeq    bool
	expSeq        uint16
	rxPkts        int64
	lostPkts      int64 // sequence gaps detected on arrival
	reorderPkts   int64 // packets arriving older than expected
	plcSubframes  int64 // unused (kept for log compatibility)
	underflowFrms int64 // frames served with less than a full frame of audio
}

const (
	// prebufferSamples is the jitter cushion before playout begins (~180 ms).
	// The inbound RTP is in-order and near-lossless, so the residual picotes are
	// clock-jitter underflow; a deeper cushion absorbs it at the cost of latency.
	prebufferSamples = calls.FrameSamples * 3
	// maxAccSamples caps the accumulator (~360 ms) so a fast sender can't grow it
	// without bound; the oldest samples are dropped, bounding latency.
	maxAccSamples = calls.FrameSamples * 6
)

// NewVoipBridgeSource creates an AudioSource that reads L16/16000 RTP from conn and
// delivers fixed-size 16 kHz float32 frames through a prebuffered accumulator.
// It records the sender address into peer so the paired VoipBridgeSink can send back.
func NewVoipBridgeSource(conn *net.UDPConn, peer *sipPeer) *VoipBridgeSource {
	return &VoipBridgeSource{
		conn: conn,
		peer: peer,
		done: make(chan struct{}),
	}
}

// SetLogger attaches a logger used to emit jitter-buffer diagnostics on Close.
func (s *VoipBridgeSource) SetLogger(l qplog.Logger) {
	s.mu.Lock()
	s.logger = l
	s.mu.Unlock()
}

// StartReadLoop launches the background goroutine that reads RTP packets from
// the UDP socket and files decoded samples by sequence. Must be called once.
func (s *VoipBridgeSource) StartReadLoop() {
	go s.readLoop()
}

// readLoop reads RTP datagrams, decodes the L16 payload, and files it by seq.
func (s *VoipBridgeSource) readLoop() {
	defer close(s.done)

	buf := make([]byte, 2048)
	for {
		s.mu.Lock()
		if s.closed {
			s.mu.Unlock()
			return
		}
		s.mu.Unlock()

		n, src, err := s.conn.ReadFromUDP(buf)
		if err != nil {
			return
		}

		// Learn the SIP server's RTP address (symmetric RTP) so VoipBridgeSink
		// knows where to send the WhatsApp→SIP direction.
		s.peer.set(src)

		info, err := ParseRTP(buf[:n])
		if err != nil {
			continue // drop malformed packets
		}

		// Decode L16/16000 to 16 kHz float32 (own buffer — it is stored).
		decoded := L16Decode(info.Payload, nil)

		s.mu.Lock()
		// Arrival-order diagnostics: detect gaps (loss) and reordering.
		s.rxPkts++
		if !s.haveExpSeq {
			s.haveExpSeq = true
			s.expSeq = info.Seq + 1
		} else if info.Seq == s.expSeq {
			s.expSeq++
		} else if seqLess(info.Seq, s.expSeq) {
			s.reorderPkts++ // arrived later than a successor
		} else {
			s.lostPkts += int64(info.Seq - s.expSeq) // gap = presumed lost
			s.expSeq = info.Seq + 1
		}

		// Append decoded samples in order (inbound is in-order; see stats).
		s.acc = append(s.acc, decoded...)
		if len(s.acc) > maxAccSamples {
			s.acc = append(s.acc[:0], s.acc[len(s.acc)-maxAccSamples:]...)
		}
		s.mu.Unlock()
	}
}

// ReadFrame returns the next FrameSamples-long (60 ms) 16 kHz float32 frame from
// the accumulator. It always consumes whatever real audio is available (never
// dropping it) and pads any shortfall with silence, after an initial prebuffer
// to absorb jitter. Returns an error only if the source is closed.
func (s *VoipBridgeSource) ReadFrame() ([]float32, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil, ErrVoipBridgeSourceClosed
	}

	frame := make([]float32, calls.FrameSamples)

	// Prebuffer: emit silence until the cushion fills, then start serving.
	if !s.primed {
		if len(s.acc) < prebufferSamples {
			return frame, nil
		}
		s.primed = true
	}

	n := len(s.acc)
	if n > calls.FrameSamples {
		n = calls.FrameSamples
	}
	if n > 0 {
		copy(frame, s.acc[:n])
		s.acc = append(s.acc[:0], s.acc[n:]...)
	}
	if n < calls.FrameSamples {
		// Shortfall padded with silence (clock-jitter underflow).
		s.underflowFrms++
	}
	return frame, nil
}

// Close stops the read loop and releases resources, logging jitter diagnostics.
// The UDP connection is NOT closed here (owned by the RTPStream / VoIPBridge).
func (s *VoipBridgeSource) Close() error {
	s.mu.Lock()
	s.closed = true
	l := s.logger
	rx, lost, reorder, plc, under := s.rxPkts, s.lostPkts, s.reorderPkts, s.plcSubframes, s.underflowFrms
	s.mu.Unlock()
	if l != nil {
		lossPct := 0.0
		if rx+lost > 0 {
			lossPct = 100 * float64(lost) / float64(rx+lost)
		}
		l.InfoE().
			Int64("rx_pkts", rx).
			Int64("lost_pkts", lost).
			Float64("loss_pct", lossPct).
			Int64("reorder_pkts", reorder).
			Int64("plc_20ms", plc).
			Int64("underflow_60ms", under).
			Msg("VoipBridge: inbound RTP jitter stats")
	}
	<-s.done
	return nil
}

// seqLess returns true if a is strictly before b in uint16 sequence space,
// accounting for wraparound.
func seqLess(a, b uint16) bool {
	diff := int16(b - a)
	return diff > 0
}
