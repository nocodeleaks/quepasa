package whatsmeow

import (
	"context"
	"fmt"

	"github.com/nocodeleaks/quepasa/ports"
	"github.com/nocodeleaks/quepasa/whatsapp"
)

// WhatsmeowDriverAdapter implements ports.WhatsappDriverFactory.
// This adapter breaks the models -> whatsmeow import cycle per PLAN P1.1.
type WhatsmeowDriverAdapter struct{}

// CreateEmptyConnection delegates to WhatsmeowService.
func (a *WhatsmeowDriverAdapter) CreateEmptyConnection() (whatsapp.IWhatsappConnection, error) {
	return WhatsmeowService.CreateEmptyConnection()
}

// CreateConnection delegates to WhatsmeowService.
func (a *WhatsmeowDriverAdapter) CreateConnection(options *whatsapp.WhatsappConnectionOptions) (whatsapp.IWhatsappConnection, error) {
	return WhatsmeowService.CreateConnection(options)
}

// GetContactManagerForWid returns a contact manager for a wid, with store fallback.
func (a *WhatsmeowDriverAdapter) GetContactManagerForWid(wid string, conn whatsapp.IWhatsappConnection) (whatsapp.WhatsappContactManagerInterface, error) {
	return GetContactManagerForWid(wid, conn)
}

// ResolveMigratedWid returns the current device wid for a migrated phone number.
func (a *WhatsmeowDriverAdapter) ResolveMigratedWid(phone string) (string, error) {
	if WhatsmeowService == nil {
		return "", fmt.Errorf("whatsmeow service is not initialised")
	}

	device, err := WhatsmeowService.GetStoreForMigrated(phone)
	if err != nil {
		return "", err
	}

	return device.ID.String(), nil
}

// ListDevices returns every device session known to the whatsmeow database.
func (a *WhatsmeowDriverAdapter) ListDevices() ([]ports.WhatsappDeviceInfo, error) {
	if WhatsmeowService == nil {
		return nil, fmt.Errorf("whatsmeow service is not initialised")
	}

	devices, err := WhatsmeowService.Container.GetAllDevices(context.TODO())
	if err != nil {
		return nil, err
	}

	infos := make([]ports.WhatsappDeviceInfo, 0, len(devices))
	for _, dev := range devices {
		if dev == nil || dev.ID == nil {
			continue
		}
		infos = append(infos, ports.WhatsappDeviceInfo{
			JID:      dev.ID.String(),
			PushName: dev.PushName,
			Platform: dev.Platform,
		})
	}

	return infos, nil
}

// Ensure WhatsmeowDriverAdapter implements ports interfaces at compile time.
var (
	_ ports.WhatsappDriverFactory = (*WhatsmeowDriverAdapter)(nil)
	_ ports.WhatsappDriverService = (*WhatsmeowDriverAdapter)(nil)
)
