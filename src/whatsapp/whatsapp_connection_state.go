package whatsapp

import "encoding/json"

type WhatsappConnectionState uint

const (
	// Unknown, not treated state
	Unknown WhatsappConnectionState = iota

	// No connection attached, possible failed on retry after connection lost
	UnPrepared

	// Not verified (not logged)
	UnVerified

	// Starting variables
	Starting

	// Connecting to whatsapp servers
	Connecting

	// Stopped requested
	Stopping

	// Stopped after request completed
	Stopped

	Restarting

	// Attempting to reconnect after connection loss (auto-reconnect active)
	Reconnecting

	/*
		<summary>
			Connected to whatsapp servers
			Start to logging with saved keys or waiting for qrcode reads
		</summary>
	*/
	Connected

	// Fetching messages from servers
	Fetching

	// Ready and Fully operational
	Ready

	// Finalizing
	Halting

	// Disconnected from whatsapp servers
	Disconnected

	// Failed state, for any reason
	Failed
)

// EnumIndex - Creating common behavior - give the type a EnumIndex function
func (s WhatsappConnectionState) EnumIndex() int {
	return int(s)
}

func (s WhatsappConnectionState) String() string {
	return [...]string{
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
	}[s]
}

// used for the status monitoring systems
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
