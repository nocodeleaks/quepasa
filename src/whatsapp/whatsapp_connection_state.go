package whatsapp

type WhatsappConnectionState uint

const (
	// Unknown, not treated state
	Unknown WhatsappConnectionState = iota

	// No connection attached
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
		"Connected",
		"Fetching",
		"Ready",
		"Halting",
		"Disconnected",
		"Failed",
	}[s]
}
