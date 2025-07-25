package whatsapp

// WhatsappContactManagerInterface defines the interface for contact management operations
// This interface should be implemented by the contact manager in the whatsmeow package
type WhatsappContactManagerInterface interface {
	// Get all contacts from WhatsApp
	GetContacts() ([]WhatsappChat, error)

	// Check if phone numbers are registered on WhatsApp
	IsOnWhatsApp(phones ...string) ([]string, error)

	// Get profile picture information
	GetProfilePicture(wid string, knowingId string) (*WhatsappProfilePicture, error)

	// Get LID from phone number
	GetLIDFromPhone(phone string) (string, error)

	// Get phone number from LID
	GetPhoneFromLID(lid string) (string, error)

	// Get comprehensive user information for given JIDs
	GetUserInfo(jids []string) ([]interface{}, error)
}

// IWhatsappConnectionWithContacts extends IWhatsappConnection with contact management
// Use this interface when you need both connection and contact operations
type IWhatsappConnectionWithContacts interface {
	IWhatsappConnection

	// GetContactManager returns the contact manager for contact operations
	GetContactManager() WhatsappContactManagerInterface
}
