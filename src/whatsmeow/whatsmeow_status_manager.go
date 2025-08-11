package whatsmeow

import (
	"fmt"
	"time"

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

// GetPlatform returns the WhatsApp connection platform
func (sm *WhatsmeowStatusManager) GetPlatform() string {
	if sm.WhatsmeowConnection != nil && sm.Client != nil && sm.Client.Store != nil {
		// Try to get platform information from store
		if sm.Client.Store.Platform != "" {
			return sm.Client.Store.Platform
		}
	}
	// Fallback to a more descriptive platform string
	return "unavailable"
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

// GetState returns current connection status
func (sm *WhatsmeowStatusManager) GetState() whatsapp.WhatsappConnectionState {
	return GetStatus(sm.WhatsmeowConnection)
}

// GetResume returns comprehensive connection status information
// This method creates a snapshot of the current connection state
func (sm *WhatsmeowStatusManager) GetResume() *whatsapp.WhatsappConnectionStatus {
	if sm.WhatsmeowConnection == nil {
		return nil
	}

	status := &whatsapp.WhatsappConnectionStatus{
		State:        GetStatus(sm.WhatsmeowConnection),
		IsConnected:  IsConnected(sm.WhatsmeowConnection),
		FailedToken:  sm.WhatsmeowConnection.failedToken,
		IsConnecting: sm.WhatsmeowConnection.IsConnecting,
		// Default values for reconnection fields (can be enhanced later)
		IsReconnecting:    false,
		ReconnectAttempts: 0,
	}

	// Get client-specific information if available
	if sm.Client != nil {
		status.IsAuthenticated = sm.Client.IsLoggedIn()
		status.IsValid = status.IsConnected && status.IsAuthenticated
		status.AutoReconnectEnabled = sm.Client.EnableAutoReconnect
		status.ReconnectErrors = uint32(sm.Client.AutoReconnectErrors)

		// Set last successful connect time if available
		if !sm.Client.LastSuccessfulConnect.IsZero() {
			status.LastSuccessfulConnect = &sm.Client.LastSuccessfulConnect
		}

		// Calculate connection uptime if connected
		if status.IsConnected && status.LastSuccessfulConnect != nil {
			uptime := time.Since(*status.LastSuccessfulConnect)
			status.ConnectionUptime = &uptime
		}

		// Get WhatsApp information (platform and SessionId)
		if sm.Client.Store != nil {
			// Try to get platform information from store for platform field
			// The platform is typically set as OSInfo in the store initialization
			if sm.Client.Store.Platform != "" {
				status.Platform = sm.Client.Store.Platform
			} else {
				// Fallback to a more descriptive platform string
				status.Platform = "unavailable"
			}

			// Get SessionId if available
			if sm.Client.Store.ID != nil {
				status.SessionId = sm.Client.Store.ID.User
			}
		}
	}

	return status
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
