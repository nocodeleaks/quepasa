package models

import (
	"fmt"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	whatsmeow "github.com/nocodeleaks/quepasa/whatsmeow"
)

// Compile-time check to ensure QpContactManager implements whatsapp.WhatsappContactManagerInterface
var _ whatsapp.WhatsappContactManagerInterface = (*QpContactManager)(nil)

// QpContactManager handles all contact-related operations for QpWhatsappServer
// Implements whatsapp.WhatsappContactManagerInterface interface
type QpContactManager struct {
	*QpWhatsappServer // embedded server for direct access
}

// NewQpContactManager creates a new QpContactManager instance
func NewQpContactManager(server *QpWhatsappServer) *QpContactManager {
	return &QpContactManager{
		QpWhatsappServer: server,
	}
}

// getContactManager is a helper function to get the contact manager from connection
func (cm *QpContactManager) getContactManager() (whatsapp.WhatsappContactManagerInterface, error) {
	conn, err := cm.GetValidConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to get valid connection: %v", err)
	}

	contactManager := conn.GetContactManager()
	if contactManager == nil {
		return nil, fmt.Errorf("contact manager not available")
	}

	return contactManager, nil
}

// GetContacts returns all contacts from WhatsApp
// Works with both active connection and stopped server (uses cached data automatically via factory)
func (cm *QpContactManager) GetContacts() ([]whatsapp.WhatsappChat, error) {
	// Use factory that handles fallback to store-only access automatically
	contactManager, err := whatsmeow.GetContactManagerForWid(cm.Wid, cm.GetConnection())
	if err != nil {
		return nil, err
	}
	return contactManager.GetContacts()
}

// IsOnWhatsApp checks if phone numbers are registered on WhatsApp
func (cm *QpContactManager) IsOnWhatsApp(phones ...string) ([]string, error) {
	contactManager, err := cm.getContactManager()
	if err != nil {
		return nil, err
	}
	return contactManager.IsOnWhatsApp(phones...)
}

// GetProfilePicture gets profile picture information
func (cm *QpContactManager) GetProfilePicture(wid string, knowingId string) (*whatsapp.WhatsappProfilePicture, error) {
	contactManager, err := cm.getContactManager()
	if err != nil {
		return nil, err
	}
	return contactManager.GetProfilePicture(wid, knowingId)
}

// GetLIDFromPhone returns the @lid for a given phone number
func (cm *QpContactManager) GetLIDFromPhone(phone string) (string, error) {
	contactManager, err := cm.getContactManager()
	if err != nil {
		return "", err
	}
	return contactManager.GetLIDFromPhone(phone)
}

// GetPhoneFromLID returns the phone number for a given @lid
func (cm *QpContactManager) GetPhoneFromLID(lid string) (string, error) {
	contactManager, err := cm.getContactManager()
	if err != nil {
		return "", err
	}
	return contactManager.GetPhoneFromLID(lid)
}

// GetPhoneFromContactId returns the phone number for a given contact Id (works with both @s.whatsapp.net and @lid formats)
func (cm *QpContactManager) GetPhoneFromContactId(contactId string) (string, error) {
	contactManager, err := cm.getContactManager()
	if err != nil {
		return "", err
	}
	return contactManager.GetPhoneFromContactId(contactId)
}

// GetUserInfo retrieves comprehensive user information for given JIDs
func (cm *QpContactManager) GetUserInfo(jids []string) ([]interface{}, error) {
	contactManager, err := cm.getContactManager()
	if err != nil {
		return nil, err
	}
	return contactManager.GetUserInfo(jids)
}
