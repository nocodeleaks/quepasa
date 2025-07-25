package whatsapp

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
)

type IWhatsappConnection interface {
	IWhatsappConnectionOptions

	GetChatTitle(string) string

	Connect() error
	Disconnect() error

	GetWhatsAppQRChannel(context.Context, chan<- string) error
	GetWhatsAppQRCode() string

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

	// Useful to check if is a member of a group before send a msg.
	// Indicates if has an open or archived chat.
	HasChat(string) bool

	GetLogger() *log.Entry

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

	// Is a valid whatsapp phone numbers
	IsOnWhatsApp(...string) ([]string, error)

	HistorySync(time.Time) error

	GetContacts() ([]WhatsappChat, error)

	PairPhone(phone string) (string, error)

	//region Send Presence
	SendChatPresence(chatId string, presenceType uint) error
	//endregion

	GetLIDFromPhone(phone string) (string, error)

	// Get phone number from LID
	GetPhoneFromLID(lid string) (string, error)

	GetUserInfo(jids []string) ([]interface{}, error)

	// NOTE: Group operations have been moved to IGroupManager
	// Access them via connection.GetGroupManager().MethodName()

	// GetStatusManager returns the status manager for status operations
	GetStatusManager() WhatsappStatusManagerInterface
}
