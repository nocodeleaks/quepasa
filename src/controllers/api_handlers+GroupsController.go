package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

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

// endregion
