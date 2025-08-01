package models

import (
	"fmt"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// Compile-time check to ensure QpGroupManager implements whatsapp.WhatsappGroupManagerInterface
var _ whatsapp.WhatsappGroupManagerInterface = (*QpGroupManager)(nil)

// QpGroupManager handles all group-related operations for QpWhatsappServer
// Implements whatsapp.WhatsappGroupManagerInterface interface
type QpGroupManager struct {
	*QpWhatsappServer // embedded server for direct access
}

// NewQpGroupManager creates a new QpGroupManager instance
func NewQpGroupManager(server *QpWhatsappServer) *QpGroupManager {
	return &QpGroupManager{
		QpWhatsappServer: server,
	}
}

// Compile-time check to ensure QpGroupManager implements whatsapp.WhatsappGroupManagerInterface
var _ whatsapp.WhatsappGroupManagerInterface = (*QpGroupManager)(nil)

// getGroupManager is a helper function to get the group manager from connection
func (gm *QpGroupManager) getGroupManager() (whatsapp.WhatsappGroupManagerInterface, error) {
	conn, err := gm.GetValidConnection()
	if err != nil {
		return nil, err
	}

	// Type assertion to access group manager
	connWithGroups, ok := conn.(whatsapp.IWhatsappConnectionWithGroups)
	if !ok {
		return nil, fmt.Errorf("connection does not support group operations")
	}

	return connWithGroups.GetGroupManager(), nil
}

// GetInvite gets the invite link for a group
func (gm *QpGroupManager) GetInvite(groupId string) (string, error) {
	conn, err := gm.GetValidConnection()
	if err != nil {
		return "", err
	}

	// Type assertion to access group manager
	connWithGroups, ok := conn.(whatsapp.IWhatsappConnectionWithGroups)
	if !ok {
		return "", fmt.Errorf("connection does not support group operations")
	}

	return connWithGroups.GetGroupManager().GetInvite(groupId)
}

// GetJoinedGroups returns all groups the user has joined
func (gm *QpGroupManager) GetJoinedGroups() ([]interface{}, error) {
	groupManager, err := gm.getGroupManager()
	if err != nil {
		return nil, err
	}

	return groupManager.GetJoinedGroups()
}

// GetGroupInfo returns information about a specific group
func (gm *QpGroupManager) GetGroupInfo(groupID string) (interface{}, error) {
	groupManager, err := gm.getGroupManager()
	if err != nil {
		return nil, err
	}

	return groupManager.GetGroupInfo(groupID)
}

// CreateGroup creates a new group with the given name and participants
func (gm *QpGroupManager) CreateGroup(name string, participants []string) (interface{}, error) {
	groupManager, err := gm.getGroupManager()
	if err != nil {
		return nil, err
	}

	return groupManager.CreateGroup(name, participants)
}

// UpdateGroupSubject updates the name/subject of a group
func (gm *QpGroupManager) UpdateGroupSubject(groupID string, name string) (interface{}, error) {
	groupManager, err := gm.getGroupManager()
	if err != nil {
		return nil, err
	}

	return groupManager.UpdateGroupSubject(groupID, name)
}

// UpdateGroupTopic updates the topic/description of a group
func (gm *QpGroupManager) UpdateGroupTopic(groupID string, topic string) (interface{}, error) {
	groupManager, err := gm.getGroupManager()
	if err != nil {
		return nil, err
	}

	return groupManager.UpdateGroupTopic(groupID, topic)
}

// UpdateGroupPhoto updates the photo of a group
func (gm *QpGroupManager) UpdateGroupPhoto(groupID string, imageData []byte) (string, error) {
	groupManager, err := gm.getGroupManager()
	if err != nil {
		return "", err
	}

	return groupManager.UpdateGroupPhoto(groupID, imageData)
}

// UpdateGroupParticipants adds, removes, promotes, or demotes participants in a group
func (gm *QpGroupManager) UpdateGroupParticipants(groupJID string, participants []string, action string) ([]interface{}, error) {
	groupManager, err := gm.getGroupManager()
	if err != nil {
		return nil, err
	}

	return groupManager.UpdateGroupParticipants(groupJID, participants, action)
}

// GetGroupJoinRequests gets pending join requests for a group
func (gm *QpGroupManager) GetGroupJoinRequests(groupJID string) ([]interface{}, error) {
	groupManager, err := gm.getGroupManager()
	if err != nil {
		return nil, err
	}

	return groupManager.GetGroupJoinRequests(groupJID)
}

// HandleGroupJoinRequests approves or rejects pending join requests
func (gm *QpGroupManager) HandleGroupJoinRequests(groupJID string, participants []string, action string) ([]interface{}, error) {
	groupManager, err := gm.getGroupManager()
	if err != nil {
		return nil, err
	}

	return groupManager.HandleGroupJoinRequests(groupJID, participants, action)
}

// CreateGroupExtended creates a new group with extended options (interface signature)
func (gm *QpGroupManager) CreateGroupExtended(title string, participants []string) (interface{}, error) {
	groupManager, err := gm.getGroupManager()
	if err != nil {
		return nil, err
	}

	return groupManager.CreateGroupExtended(title, participants)
}

// CreateGroupExtendedWithOptions creates a new group with extended options (map-based)
func (gm *QpGroupManager) CreateGroupExtendedWithOptions(options map[string]interface{}) (interface{}, error) {
	groupManager, err := gm.getGroupManager()
	if err != nil {
		return nil, err
	}

	// Extract parameters
	title, _ := options["title"].(string)
	participantsRaw, _ := options["participants"].([]string)

	return groupManager.CreateGroupExtended(title, participantsRaw)
}

// LeaveGroup leaves a group by group ID
func (gm *QpGroupManager) LeaveGroup(groupID string) error {
	groupManager, err := gm.getGroupManager()
	if err != nil {
		return err
	}

	return groupManager.LeaveGroup(groupID)
}
