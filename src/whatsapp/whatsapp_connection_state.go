package whatsapp

import "encoding/json"

type WhatsappConnectionState uint

const (
	// Unknown is a fallback for invalid, missing, or unmapped state values.
	// It is currently returned by the server wrapper only for invalid/nil server references.
	Unknown WhatsappConnectionState = iota

	// UnPrepared means the server exists but has no active connection object attached.
	// This usually happens before a start attempt or after a connection was disposed.
	// It is currently emitted by both the server wrapper and the whatsmeow status provider.
	UnPrepared

	// UnVerified means the server is not authenticated with WhatsApp yet.
	// It is expected before pairing/login and is not, by itself, a transport failure.
	// It is currently emitted when no authenticated client/session is available.
	UnVerified

	// Starting means local startup and dependency initialization are in progress.
	// Reserved for finer-grained lifecycle reporting. It is not currently emitted by
	// the active status calculation path.
	Starting

	// Connecting means the client is trying to establish a session with WhatsApp servers.
	// It is currently emitted by the whatsmeow status provider while IsConnecting is true.
	Connecting

	// Stopping means a stop was requested and the active connection is still being released.
	// This is a transitional state, not a failure state.
	// It is currently emitted by the server wrapper when StopRequested is true but the
	// transport is still connected.
	Stopping

	// Stopped means the server is intentionally offline after a stop request completed.
	// In this project, it is considered operationally healthy because the server is
	// in a stable requested state and can be started again.
	//
	// This state is driven by StopRequested and is commonly reached by explicit user
	// actions such as toggling stop, but it may also appear in controlled internal
	// flows that call Stop(), such as restart.
	// It is currently emitted by the server wrapper and is considered healthy by the
	// health endpoint.
	Stopped

	// Restarting means a controlled stop/start cycle is being executed.
	// Reserved for finer-grained lifecycle reporting. It is not currently emitted by
	// the active status calculation path.
	Restarting

	// Reconnecting means the connection was lost and the auto-reconnect flow is trying
	// to restore the session without requiring a new manual start.
	// Reserved for finer-grained lifecycle reporting. It is not currently emitted by
	// the active status calculation path.
	Reconnecting

	/*
		<summary>
			Connected to WhatsApp servers.
			The transport is established, but the session may still be completing login,
			loading saved credentials, or waiting for QR code confirmation.
			This state is currently emitted by the whatsmeow status provider when the
			transport is connected but the client is not yet fully logged in.
		</summary>
	*/
	Connected

	// Fetching means the session is connected and synchronizing initial data/history.
	// Reserved for finer-grained lifecycle reporting. It is not currently emitted by
	// the active status calculation path.
	Fetching

	// Ready means the server is connected, authenticated, and fully operational.
	// It is currently emitted by the whatsmeow status provider and is considered healthy
	// by the health endpoint.
	Ready

	// Halting means the connection is shutting down as part of a finalization flow.
	// Reserved for finer-grained lifecycle reporting. It is not currently emitted by
	// the active status calculation path.
	Halting

	// Disconnected means the connection to WhatsApp servers was lost unexpectedly or
	// ended outside the intentional stopped state.
	// It is currently emitted by the whatsmeow status provider for non-intentional
	// offline states.
	Disconnected

	// Failed means the server entered an error state that prevented normal operation.
	// It is currently emitted by the whatsmeow status provider when a connection/token
	// failure is flagged.
	Failed
)

// EnumIndex - Creating common behavior - give the type a EnumIndex function
func (s WhatsappConnectionState) EnumIndex() int {
	return int(s)
}

func (s WhatsappConnectionState) String() string {
	names := [...]string{
		"Unknown",
		"UnPrepared",
		"UnVerified",
		"Starting",
		"Connecting",
		"Stopping",
		"Stopped",
		"Restarting",
		"Reconnecting",
		"Connected",
		"Fetching",
		"Ready",
		"Halting",
		"Disconnected",
		"Failed",
	}

	if int(s) < 0 || int(s) >= len(names) {
		return "Unknown"
	}

	return names[s]
}

// IsHealthy reports whether the server is in an operationally acceptable state for
// health monitoring. Ready is healthy because the server is fully available, and
// Stopped is also healthy because it represents an intentional, stable stop state
// rather than a fault condition.
func (s WhatsappConnectionState) IsHealthy() bool {
	return s == Ready || s == Stopped
}

func (s WhatsappConnectionState) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *WhatsappConnectionState) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	states := map[string]WhatsappConnectionState{
		"Unknown":      Unknown,
		"UnPrepared":   UnPrepared,
		"UnVerified":   UnVerified,
		"Starting":     Starting,
		"Connecting":   Connecting,
		"Stopping":     Stopping,
		"Stopped":      Stopped,
		"Restarting":   Restarting,
		"Reconnecting": Reconnecting,
		"Connected":    Connected,
		"Fetching":     Fetching,
		"Ready":        Ready,
		"Halting":      Halting,
		"Disconnected": Disconnected,
		"Failed":       Failed,
	}

	if state, ok := states[str]; ok {
		*s = state
		return nil
	}

	// Default to Unknown if not found
	*s = Unknown
	return nil
}
