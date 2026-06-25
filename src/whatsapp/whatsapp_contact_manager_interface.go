package whatsapp

// WhatsappContactManagerInterface defines the interface for contact management operations
// This interface should be implemented by the contact manager in the whatsmeow package
type WhatsappContactManagerInterface interface {
	// Get all contacts from WhatsApp
	GetContacts() ([]WhatsappChat, error)

	// IsOnWhatsApp checks whether the given phone numbers are registered on WhatsApp.
	//
	// WARNING: This method performs a live query against WhatsApp servers.
	// Calling it too frequently (e.g. on every outbound message) will trigger
	// WhatsApp's anti-abuse detection and may result in the account being banned.
	//
	// Use only when strictly necessary (e.g. one-time digit-9 resolution during
	// REMOVEDIGIT9 normalization). Do NOT call in hot paths or per-message loops.
	IsOnWhatsApp(phones ...string) ([]string, error)

	// Get profile picture information
	GetProfilePicture(wid string, knowingId string) (*WhatsappProfilePicture, error)

	// Get LID from phone number
	GetLIDFromPhone(phone string) (string, error)

	// Get phone number from LID
	GetPhoneFromLID(lid string) (string, error)

	// Get phone number from contact Id (works with both @s.whatsapp.net and @lid formats)
	GetPhoneFromContactId(contactId string) (string, error)

	// Get comprehensive user information for given JIDs
	GetUserInfo(jids []string) ([]interface{}, error)

	// BlockContact blocks a contact by WID/JID.
	BlockContact(wid string) error

	// UnblockContact unblocks a previously blocked contact by WID/JID.
	UnblockContact(wid string) error
}

// IWhatsappConnectionWithContacts extends IWhatsappConnection with contact management
// Use this interface when you need both connection and contact operations
type IWhatsappConnectionWithContacts interface {
	IWhatsappConnection

	// GetContactManager returns the contact manager for contact operations
	GetContactManager() WhatsappContactManagerInterface
}
