package whatsapp

import "strings"

// VoIPMode controls how inbound WhatsApp calls are handled per instance.
//
// This is a per-instance (per-bot) setting, persisted in the server metadata,
// that supersedes the legacy global VOIP_ENABLED environment switch.
type VoIPMode string

const (
	// VoIPModeDisabled keeps the default QuePasa behavior: inbound calls are
	// rejected (or relayed) based on the legacy Calls option. The VoIP/SIP
	// bridge is NOT activated for this instance.
	VoIPModeDisabled VoIPMode = "disabled"

	// VoIPModeExclusive makes the SIP proxy the ONLY endpoint for the call.
	// QuePasa answers the WhatsApp call natively and bridges it to SIP. If the
	// SIP server rejects the INVITE (e.g. 401/4xx) or times out, QuePasa hangs
	// up the WhatsApp call so it is NOT left marked as "answered".
	VoIPModeExclusive VoIPMode = "exclusive"

	// VoIPModeAdditional treats the SIP proxy as an EXTRA device. QuePasa does
	// NOT answer and does NOT reject the call: it forwards the INVITE to SIP and
	// lets the native WhatsApp call ring on other paired devices, so a human can
	// pick it up on a phone while the SIP endpoint also rings.
	VoIPModeAdditional VoIPMode = "additional"
)

// MetadataKeyVoIPMode is the metadata key used to persist the per-instance VoIP
// mode in the server metadata JSON blob.
const MetadataKeyVoIPMode = "voipmode"

// ParseVoIPMode normalizes a free-form string into a VoIPMode.
// Unknown/empty values fall back to VoIPModeDisabled so production stays safe by
// default. Accepts a few friendly aliases for convenience.
func ParseVoIPMode(value string) VoIPMode {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(VoIPModeExclusive), "sip", "only", "single", "proxy":
		return VoIPModeExclusive
	case string(VoIPModeAdditional), "extra", "device", "additional-device":
		return VoIPModeAdditional
	default:
		return VoIPModeDisabled
	}
}

// IsActive reports whether the VoIP/SIP bridge should be activated for this
// mode (i.e. any mode other than disabled).
func (mode VoIPMode) IsActive() bool {
	return mode == VoIPModeExclusive || mode == VoIPModeAdditional
}

// String returns the canonical lowercase string representation.
func (mode VoIPMode) String() string {
	if mode == "" {
		return string(VoIPModeDisabled)
	}
	return string(mode)
}
