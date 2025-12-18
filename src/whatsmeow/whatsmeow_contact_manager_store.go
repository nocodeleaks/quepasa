package whatsmeow

import (
	"fmt"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	"go.mau.fi/whatsmeow/store"
)

// Compile-time interface check
var _ whatsapp.WhatsappContactManagerInterface = (*WhatsmeowStoreContactManager)(nil)

// WhatsmeowStoreContactManager reads contacts directly from whatsmeow store
// Used when connection is not available (server stopped) but we need to access cached contacts
type WhatsmeowStoreContactManager struct {
	Device *store.Device
}

// NewStoreContactManagerFromWid creates a contact manager that accesses store directly
// This allows reading cached contacts even when the WhatsApp connection is stopped
func NewStoreContactManagerFromWid(wid string) (whatsapp.WhatsappContactManagerInterface, error) {
	device, err := WhatsmeowService.GetOrCreateStore(wid)
	if err != nil {
		return nil, fmt.Errorf("failed to get store for wid %s: %v", wid, err)
	}

	if device == nil || device.Contacts == nil {
		return nil, fmt.Errorf("store or contacts not available for wid %s", wid)
	}

	return &WhatsmeowStoreContactManager{
		Device: device,
	}, nil
}

// GetContacts reads contacts directly from the store (no active connection required)
func (scm *WhatsmeowStoreContactManager) GetContacts() (chats []whatsapp.WhatsappChat, err error) {
	if scm.Device == nil {
		return nil, fmt.Errorf("device not available")
	}

	// Delegate to shared helper function
	return GetContactsFromDevice(scm.Device)
}

// Methods below are not supported for store-only access (return errors)

func (scm *WhatsmeowStoreContactManager) IsOnWhatsApp(phones ...string) ([]string, error) {
	return nil, fmt.Errorf("IsOnWhatsApp not supported for store-only access (requires active connection)")
}

func (scm *WhatsmeowStoreContactManager) GetProfilePicture(wid string, knowingId string) (*whatsapp.WhatsappProfilePicture, error) {
	return nil, fmt.Errorf("GetProfilePicture not supported for store-only access (requires active connection)")
}

func (scm *WhatsmeowStoreContactManager) GetLIDFromPhone(phone string) (string, error) {
	return "", fmt.Errorf("GetLIDFromPhone not supported for store-only access (requires active connection)")
}

func (scm *WhatsmeowStoreContactManager) GetPhoneFromLID(lid string) (string, error) {
	return "", fmt.Errorf("GetPhoneFromLID not supported for store-only access (requires active connection)")
}

func (scm *WhatsmeowStoreContactManager) GetPhoneFromContactId(contactId string) (string, error) {
	return "", fmt.Errorf("GetPhoneFromContactId not supported for store-only access (requires active connection)")
}

func (scm *WhatsmeowStoreContactManager) GetUserInfo(jids []string) ([]interface{}, error) {
	return nil, fmt.Errorf("GetUserInfo not supported for store-only access (requires active connection)")
}
