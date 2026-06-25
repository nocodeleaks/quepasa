package whatsmeow

import (
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// GetContactManagerForWid returns a contact manager for a given wid and optional connection
// If connection is available and valid, returns standard WhatsmeowContactManager
// If connection is nil/unavailable, creates a store-only contact manager for cached data access
// This factory function encapsulates the decision logic and keeps it within the whatsmeow package
func GetContactManagerForWid(wid string, conn whatsapp.IWhatsappConnection) (whatsapp.WhatsappContactManagerInterface, error) {
	// Try to get contact manager from active connection first
	if conn != nil && !conn.IsInterfaceNil() {
		contactManager := conn.GetContactManager()
		if contactManager != nil {
			return contactManager, nil
		}
	}

	// Connection not available - fallback to store-only access
	return NewStoreContactManagerFromWid(wid)
}

// GetContactManagerForConnection returns a contact manager for a given connection
// If connection is nil or unavailable, tries to return a store-only contact manager
// This allows reading cached contacts even when the server is stopped
// Deprecated: Use GetContactManagerForWid instead as it's more explicit
func GetContactManagerForConnection(conn whatsapp.IWhatsappConnection, wid string) (whatsapp.WhatsappContactManagerInterface, error) {
	return GetContactManagerForWid(wid, conn)
}
