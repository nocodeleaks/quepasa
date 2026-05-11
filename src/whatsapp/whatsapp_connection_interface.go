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

	UpdateHandler(IWhatsappHandlers)
	UpdatePairedCallBack(func(string))

	// Download message attachment if exists
	DownloadData(IWhatsappMessage) ([]byte, error)

	// Download message attachment if exists and informations
	Download(IWhatsappMessage, bool) (*WhatsappAttachment, error)

	Revoke(IWhatsappMessage) error

	// Edit an existing message with new content
	Edit(IWhatsappMessage, string) error

	MarkRead(IWhatsappMessage) error

	// Default send message method
	Send(*WhatsappMessage) (IWhatsappSendResponse, error)

	// SendReaction sends an emoji reaction to a specific message.
	// Pass emoji="" to remove an existing reaction.
	SendReaction(chatID, targetMsgID string, fromMe bool, emoji string) error

	// PublishStatus publishes a WhatsApp status (story) visible to contacts.
	// Provide text for text-only status; provide attachment for media status.
	// Returns the sent message ID on success.
	PublishStatus(text string, attachment *WhatsappAttachment) (string, error)

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

	HistorySync(time.Time) error

	PairPhone(phone string) (string, error)

	//region Send Presence
	SendChatPresence(chatId string, presenceType uint) error
	//endregion

	// NOTE: Group operations have been moved to IGroupManager
	// Access them via connection.GetGroupManager().MethodName()

	// NOTE: Contact operations have been moved to IContactManager
	// Access them via connection.GetContactManager().MethodName()

	// GetStatusManager returns the status manager for status operations
	GetStatusManager() WhatsappStatusManagerInterface

	// GetContactManager returns the contact manager for contact operations
	GetContactManager() WhatsappContactManagerInterface

	// GetResume returns detailed connection status information
	// This consolidates all status management functionality in a single method
	GetResume() *WhatsappConnectionStatus
}
