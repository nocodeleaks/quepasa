package models

import (
	"fmt"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// Compile-time check to ensure QpStatusManager implements whatsapp.WhatsappStatusManagerInterface
var _ whatsapp.WhatsappStatusManagerInterface = (*QpStatusManager)(nil)

// QpStatusManager handles all status and connection information operations for QpWhatsappServer
// Implements whatsapp.WhatsappStatusManagerInterface interface
type QpStatusManager struct {
	*QpWhatsappServer // embedded server for direct access
}

// NewQpStatusManager creates a new QpStatusManager instance
func NewQpStatusManager(server *QpWhatsappServer) *QpStatusManager {
	return &QpStatusManager{
		QpWhatsappServer: server,
	}
}

// getStatusManager is a helper function to get the status manager from connection
func (sm *QpStatusManager) getStatusManager() (whatsapp.WhatsappStatusManagerInterface, error) {
	conn, err := sm.GetValidConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to get valid connection: %v", err)
	}

	statusManager := conn.GetStatusManager()
	if statusManager == nil {
		return nil, fmt.Errorf("status manager not available")
	}

	return statusManager, nil
}

// GetVersion returns the WhatsApp connection version
func (sm *QpStatusManager) GetVersion() string {
	statusManager, err := sm.getStatusManager()
	if err != nil {
		return ""
	}
	return statusManager.GetVersion()
}

// GetWid returns the WhatsApp ID (WID) as string
func (sm *QpStatusManager) GetWid() string {
	statusManager, err := sm.getStatusManager()
	if err != nil {
		return ""
	}
	return statusManager.GetWid()
}

// GetWidInternal returns the WhatsApp ID with error handling
func (sm *QpStatusManager) GetWidInternal() (string, error) {
	statusManager, err := sm.getStatusManager()
	if err != nil {
		return "", err
	}
	return statusManager.GetWidInternal()
}

// IsValid checks if connection is valid (connected and logged in)
func (sm *QpStatusManager) IsValid() bool {
	statusManager, err := sm.getStatusManager()
	if err != nil {
		return false
	}
	return statusManager.IsValid()
}

// IsConnected checks if connection is established
func (sm *QpStatusManager) IsConnected() bool {
	statusManager, err := sm.getStatusManager()
	if err != nil {
		return false
	}
	return statusManager.IsConnected()
}

// GetStatus returns current connection status
func (sm *QpStatusManager) GetStatus() whatsapp.WhatsappConnectionState {
	statusManager, err := sm.getStatusManager()
	if err != nil {
		return whatsapp.WhatsappConnectionState(0) // Default state
	}
	return statusManager.GetStatus()
}

// GetReconnect returns auto-reconnect setting
func (sm *QpStatusManager) GetReconnect() bool {
	statusManager, err := sm.getStatusManager()
	if err != nil {
		return false
	}
	return statusManager.GetReconnect()
}
