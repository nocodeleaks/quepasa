package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	library "github.com/nocodeleaks/quepasa/library"
	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

//region CONTROLLER - GET GROUP

func GetGroupController(w http.ResponseWriter, r *http.Request) {

	// setting default response type as json
	w.Header().Set("Content-Type", "application/json")

	response := &models.QpSingleGroupResponse{}

	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	groupId := library.GetRequestParameter(r, "groupid")
	valid := whatsapp.IsValidGroupId(groupId)
	if !valid {
		response.ParseError(fmt.Errorf("seams to be an invalid group id: %s", groupId))
		RespondInterface(w, response)
		return
	}

	group, err := server.GetGroupInfo(groupId)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}
	response.GroupInfo = group

	RespondSuccess(w, response)
}

//endregion

//region CONTROLLER - FETCH ALL GROUPS

func FetchAllGroupsController(w http.ResponseWriter, r *http.Request) {
	// setting default response type as json
	w.Header().Set("Content-Type", "application/json")

	response := &models.QpGroupsResponse{}

	// Get the server from the request
	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Get all joined groups
	groups, err := server.GetJoinedGroups()
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Set the response with the groups and the amount of groups
	response.Total = len(groups)
	response.Groups = groups

	RespondSuccess(w, response)
}

//endregion

//region CONTROLLER - CREATE GROUP

func CreateGroupController(w http.ResponseWriter, r *http.Request) {
	// Setting default response type as json
	w.Header().Set("Content-Type", "application/json")

	response := &models.QpSingleGroupResponse{}

	// Get server
	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Parse request body with extended options
	var request struct {
		Title        string   `json:"title"`
		Participants []string `json:"participants"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Validate request
	if request.Title == "" {
		response.ParseError(fmt.Errorf("title is required"))
		RespondInterface(w, response)
		return
	}

	// WhatsApp enforces a 25 character limit on group names
	if len(request.Title) > 25 {
		response.ParseError(fmt.Errorf("group title is limited to 25 characters"))
		RespondInterface(w, response)
		return
	}

	if len(request.Participants) == 0 {
		response.ParseError(fmt.Errorf("participants are required"))
		RespondInterface(w, response)
		return
	}

	// Convert phone numbers to proper JID format
	formattedParticipants, err := convertToJIDs(request.Participants)
	if err != nil {
		response.ParseError(fmt.Errorf("failed to format participant numbers: %v", err))
		RespondInterface(w, response)
		return
	}

	// Build extended options for group creation
	options := map[string]interface{}{
		"title":        request.Title,
		"participants": formattedParticipants,
	}

	// Create group using the interface method with properly formatted participants and options
	groupInfo, err := server.CreateGroupExtended(options)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Set response and return
	response.GroupInfo = groupInfo
	RespondSuccess(w, response)
}

func SetGroupNameController(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := &models.QpSingleGroupResponse{}

	type setGroupNameStruct struct {
		GroupJID string `json:"group_jid"`
		Name     string `json:"name"`
	}

	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var t setGroupNameStruct
	err = decoder.Decode(&t)
	if err != nil {
		response.ParseError(fmt.Errorf("could not decode payload: %v", err))
		RespondInterface(w, response)
		return
	}

	if t.GroupJID == "" {
		response.ParseError(fmt.Errorf("group JID is required"))
		RespondInterface(w, response)
		return
	}

	if t.Name == "" {
		response.ParseError(fmt.Errorf("name is required"))
		RespondInterface(w, response)
		return
	}

	// Convert string JID to appropriate format
	groupID := t.GroupJID

	// Call UpdateGroupSubject with the correct parameters and capture both return values
	updatedGroup, err := server.UpdateGroupSubject(groupID, t.Name)
	if err != nil {
		response.ParseError(fmt.Errorf("failed to set group name: %v", err))
		RespondInterface(w, response)
		return
	}

	response.GroupInfo = updatedGroup

	RespondSuccess(w, response)
}

func SetGroupPhotoController(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := &models.QpResponse{}

	type setGroupPhotoStruct struct {
		GroupJID  string `json:"group_jid"`
		ImageURL  string `json:"image_url"`  // image URL
		RemoveImg bool   `json:"remove_img"` // Option to remove existing photo
	}

	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var t setGroupPhotoStruct
	err = decoder.Decode(&t)
	if err != nil {
		response.ParseError(fmt.Errorf("could not decode payload: %v", err))
		RespondInterface(w, response)
		return
	}

	if t.GroupJID == "" {
		response.ParseError(fmt.Errorf("group JID is required"))
		RespondInterface(w, response)
		return
	}

	// Handle image removal request
	if t.RemoveImg {
		_, err := server.UpdateGroupPhoto(t.GroupJID, nil)
		if err != nil {
			response.ParseError(fmt.Errorf("failed to remove group photo: %v", err))
			RespondInterface(w, response)
			return
		}
		response.ParseSuccess("Group photo removed successfully")
		RespondSuccess(w, response)
		return
	}

	// Check if an image URL is provided
	if t.ImageURL == "" {
		response.ParseError(fmt.Errorf("image_url is required"))
		RespondInterface(w, response)
		return
	}

	// Download the image from the URL
	httpClient := &http.Client{
		Timeout: time.Second * 30, // Set a timeout for the request
	}

	resp, err := httpClient.Get(t.ImageURL)
	if err != nil {
		response.ParseError(fmt.Errorf("failed to download image: %v", err))
		RespondInterface(w, response)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		response.ParseError(fmt.Errorf("failed to download image, server returned status: %s", resp.Status))
		RespondInterface(w, response)
		return
	}

	// Read the image data
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		response.ParseError(fmt.Errorf("failed to read image data: %v", err))
		RespondInterface(w, response)
		return
	}

	// Check if the image meets WhatsApp's size requirements
	if len(imageData) < 10000 || len(imageData) > 500000 {
		response.ParseError(fmt.Errorf("image size not ideal for WhatsApp (should be between 10KB and 500KB)"))
		RespondInterface(w, response)
		return
	}

	// Call the service to update the group photo
	pictureID, err := server.UpdateGroupPhoto(t.GroupJID, imageData)
	if err != nil {
		response.ParseError(fmt.Errorf("failed to set group photo: %v", err))
		RespondInterface(w, response)
		return
	}

	// Prepare response
	response.ParseSuccess(fmt.Sprintf("Group photo updated successfully. Picture ID: %s", pictureID))
	RespondSuccess(w, response)
}

// endregion

// UpdateGroupParticipantsController handles adding, removing, promoting, and demoting group members
func UpdateGroupParticipantsController(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := &models.QpParticipantResponse{}

	type participantUpdate struct {
		GroupJID     string   `json:"group_jid"`
		Participants []string `json:"participants"` // Phone numbers or JIDs
		Action       string   `json:"action"`       // add, remove, promote, demote
	}

	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var req participantUpdate
	err = decoder.Decode(&req)
	if err != nil {
		response.ParseError(fmt.Errorf("could not decode payload: %v", err))
		RespondInterface(w, response)
		return
	}

	if req.GroupJID == "" {
		response.ParseError(fmt.Errorf("group_jid is required"))
		RespondInterface(w, response)
		return
	}

	if len(req.Participants) == 0 {
		response.ParseError(fmt.Errorf("at least one participant is required"))
		RespondInterface(w, response)
		return
	}

	// Validate action
	validActions := map[string]bool{"add": true, "remove": true, "promote": true, "demote": true}
	if !validActions[strings.ToLower(req.Action)] {
		response.ParseError(fmt.Errorf("invalid action, must be one of: add, remove, promote, demote"))
		RespondInterface(w, response)
		return
	}

	// Convert participants to JIDs if needed
	participantJIDs, err := convertToJIDs(req.Participants)
	if err != nil {
		response.ParseError(fmt.Errorf("failed to parse participant identifiers: %v", err))
		RespondInterface(w, response)
		return
	}

	// Perform the requested action
	result, err := server.UpdateGroupParticipants(req.GroupJID, participantJIDs, strings.ToLower(req.Action))
	if err != nil {
		response.ParseError(fmt.Errorf("failed to update group participants: %v", err))
		RespondInterface(w, response)
		return
	}

	response.Total = len(result)
	response.Participants = result

	// Create success response with participant statuses
	response.ParseSuccess(fmt.Sprintf("Group participants updated successfully. Action: %s", req.Action))
	RespondSuccess(w, response)
}

// GroupMembershipRequestsController handles retrieving and managing join requests for groups
func GroupMembershipRequestsController(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := &models.QpRequestResponse{}

	type membershipRequest struct {
		GroupJID     string   `json:"group_jid"`
		Participants []string `json:"participants"` // Only needed for approve/reject actions
		Action       string   `json:"action"`       // get, approve, reject
	}

	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	var req membershipRequest
	if r.Method == http.MethodPost {
		decoder := json.NewDecoder(r.Body)
		err = decoder.Decode(&req)
		if err != nil {
			response.ParseError(fmt.Errorf("could not decode payload: %v", err))
			RespondInterface(w, response)
			return
		}
	} else {
		// For GET requests, extract parameters from query string
		req.GroupJID = library.GetRequestParameter(r, "group_jid")
		req.Action = "get" // Default action for GET requests
	}

	if req.GroupJID == "" {
		response.ParseError(fmt.Errorf("group_jid is required"))
		RespondInterface(w, response)
		return
	}

	// Handle different actions
	switch strings.ToLower(req.Action) {
	case "get", "":
		// Get list of pending join requests
		requests, err := server.GetGroupJoinRequests(req.GroupJID)
		if err != nil {
			response.ParseError(fmt.Errorf("failed to get group join requests: %v", err))
			RespondInterface(w, response)
			return
		}

		// Set the response with the requests and the amount of requests
		response.Requests = requests
		response.Total = len(requests)

		response.ParseSuccess("Group join requests retrieved successfully")
		RespondInterface(w, response)
		return

	case "approve", "reject":
		if len(req.Participants) == 0 {
			response.ParseError(fmt.Errorf("at least one participant is required"))
			RespondInterface(w, response)
			return
		}

		// Process the approval/rejection
		result, err := server.HandleGroupJoinRequests(req.GroupJID, req.Participants, req.Action)
		if err != nil {
			response.ParseError(fmt.Errorf("failed to process join requests: %v", err))
			RespondInterface(w, response)
			return
		}
		response.ParseSuccess("Group join requests processed successfully")
		response.Requests = result
		RespondInterface(w, response)
		return

	default:
		response.ParseError(fmt.Errorf("invalid action, must be one of: get, approve, reject"))
		RespondInterface(w, response)
		return
	}
}

// Helper function to convert phone numbers or partial JIDs to full JIDs
func convertToJIDs(participants []string) ([]string, error) {
	result := make([]string, len(participants))

	for i, participant := range participants {
		// If it already contains @, assume it's a JID
		if strings.Contains(participant, "@") {
			result[i] = participant
		} else {
			// Otherwise, treat as a phone number and convert to JID format
			result[i] = participant + "@s.whatsapp.net"
		}
	}

	return result, nil
}

func SetGroupTopicController(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := &models.QpSingleGroupResponse{}

	type setGroupTopicStruct struct {
		GroupJID string `json:"group_jid"`
		Topic    string `json:"topic"`
	}

	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var t setGroupTopicStruct
	err = decoder.Decode(&t)
	if err != nil {
		response.ParseError(fmt.Errorf("could not decode payload: %v", err))
		RespondInterface(w, response)
		return
	}

	if t.GroupJID == "" {
		response.ParseError(fmt.Errorf("group JID is required"))
		RespondInterface(w, response)
		return
	}

	// Convert string JID to appropriate format
	groupID := t.GroupJID

	updatedGroup, err := server.UpdateGroupTopic(groupID, t.Topic)
	if err != nil {
		response.ParseError(fmt.Errorf("failed to set group topic: %v", err))
		RespondInterface(w, response)
		return
	}

	response.GroupInfo = updatedGroup

	RespondSuccess(w, response)
}
