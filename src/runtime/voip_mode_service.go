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
// metadata and saves the instance to the database. The new mode is applied to
// the live connection on its next (re)connection, since the VoIP manager reads
// the mode at connection setup.
func SetSessionVoIPMode(server *models.QpWhatsappServer, mode whatsapp.VoIPMode) error {
	if server == nil {
		return ErrNilSession
	}

	server.SetVoIPMode(mode)
	if err := server.Save(fmt.Sprintf("voip mode set to %s", mode.String())); err != nil {
		return fmt.Errorf("failed to persist voip mode: %w", err)
	}
	return nil
}
