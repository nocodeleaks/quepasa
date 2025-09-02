package whatsmeow

import (
	"context"
	"fmt"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
	whatsmeow "go.mau.fi/whatsmeow"
	types "go.mau.fi/whatsmeow/types"
)

// Compile-time interface check
var _ whatsapp.WhatsappGroupManagerInterface = (*WhatsmeowGroupManager)(nil)

// WhatsmeowGroupManager handles all group-related operations for WhatsmeowConnection
type WhatsmeowGroupManager struct {
	*WhatsmeowConnection // embedded connection instead of property
}

// NewWhatsmeowGroupManager creates a new WhatsmeowGroupManager instance
func NewWhatsmeowGroupManager(conn *WhatsmeowConnection) *WhatsmeowGroupManager {
	return &WhatsmeowGroupManager{
		WhatsmeowConnection: conn,
	}
}

// GetClient returns the whatsmeow client from the embedded connection
func (gm *WhatsmeowGroupManager) GetClient() *whatsmeow.Client {
	if gm.WhatsmeowConnection != nil {
		return gm.WhatsmeowConnection.Client
	}
	return nil
}

// GetLogger returns the logger from the embedded connection
func (gm *WhatsmeowGroupManager) GetLogger() *log.Entry {
	if gm.WhatsmeowConnection != nil {
		return gm.WhatsmeowConnection.GetLogger()
	}
	return log.NewEntry(log.StandardLogger())
}

// GetInvite gets the invite link for a group
func (gm *WhatsmeowGroupManager) GetInvite(groupId string) (link string, err error) {
	client := gm.GetClient()
	if client == nil {
		return "", fmt.Errorf("client not defined")
	}

	jid, err := types.ParseJID(groupId)
	if err != nil {
		gm.GetLogger().Infof("getting invite error on parse jid: %s", err)
		return "", err
	}

	link, err = client.GetGroupInviteLink(jid, false)
	return
}

// GetJoinedGroups returns all groups the user has joined
func (gm *WhatsmeowGroupManager) GetJoinedGroups() ([]interface{}, error) {
	client := gm.GetClient()
	if client == nil {
		return nil, fmt.Errorf("client not defined")
	}

	// Get the group info slice
	groupInfos, err := client.GetJoinedGroups()
	if err != nil {
		return nil, err
	}

	// Iterate over groupInfos and set the DisplayName for each participant
	for _, groupInfo := range groupInfos {
		if groupInfo.Participants != nil {
			for i, participant := range groupInfo.Participants {
				// Get the contact info from the store
				contact, err := client.Store.Contacts.GetContact(context.TODO(), participant.JID)
				if err != nil {
					// If no contact info is found, fallback to JID user part
					groupInfo.Participants[i].DisplayName = participant.JID.User
				} else {
					// Set the DisplayName field to the contact's full name or push name
					if len(contact.FullName) > 0 {
						groupInfo.Participants[i].DisplayName = contact.FullName
					} else if len(contact.PushName) > 0 {
						groupInfo.Participants[i].DisplayName = contact.PushName
					} else {
						groupInfo.Participants[i].DisplayName = "" // Fallback to JID user part
					}
				}
			}
		} else {
			// If Participants is nil, initialize it to an empty slice
			groupInfo.Participants = []types.GroupParticipant{}
			// You might want to log this or handle it differently
			gm.GetLogger().Warnf("Group %s has nil Participants, initializing to empty slice", groupInfo.JID.String())
		}
	}

	groups := make([]interface{}, len(groupInfos))
	for i, group := range groupInfos {
		groups[i] = group
	}

	return groups, nil
}

// GetGroupInfo returns information about a specific group
func (gm *WhatsmeowGroupManager) GetGroupInfo(groupId string) (interface{}, error) {
	client := gm.GetClient()
	if client == nil {
		return nil, fmt.Errorf("client not defined")
	}

	jid, err := types.ParseJID(groupId)
	if err != nil {
		return nil, err
	}

	groupInfo, err := client.GetGroupInfo(jid)
	if err != nil {
		return nil, err
	}

	// Fill contact names for participants
	if groupInfo.Participants != nil {
		for i, participant := range groupInfo.Participants {
			// Get the contact info from the store
			contact, err := client.Store.Contacts.GetContact(context.TODO(), participant.JID)
			if err != nil {
				// If no contact info is found, fallback to JID user part
				groupInfo.Participants[i].DisplayName = participant.JID.User
			} else {
				// Set the DisplayName field to the contact's full name or push name
				if len(contact.FullName) > 0 {
					groupInfo.Participants[i].DisplayName = contact.FullName
				} else if len(contact.PushName) > 0 {
					groupInfo.Participants[i].DisplayName = contact.PushName
				} else {
					groupInfo.Participants[i].DisplayName = "" // Fallback to JID user part
				}
			}
		}
	} else {
		// If Participants is nil, initialize it to an empty slice
		groupInfo.Participants = []types.GroupParticipant{}
		gm.GetLogger().Warnf("Group %s has nil Participants, initializing to empty slice", groupInfo.JID.String())
	}

	return groupInfo, nil
}

// CreateGroup creates a new group with the given name and participants
func (gm *WhatsmeowGroupManager) CreateGroup(name string, participants []string) (interface{}, error) {
	client := gm.GetClient()
	if client == nil {
		return nil, fmt.Errorf("client not defined")
	}

	// Convert participants to JID format
	var participantsJID []types.JID
	for _, participant := range participants {

		phoneJID, err := PhoneToJID(participant)
		if err != nil {
			return nil, fmt.Errorf("invalid participant format: %v", err)
		}

		participantsJID = append(participantsJID, phoneJID)
	}

	// Create the request struct
	groupConfig := whatsmeow.ReqCreateGroup{
		Name:         name,
		Participants: participantsJID,
	}

	// Call the existing method with the constructed request
	return client.CreateGroup(context.TODO(), groupConfig)
}

// CreateGroupExtended creates a new group with extended options
func (gm *WhatsmeowGroupManager) CreateGroupExtended(title string, participants []string) (interface{}, error) {
	client := gm.GetClient()
	if client == nil {
		return nil, fmt.Errorf("client not defined")
	}

	// Convert participants to JIDs
	participantJIDs := make([]types.JID, len(participants))
	for i, participant := range participants {
		jid, err := types.ParseJID(participant)
		if err != nil {
			return nil, fmt.Errorf("invalid participant JID: %v", err)
		}
		participantJIDs[i] = jid
	}

	// Create request structure
	req := whatsmeow.ReqCreateGroup{
		Name:         title,
		Participants: participantJIDs,
	}

	// Call the WhatsApp method
	return client.CreateGroup(context.TODO(), req)
}

// UpdateGroupSubject updates the name/subject of a group
func (gm *WhatsmeowGroupManager) UpdateGroupSubject(groupID string, name string) (interface{}, error) {
	client := gm.GetClient()
	if client == nil {
		return nil, fmt.Errorf("client not defined")
	}

	// Parse the group ID to JID format
	jid, err := types.ParseJID(groupID)
	if err != nil {
		return nil, fmt.Errorf("invalid group JID format: %v", err)
	}

	// Update the group subject
	err = client.SetGroupName(jid, name)
	if err != nil {
		return nil, fmt.Errorf("failed to update group subject: %v", err)
	}

	// Return the updated group info
	return client.GetGroupInfo(jid)
}

// UpdateGroupPhoto updates the photo of a group
func (gm *WhatsmeowGroupManager) UpdateGroupPhoto(groupID string, imageData []byte) (string, error) {
	client := gm.GetClient()
	if client == nil {
		return "", fmt.Errorf("client not defined")
	}

	// Parse the group ID to JID format
	jid, err := types.ParseJID(groupID)
	if err != nil {
		return "", fmt.Errorf("invalid group JID format: %v", err)
	}

	// Update the group photo
	pictureID, err := client.SetGroupPhoto(jid, imageData)
	if err != nil {
		return "", fmt.Errorf("failed to update group photo: %v", err)
	}

	return pictureID, nil
}

// UpdateGroupTopic updates the topic/description of a group
func (gm *WhatsmeowGroupManager) UpdateGroupTopic(groupID string, topic string) (interface{}, error) {
	client := gm.GetClient()
	if client == nil {
		return nil, fmt.Errorf("client not defined")
	}

	// Parse the group ID to JID format
	jid, err := types.ParseJID(groupID)
	if err != nil {
		return nil, fmt.Errorf("invalid group JID format: %v", err)
	}

	// Update the group topic (description)
	// SetGroupTopic requires: jid, previousID, newID, topic
	// Let the whatsmeow library handle previousID and newID automatically by passing empty strings
	err = client.SetGroupTopic(jid, "", "", topic)
	if err != nil {
		return nil, fmt.Errorf("failed to update group topic: %v", err)
	}

	// Return the updated group info
	return client.GetGroupInfo(jid)
}

// UpdateGroupParticipants adds, removes, promotes, or demotes participants in a group
func (gm *WhatsmeowGroupManager) UpdateGroupParticipants(groupJID string, participants []string, action string) ([]interface{}, error) {
	client := gm.GetClient()
	if client == nil {
		return nil, fmt.Errorf("client not defined")
	}

	// Parse the group JID
	jid, err := types.ParseJID(groupJID)
	if err != nil {
		return nil, fmt.Errorf("invalid group JID format: %v", err)
	}

	// Convert participant strings to JIDs
	participantJIDs := make([]types.JID, len(participants))
	for i, participant := range participants {
		participantJIDs[i], err = types.ParseJID(participant)
		if err != nil {
			return nil, fmt.Errorf("invalid participant JID format for %s: %v", participant, err)
		}
	}

	// Map the action string to the ParticipantChange type
	var participantAction whatsmeow.ParticipantChange
	switch action {
	case "add":
		participantAction = whatsmeow.ParticipantChangeAdd
	case "remove":
		participantAction = whatsmeow.ParticipantChangeRemove
	case "promote":
		participantAction = whatsmeow.ParticipantChangePromote
	case "demote":
		participantAction = whatsmeow.ParticipantChangeDemote
	default:
		return nil, fmt.Errorf("invalid action %s", action)
	}

	// Call the whatsmeow method
	result, err := client.UpdateGroupParticipants(jid, participantJIDs, participantAction)
	if err != nil {
		return nil, fmt.Errorf("failed to update group participants: %v", err)
	}

	// Convert to interface array for the generic return type
	interfaceResults := make([]interface{}, len(result))
	for i, r := range result {
		interfaceResults[i] = r
	}

	return interfaceResults, nil
}

// GetGroupJoinRequests gets pending join requests for a group
func (gm *WhatsmeowGroupManager) GetGroupJoinRequests(groupJID string) ([]interface{}, error) {
	client := gm.GetClient()
	if client == nil {
		return nil, fmt.Errorf("client not defined")
	}

	// Parse the group JID
	jid, err := types.ParseJID(groupJID)
	if err != nil {
		return nil, fmt.Errorf("invalid group JID format: %v", err)
	}

	// Call the whatsmeow method
	requests, err := client.GetGroupRequestParticipants(jid)
	if err != nil {
		return nil, fmt.Errorf("failed to get group join requests: %v", err)
	}

	// Convert to interface array for the generic return type
	interfaceResults := make([]interface{}, len(requests))
	for i, r := range requests {
		interfaceResults[i] = r
	}

	return interfaceResults, nil
}

// HandleGroupJoinRequests approves or rejects pending join requests
func (gm *WhatsmeowGroupManager) HandleGroupJoinRequests(groupJID string, participants []string, action string) ([]interface{}, error) {
	client := gm.GetClient()
	if client == nil {
		return nil, fmt.Errorf("client not defined")
	}

	// Parse the group JID
	jid, err := types.ParseJID(groupJID)
	if err != nil {
		return nil, fmt.Errorf("invalid group JID format: %v", err)
	}

	// Convert participant strings to JIDs
	participantJIDs := make([]types.JID, len(participants))
	for i, participant := range participants {
		participantJIDs[i], err = types.ParseJID(participant)
		if err != nil {
			return nil, fmt.Errorf("invalid participant JID format for %s: %v", participant, err)
		}
	}

	// Map the action string to the ParticipantRequestChange type
	var requestAction whatsmeow.ParticipantRequestChange
	switch action {
	case "approve":
		requestAction = whatsmeow.ParticipantChangeApprove
	case "reject":
		requestAction = whatsmeow.ParticipantChangeReject
	default:
		return nil, fmt.Errorf("invalid action %s", action)
	}

	// Call the correct WhatsApp method which returns participant results
	result, err := client.UpdateGroupRequestParticipants(jid, participantJIDs, requestAction)
	if err != nil {
		return nil, fmt.Errorf("failed to handle group join requests: %v", err)
	}

	// Convert the typed results to interface array
	interfaceResults := make([]interface{}, len(result))
	for i, r := range result {
		interfaceResults[i] = r
	}

	return interfaceResults, nil
}

// CreateGroupExtendedWithOptions creates a new group with extended options (map-based)
func (gm *WhatsmeowGroupManager) CreateGroupExtendedWithOptions(options map[string]interface{}) (interface{}, error) {
	// Extract parameters from options map
	title, _ := options["title"].(string)
	participantsRaw, _ := options["participants"].([]string)

	// Call the existing CreateGroupExtended method
	return gm.CreateGroupExtended(title, participantsRaw)
}

// LeaveGroup leaves a group by group ID
func (gm *WhatsmeowGroupManager) LeaveGroup(groupID string) error {
	client := gm.GetClient()
	if client == nil {
		return fmt.Errorf("client not defined")
	}

	logger := gm.GetLogger()

	// Parse group JID
	jid, err := types.ParseJID(groupID)
	if err != nil {
		logger.Errorf("failed to parse group JID %s: %v", groupID, err)
		return fmt.Errorf("invalid group JID: %v", err)
	}

	// Validate that it's a group JID
	if jid.Server != types.GroupServer {
		return fmt.Errorf("JID %s is not a group", groupID)
	}

	// Leave the group
	err = client.LeaveGroup(jid)
	if err != nil {
		logger.Errorf("failed to leave group %s: %v", groupID, err)
		return fmt.Errorf("failed to leave group: %v", err)
	}

	logger.Infof("successfully left group %s", groupID)
	return nil
}
