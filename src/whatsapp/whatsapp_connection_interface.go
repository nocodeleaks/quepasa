package whatsapp

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	types "go.mau.fi/whatsmeow/types"
)

type IWhatsappConnection interface {
	IWhatsappConnectionOptions

	GetStatus() WhatsappConnectionState

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

	// Is connected and logged, valid verification
	IsValid() bool

	IsConnected() bool

	// Is a valid whatsapp phone numbers
	IsOnWhatsApp(...string) ([]string, error)

	HistorySync(time.Time) error

	GetContacts() ([]WhatsappChat, error)

	PairPhone(phone string) (string, error)

	//region Group Methods

	// Get a list of all groups
	GetJoinedGroups() ([]*types.GroupInfo, error)

	// Get a specific group
	GetGroupInfo(string) (*types.GroupInfo, error)

	// Create a group
	CreateGroup(string, []string) (*types.GroupInfo, error)

	// Update Group Name
	UpdateGroupSubject(string, string) (*types.GroupInfo, error)

	// Update Group Photo
	UpdateGroupPhoto(string, []byte) (string, error)

	// Update group participants (add, remove, promote, demote)
	UpdateGroupParticipants(groupJID string, participants []string, action string) ([]interface{}, error)

	// Get list of pending join requests for a group
	GetGroupJoinRequests(groupJID string) ([]interface{}, error)

	// Handle join requests (approve/reject)
	HandleGroupJoinRequests(groupJID string, participants []string, action string) ([]interface{}, error)

	// Add to the IWhatsappConnection interface in whatsapp/interfaces.go
	CreateGroupExtended(title string, participants []string) (*types.GroupInfo, error)

	//endregion Group Methods
}
