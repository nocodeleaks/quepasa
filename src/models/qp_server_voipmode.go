package models

import "github.com/nocodeleaks/quepasa/whatsapp"

// GetVoIPMode returns the per-instance VoIP mode for this server, read from the
// server metadata. Falls back to VoIPModeDisabled when not set, keeping the
// default (safe) behavior for instances that never configured VoIP.
func (server *QpServer) GetVoIPMode() whatsapp.VoIPMode {
	if server == nil {
		return whatsapp.VoIPModeDisabled
	}

	raw := server.GetMetadataValue(whatsapp.MetadataKeyVoIPMode)
	if raw == nil {
		return whatsapp.VoIPModeDisabled
	}

	if value, ok := raw.(string); ok {
		return whatsapp.ParseVoIPMode(value)
	}

	return whatsapp.VoIPModeDisabled
}

// SetVoIPMode persists the per-instance VoIP mode into the server metadata.
// Passing VoIPModeDisabled removes the key to keep the metadata clean.
func (server *QpServer) SetVoIPMode(mode whatsapp.VoIPMode) {
	if server == nil {
		return
	}

	if mode == whatsapp.VoIPModeDisabled {
		server.RemoveMetadataValue(whatsapp.MetadataKeyVoIPMode)
		return
	}

	server.SetMetadataValue(whatsapp.MetadataKeyVoIPMode, mode.String())
}
