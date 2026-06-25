package calls

import (
	"sync"

	"go.mau.fi/whatsmeow/types"
)

// Call is one live 1:1 call. Place one with Client.Call, or receive one (unanswered)
// in an OnIncomingCall listener. Attach outbound audio with Subscribe/Play and inbound
// audio with Receive, and lifecycle listeners with OnReady/OnEnd/OnStateChange. All
// methods are safe for concurrent use.
type Call struct {
	eng  *engine
	id   string
	peer types.JID

	mu      sync.Mutex
	phase   CallPhase
	player  *Player
	sink    AudioSink
	onReady func()
	onEnd   func(reason string)
	onState func(CallPhase)
}

// ID returns the call-id (32 uppercase hex chars).
func (c *Call) ID() string { return c.id }

// Peer returns the remote party's LID.
func (c *Call) Peer() types.JID { return c.peer }

// State returns the call's current phase.
func (c *Call) State() CallPhase {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.phase
}

// Answer accepts an inbound call (preaccept + accept) and brings media up. No-op error
// if the call is not in a ringing state.
func (c *Call) Answer() error { return c.eng.answer(c) }

// Reject declines an inbound call.
func (c *Call) Reject() error { return c.eng.reject(c) }

// Hangup ends the call (either direction) and tears down media.
func (c *Call) Hangup() error { return c.eng.hangup(c) }

// Subscribe attaches p as the call's outbound audio player, replacing any previous one.
// While the player is Playing, its source frames are encoded and sent to the peer;
// otherwise silence is sent (the call must keep sending to hold the relay bridge).
func (c *Call) Subscribe(p *Player) {
	c.mu.Lock()
	c.player = p
	c.mu.Unlock()
}

// Play is a shortcut: it creates a Player, subscribes it, starts src, and returns the
// Player (use it for Pause/Stop/OnFinish).
func (c *Call) Play(src AudioSource) *Player {
	p := NewPlayer()
	c.Subscribe(p)
	p.Play(src)
	return p
}

// Receive attaches a sink for the peer's decoded audio (16 kHz mono frames), replacing
// any previous one. Without a sink the inbound audio is decoded and discarded.
func (c *Call) Receive(sink AudioSink) {
	c.mu.Lock()
	c.sink = sink
	c.mu.Unlock()
}

// OnReady registers a callback fired once media is flowing (relay bound, first frames
// exchanged).
func (c *Call) OnReady(fn func()) {
	c.mu.Lock()
	c.onReady = fn
	c.mu.Unlock()
}

// OnEnd registers a callback fired when the call ends, with a short reason string.
func (c *Call) OnEnd(fn func(reason string)) {
	c.mu.Lock()
	c.onEnd = fn
	c.mu.Unlock()
}

// OnStateChange registers a callback fired on each phase transition.
func (c *Call) OnStateChange(fn func(CallPhase)) {
	c.mu.Lock()
	c.onState = fn
	c.mu.Unlock()
}

// setPhase advances the call's phase and fires OnStateChange (used by the engine).
func (c *Call) setPhase(next CallPhase) {
	c.mu.Lock()
	if c.phase == next {
		c.mu.Unlock()
		return
	}
	c.phase = next
	fn := c.onState
	c.mu.Unlock()
	if fn != nil {
		fn(next)
	}
}
