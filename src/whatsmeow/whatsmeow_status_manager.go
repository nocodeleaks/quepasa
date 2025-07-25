package whatsmeow

import (
	"fmt"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// Compile-time interface check
var _ whatsapp.WhatsappStatusManagerInterface = (*WhatsmeowStatusManager)(nil)

// WhatsmeowStatusManager handles all status and connection information operations for WhatsmeowConnection
type WhatsmeowStatusManager struct {
	*WhatsmeowConnection // embedded connection for direct access
}

// NewWhatsmeowStatusManager creates a new WhatsmeowStatusManager instance
func NewWhatsmeowStatusManager(conn *WhatsmeowConnection) *WhatsmeowStatusManager {
	return &WhatsmeowStatusManager{
		WhatsmeowConnection: conn,
	}
}

// GetVersion returns the WhatsApp connection version
func (sm *WhatsmeowStatusManager) GetVersion() string {
	return "multi"
}

// GetWid returns the WhatsApp ID (WID) as string
func (sm *WhatsmeowStatusManager) GetWid() string {
	if sm.WhatsmeowConnection != nil {
		wid, err := sm.GetWidInternal()
		if err != nil {
			return wid
		}
	}
	return ""
}

// GetWidInternal returns the WhatsApp ID with error handling
func (sm *WhatsmeowStatusManager) GetWidInternal() (string, error) {
	if sm.Client == nil {
		err := fmt.Errorf("client not defined on trying to get wid")
		return "", err
	}

	if sm.Client.Store == nil {
		err := fmt.Errorf("device store not defined on trying to get wid")
		return "", err
	}

	if sm.Client.Store.ID == nil {
		err := fmt.Errorf("device id not defined on trying to get wid")
		return "", err
	}

	wid := sm.Client.Store.ID.User
	return wid, nil
}

// IsValid checks if connection is valid (connected and logged in)
func (sm *WhatsmeowStatusManager) IsValid() bool {
	if sm.WhatsmeowConnection != nil {
		if sm.Client != nil {
			if sm.Client.IsConnected() {
				if sm.Client.IsLoggedIn() {
					return true
				}
			}
		}
	}
	return false
}

// IsConnected checks if connection is established
func (sm *WhatsmeowStatusManager) IsConnected() bool {
	return IsConnected(sm.WhatsmeowConnection)
}

// GetStatus returns current connection status
func (sm *WhatsmeowStatusManager) GetStatus() whatsapp.WhatsappConnectionState {
	return GetStatus(sm.WhatsmeowConnection)
}

// GetReconnect returns auto-reconnect setting
func (sm *WhatsmeowStatusManager) GetReconnect() bool {
	if sm.WhatsmeowConnection != nil {
		if sm.Client != nil {
			return sm.Client.EnableAutoReconnect
		}
	}
	return false
}

// SetReconnect sets auto-reconnect setting
func (sm *WhatsmeowStatusManager) SetReconnect(value bool) {
	if sm.WhatsmeowConnection != nil {
		if sm.Client != nil {
			sm.Client.EnableAutoReconnect = value
		}
	}
}
