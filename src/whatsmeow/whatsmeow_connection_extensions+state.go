package whatsmeow

import (
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

func IsConnected(source *WhatsmeowConnection) bool {
	if source != nil {

		// manual checks for avoid thread locking
		if source.IsConnecting {
			return false
		}

		if source.Client != nil {
			if source.Client.IsConnected() {
				return true
			}
		}
	}
	return false
}

func GetStatus(source *WhatsmeowConnection) whatsapp.WhatsappConnectionState {
	if source != nil {
		if source.Client == nil {
			return whatsapp.UnVerified
		} else {

			// manual checks for avoid thread locking
			if source.IsConnecting {
				return whatsapp.Connecting
			}

			// this is connected method locks the socket thread, so, if its in connecting state, it will be blocked here
			if source.Client.IsConnected() {
				if source.Client.IsLoggedIn() {
					return whatsapp.Ready
				} else {
					return whatsapp.Connected
				}
			} else {
				if source.failedToken {
					return whatsapp.Failed
				} else {
					return whatsapp.Disconnected
				}
			}
		}
	} else {
		return whatsapp.UnPrepared
	}
}
