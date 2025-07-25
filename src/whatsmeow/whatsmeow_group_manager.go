package whatsmeow

import (
	"context"
	"fmt"
	"strings"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
	whatsmeow "go.mau.fi/whatsmeow"
	types "go.mau.fi/whatsmeow/types"
)

// GroupManager handles all group-related operations for WhatsmeowConnection
type GroupManager struct {
	connection *WhatsmeowConnection
}

// NewGroupManager creates a new GroupManager instance
func NewGroupManager(conn *WhatsmeowConnection) *GroupManager {
	return &GroupManager{
		connection: conn,
	}
}

// GetClient returns the whatsmeow client from the connection
func (gm *GroupManager) GetClient() *whatsmeow.Client {
	if gm.connection != nil {
		return gm.connection.Client
	}
	return nil
}

// GetLogger returns the logger from the connection
func (gm *GroupManager) GetLogger() *log.Entry {
	if gm.connection != nil {
		return gm.connection.GetLogger()
	}
	return log.NewEntry(log.StandardLogger())
}

// GetInvite gets the invite link for a group
func (gm *GroupManager) GetInvite(groupId string) (link string, err error) {
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
func (gm *GroupManager) GetJoinedGroups() ([]interface{}, error) {
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
func (gm *GroupManager) GetGroupInfo(groupId string) (interface{}, error) {
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
func (gm *GroupManager) CreateGroup(name string, participants []string) (interface{}, error) {
	client := gm.GetClient()
	if client == nil {
		return nil, fmt.Errorf("client not defined")
	}

	// Convert participants to JID format
	var participantsJID []types.JID
	for _, participant := range participants {
		// Check if it's already in JID format
		if strings.Contains(participant, "@") {
			jid, err := types.ParseJID(participant)
			if err != nil {
				return nil, fmt.Errorf("invalid JID format for participant %s: %v", participant, err)
			}
			participantsJID = append(participantsJID, jid)
		} else {
			// Assume it's a phone number and convert to JID
			jid := types.JID{
				User:   participant,
				Server: whatsapp.WHATSAPP_SERVERDOMAIN_USER, // Use the standard WhatsApp server
			}
			participantsJID = append(participantsJID, jid)
		}
	}

	// Create the request struct
	groupConfig := whatsmeow.ReqCreateGroup{
		Name:         name,
		Participants: participantsJID,
	}

	// Call the existing method with the constructed request
	return client.CreateGroup(groupConfig)
}

// CreateGroupExtended creates a new group with extended options
func (gm *GroupManager) CreateGroupExtended(title string, participants []string) (interface{}, error) {
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
	return client.CreateGroup(req)
}

// UpdateGroupSubject updates the name/subject of a group
func (gm *GroupManager) UpdateGroupSubject(groupID string, name string) (interface{}, error) {
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
func (gm *GroupManager) UpdateGroupPhoto(groupID string, imageData []byte) (string, error) {
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
func (gm *GroupManager) UpdateGroupTopic(groupID string, topic string) (interface{}, error) {
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
func (gm *GroupManager) UpdateGroupParticipants(groupJID string, participants []string, action string) ([]interface{}, error) {
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
func (gm *GroupManager) GetGroupJoinRequests(groupJID string) ([]interface{}, error) {
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
func (gm *GroupManager) HandleGroupJoinRequests(groupJID string, participants []string, action string) ([]interface{}, error) {
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
