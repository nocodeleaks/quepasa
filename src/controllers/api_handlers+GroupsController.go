package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	models "github.com/nocodeleaks/quepasa/models"
	types "go.mau.fi/whatsmeow/types"
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

	groupId := models.GetRequestParameter(r, "groupId")
	group, err := server.GetGroupInfo(groupId)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}
	response.GroupInfo = []*types.GroupInfo{group}

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

	// Parse request body
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

	if len(request.Participants) == 0 {
		response.ParseError(fmt.Errorf("participants are required"))
		RespondInterface(w, response)
		return
	}

	// Create group using the interface method
	groupInfo, err := server.CreateGroup(request.Title, request.Participants)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Set response and return
	response.GroupInfo = []*types.GroupInfo{groupInfo}
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

	response.GroupInfo = []*types.GroupInfo{updatedGroup}

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
