package models

// SessionIntent represents the current application-level intent for a WhatsApp session.
// It replaces the pair of boolean flags (StopRequested, DeleteRequested) with an
// explicit enum to eliminate invalid state combinations and simplify GetState logic.
type SessionIntent uint8

const (
	// SessionIntentNone means no special intent — the session is running normally.
	SessionIntentNone SessionIntent = iota

	// SessionIntentStop means a stop was requested but no deletion is pending.
	SessionIntentStop

	// SessionIntentDelete means a delete was requested, which always implies a stop.
	// Callers should use IsStopRequested() to test both Stop and Delete intents.
	SessionIntentDelete
)

// IsStopRequested reports whether the session has a pending stop or delete intent.
// Both SessionIntentStop and SessionIntentDelete require the connection to be stopped.
func (i SessionIntent) IsStopRequested() bool {
	return i == SessionIntentStop || i == SessionIntentDelete
}

// IsDeleteRequested reports whether the session has a pending delete intent.
func (i SessionIntent) IsDeleteRequested() bool {
	return i == SessionIntentDelete
}

// String returns a human-readable name for the intent.
func (i SessionIntent) String() string {
	switch i {
	case SessionIntentNone:
		return "None"
	case SessionIntentStop:
		return "Stop"
	case SessionIntentDelete:
		return "Delete"
	default:
		return "Unknown"
	}
}
