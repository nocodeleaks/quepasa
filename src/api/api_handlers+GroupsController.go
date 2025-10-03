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

// GetGroupController retrieves information about a specific WhatsApp group
// @Summary Get group information
// @Description Retrieves detailed information about a specific WhatsApp group
// @Tags Groups
// @Accept json
// @Produce json
// @Param groupId query string true "Group ID"
// @Success 200 {object} models.QpSingleGroupResponse
// @Failure 400 {object} models.QpResponse
// @Security ApiKeyAuth
// @Router /groups/get [get]
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

	group, err := server.GetGroupManager().GetGroupInfo(groupId)
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

// FetchAllGroupsController retrieves all WhatsApp groups the bot has joined
// @Summary Get all groups
// @Description Retrieves a list of all WhatsApp groups that the bot is currently a member of
// @Tags Groups
// @Accept json
// @Produce json
// @Success 200 {object} models.QpGroupsResponse
// @Failure 400 {object} models.QpResponse
// @Security ApiKeyAuth
// @Router /groups/getall [get]
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
	groups, err := server.GetGroupManager().GetJoinedGroups()
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

// CreateGroupController creates a new WhatsApp group
// @Summary Create a new group
// @Description Creates a new WhatsApp group with specified title and participants
// @Tags Groups
// @Accept json
// @Produce json
// @Param request body object{title=string,participants=[]string} true "Group creation request"
// @Success 200 {object} models.QpSingleGroupResponse
// @Failure 400 {object} models.QpResponse
// @Security ApiKeyAuth
// @Router /groups/create [post]
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

	// Convert phone numbers to proper WID format
	formattedParticipants := whatsapp.PhonesToWids(request.Participants)

	// Build extended options for group creation
	options := map[string]interface{}{
		"title":        request.Title,
		"participants": formattedParticipants,
	}

	// Create group using the interface method with properly formatted participants and options
	groupInfo, err := server.GetGroupManager().CreateGroupExtendedWithOptions(options)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Set response and return
	response.GroupInfo = groupInfo
	RespondSuccess(w, response)
}

// SetGroupNameController updates the name of a WhatsApp group
// @Summary Set group name
// @Description Updates the name of a specific WhatsApp group
// @Tags Groups
// @Accept json
// @Produce json
// @Param request body object{group_jid=string,name=string} true "Group name update request"
// @Success 200 {object} models.QpSingleGroupResponse
// @Failure 400 {object} models.QpResponse
// @Security ApiKeyAuth
// @Router /groups/name [put]
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
	updatedGroup, err := server.GetGroupManager().UpdateGroupSubject(groupID, t.Name)
	if err != nil {
		response.ParseError(fmt.Errorf("failed to set group name: %v", err))
		RespondInterface(w, response)
		return
	}

	response.GroupInfo = updatedGroup

	RespondSuccess(w, response)
}

// SetGroupPhotoController updates or removes the photo of a WhatsApp group
// @Summary Set/Remove group photo
// @Description Updates or removes the photo of a specific WhatsApp group
// @Tags Groups
// @Accept json
// @Produce json
// @Param request body object{group_jid=string,remove_img=boolean} true "Group photo update request"
// @Success 200 {object} models.QpSingleGroupResponse
// @Failure 400 {object} models.QpResponse
// @Security ApiKeyAuth
// @Router /groups/photo [put]
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
		_, err := server.GetGroupManager().UpdateGroupPhoto(t.GroupJID, nil)
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
	pictureID, err := server.GetGroupManager().UpdateGroupPhoto(t.GroupJID, imageData)
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
// @Summary Update group participants
// @Description Add, remove, promote, or demote participants in a WhatsApp group
// @Tags Groups
// @Accept json
// @Produce json
// @Param request body object{group_jid=string,participants=[]string,action=string} true "Participants update request"
// @Success 200 {object} models.QpParticipantResponse
// @Failure 400 {object} models.QpResponse
// @Security ApiKeyAuth
// @Router /groups/participants [put]
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

	// Convert participants to WIDs if needed
	participantWIDs := whatsapp.PhonesToWids(req.Participants)

	// Perform the requested action
	result, err := server.GetGroupManager().UpdateGroupParticipants(req.GroupJID, participantWIDs, strings.ToLower(req.Action))
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
// @Summary Handle group join requests
// @Description Get, approve, or reject join requests for WhatsApp groups
// @Tags Groups
// @Accept json
// @Produce json
// @Param request body object{group_jid=string,participants=[]string,action=string} true "Membership request"
// @Param group_jid query string false "Group JID (for GET requests)"
// @Success 200 {object} models.QpRequestResponse
// @Failure 400 {object} models.QpResponse
// @Security ApiKeyAuth
// @Router /groups/requests [get]
// @Router /groups/requests [post]
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
		requests, err := server.GetGroupManager().GetGroupJoinRequests(req.GroupJID)
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
		result, err := server.GetGroupManager().HandleGroupJoinRequests(req.GroupJID, req.Participants, req.Action)
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

// SetGroupTopicController updates the topic/description of a WhatsApp group
// @Summary Set group topic
// @Description Updates the topic/description of a specific WhatsApp group
// @Tags Groups
// @Accept json
// @Produce json
// @Param request body object{group_jid=string,topic=string} true "Group topic update request"
// @Success 200 {object} models.QpSingleGroupResponse
// @Failure 400 {object} models.QpResponse
// @Security ApiKeyAuth
// @Router /groups/description [put]
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

	updatedGroup, err := server.GetGroupManager().UpdateGroupTopic(groupID, t.Topic)
	if err != nil {
		response.ParseError(fmt.Errorf("failed to set group topic: %v", err))
		RespondInterface(w, response)
		return
	}

	response.GroupInfo = updatedGroup

	RespondSuccess(w, response)
}

// LeaveGroupController allows the bot to leave a WhatsApp group
// @Summary Leave group
// @Description Leave a specific WhatsApp group
// @Tags Groups
// @Accept json
// @Produce json
// @Param request body object{chatId=string} true "Leave group request"
// @Success 200 {object} models.QpResponse
// @Failure 400 {object} models.QpResponse
// @Security ApiKeyAuth
// @Router /groups/leave [post]
func LeaveGroupController(w http.ResponseWriter, r *http.Request) {
	response := &models.QpResponse{}

	// Declare a new request struct inline
	var request struct {
		ChatId string `json:"chatId"`
	}

	// Decode the JSON body into the request struct
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		response.ParseError(fmt.Errorf("invalid JSON in request body: %s", err.Error()))
		RespondInterface(w, response)
		return
	}

	// Validate required fields
	if request.ChatId == "" {
		response.ParseError(fmt.Errorf("chatId is required"))
		RespondInterface(w, response)
		return
	}

	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Validate group JID format
	valid := whatsapp.IsValidGroupId(request.ChatId)
	if !valid {
		response.ParseError(fmt.Errorf("invalid group JID format: %s", request.ChatId))
		RespondInterface(w, response)
		return
	}

	// Leave the group
	err = server.GetGroupManager().LeaveGroup(request.ChatId)
	if err != nil {
		response.ParseError(fmt.Errorf("failed to leave group: %s", err.Error()))
		RespondInterface(w, response)
		return
	}

	response.ParseSuccess("successfully left the group")
	RespondSuccess(w, response)
}
