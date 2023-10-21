package whatsapp

import (
	"context"

	log "github.com/sirupsen/logrus"
)

type IWhatsappConnection interface {
	GetStatus() WhatsappConnectionState

	// Retorna o ID do controlador whatsapp
	GetWid() (string, error)
	GetChatTitle(string) string

	Connect() error
	Disconnect() error

	GetWhatsAppQRChannel(context.Context, chan<- string) error
	GetWhatsAppQRCode() string

	// Get group invite link
	GetInvite(groupId string) (string, error)

	// Get info to download profile picture
	GetProfilePicture(wid string, knowingId string) (*WhatsappProfilePicture, error)

	UpdateHandler(IWhatsappHandlers)
	UpdatePairedCallBack(func(string))

	// Download message attachment if exists
	DownloadData(IWhatsappMessage) ([]byte, error)

	// Download message attachment if exists and informations
	Download(IWhatsappMessage, bool) (*WhatsappAttachment, error)

	Revoke(IWhatsappMessage) error

	// Default send message method
	Send(*WhatsappMessage) (IWhatsappSendResponse, error)

	// Define the log level for this connection
	UpdateLog(*log.Entry)

	/*
		<summary>
			Disconnect if connected
			Cleanup Handlers
			Dispose resources
			Does not erase permanent data !
		</summary>
	*/
	Dispose(string)

	/*
		<summary>
			Erase permanent data + Dispose !
		</summary>
	*/
	Delete() error

	IsInterfaceNil() bool

	// Is connected and logged, valid verification
	IsValid() bool

	IsConnected() bool

	// Is a valid whatsapp phone numbers
	IsOnWhatsApp(...string) ([]string, error)
}
