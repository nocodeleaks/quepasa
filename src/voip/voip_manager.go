package voip

// VoIP VoipManager: orchestrates the full inbound WhatsApp-call → SIP bridge.
//
// The manager ties together three layers:
//
//   1. calls.Client (src/calls)   — native WhatsApp call signaling, SRTP/RTP,
//                                    MLOW codec, and 16 kHz mono float32 audio.
//   2. sipproxy.SIPProxyManager   — SIP proxy that sends INVITE to a configured
//                                    SIP server and exposes an RTP stream.
//   3. VoipBridgeSink / VoipBridgeSource  — audio adapters between the calls module
//                                    audio contract (16 kHz float32) and the
//                                    SIP side (G.711 μ-law 8 kHz RTP).
//
// Flow (inbound WhatsApp call):
//
//   whatsmeow fires CallOffer
//     → CallMessage checks source.voip != nil
//     → VoIPManager.OnIncomingCall is invoked by calls.Client
//     → VoipManager answers the WhatsApp call natively
//     → VoipManager calls sipproxy.BridgeInboundWhatsAppCall
//     → VoipManager calls sipproxy.SendSIPInvite (INVITE → SIP server)
//     → On call.OnReady (media path up):
//         - VoipBridgeSource reads μ-law RTP from SIP side → feeds calls AudioSource
//         - VoipBridgeSink receives 16 kHz float32 → encodes μ-law → sends to SIP side

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/emiago/sipgo/sip"
	environment "github.com/nocodeleaks/quepasa/environment"
	qplog "github.com/nocodeleaks/quepasa/qplog"
	sipproxy "github.com/nocodeleaks/quepasa/sipproxy"
	calls "github.com/nocodeleaks/quepasa/voip/calls"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	"go.mau.fi/whatsmeow"
	types "go.mau.fi/whatsmeow/types"
)

// EnvVoIPEnabled is the opt-in switch for the VoIP VoipManager.
// When set to "true" or "1", inbound WhatsApp calls are answered natively
// and forwarded to the configured SIP server via sipproxy.
// When unset/empty, the manager stays disabled and QuePasa keeps its default
// behavior (reject/relay in CallMessage).
const EnvVoIPEnabled = "VOIP_ENABLED"

// VoipManager wraps a connected whatsmeow client, a calls.Client, and the
// sipproxy.SIPProxyManager to provide full WhatsApp → SIP call bridging.
//
// Lifecycle:
//   - Construct with NewVoipManager(client) right after the whatsmeow client is
//     created (and BEFORE it connects, so the low-level <ack>/<call>
//     interception is in place).
//   - Call Enable() to activate the inbound-call listener.
//   - Keep the *VoipManager alive (store it on the connection) so its
//     OnIncomingCall closure is not garbage-collected.
type VoipManager struct {
	mc      *calls.Client
	log     qplog.Logger
	proxy   *sipproxy.SIPProxyManager
	mu      sync.Mutex
	enabled bool

	// mode is the per-instance VoIP behavior (exclusive or additional).
	// It controls whether QuePasa answers the WhatsApp call (exclusive) or
	// leaves it ringing on other devices (additional), and whether a SIP
	// failure should hang up the WhatsApp call.
	mode whatsapp.VoIPMode

	// sipAcceptTimeout bounds how long the manager waits for the SIP server to
	// answer (200 OK) before considering the call failed (exclusive mode).
	sipAcceptTimeout time.Duration

	// activeCalls maps call-id → *bridgeContext for tear-down.
	activeCalls sync.Map // map[string]*bridgeContext

	// sectionID is an opaque QuePasa section/session identifier propagated to
	// SIP gateways so they can route calls by WhatsApp section.
	sectionID string
}

// bridgeContext holds per-call resources for cleanup on hangup.
type bridgeContext struct {
	callID     string
	sinkConn   *net.UDPConn      // SIP-bound UDP socket
	sourceConn *net.UDPConn      // WhatsApp-side UDP socket
	source     *VoipBridgeSource // jitter-buffer reader (Close logs diagnostics)
	sink       *VoipBridgeSink   // paced sender (Close stops its goroutine)
}

// NewVoipManager builds the VoIP VoipManager around an already-connected
// *whatsmeow.Client. The calls.Client is created with a no-op logger unless
// one is provided via NewVoipManagerWithLogger.
func NewVoipManager(client *whatsmeow.Client) *VoipManager {
	return NewVoipManagerWithLogger(client, qplog.Nop())
}

// NewVoipManagerWithLogger is NewVoipManager with an explicit logger so the
// vendored calls stack's debug/trace can be surfaced during validation.
func NewVoipManagerWithLogger(client *whatsmeow.Client, log qplog.Logger) *VoipManager {
	return &VoipManager{
		mc:  calls.NewClient(client, calls.WithLogger(log)),
		log: log,
	}
}

// Enable registers an inbound-call listener that auto-answers every 1:1
// WhatsApp call and bridges it to the SIP server via sipproxy.
//
// It must be called once; subsequent calls are no-ops.
func (m *VoipManager) Enable() error {
	if m == nil || m.mc == nil {
		return fmt.Errorf("voip manager: nil client")
	}

	m.mu.Lock()
	already := m.enabled
	m.enabled = true
	m.mu.Unlock()

	if already {
		return nil
	}

	m.mc.OnIncomingCall(func(call *calls.Call) {
		m.handleIncoming(call)
	})

	m.log.InfoE().Msg("VoipManager: inbound call listener registered")
	return nil
}

// handleIncoming processes a single inbound WhatsApp call:
//  1. Answer the call natively via calls.Client.
//  2. Create an RTP stream via sipproxy.BridgeInboundWhatsAppCall.
//  3. Send SIP INVITE to the SIP server.
//  4. On media-ready, wire VoipBridgeSink/VoipBridgeSource for bidirectional audio.
func (m *VoipManager) handleIncoming(call *calls.Call) {
	callID := call.ID()
	peer := call.Peer()

	m.log.InfoE().
		Str("call_id", callID).
		Str("peer", peer.String()).
		Msg("VoipManager: handleIncoming — incoming call")

	// Extract phone numbers for SIP routing.
	fromPhone := jidToPhone(peer)
	toPhone := m.getOwnPhone()

	m.log.InfoE().
		Str("call_id", callID).
		Str("from_phone", fromPhone).
		Str("to_phone", toPhone).
		Msg("VoipManager: call routing info")

	// Register OnEnd for cleanup (applies to all modes).
	call.OnEnd(func(reason string) {
		m.log.InfoE().
			Str("call_id", callID).
			Str("reason", reason).
			Msg("VoipManager: call ended, cleaning up")
		m.cleanupCall(callID)
	})

	switch m.mode {
	case whatsapp.VoIPModeAdditional:
		m.handleAdditional(call, callID, fromPhone, toPhone)
	default:
		// VoIPModeExclusive (and any active fallback) behaves exclusively.
		m.handleExclusive(call, callID, fromPhone, toPhone)
	}
}

// handleExclusive answers the WhatsApp call natively and bridges it to SIP.
// SIP is the only endpoint. If the SIP server rejects the INVITE or never
// answers within sipAcceptTimeout, the WhatsApp call is hung up so it is NOT
// left marked as "answered".
func (m *VoipManager) handleExclusive(call *calls.Call, callID, fromPhone, toPhone string) {
	// Register OnReady BEFORE answering — OnReady fires on the first inbound
	// RTP frame and may arrive immediately after Answer() returns.
	call.OnReady(func() {
		m.log.InfoE().
			Str("call_id", callID).
			Msg("VoipManager[exclusive]: media path ready, wiring SIP bridge")

		if err := m.wireSIPBridge(call, callID, fromPhone, toPhone); err != nil {
			m.log.ErrorE().Err(err).
				Str("call_id", callID).
				Msg("VoipManager[exclusive]: failed to wire SIP bridge")
			// Attempt graceful hangup since the bridge failed.
			_ = call.Hangup()
			return
		}

		// Guard: ensure the SIP server actually accepts the call. If it
		// rejects or times out, hang up the WhatsApp call so it is not left
		// marked as answered on the phone.
		m.guardSIPAcceptance(call, callID)
	})

	m.log.InfoE().
		Str("call_id", callID).
		Msg("VoipManager[exclusive]: answering call")
	if err := call.Answer(); err != nil {
		m.log.ErrorE().Err(err).
			Str("call_id", callID).
			Msg("VoipManager[exclusive]: call.Answer() failed")
		return
	}

	m.log.InfoE().
		Str("call_id", callID).
		Msg("VoipManager[exclusive]: call answered, awaiting media-ready")
}

// handleAdditional forwards the call to SIP as an extra device WITHOUT
// answering or rejecting the native WhatsApp call. This leaves the call
// ringing on the user's other paired WhatsApp devices, so a human can pick it
// up on a phone while the SIP endpoint also rings.
//
// Because we never call Answer(), the native media path (OnReady) is not
// established by QuePasa; we only fire the SIP INVITE so the SIP side rings.
func (m *VoipManager) handleAdditional(call *calls.Call, callID, fromPhone, toPhone string) {
	m.log.InfoE().
		Str("call_id", callID).
		Msg("VoipManager[additional]: forwarding to SIP without answering (call keeps ringing on other devices)")

	if m.proxy == nil {
		m.log.WarnE().
			Str("call_id", callID).
			Msg("VoipManager[additional]: sipproxy not configured, leaving call to native devices")
		return
	}

	// Send the SIP INVITE so the SIP endpoint rings in parallel. We do NOT
	// answer the WhatsApp call: no native media bridge is established here.
	if err := m.proxy.SendSIPInviteWithHeaders(callID, fromPhone, toPhone, m.sipHeaders()); err != nil {
		m.log.ErrorE().Err(err).
			Str("call_id", callID).
			Msg("VoipManager[additional]: SendSIPInvite failed; call still rings on native devices")
		return
	}

	m.log.InfoE().
		Str("call_id", callID).
		Str("from", fromPhone).
		Str("to", toPhone).
		Msg("VoipManager[additional]: SIP INVITE sent; native call left ringing")
}

// guardSIPAcceptance starts a watchdog that hangs up the WhatsApp call if the
// SIP server does not accept (200 OK) within sipAcceptTimeout. It relies on
// the proxy's accepted/rejected callbacks to short-circuit the timer.
//
// In exclusive mode this prevents the WhatsApp call from staying "answered"
// when the SIP side rejects (e.g. 401/4xx) or never responds.
func (m *VoipManager) guardSIPAcceptance(call *calls.Call, callID string) {
	timeout := m.sipAcceptTimeout
	if timeout <= 0 {
		timeout = 15 * time.Second
	}

	accepted := make(chan struct{}, 1)
	rejected := make(chan struct{}, 1)

	// Wire transient handlers scoped to this call id. The proxy invokes them
	// when it sees the SIP final response for this INVITE.
	// Use per-call handlers to avoid being overwritten by global handlers.
	if m.proxy != nil {
		// Get the sipgo call manager directly for per-call handler registration
		if sipgoMgr := m.proxy.GetSipgoCallManager(); sipgoMgr != nil {
			sipgoMgr.RegisterCallAcceptedHandler(callID, func(id, _, _ string, _ *sip.Response) {
				if id == callID {
					select {
					case accepted <- struct{}{}:
					default:
					}
					// Clear handlers after acceptance
					sipgoMgr.ClearCallHandlers(callID)
				}
			})
			sipgoMgr.RegisterCallRejectedHandler(callID, func(id, _, _ string, _ *sip.Response) {
				if id == callID {
					select {
					case rejected <- struct{}{}:
					default:
					}
					// Clear handlers after rejection
					sipgoMgr.ClearCallHandlers(callID)
				}
			})
			m.log.InfoE().Str("call_id", callID).Msg("VoipManager[exclusive]: per-call handlers registered for SIP acceptance monitoring")
		} else {
			m.log.WarnE().Str("call_id", callID).Msg("VoipManager[exclusive]: sipgo call manager not available, using global handlers (may be overwritten)")
			// Fallback to global handlers (may be overwritten by other calls)
			m.proxy.SetCallAcceptedHandler(func(id, _, _ string, _ *sip.Response) {
				if id == callID {
					select {
					case accepted <- struct{}{}:
					default:
					}
				}
			})
			m.proxy.SetCallRejectedHandler(func(id, _, _ string, _ *sip.Response) {
				if id == callID {
					select {
					case rejected <- struct{}{}:
					default:
					}
				}
			})
		}
	}

	go func() {
		select {
		case <-accepted:
			m.log.InfoE().
				Str("call_id", callID).
				Msg("VoipManager[exclusive]: SIP accepted (200 OK), call established")
		case <-rejected:
			m.log.WarnE().
				Str("call_id", callID).
				Msg("VoipManager[exclusive]: SIP rejected the call, hanging up WhatsApp call")
			_ = call.Hangup()
		case <-time.After(timeout):
			m.log.WarnE().
				Str("call_id", callID).
				Dur("timeout", timeout).
				Msg("VoipManager[exclusive]: SIP did not accept in time, hanging up WhatsApp call")
			_ = call.Hangup()
		}
	}()
}

func (m *VoipManager) sipHeaders() map[string]string {
	if m == nil {
		return nil
	}

	sectionID := strings.TrimSpace(m.sectionID)
	if sectionID == "" {
		return nil
	}

	return map[string]string{
		"X-QuePasa-SectionId": sectionID,
		"X-QuePasa-Token":     sectionID,
	}
}

// wireSIPBridge creates the RTP stream, sends SIP INVITE, and wires audio.
// This is called from the call's OnReady callback once media is flowing.
func (m *VoipManager) wireSIPBridge(call *calls.Call, callID, fromPhone, toPhone string) error {
	if m.proxy == nil {
		return fmt.Errorf("voip manager: sipproxy not configured")
	}

	// --- 2. Create RTP stream via sipproxy ---
	stream, err := m.proxy.BridgeInboundWhatsAppCall(callID, fromPhone, toPhone)
	if err != nil {
		return fmt.Errorf("wireSIPBridge: BridgeInboundWhatsAppCall: %w", err)
	}

	m.log.InfoE().
		Str("call_id", callID).
		Int("wa_port", stream.WhatsAppPort).
		Int("sip_port", stream.SIPPort).
		Str("sip_remote", fmt.Sprintf("%s:%d", stream.RemoteHost, stream.RemotePort)).
		Msg("VoipManager: RTP stream created")

	// The SIP RTP session uses a SINGLE UDP socket (stream.WhatsAppConn): the
	// SDP offer advertises its port, asterisk sends its RTP there, and we send
	// the WhatsApp→SIP direction back from the same socket (symmetric RTP). The
	// SDP MUST advertise this exact port or asterisk's RTP is lost — register
	// it before sending the INVITE.
	sipConn := stream.WhatsAppConn
	if sipConn == nil {
		return fmt.Errorf("wireSIPBridge: SIP RTP connection is nil")
	}
	m.proxy.SetLocalRTPPort(callID, stream.WhatsAppPort)

	// --- 3. Send SIP INVITE to the SIP server ---
	if err := m.proxy.SendSIPInviteWithHeaders(callID, fromPhone, toPhone, m.sipHeaders()); err != nil {
		return fmt.Errorf("wireSIPBridge: SendSIPInvite: %w", err)
	}

	m.log.InfoE().
		Str("call_id", callID).
		Str("from", fromPhone).
		Str("to", toPhone).
		Msg("VoipManager: SIP INVITE sent")

	// --- 4. Wire bidirectional audio bridge over the single SIP RTP socket ---
	// peer is the SIP server's RTP address used by VoipBridgeSink. It is seeded from
	// the 200 OK SDP answer (so we send first and don't deadlock waiting for the
	// server's RTP) and also learned by VoipBridgeSource from inbound packets.
	peer := &sipPeer{}

	// Seed the sink target from the SDP answer as soon as the call is accepted.
	// The 200 OK arrives shortly after the INVITE we just sent; poll briefly.
	go func() {
		for i := 0; i < 200; i++ { // up to ~10s
			if addrStr, ok := m.proxy.GetRemoteRTPAddr(callID); ok {
				if ua, err := net.ResolveUDPAddr("udp", addrStr); err == nil {
					peer.set(ua)
					m.log.InfoE().
						Str("call_id", callID).
						Str("sip_rtp", addrStr).
						Msg("VoipManager: SIP RTP sink target set from SDP answer")
				}
				return
			}
			time.Sleep(50 * time.Millisecond)
		}
	}()

	// VoipBridgeSource: SIP (L16 RTP) → WhatsApp audio (16 kHz float32).
	bridgeSource := NewVoipBridgeSource(sipConn, peer)
	bridgeSource.SetLogger(m.log)
	bridgeSource.StartReadLoop()

	// Feed SIP audio into the call as an AudioSource via a Player.
	player := calls.NewPlayer()
	call.Subscribe(player)
	player.Play(bridgeSource)

	// VoipBridgeSink: WhatsApp audio (16 kHz float32) → SIP (μ-law RTP).
	bridgeSink := NewVoipBridgeSink(sipConn, peer, 0) // SSRC 0 — VoipRTPBuilder assigns
	call.Receive(bridgeSink)

	// Store context for cleanup.
	ctx := &bridgeContext{
		callID:     callID,
		sinkConn:   sipConn,
		sourceConn: sipConn,
		source:     bridgeSource,
		sink:       bridgeSink,
	}
	m.activeCalls.Store(callID, ctx)

	m.log.InfoE().
		Str("call_id", callID).
		Msg("VoipManager: bidirectional audio bridge wired")

	return nil
}

// cleanupCall stops the RTP stream and releases resources for a call.
func (m *VoipManager) cleanupCall(callID string) {
	val, ok := m.activeCalls.LoadAndDelete(callID)
	if !ok {
		return
	}
	ctx := val.(*bridgeContext)

	// The WhatsApp leg ended — tear down the SIP leg too by sending a BYE to the
	// SIP server. Without this the asterisk call stays up after the WhatsApp
	// caller hangs up. Safe/no-op if the SIP call was already removed.
	if m.proxy != nil {
		m.proxy.HangupCall(callID)
	}

	// Close the single SIP RTP socket owned by the RTPStream. sinkConn and
	// sourceConn point at the same socket now (symmetric RTP), so close once.
	// Close the conn FIRST so the VoipBridgeSource read loop unblocks, then Close
	// the source (which waits for that loop and logs jitter diagnostics).
	if ctx.sinkConn != nil {
		_ = ctx.sinkConn.Close()
	}
	if ctx.sourceConn != nil && ctx.sourceConn != ctx.sinkConn {
		_ = ctx.sourceConn.Close()
	}
	if ctx.sink != nil {
		_ = ctx.sink.Close()
	}
	if ctx.source != nil {
		_ = ctx.source.Close()
	}

	// Remove the call from sipproxy tracking.
	if m.proxy != nil {
		m.proxy.RemoveCall(callID)
	}

	m.log.InfoE().
		Str("call_id", callID).
		Msg("VoipManager: cleanup complete")
}

// SetSIPProxy configures the SIP proxy manager to use for call bridging.
// Must be called before Enable().
func (m *VoipManager) SetSIPProxy(proxy *sipproxy.SIPProxyManager) {
	m.mu.Lock()
	m.proxy = proxy
	m.mu.Unlock()
}

// getOwnPhone extracts the phone number of the WhatsApp account (the receiver)
// from the calls client's underlying whatsmeow store.
func (m *VoipManager) getOwnPhone() string {
	if m == nil || m.mc == nil {
		return ""
	}
	return m.mc.OwnPhone()
}

// jidToPhone extracts the phone number from a JID like "557138388109@s.whatsapp.com"
// or "557138388109.8:50@s.whatsapp.net". Returns the raw digits only.
func jidToPhone(jid types.JID) string {
	user := jid.User
	if user == "" {
		// Try splitting the full JID string
		s := jid.String()
		if idx := strings.IndexAny(s, "@."); idx > 0 {
			user = s[:idx]
		}
	}
	return user
}

// PlaceCall is a thin pass-through to place an outbound 1:1 call, exposed so
// the manager can also exercise the outbound path during validation.
func (m *VoipManager) PlaceCall(ctx context.Context, target string) (*calls.Call, error) {
	if m == nil || m.mc == nil {
		return nil, fmt.Errorf("voip manager: nil client")
	}
	return m.mc.Call(ctx, target)
}

// adaptSIPProxySettings converts environment.SIPProxySettings (loaded from
// environment variables) into sipproxy.SIPProxySettings (used by the SIP proxy
// manager). The two types live in different packages with slightly different
// field names; this function bridges them.
func adaptSIPProxySettings(env environment.SIPProxySettings) sipproxy.SIPProxySettings {
	localIP := env.PublicIP // fallback; sipproxy resolves its own if empty
	_ = localIP             // intentionally unused for now; network manager resolves

	return sipproxy.SIPProxySettings{
		ServerHost:     env.Host,
		ServerPort:     int(env.Port),
		ListenerPort:   int(env.LocalPort),
		Protocol:       env.Protocol,
		UserAgent:      env.UserAgent,
		SDPSessionName: env.SDPSessionName,
		SIPProxyNetworkManagerSettings: sipproxy.SIPProxyNetworkManagerSettings{
			StunServer: env.STUNServer,
			SIPServer:  env.Host,
			SIPPort:    int(env.Port),
			LocalPort:  int(env.LocalPort),
			PublicIP:   env.PublicIP,
		},
	}
}

// MaybeEnableManager activates the VoIP VoipManager for a connection based on the
// per-instance VoIPMode.
//
// As a per-instance setting it is OPT-IN: a disabled mode is a no-op (returns
// nil, nil) and QuePasa keeps its default reject/relay behavior. For backward
// compatibility, when mode is disabled the legacy global VOIP_ENABLED env var
// is still honored as a fallback (treated as exclusive mode).
//
// The returned *VoipManager must be kept alive by the caller (store it on the
// connection) so its OnIncomingCall closure is not garbage-collected.
func MaybeEnableManager(client *whatsmeow.Client, mode whatsapp.VoIPMode, sectionID string) (*VoipManager, error) {
	// Backward-compatibility: if no per-instance mode is set but the legacy
	// global switch is on, behave like exclusive mode.
	if !mode.IsActive() {
		if environment.NewSIPProxySettings().LegacyVoIP {
			mode = whatsapp.VoIPModeExclusive
		} else {
			return nil, nil
		}
	}
	if client == nil {
		return nil, fmt.Errorf("voip manager: nil whatsmeow client")
	}

	// Logger via the qplog facade, tagged
	// with the VoIP component and mode so call diagnostics share the app's output.
	meowLogger := qplog.New().
		WithField("component", "voip").
		WithField("voip_mode", mode.String())

	mgr := NewVoipManagerWithLogger(client, meowLogger)
	mgr.mode = mode
	mgr.sectionID = strings.TrimSpace(sectionID)
	mgr.sipAcceptTimeout = 15 * time.Second

	// Try to connect to the sipproxy manager singleton.
	// Settings come from the environment package and are converted to the
	// sipproxy package's own settings type.
	envSettings := environment.NewSIPProxySettings()
	if envSettings.Enabled && envSettings.Host != "" {
		sipSettings := adaptSIPProxySettings(envSettings)
		proxy := sipproxy.GetSIPProxyManager(sipSettings)
		if err := proxy.Start(); err != nil {
			// "already running" is not a real error — the singleton proxy was
			// already started by a previous connection and is fully functional.
			// We must still wire SetSIPProxy so this manager can use it.
			if strings.Contains(err.Error(), "já está rodando") {
				mgr.SetSIPProxy(proxy)
				mgr.log.InfoE().
					Str("sip_host", envSettings.Host).
					Uint32("sip_port", envSettings.Port).
					Msg("VoipManager: SIP proxy already running (singleton), wired to manager")
			} else {
				mgr.log.ErrorE().Err(err).Msg("VoipManager: failed to start SIP proxy")
			}
		} else {
			mgr.SetSIPProxy(proxy)
			mgr.log.InfoE().
				Str("sip_host", envSettings.Host).
				Uint32("sip_port", envSettings.Port).
				Msg("VoipManager: SIP proxy started and connected")
		}
	} else {
		mgr.log.WarnE().Msg("VoipManager: SIP server host not configured (SIPPROXY_HOST), SIP bridge will be inactive")
	}

	if err := mgr.Enable(); err != nil {
		return nil, err
	}

	mgr.log.InfoE().Msg("VoipManager: VoIP VoipManager enabled (WhatsApp → SIP bridging)")
	return mgr, nil
}

// isVoIPEnabled checks if the environment value enables the VoIP VoipManager.
func isVoIPEnabled(val string) bool {
	val = strings.ToLower(strings.TrimSpace(val))
	return val == "true" || val == "1" || val == "yes" || val == "on"
}

// Close shuts down the manager and all active calls.
func (m *VoipManager) Close() error {
	if m == nil {
		return nil
	}

	// Clean up all active calls.
	m.activeCalls.Range(func(key, _ interface{}) bool {
		m.cleanupCall(key.(string))
		return true
	})

	return nil
}
