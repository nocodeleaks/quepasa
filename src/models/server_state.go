package models

import (
	"fmt"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

func (source *QpWhatsappServer) GetValidConnection() (whatsapp.IWhatsappConnection, error) {
	if source == nil || source.connection == nil || source.connection.IsInterfaceNil() {
		return nil, ErrInvalidConnection
	}

	return source.connection, nil
}

//region IMPLEMENTING WHATSAPP OPTIONS INTERFACE

func (source *QpWhatsappServer) GetOptions() *whatsapp.WhatsappOptions {
	if source == nil {
		return nil
	}

	return &source.WhatsappOptions
}

func (source *QpWhatsappServer) SetOptions(options *whatsapp.WhatsappOptions) error {
	source.WhatsappOptions = *options

	reason := fmt.Sprintf("options updated: %v", source.WhatsappOptions)
	return source.Save(reason)
}

//#endregion

// Ensure default handler
func (server *QpWhatsappServer) HandlerEnsure() {
	if server == nil {
		return // invalid state
	}

	if server.Handler == nil {
		handler := &DispatchingHandler{
			server:             server,
			lifecyclePublisher: DefaultDispatchingLifecyclePublisher(),
		}

		logentry := server.GetLogger()
		logentry.Debug("ensuring messages handler for server")

		// logging
		handler.LogEntry = logentry

		// Inject cache backend from centralized cache service
		InjectCacheBackendIntoHandler(handler)

		// updating
		server.Handler = handler
	}
}

func (server *QpWhatsappServer) HasSignalRActiveConnections() bool {
	if server == nil {
		return false // invalid state
	}

	return HasActiveRealtimeConnections(server.Token)
}

//region IMPLEMENT OF INTERFACE STATE RECOVERY

func (server *QpWhatsappServer) GetStatus() whatsapp.WhatsappConnectionState {
	return server.GetState()
}

// GetState retrieves the current calculated connection state of the WhatsApp server
func (server *QpWhatsappServer) GetState() whatsapp.WhatsappConnectionState {
	if server == nil {
		return whatsapp.Unknown // invalid state
	}

	if server.Intent.IsDeleteRequested() {
		return whatsapp.Stopping
	}

	if server.connection == nil {
		if server.Verified {
			if server.Intent.IsStopRequested() {
				return whatsapp.Stopped
			}
			return whatsapp.UnPrepared
		}

		return whatsapp.UnVerified
	} else {
		if server.Intent.IsStopRequested() {
			statusManager := server.GetStatusManager()
			if server.connection != nil && !server.connection.IsInterfaceNil() && statusManager.IsConnected() {
				return whatsapp.Stopping
			} else {
				return whatsapp.Stopped
			}
		} else {
			statusManager := server.GetStatusManager()
			state := statusManager.GetState()
			if state == whatsapp.Disconnected && !server.Verified {
				return whatsapp.UnVerified
			}
			return state
		}
	}
}

//#endregion
//region IMPLEMENT OF INTERFACE QUEPASA SERVER

// Returns whatsapp controller id on E164
// Ex: 5521967609494
func (server QpWhatsappServer) GetWId() string {
	return server.QpServer.GetWId()
}
