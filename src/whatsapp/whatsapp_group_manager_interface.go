package whatsapp

// WhatsappGroupManagerInterface defines the interface for group management operations
// This interface should be implemented by the group manager in the whatsmeow package
type WhatsappGroupManagerInterface interface {
	// Get group invite link
	GetInvite(groupId string) (string, error)

	// Get a list of all groups
	GetJoinedGroups() ([]interface{}, error)

	// Get a specific group
	GetGroupInfo(string) (interface{}, error)

	// Create a group
	CreateGroup(string, []string) (interface{}, error)

	// Create a group with extended options (title and participants)
	CreateGroupExtended(title string, participants []string) (interface{}, error)

	// Create group with extended options (map-based for QP level)
	CreateGroupExtendedWithOptions(options map[string]interface{}) (interface{}, error)

	// Update Group Name
	UpdateGroupSubject(string, string) (interface{}, error)

	// Update Group Topic (Description)
	UpdateGroupTopic(string, string) (interface{}, error)

	// Update Group Photo
	UpdateGroupPhoto(string, []byte) (string, error)

	// Update group participants (add, remove, promote, demote)
	UpdateGroupParticipants(groupJID string, participants []string, action string) ([]interface{}, error)

	// Get list of pending join requests for a group
	GetGroupJoinRequests(groupJID string) ([]interface{}, error)

	// Handle join requests (approve/reject)
	HandleGroupJoinRequests(groupJID string, participants []string, action string) ([]interface{}, error)

	// Leave a group
	LeaveGroup(groupID string) error
}

// IWhatsappConnectionWithGroups extends IWhatsappConnection with group management
// Use this interface when you need both connection and group operations
type IWhatsappConnectionWithGroups interface {
	IWhatsappConnection

	// GetGroupManager returns the group manager for group operations
	GetGroupManager() WhatsappGroupManagerInterface
}
