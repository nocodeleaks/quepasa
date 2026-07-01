package runtime

import (
	"fmt"

	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// GetSessionVoIPMode returns the per-instance VoIP mode persisted in the server
// metadata. Defaults to disabled when never configured.
func GetSessionVoIPMode(server *models.QpWhatsappServer) (whatsapp.VoIPMode, error) {
	if server == nil {
		return whatsapp.VoIPModeDisabled, ErrNilSession
	}
	return server.GetVoIPMode(), nil
}

// SetSessionVoIPMode persists the per-instance VoIP mode into the server
// metadata, updates the live options object, and saves the instance to the
// database.
func SetSessionVoIPMode(server *models.QpWhatsappServer, mode whatsapp.VoIPMode) error {
	if server == nil {
		return ErrNilSession
	}

	server.SetVoIPMode(mode)
	server.WhatsappOptions.VoIPMode = mode
	if err := server.Save(fmt.Sprintf("voip mode set to %s", mode.String())); err != nil {
		return fmt.Errorf("failed to persist voip mode: %w", err)
	}
	return nil
}
